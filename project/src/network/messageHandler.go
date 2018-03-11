package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	def "../definitions"
	"./localip"
	//elevio "../driver/elevio"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

type Message struct {
	id    string
	State def.ElevInfo
}

func main() {

	var id string
	var ID string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	ID = "1"

	go func() {
		var curState def.ElevInfo
		curState.ID = ID
		curState.PrevFloor = 1
		curState.Dir = -1
		elevInfoMsg := Message{"Hello from " + id, curState}
		for {

			elevInfoTx <- elevInfoMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-elevInfoRx:
			fmt.Printf("Received: ID: %v, PrevFloor: %v, dir: %v\n", a.State.ID, a.State.PrevFloor, a.State.Dir)
			fmt.Printf("Queue:\n")
			for i := 0; i < def.NumFloors; i++ {
				for j := 0; j < def.NumButtons; j++ {
					if a.State.QueueMat.Matrix[i][j] {
						fmt.Printf("1 ")
					}
					if !a.State.QueueMat.Matrix[i][j] {
						fmt.Printf("0 ")
					}
				}
				fmt.Printf("\n")
			}
		}
	}
}
