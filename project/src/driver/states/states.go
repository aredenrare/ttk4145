package states

import ".././elevio"

func Init(a int, initCheck bool){
	if a==0 && initCheck==false {
		elevio.SetMotorDirection(0)
		initCheck = true
	}
}
