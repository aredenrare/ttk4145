package main

import(
	"./driver/elevio"
	"./driver/states"
	"fmt"
)

func main(){
	
	// Get channels
	//getOrderCh	 	:= 	make(chan elevio.ButtonEvent)
	//getCostCh 		:= 	make(chan int)
	//getHeartbeatCh 	:= 	make(chan bool)

	//Send channels
	//sendOrderCh 	:= 	make(chan elevio.ButtonEvent)
	//sendCostCh		:=	make(chan int)
	//sendHeartbeatCh := 	make(chan bool)

    numFloors := 4

    elevio.Init("localhost:15657", numFloors)

	var init bool = false
	var d elevio.MotorDirection = elevio.MD_Down
    elevio.SetMotorDirection(d)
	
	drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    for {
        select {
        case a := <- drv_buttons:
            fmt.Printf("%+v\n", a)
            elevio.SetButtonLamp(a.Button, a.Floor, true)
            
        case a := <- drv_floors:

            fmt.Printf("%+v\n", a)
            states.Init(a,init)
            elevio.SetFloorIndicator(a)
            
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }
            
        case a := <- drv_stop:
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                }
            }
        }
    }

}
