package queue

import (
	"fmt"

	def "../definitions"
	elevio "../driver/elevio"
)

//type QueueMatrix struct {
//	Matrix [def.NumFloors][def.NumButtons]bool
//}

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
	// Skru på lys (fjernes?)
	elevio.SetButtonLamp(button.Button, button.Floor, true)
	return temp
}

func RemoveFromQueue(queueMat def.QueueMatrix, floor int, button elevio.ButtonType) def.QueueMatrix {
	temp := queueMat
	temp.Matrix[floor][int(button)] = false
	// Skru av lys (fjernes?)
	newButton := elevio.ButtonEvent{Floor: floor, Button: button}
	elevio.SetButtonLamp(newButton.Button, newButton.Floor, false)
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

func CheckInQueue(queueMat def.QueueMatrix, floor int, button elevio.ButtonType) bool {
	return queueMat.Matrix[floor][int(button)]
}

func AddOrdersToCurrentQueue(queueMat def.QueueMatrix, orderMat def.QueueMatrix) def.QueueMatrix {
	temp := queueMat
	for flr := 0; flr < def.NumFloors; flr++ {
		for btn := 0; btn < def.NumButtons; btn++ {
			if orderMat.Matrix[flr][btn] == true {
				temp.Matrix[flr][btn] = true
			}
		}
	}
	return temp
}
