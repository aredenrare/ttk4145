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

var elevatorPort1 int
var elevatorPort2 int
var elevatorPort3 int
var simulatorPort string

func main() {
	flag.IntVar(&elevatorPort1, "port1", 30001, "port for peers")
	flag.IntVar(&elevatorPort2, "port2", 31001, "port for elevinfo")
	flag.IntVar(&elevatorPort3, "port3", 32001, "port for orders")
	flag.StringVar(&simulatorPort, "simulator", "localhost:15657", "address of simulator")
	flag.Parse()

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

	// Hardware channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	// Peer check
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	// heartbeat
	elevInfoTx := make(chan def.Message)
	elevInfoRx := make(chan def.Message)
	// orders
	orderTx := make(chan def.Message)
	orderRx := make(chan def.Message)

	elevio.Init(simulatorPort, def.NumFloors)

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
		peerUpdateCh, peerTxEnable, elevInfoTx, elevInfoRx, orderTx, orderRx, id)

	select {}

}
