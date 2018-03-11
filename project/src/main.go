package main

import (
	"flag"
	"fmt"

	bcast "./network/bcast"
	"./network/localip"
	peers "./network/peers"

	def "./definitions"
	elevio "./driver/elevio"
	evhandler "./driver/eventhandler"
)

func main() {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("%s", localIP)
	}
	localID := id

	// Hardware channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	// Network channels
	// peer check
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	// messages
	elevInfoTx := make(chan def.Message)
	elevInfoRx := make(chan def.Message)
	// orders
	orderTx := make(chan def.Message)
	orderRx := make(chan def.Message)
	// driver routines
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	// Message routines
	go peers.Transmitter(30001, id, peerTxEnable)
	go peers.Receiver(30001, peerUpdateCh)
	go bcast.Transmitter(31001, elevInfoTx)
	go bcast.Receiver(31001, elevInfoRx)
	go bcast.Transmitter(32001, orderTx)
	go bcast.Receiver(32001, orderRx)

	go evhandler.EventHandlerMain(drv_buttons, drv_floors, drv_obstr, drv_stop,
		peerUpdateCh, peerTxEnable, elevInfoTx, elevInfoRx, orderTx, orderRx, localID)

	select {}

}
