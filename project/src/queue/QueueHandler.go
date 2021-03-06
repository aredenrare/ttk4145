package queue

import (
	"fmt"

	def "../definitions"
	elevio "../driver/elevio"
)

var queue def.QueueMatrix

func InitQueue() def.QueueMatrix {
	var temp def.QueueMatrix
	for i := 0; i < def.NumFloors; i++ {
		for j := 0; j < def.NumButtons; j++ {
			temp.Matrix[i][j] = false
		}
	}
	return temp
}

func AddToQueue(queueMat def.QueueMatrix, button elevio.ButtonEvent) def.QueueMatrix {
	temp := queueMat
	temp.Matrix[button.Floor][button.Button] = true
	return temp
}

func RemoveFromQueue(queueMat def.QueueMatrix, floor int, button elevio.ButtonType) def.QueueMatrix {
	temp := queueMat
	temp.Matrix[floor][int(button)] = false
	// Skru av lys (fjernes?)
	tempButton := elevio.ButtonEvent{Floor: floor, Button: button}
	elevio.SetButtonLamp(tempButton.Button, tempButton.Floor, false)
	return temp
}

func PrintQueue(queueMat def.QueueMatrix) {
	fmt.Println("Queue matrix: ")
	for i := 0; i < def.NumFloors; i++ {
		for j := 0; j < def.NumButtons; j++ {
			if queueMat.Matrix[i][j] {
				fmt.Printf("1 ")
			}
			if !queueMat.Matrix[i][j] {
				fmt.Printf("0 ")
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}

func AddOrdersToCurrentQueue(queueMat def.QueueMatrix, orderMat def.QueueMatrix) def.QueueMatrix {
	temp := queueMat
	for flr := 0; flr < def.NumFloors; flr++ {
		for btn := 0; btn < def.NumButtons-1; btn++ {
			if orderMat.Matrix[flr][btn] == true {
				temp.Matrix[flr][btn] = true
			}
		}
	}
	return temp
}

func SetHallLampsInQueue(queueMat def.QueueMatrix) {
	for flr := 0; flr < def.NumFloors; flr++ {
		for btn := 0; btn < def.NumButtons-1; btn++ {
			tempBtn := elevio.ButtonType(btn)
			tempButton := elevio.ButtonEvent{Floor: flr, Button: tempBtn}
			if queueMat.Matrix[flr][btn] {
				elevio.SetButtonLamp(tempButton.Button, tempButton.Floor, true)
			}
		}
	}
}

func CheckEmptyQueue(queueMat def.QueueMatrix) bool {
	for flr:=0; flr < def.NumFloors; flr++ {
		for btn:=0; btn<def.NumButtons; btn++ {
			if queueMat.Matrix[flr][btn] {
				return false
			}
		}
	}
	return true
}

func ResetHallCalls(queueMat def.QueueMatrix) def.QueueMatrix {
	temp := queueMat
	for flr := 0; flr < def.NumFloors; flr++ {
		for btn := 0; btn < def.NumButtons-1; btn++ {
			temp.Matrix[flr][btn] = false
		}
	}
	return temp

}