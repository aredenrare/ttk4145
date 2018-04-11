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
var initFlag bool = false
var doorOpen = false

func EventHandlerMain(drv_buttons <-chan elevio.ButtonEvent, drv_floors <-chan int, drv_obstr <-chan bool,
	drv_stop <-chan bool, peerUpdateCh <-chan peers.PeerUpdate, peerTxEnable chan<- bool,
	elevInfoTx chan<- def.Message, elevInfoRx <-chan def.Message, orderTx chan<- def.Message,
	orderRx <-chan def.Message, id string) {

	elevtr.InitializeElevTracker()

	var doorTimeout <-chan time.Time
	heartBeat := time.Tick(def.HeartBeatTime)

	var dir elevio.MotorDirection = elevio.MD_Down
	var tempDir elevio.MotorDirection
	var tempMat def.QueueMatrix
	var tempElev def.ElevInfo

	elevio.SetMotorDirection(dir)

	for {
		// Initializing the elevator on the first floor
		for !initFlag {
			doorOpen = false
			elevtr.ResetAllLamps()
			select {
			case a := <-drv_floors:
				initFlag = Init(a)
				curState.ID, _ = ip.LocalIP()
				curState.PrevFloor = 0
				curState.Dir = elevio.MD_Stop
				curState.QueueMat = q.InitQueue()
				if initFlag {
					// add cab calls from disk to current queuematrix
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
			// standing still at ordered floor, complete order
			if curState.Dir == elevio.MD_Stop && curState.PrevFloor == btn.Floor {
				tempMat = curState.QueueMat
				elevio.SetButtonLamp(btn.Button, btn.Floor, false)
				doorOpen = true
				elevio.SetDoorOpenLamp(doorOpen)
				doorTimeout = time.After(def.DoorOpenTime)

			} else if btn.Button == elevio.BT_Cab { // deal with cab calls
				// write new cab calls to disk
				tempMat = curState.QueueMat
				curState.QueueMat = q.AddToQueue(tempMat, btn)
				elevio.SetButtonLamp(btn.Button, btn.Floor, true)
			} else { // a Hall call is assigned to Cheapest elevator
				tempElev = cost.ChooseCheapestElevator(btn)
				tempMat = q.AddToQueue(tempElev.QueueMat, btn)
				tempElev.QueueMat = tempMat
				TrOrder := def.Message{ID: "", State: tempElev}
				orderTx <- TrOrder
			}

		// A floor is reached
		case flr := <-drv_floors:
			curState.PrevFloor = flr
			elevio.SetFloorIndicator(flr)
			// the elevator should stop and complete order
			if q.ShouldStop(curState.QueueMat, flr, dir) {
				// if a cab call is cleared, update the file on disk
				tempDir = elevio.MD_Stop
				tempMat = q.ClearAtCurrentFloor(curState.QueueMat, flr, dir)
				curState.QueueMat = tempMat
				doorOpen = true
				elevio.SetDoorOpenLamp(doorOpen)
				doorTimeout = time.After(def.DoorOpenTime)
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
			if !stop {
				initFlag = false
			}

		// peers are added or lost from the network
		case pUpdt := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", pUpdt.Peers)
			fmt.Printf("  New:      %q\n", pUpdt.New)
			fmt.Printf("  Lost:     %q\n", pUpdt.Lost)

			// If an elevator is lost from the network, this elevators takes its hall calls
			// Implicit -> all other elevators takes its calls
			if len(pUpdt.Lost) != 0 {
				// a lost elevators is detected
				for i := 0; i < len(pUpdt.Lost); i++ {
					// finding the elevator(s) that are lost in the map
					for key, value := range elevtr.ElevMap {
						// comparing the lost elevators ID with IDs from the keys in the map
						if key == pUpdt.Lost[i] {
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
		case msgRec := <-elevInfoRx:
			elevtr.UpdateMap(msgRec)
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
					}
				}
			}
		// Update the elevators state to the network
		case <-heartBeat:
			TrState := def.Message{ID: "", State: curState}
			elevInfoTx <- TrState
			elevtr.ResetEmptyHallCalls()
		}
	}
}
