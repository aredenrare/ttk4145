package singleelevatorcontroller

import(
	elevio  "../elevio"
	q       "../../queue"
	"time"
	"fmt"
	state "../../states"
	bcast   "../../network/bcast"
	//msgHandler "./network"
)

var numFloors int = 4
var initFlag bool = false
var cur_floor int
var doorOpenTime = time.Second*2
var doorOpen = false

func Main(){
		// Get channels
		RecOrderCh	 	:= 	make(chan elevio.ButtonEvent)
		RecCostCh 		:= 	make(chan int)
		//getHeartbeatCh 	:= 	make(chan bool)
	
		//Send channels
		TrOrderCh 	:= 	make(chan elevio.ButtonEvent)
		TrCostCh		:=	make(chan int)
		//sendHeartbeatCh := 	make(chan bool)
	
		elevio.Init("localhost:15657", numFloors)
		var d elevio.MotorDirection = elevio.MD_Down
		var d_temp elevio.MotorDirection
	
		elevio.SetMotorDirection(d)
		
		drv_buttons := make(chan elevio.ButtonEvent)
		drv_floors  := make(chan int)
		drv_obstr   := make(chan bool)
		drv_stop    := make(chan bool)    
		
		// driver routines
		go elevio.PollButtons(drv_buttons)
		go elevio.PollFloorSensor(drv_floors)
		go elevio.PollObstructionSwitch(drv_obstr)
		go elevio.PollStopButton(drv_stop)
	
		// Message routines
		go bcast.Transmitter(30001, TrOrderCh)
		go bcast.Receiver(30001, RecOrderCh)

		go bcast.Transmitter(30001, TrCostCh)
		go bcast.Receiver(30001, RecCostCh)

		
	
		// Initializing the elevator on the first floor
		for (!initFlag){
			select {
			case a := <- drv_floors:
				initFlag = state.Init(a)
			}
		}
	
		
		for {
			if doorOpen {
				d_temp = elevio.MD_Stop
			} else {
				d_temp = q.ChooseDirection(cur_floor, d)
			}
			d = d_temp
			elevio.SetMotorDirection(d)
			//fmt.Printf("doorOpen = %+v, dir = %+v\n",doorOpen,d)
			
			select {
				// A button is triggered
			case a := <- drv_buttons:
				if a.Button == elevio.BT_Cab {
					q.AddToQueue(a)
				} else {
					CalculateCost()
				}
			
				// A floor is reached
			case a := <- drv_floors:
				cur_floor = a
				//ElevInfoMsg.prevFloor = cur_floor
				elevio.SetFloorIndicator(a)
				if q.ShouldStop(a,d) {
					d_temp = elevio.MD_Stop
					q.ClearAtCurrentFloor(a, d)
	
					doorOpen = true
					elevio.SetDoorOpenLamp(doorOpen)
					
					q.PrintQueue()
				} else {
					d_temp = q.ChooseDirection(a, d)
				}
				d = d_temp
				//ElevInfoMsg.dir = d_temp
				elevio.SetMotorDirection(d)
	
			case <- time.After(doorOpenTime):
				doorOpen = false
				elevio.SetDoorOpenLamp(doorOpen)
	
				// Obstruction triggered
			case a := <- drv_obstr:
				fmt.Printf("Obstr = %+v\n", a)
				if a {
					elevio.SetMotorDirection(elevio.MD_Stop)
				} else {
					elevio.SetMotorDirection(d)
				}
				
				// Stop triggered
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