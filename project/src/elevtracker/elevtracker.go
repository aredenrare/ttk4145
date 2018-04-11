package elevtracker

import (
	def "../definitions"
	elevio "../driver/elevio"
	peers "../network/peers"
	q "../queue"
)

var ElevMap def.ElevMap

func InitMap(peer peers.PeerUpdate) {
	// tempMap := make(map[string]def.ElevInfo)
	ElevMap = make(def.ElevMap)
	initMat := q.InitQueue()
	initState := def.ElevInfo{ID: peer.New, PrevFloor: 0, Dir: elevio.MD_Stop, QueueMat: initMat}
	ElevMap[peer.New] = initState
}

func RemoveFromMap(peer peers.PeerUpdate) {
	numLost := len(peer.Lost)
	for key := range ElevMap {
		for i := 0; i < numLost; i++ {
			if key == peer.Lost[i] {
				delete(ElevMap, key)
			}
		}
	}
}

func UpdateMap(message def.Message) {
	ElevMap[message.State.ID] = message.State
}

func CheckEmptyMap() bool {
	if len(ElevMap) == 0 {
		return true
	}
	return false
}

func ResetEmptyHallCalls() {
	for flr := 0; flr < def.NumFloors; flr++ {
		for btn := 0; btn < def.NumButtons-1; btn++ {
			isOrder := false
			tempBtn := elevio.ButtonType(btn)
			tempButton := elevio.ButtonEvent{Floor: flr, Button: tempBtn}
			for _, value := range ElevMap {
				if value.QueueMat.Matrix[flr][btn] {
					isOrder = true
				}
			}
			elevio.SetButtonLamp(tempButton.Button, tempButton.Floor, isOrder)
		}
	}
}

func SetDoorLampHallCalls(button elevio.ButtonEvent) bool {
	isStoppedAtFloor := false
	for _, value := range ElevMap {
		if value.PrevFloor == button.Floor && value.Dir == elevio.MD_Stop {
			isStoppedAtFloor = true
		}
	}
	return isStoppedAtFloor
}

/*
func ResolveLostPeersFromMap(pUpdt peers.PeerUpdate, i int, curState def.ElevInfo) {
	for key, value := range ElevMap {
		if key == pUpdt.Lost[i] {
			// adding the lost elevator orders to this elevators queue matrix
			tempMat := q.AddOrdersToCurrentQueue(curState.QueueMat, value.State.QueueMat)
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
*/
