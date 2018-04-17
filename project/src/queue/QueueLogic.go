package queue

import (
	def "../definitions"
	elevio "../driver/elevio"
)

func RequestAbove(queueMat def.QueueMatrix, floor int) bool {
	for f := floor + 1; f < def.NumFloors; f++ {
		for b := 0; b < def.NumButtons; b++ {
			if queueMat.Matrix[f][b] == true {
				return true
			}
		}
	}
	return false
}

func RequestBelow(queueMat def.QueueMatrix, floor int) bool {
	for f := 0; f < floor; f++ {
		for b := 0; b < def.NumButtons; b++ {
			if queueMat.Matrix[f][b] == true {
				return true
			}
		}
	}
	return false
}

func ChooseDirection(queueMat def.QueueMatrix, floor int, dir elevio.MotorDirection) elevio.MotorDirection {
	switch dir {
	case elevio.MD_Up:
		if RequestAbove(queueMat, floor) {
			return elevio.MD_Up
		}
		if RequestBelow(queueMat, floor) {
			return elevio.MD_Down
		}
	case elevio.MD_Down, elevio.MD_Stop:
		if RequestBelow(queueMat, floor) {
			return elevio.MD_Down
		}
		if RequestAbove(queueMat, floor) {
			return elevio.MD_Up
		}
	}
	return elevio.MD_Stop
}

func ShouldStop(queueMat def.QueueMatrix, floor int, dir elevio.MotorDirection) bool {
	switch dir {
	case elevio.MD_Down:
		return queueMat.Matrix[floor][elevio.BT_HallDown] ||
			queueMat.Matrix[floor][elevio.BT_Cab] || !RequestBelow(queueMat, floor)
	case elevio.MD_Up:
		return queueMat.Matrix[floor][elevio.BT_HallUp] ||
			queueMat.Matrix[floor][elevio.BT_Cab] || !RequestAbove(queueMat, floor)
	case elevio.MD_Stop:
		return queueMat.Matrix[floor][elevio.BT_HallDown] ||
			queueMat.Matrix[floor][elevio.BT_Cab] || queueMat.Matrix[floor][elevio.BT_HallUp]
	}
	return false
}

func ClearAtCurrentFloor(queueMat def.QueueMatrix, floor int, dir elevio.MotorDirection) def.QueueMatrix {
	temp := RemoveFromQueue(queueMat, floor, elevio.BT_Cab)
	switch dir {
	case elevio.MD_Up:
		temp2 := RemoveFromQueue(temp, floor, elevio.BT_HallUp)
		if !RequestAbove(temp2, floor) {
			temp3 := RemoveFromQueue(temp2, floor, elevio.BT_HallDown)
			return temp3
		}
		return temp2

	case elevio.MD_Down:
		temp2 := RemoveFromQueue(temp, floor, elevio.BT_HallDown)
		if !RequestBelow(temp2, floor) {
			temp3 := RemoveFromQueue(temp2, floor, elevio.BT_HallUp)
			return temp3
		}
		return temp2

	case elevio.MD_Stop:
	default:
		temp2 := RemoveFromQueue(temp, floor, elevio.BT_HallUp)
		temp3 := RemoveFromQueue(temp2, floor, elevio.BT_HallDown)
		return temp3
	}
	return temp
}

