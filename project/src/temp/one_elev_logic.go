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
		q.AddToQueue(a)
	
		// A floor is reached
	case a := <- drv_floors:
		cur_floor = a
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