package definitions

import (
	"time"
	elevio "../driver/elevio"
)
// Flyttet denne hit for å ha den med i ElevInfo structen
type QueueMatrix struct {
	Matrix [NumFloors][NumButtons]bool
}

type ElevInfo struct{
	ID 			string
	PrevFloor 	int
	Dir 		elevio.MotorDirection
	QueueMat	QueueMatrix
}
var localIP string

const NumButtons = 3
const NumFloors = 4
const DoorOpenSec = 3 * time.Second
const HeartbeatInterval = 800 * time.Millisecond
