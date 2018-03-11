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

// var numFloors int = 4
// var doorOpenTime = time.Second * 2

// var heartBeatTime = time.Millisecond * 200
var curState def.ElevInfo
var initFlag bool = false
var doorOpen = false

func EventHandlerMain(drv_buttons <-chan elevio.ButtonEvent, drv_floors <-chan int, drv_obstr <-chan bool,
	drv_stop <-chan bool, peerUpdateCh <-chan peers.PeerUpdate, peerTxEnable chan<- bool,
	elevInfoTx chan<- def.Message, elevInfoRx <-chan def.Message, orderTx chan<- def.Message,
	orderRx <-chan def.Message, id string) {

	var doorTimeout <-chan time.Time
	heartBeat := time.Tick(def.HeartBeatTime)

	var d elevio.MotorDirection = elevio.MD_Down
	var tempDir elevio.MotorDirection
	var tempMat def.QueueMatrix
	var tempElev def.ElevInfo

	elevio.SetMotorDirection(d)
	// Initializing the elevator on the first floor
	for !initFlag {
		select {
		case a := <-drv_floors:
			initFlag = Init(a)
			curState.ID, _ = ip.LocalIP()
			curState.PrevFloor = 0
			curState.Dir = elevio.MD_Stop
			curState.QueueMat = q.InitQueue()
		}
	}

	for {
		// Prevents the elevator from moving while the door is open
		// if doors not open, it chooses direction depending on its orders
		if doorOpen {
			tempDir = elevio.MD_Stop
		} else {
			tempDir = q.ChooseDirection(curState.QueueMat, curState.PrevFloor, d)
		}
		d = tempDir
		elevio.SetMotorDirection(d)
		curState.Dir = d

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
			if q.ShouldStop(curState.QueueMat, flr, d) {
				tempDir = elevio.MD_Stop
				tempMat = q.ClearAtCurrentFloor(curState.QueueMat, flr, d)
				curState.QueueMat = tempMat
				q.PrintQueue(curState.QueueMat)
				doorOpen = true
				elevio.SetDoorOpenLamp(doorOpen)
				doorTimeout = time.After(def.DoorOpenTime)
			} else { // elevator should continue its journey (not stopping believing)
				tempDir = q.ChooseDirection(curState.QueueMat, flr, d)
			}
			d = tempDir
			elevio.SetMotorDirection(d)
			curState.Dir = d
		// closes door after timeout
		case <-doorTimeout:
			doorOpen = false
			elevio.SetDoorOpenLamp(doorOpen)

			// Obstruction triggered
		case obstr := <-drv_obstr:
			if obstr {
				tempDir = elevio.MD_Stop
			} else {
				tempDir = d
			}
			elevio.SetMotorDirection(d)
			curState.Dir = d

			// Stop triggered
		case stop := <-drv_stop:
			fmt.Printf("%+v\n", stop)
			for f := 0; f < def.NumFloors; f++ {
				for b := elevio.ButtonType(0); b < def.NumButtons; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
			curState.QueueMat = q.InitQueue()
		// peers are added or lost from the network
		case pUpdt := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", pUpdt.Peers)
			fmt.Printf("  New:      %q\n", pUpdt.New)
			fmt.Printf("  Lost:     %q\n", pUpdt.Lost)
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
						fmt.Printf("order at elevs floor standing still\n")
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
