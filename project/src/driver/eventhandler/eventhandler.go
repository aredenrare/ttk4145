package eventhandler

import (
	"fmt"
	"time"

	ip "../../network/localip"
	peers "../../network/peers"

	cost "../../assigner"
	def "../../definitions"
	elevtr "../../elevtracker"
	q "../../queue"
	elevio "../elevio"
)

var curState def.ElevInfo
var initFlag = false
var doorOpen = false
var redundancyFlag = false

func EventHandlerMain(drv_buttons <-chan elevio.ButtonEvent, drv_floors <-chan int, drv_obstr <-chan bool,
	drv_stop <-chan bool, peerUpdateCh <-chan peers.PeerUpdate, peerTxEnable chan<- bool,
	elevInfoTx chan<- def.Message, elevInfoRx <-chan def.Message, orderTx chan<- def.Message,
	orderRx <-chan def.Message, id string) {

	elevtr.InitializeElevTracker()
	var doorTimeout <-chan time.Time
	var OrderTimeOut <-chan time.Time
	heartBeat := time.Tick(def.HeartBeatTime)

	var dir elevio.MotorDirection = elevio.MD_Down
	var tempDir elevio.MotorDirection
	var tempMat def.QueueMatrix
	var tempElev def.ElevInfo

	elevio.SetMotorDirection(dir)

	for {
		// Initializing the elevator on the first floor
		for !initFlag {

			dir = elevio.MD_Down
			elevio.SetMotorDirection(dir)
			doorOpen = false
			elevio.SetDoorOpenLamp(doorOpen)
			elevtr.ResetAllLamps()
			if curState.Alive != true {
				curState.PrevFloor = -1
			}
			select {
			case a := <-drv_floors:
				elevio.SetFloorIndicator(a)
				
				if a == 0{
					elevio.SetMotorDirection(elevio.MD_Stop)
					curState.Dir = elevio.MD_Stop
					curState.ID, _ = ip.LocalIP()
					curState.PrevFloor = 0
					curState.QueueMat = q.InitQueue()
					initFlag = true
					fmt.Println("Initialized")
				}
				if initFlag && curState.PrevFloor == 0{
					curState.Alive = true
				}
			}
		}
		// Prevents the elevator from moving while the door is open
		// if doors not open, it chooses direction depending on its orders
		if doorOpen {
			tempDir = elevio.MD_Stop
		} else {
			tempDir = q.ChooseDirection(curState.QueueMat, curState.PrevFloor, dir)
		}
		dir = tempDir
		elevio.SetMotorDirection(dir)
		curState.Dir = dir

		select {
		// An order is noticed on this elevator
		case btn := <-drv_buttons:
			
			// To assure redundancy, the elevator will not take hall calls if its the only elevator on the network
			if !redundancyFlag && btn.Button!=elevio.BT_Cab{
				break
			}

			// standing still at ordered floor, complete order
			if curState.Dir == elevio.MD_Stop && curState.PrevFloor == btn.Floor {
				tempMat = curState.QueueMat
				elevio.SetButtonLamp(btn.Button, btn.Floor, false)
				doorOpen = true
				elevio.SetDoorOpenLamp(doorOpen)
				doorTimeout = time.After(def.DoorOpenTime)
				OrderTimeOut = time.After(def.OrderTime)
				curState.Alive = true

			} else if btn.Button == elevio.BT_Cab { // deal with cab call
				tempMat = curState.QueueMat
				curState.QueueMat = q.AddToQueue(tempMat, btn)
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)

			} else { // a Hall call is assigned to Cheapest elevator
				if !elevtr.CheckIfOrderTaken(btn) {
					tempElev = cost.ChooseCheapestElevator(btn)
					tempMat = q.AddToQueue(tempElev.QueueMat, btn)
					tempElev.QueueMat = tempMat
					TrOrder := def.Message{ID: "", State: tempElev}
					orderTx <- TrOrder
				}

			}

		// A floor is reached
		case flr := <-drv_floors:
			curState.PrevFloor = flr
			elevio.SetFloorIndicator(flr)
			// the elevator should stop and complete order
			if q.ShouldStop(curState.QueueMat, flr, dir) {
				tempDir = elevio.MD_Stop
				tempMat = q.ClearAtCurrentFloor(curState.QueueMat, flr, dir)
				curState.QueueMat = tempMat
				doorOpen = true
				elevio.SetDoorOpenLamp(doorOpen)
				doorTimeout = time.After(def.DoorOpenTime)
				OrderTimeOut = time.After(def.OrderTime)
				curState.Alive = true
			} else { // elevator should continue its journey (not stopping believing)
				tempDir = q.ChooseDirection(curState.QueueMat, flr, dir)
			}
			dir = tempDir
			elevio.SetMotorDirection(dir)
			curState.Dir = dir
		// closes door after timeout
		case <-doorTimeout:
			doorOpen = false
			elevio.SetDoorOpenLamp(doorOpen)
		
			// peers are added or lost from the network
		case pUpdt := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", pUpdt.Peers)
			fmt.Printf("  New:      %q\n", pUpdt.New)
			fmt.Printf("  Lost:     %q\n", pUpdt.Lost)
			fmt.Printf("num peers: %+v\n", len(pUpdt.Peers))
			if len(pUpdt.Peers) > 1 {
				redundancyFlag = true
			} else {
				redundancyFlag = false
			}
			// If an elevator is lost from the network, this elevators takes its hall calls
			// Implicit -> all other elevators takes its calls
			if len(pUpdt.Lost) != 0 {
				// a lost elevators is detected
				for i := 0; i < len(pUpdt.Lost); i++ {
					// finding the elevator(s) that are lost in the map
					for key, value := range elevtr.ElevMap {
						// comparing the lost elevators ID with IDs from the keys in the map
						if key == pUpdt.Lost[i] && curState.Alive {
							// adding the lost elevator orders to this elevators queue matrix
							tempMat = q.AddOrdersToCurrentQueue(curState.QueueMat, value.QueueMat)
							curState.QueueMat = tempMat
							for btn := 0; btn < def.NumButtons; btn++ {
								// resolves orders that are on this elevators floor if it stands still
								if curState.QueueMat.Matrix[curState.PrevFloor][btn] && curState.Dir == elevio.MD_Stop {
									curState.QueueMat.Matrix[curState.PrevFloor][btn] = false
									doorOpen = true
									elevio.SetDoorOpenLamp(doorOpen)
									doorTimeout = time.After(def.DoorOpenTime)
									OrderTimeOut = time.After(def.OrderTime)
									curState.Alive = true
								}
							}
						}
					}
				}
			}
			if elevtr.CheckEmptyMap() {
				elevtr.InitMap(pUpdt)
			} else {
				elevtr.RemoveFromMap(pUpdt)
			}

		// The elevator receives a state update from other elevators on the network
		// will happen when a heartbeat is sent
		case msgRec := <-elevInfoRx:
			elevtr.UpdateMap(msgRec)
			// Checks if some elevators on the network has timed out and can not be considered alive
			// This elevator (implicit all other elevators alive) will then copy its hall calls and add them to its own queue matrix
			for _, value := range elevtr.ElevMap {
				if value.Alive != true && curState.Alive == true {
					fmt.Println("A lost elevator is detected and dealt with, elevator found:")
					fmt.Println(value.ID)
					// adding the lost elevator orders to this elevators queue matrix
					tempMat = q.AddOrdersToCurrentQueue(curState.QueueMat, value.QueueMat)
					curState.QueueMat = tempMat
					for btn := 0; btn < def.NumButtons; btn++ {
						// resolves orders that are on this elevators floor if it stands still
						if curState.QueueMat.Matrix[curState.PrevFloor][btn] && curState.Dir == elevio.MD_Stop {
							curState.QueueMat.Matrix[curState.PrevFloor][btn] = false
							doorOpen = true
							elevio.SetDoorOpenLamp(doorOpen)
							doorTimeout = time.After(def.DoorOpenTime)
							OrderTimeOut = time.After(def.OrderTime)
							curState.Alive = true
						}
					}
					value.QueueMat = q.InitQueue()
				}
			}
			

		// an order is received and the elevator checks if it should take it
		case ordRec := <-orderRx:
			if ordRec.State.ID == id {
				fmt.Println("Received an order")
				tempMat = q.AddOrdersToCurrentQueue(curState.QueueMat, ordRec.State.QueueMat)
				curState.QueueMat = tempMat
				// resolves orders at the elevators current floor, opens door
				for btn := 0; btn < def.NumButtons; btn++ {
					if curState.QueueMat.Matrix[curState.PrevFloor][btn] && curState.Dir == elevio.MD_Stop {
						curState.QueueMat.Matrix[curState.PrevFloor][btn] = false
						doorOpen = true
						elevio.SetDoorOpenLamp(doorOpen)
						doorTimeout = time.After(def.DoorOpenTime)
						OrderTimeOut = time.After(def.OrderTime)
						curState.Alive = true
					}
				}
			}

		// Update the elevator's state to the network
		case <-heartBeat:
			if q.CheckEmptyQueue(curState.QueueMat){
				OrderTimeOut = time.After(def.OrderTime)
				curState.Alive = true
			}
			TrState := def.Message{ID: "", State: curState}
			elevInfoTx <- TrState
			elevtr.ResetEmptyHallCalls()

		// The elevator has orders it has not resolved within the OrderTime time limit
		case <- OrderTimeOut:
			curState.Alive = false
			fmt.Println("Elevator has timed out")
		}
	}
}
