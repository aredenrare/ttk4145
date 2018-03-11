package definitions

import (
	"time"

	elevio "../driver/elevio"
)

// Flyttet denne hit for å ha den med i ElevInfo structen
type QueueMatrix struct {
	Matrix [NumFloors][NumButtons]bool
}

type ElevInfo struct {
	ID        string
	PrevFloor int
	Dir       elevio.MotorDirection
	QueueMat  QueueMatrix
}
type Message struct {
	ID    string
	State ElevInfo
}

type ElevMap map[string]ElevInfo

var localIP string

const NumButtons = 3
const NumFloors = 4

var DoorOpenTime = time.Second * 2
var HeartBeatTime = time.Millisecond * 50
