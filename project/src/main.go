package main

import(
	elevio  "./driver/elevio"
    q       "./queue"
    state   "./states"
    bcast   "./network/bcast"
    singleElev "./driver/singleElevatorController"
    "fmt"
    "time"
)

func main(){
    go singleElev.Main()
    select{

    }
}
