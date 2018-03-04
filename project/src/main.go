package main

import(
	"./driver/elevio"
    //"./driver/states"
    q "./queue"
    state "./states"
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

	//var init bool = false
    var d elevio.MotorDirection = elevio.MD_Down
    var d_temp elevio.MotorDirection
    var initFlag bool = false
    var cur_floor int
    elevio.SetMotorDirection(d)
	
	drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)

    // Initializing the elevator on the first floor
    for (!initFlag){
        select {
        case a := <- drv_floors:
            initFlag = state.Init(a)
        }
    }

    
    for {
        d_temp = q.ChooseDirection(cur_floor, d)
        d = d_temp
        elevio.SetMotorDirection(d)
        select {
        case a := <- drv_buttons:
            //fmt.Printf("%+v\n", a)
            q.AddToQueue(a)
            //q.PrintQueue()


        case a := <- drv_floors:
            cur_floor = a
            elevio.SetFloorIndicator(a)
            if q.ShouldStop(a,d) {
                d_temp = elevio.MD_Stop
                q.ClearAtCurrentFloor(a)
                q.PrintQueue()
            } else {
                d_temp = q.ChooseDirection(a, d)
            }
            d = d_temp
            //fmt.Printf("d_temp=%+v\n",d_temp)
            elevio.SetMotorDirection(d)

            
        case a := <- drv_obstr:
            fmt.Printf("Obstr = %+v\n", a)
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
            q.ResetQueue()
        }
    }

}
