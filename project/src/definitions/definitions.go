package definitions

import (
	"time"

	elevio "../driver/elevio"
)

type QueueMatrix struct {
	Matrix [NumFloors][NumButtons]bool `json:"tempMat"`
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
var OrderTime = time.Second * 30

const FILE_NAME = "backup.txt"
