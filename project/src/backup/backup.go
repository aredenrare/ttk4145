package backup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	def "../definitions"
)

func InitFile() {
	_, errOpen := os.Open(def.FILE_NAME)
	if errOpen != nil {
		//log.Fatal(errOpen)
		fmt.Println("Cant open file")
		_, errCreate := os.Create(def.FILE_NAME)
		if errCreate != nil {
			fmt.Println(errCreate)
		}
		fmt.Println("File is created")
	}
}

func SaveQueueMatInFile(elevinfo def.ElevInfo) {
	backupfile, err := os.Create(def.FILE_NAME)
	defer backupfile.Close()

	if err != nil {
		log.Printf("Error, backupfile Create - %v", err)
		fmt.Println("Couldn't make a file")
	}

	data, err := json.Marshal(elevinfo.QueueMat)
	if err != nil {
		log.Printf("Error, marshall - %v", err)
		fmt.Println("Couldn't make file")
	}
	backupfile.Write(data)
}

func ReadElevQFromFile() [def.NumFloors][def.NumButtons]bool {
	backupfile, err_read := ioutil.ReadFile(def.FILE_NAME)

	if err_read != nil {
		log.Printf("read error: %v", err_read)
	}

	var tempMat [def.NumFloors][def.NumButtons]bool
	err_json := json.Unmarshal(backupfile, &tempMat)

	if err_json != nil {
		log.Printf("json error: %v", err_json)
	}
	fmt.Printf("printing from file - %+v \n", tempMat)
	return tempMat
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

func boolToInt(a bool) int {
	var b int = 0
	if a {
		b = 1
	}
	return b
}

func intToBool(a int) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
