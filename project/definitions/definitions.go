package definitions

import "time"

var localIP string

const NumButtons = 3
const NumFloors = 4
const DoorOpenSec = 3 * time.Second
const HeartbeatInterval = 800 * time.Millisecond

const(
	DirDown int = iota -1
	DirStop	
	DirUp
)

const(
	ButHallUp int = iota
	ButHallDown
	ButCab
)

type ButPress struct{
	Floor int
	Button int
}

type Light struct{
	Floor int
	Button int
	ShouldUpdate bool
}

