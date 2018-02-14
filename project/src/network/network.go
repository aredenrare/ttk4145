package network

import(
	"fmt"
	"os"
	"time"
)

// Function needs to be run in loop
// 
func setheartbeat(sendHeartbeatCh <- chan bool, check bool){
	sendHeartbeatCh <- check
}

