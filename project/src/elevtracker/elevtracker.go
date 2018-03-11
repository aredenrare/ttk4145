package elevtracker

import (
	def "../definitions"
	elevio "../driver/elevio"
	peers "../network/peers"
	q "../queue"
)

var ElevMap def.ElevMap

// ElevMap = make(map[strinf]def.ElevInfo)

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
