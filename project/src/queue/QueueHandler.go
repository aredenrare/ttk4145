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

func AddToQueue(button elevio.ButtonEvent) {
	queue.Matrix[button.Floor][button.Button] = true
	
	// Skru på lys (fjernes?)
	elevio.SetButtonLamp(button.Button, button.Floor, true)

}

func RemoveFromQueue(button elevio.ButtonEvent) {
	queue.Matrix[button.Floor][button.Button] = false

	// Skru av lys (fjernes?)
	elevio.SetButtonLamp(button.Button, button.Floor, false)
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
			fmt.Printf("%+v ",queue.Matrix[i][j])
		}
		fmt.Printf("\n")
	}
}