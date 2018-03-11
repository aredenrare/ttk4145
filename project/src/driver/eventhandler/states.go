package eventhandler

import (
	"fmt"

	q "../../queue"
	"../elevio"
)

func Init(floor int) bool {
	elevio.SetFloorIndicator(floor)
	var initFlag bool
	if floor == 0 {
		elevio.SetMotorDirection(elevio.MD_Stop)
		q.InitQueue()
		fmt.Println("Initialized")
		initFlag = true
	}
	if floor != 0 {
		elevio.SetMotorDirection(elevio.MD_Down)
		initFlag = false
	}
	return initFlag
}

func DoorOpen(isDoorOpen bool) {
	elevio.SetDoorOpenLamp(isDoorOpen)
}
