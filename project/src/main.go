package main

import(
	"./driver/elevio"
	
)

func main(){
	// Get channels
	getOrderCh	 	:= 	make(chan elevio.ButtonEvent)
	getCostCh 		:= 	make(chan int)
	getHeartbeatCh 	:= 	make(chan bool)

	//Send channels
	sendOrderCh 	:= 	make(chan elevio.ButtonEvent)
	sendCostCh		:=	make(chan int)
	sendHeartbeatCh := 	make(chan bool)

	

}
