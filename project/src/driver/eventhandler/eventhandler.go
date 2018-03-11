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

	elevio.Init("localhost:15657", def.NumFloors)
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
		if doorOpen {
			tempDir = elevio.MD_Stop
		} else {
			tempDir = q.ChooseDirection(curState.QueueMat, curState.PrevFloor, d)
		}
		d = tempDir
		elevio.SetMotorDirection(d)
		curState.Dir = d

		select {
		// An order is noticed
		case btn := <-drv_buttons:
			if btn.Button == elevio.BT_Cab {
				tempMat = curState.QueueMat
				curState.QueueMat = q.AddToQueue(tempMat, btn)
			} else {
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
			if q.ShouldStop(curState.QueueMat, flr, d) {
				tempDir = elevio.MD_Stop
				tempMat = q.ClearAtCurrentFloor(curState.QueueMat, flr, d)
				curState.QueueMat = tempMat
				doorOpen = true
				elevio.SetDoorOpenLamp(doorOpen)
				doorTimeout = time.After(def.DoorOpenTime)
				// q.PrintQueue(curState.QueueMat)
			} else {
				tempDir = q.ChooseDirection(curState.QueueMat, flr, d)
			}
			d = tempDir
			elevio.SetMotorDirection(d)
			curState.Dir = d

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

		case msgRec := <-elevInfoRx:
			elevtr.UpdateMap(msgRec)

		case ordRec := <-orderRx:
			// fmt.Println("received order")
			if ordRec.State.ID == id {
				tempMat = q.AddOrdersToCurrentQueue(curState.QueueMat, ordRec.State.QueueMat)
				curState.QueueMat = tempMat
				q.PrintQueue(curState.QueueMat)
			}
		case <-heartBeat:
			TrState := def.Message{ID: "", State: curState}
			elevInfoTx <- TrState
		}
	}
}
