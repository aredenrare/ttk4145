package costlogic

import (
	"fmt"

	def "../definitions"
	"../driver/elevio"
	elevtr "../elevtracker"
)

type costMsg struct {
	cost int
	id   string
}

// Used in this module
const UP int = 1
const STOP int = 0
const DOWN int = -1

// Tar inn heisen State (prevFloor og dir), en ordre (buttonType og floor) og heisen kømatrise

func CalculateCost(curState def.ElevInfo, ordBtn elevio.ButtonEvent) int {
	totCost := 0
	dir := int(curState.Dir)
	checkDir := int(ordBtn.Floor - curState.PrevFloor)

	if curState.PrevFloor == -1 {
		totCost++
	} else if dir != STOP {
		totCost += 2
	}
	if dir != STOP {
		if checkDir != dir {
			totCost += 10
		}
	}
	// Adding +1 to totCost for every stop on the way to ordered floor
	if checkDir > 0 && dir == UP || dir == STOP {
		for f := curState.PrevFloor; f < ordBtn.Floor || f == def.NumFloors; f++ {
			if curState.QueueMat.Matrix[f][ordBtn.Button] || curState.QueueMat.Matrix[f][elevio.BT_Cab] {
				totCost++
			}
			totCost++
		}
	}
	if checkDir < 0 && dir == DOWN || dir == STOP {
		for f := curState.PrevFloor; f > ordBtn.Floor || f == 0; f-- {
			if curState.QueueMat.Matrix[f][ordBtn.Button] || curState.QueueMat.Matrix[f][elevio.BT_Cab] {
				totCost++
			}
			totCost++
		}
	}
	fmt.Println("Cost = ", totCost)
	return totCost
}

func ChooseCheapestElevator(ordBtn elevio.ButtonEvent) def.ElevInfo {
	lowestCost := 1000000
	var bestElev def.ElevInfo
	var curCost int
	fmt.Printf("len(map) = %+v\n", len(elevtr.ElevMap))
	for _, value := range elevtr.ElevMap {
		curCost = CalculateCost(value, ordBtn)
		if curCost < lowestCost {
			lowestCost = curCost
			bestElev = value
		}
		if len(elevtr.ElevMap) == 1 {
			return value
		}
	}
	return bestElev
}
