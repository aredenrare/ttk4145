package queue
import(
	//"fmt"
	elevio "../driver/elevio"
	def "../definitions"
	"fmt"
)
//type QueueType struct {
//	Status bool
//}

type QueueMatrix struct {
	Matrix [def.NumFloors][def.NumButtons]bool
}

var queue QueueMatrix

func InitQueue() {
	for i := 0; i < def.NumFloors; i++ {
		for j := 0; j < def.NumButtons; j++{
			queue.Matrix[i][j] = false
		}
	}
}

func AddToQueue(button elevio.ButtonEvent) {
	queue.Matrix[button.Floor][button.Button] = true
	
	// Skru på lys (fjernes?)
	elevio.SetButtonLamp(button.Button, button.Floor, true)

}

func RemoveFromQueue(floor int, button elevio.ButtonType) {
	queue.Matrix[floor][int(button)] = false

	// Skru av lys (fjernes?)
	newButton := elevio.ButtonEvent{Floor: floor, Button: button}
	elevio.SetButtonLamp(newButton.Button, newButton.Floor, false)
}

func ResetQueue(){
	for i := 0; i < def.NumFloors; i++ {
		for j := 0; j < def.NumButtons; j++{
			queue.Matrix[i][j] = false
		}
	}
}

func PrintQueue(){
	fmt.Println("Queue matrix: ")
	for i := 0; i < def.NumFloors; i++ {
		for j := 0; j < def.NumButtons; j++{
			if queue.Matrix[i][j] {
				fmt.Printf("1 ")	
			}
			if !queue.Matrix[i][j]{
				fmt.Printf("0 ")
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}