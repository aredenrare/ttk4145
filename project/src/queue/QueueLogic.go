package queue

import(
	elevio "../driver/elevio"
	def "../definitions"

)

func RequestAbove(floor int) bool {
	for f := floor+1; f<def.NumFloors; f++ {
		for b := 0; b<def.NumButtons; b++ {
			if queue.Matrix[f][b]==true  {
				return true
			}
		}
	}
	return false
}

func RequestBelow(floor int) bool {
	for f := 0; f<floor; f++ {
		for b := 0; b<def.NumButtons; b++ {
			if queue.Matrix[f][b] == true  {
				return true
			}
		}
	}
	return false
}

func ChooseDirection(floor int, dir elevio.MotorDirection) elevio.MotorDirection {
	switch(dir){
	case elevio.MD_Up:
		if RequestAbove(floor) {
			return elevio.MD_Up
		} else if RequestBelow(floor) {
			return elevio.MD_Down
		}
	case elevio.MD_Down, elevio.MD_Stop:
		if RequestBelow(floor) {
			return elevio.MD_Down
		} else if RequestAbove(floor) {
			return elevio.MD_Up
		}
	}
	return elevio.MD_Stop
}

func ShouldStop(floor int, dir elevio.MotorDirection) bool {
	switch dir{
	case elevio.MD_Down:
		return queue.Matrix[floor][elevio.BT_HallDown] ||
		queue.Matrix[floor][elevio.BT_Cab] 
		// || !RequestBelow(floor)
	case elevio.MD_Up:
		return queue.Matrix[floor][elevio.BT_HallUp] ||
		queue.Matrix[floor][elevio.BT_Cab]
		// || !RequestAbove(floor)
	}
	return false
}

func ClearAtCurrentFloor(floor int) {
	
	for b := 0; b < def.NumButtons; b++ {
		button := elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(b)}
		queue.Matrix[floor][b] = false
		elevio.SetButtonLamp(button.Button, button.Floor, false)
	}
}