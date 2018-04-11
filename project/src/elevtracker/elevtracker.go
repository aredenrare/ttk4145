package elevtracker

import (
	def "../definitions"
	elevio "../driver/elevio"
	peers "../network/peers"
	q "../queue"
)

var ElevMap def.ElevMap

func InitializeElevTracker() {
	ElevMap = make(def.ElevMap)
}

func InitMap(peer peers.PeerUpdate) {
	// tempMap := make(map[string]def.ElevInfo)
	initMat := q.InitQueue()
	initState := def.ElevInfo{ID: peer.New, PrevFloor: 0, Dir: elevio.MD_Stop, QueueMat: initMat}
	ElevMap[peer.New] = initState
}

func RemoveFromMap(peer peers.PeerUpdate) {

	for _, lost := range peer.Lost {
		delete(ElevMap, lost)
	}
	// numLost := len(peer.Lost)
	// for key := range ElevMap {
	// 	for i := 0; i < numLost; i++ {
	// 		if key == peer.Lost[i] {
	// 			delete(ElevMap, key)
	// 		}
	// 	}
	// }
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

func ResetAllLamps() {
	for flr := 0; flr < def.NumFloors; flr++ {
		for btn := 0; btn < def.NumButtons; btn++ {
			tempBtn := elevio.ButtonType(btn)
			tempButton := elevio.ButtonEvent{Floor: flr, Button: tempBtn}
			elevio.SetButtonLamp(tempButton.Button, tempButton.Floor, false)
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
