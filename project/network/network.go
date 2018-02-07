package network

import (
    "fmt"
    "net"
    "os"
)

func NetCheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}

