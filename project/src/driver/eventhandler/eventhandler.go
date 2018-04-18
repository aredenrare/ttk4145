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

// This variable is used to avoid the same order to be calculated in the costfunction
// if one button is pressed multiple times

func EventHandlerMain(drv_buttons <-chan elevio.ButtonEvent, drv_floors <-chan int, drv_obstr <-chan bool,
	drv_stop <-chan bool, peerUpdateCh <-chan peers.PeerUpdate, peerTxEnable chan<- bool,
	elevInfoTx chan<- def.Message, elevInfoRx <-chan def.Message, orderTx chan<- def.Message,
	orderRx <-chan def.Message, id string) {

	elevtr.InitializeElevTracker()
	// creates a backup file on disk
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
			elevtr.ResetHallLamps()
			fmt.Printf("while Initializing: curState.ALive = %+v\n",curState.Alive)
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
					if value, ok := elevtr.ElevMap[curState.ID]; ok {
						curState.QueueMat = value.QueueMat
					} else {
						curState.QueueMat = q.InitQueue()
					}
					initFlag = true
					fmt.Println("Initialized")
				}
				if initFlag && curState.PrevFloor == 0{
					curState.Alive = true
					fmt.Println("in init: curState.Alive = true")
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
			
			if !redundancyFlag {
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
				fmt.Println("standing at ordered floor: curState.Alive = true")

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
				fmt.Println("in shouldStop: curState.Alive = true")
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

			// Obstruction triggered
		case obstr := <-drv_obstr:
			if obstr {
				tempDir = elevio.MD_Stop
			} else {
				tempDir = dir
			}
			elevio.SetMotorDirection(dir)
			curState.Dir = dir

			// Stop triggered
		case stop := <-drv_stop:
			fmt.Println(stop)
			initFlag = false
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
									fmt.Println("taking over orders from an elev lost from network, standing still at ordered floor: curState.Alive = true")
								}
							}
							value.Alive = false
							tempMat = q.ResetHallCalls(value.QueueMat)
							value.QueueMat = tempMat
							TrState := def.Message{ID:"", State: value}
							elevInfoTx <- TrState
						}
					}
				}
			}
			if elevtr.CheckEmptyMap() {
				elevtr.InitMap(pUpdt)
			} //else {
				//elevtr.RemoveFromMap(pUpdt)
			//}

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
							fmt.Println("taking over orders from an !Alive elevator still on the network: curState.Alive = true")
						}
					}
					tempMat = q.ResetHallCalls(value.QueueMat)
					value.QueueMat = tempMat
					TrState := def.Message{ID:"", State: value}
					elevInfoTx <- TrState
				}
			}
			

		// an order is received and the elevator checks if it should take it
		case ordRec := <-orderRx:
			if ordRec.State.ID == id {
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
						fmt.Println("Resolves its own orders at the floor its standing on: curState.Alive = true")
					}
				}
			}

		// Update the elevator's state to the network
		case <-heartBeat:
			if q.CheckEmptyQueue(curState.QueueMat){
				OrderTimeOut = time.After(def.OrderTime)
				curState.Alive = true
				fmt.Println("Empty queue: curState.Alive = true")
			}
			TrState := def.Message{ID: "", State: curState}
			elevInfoTx <- TrState
			elevtr.ResetEmptyHallCalls()

		// The elevator has orders it has not resolved within the OrderTime time limit
		case <- OrderTimeOut:
			curState.Alive = false
			initFlag = false

			fmt.Println("Elevator has timed out, the elevator will initialize when back")
		}
	}
}
