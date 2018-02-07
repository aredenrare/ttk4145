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

func UDPServer() {
    /* preparing any address at port 30000*/   
    ServerAddr,err := net.ResolveUDPAddr("udp",":30000")
    NetCheckError(err)
 
    /* Listening to port */
    ServerConn, err := net.ListenUDP("udp", ServerAddr)
    NetCheckError(err)
    defer ServerConn.Close()
 
    buf := make([]byte, 1024)
 
    for {
        n,addr,err := ServerConn.ReadFromUDP(buf)
        fmt.Println("Received ",string(buf[0:n]), " from ",addr)
 
        NetCheckError(err) 
    }
}

func UDPClient() {
	strEcho := "Forbløffis McNøffis at your sørviss"
	servAddr := "129.241.187.38:20001"
	udpAddr, err := net.ResolveUDPAddr("udp", servAddr)
	NetCheckError(err)

	conn, err := net.DialUDP("udp", nil, udpAddr)
	NetCheckError(err)
	
	defer conn.Close()

	_, err = conn.Write([]byte(strEcho))
	NetCheckError(err)
	println("write to server = ", strEcho)

	reply := make([]byte, 1024)
	_, err = conn.Read(reply)
	NetCheckError(err)
	println("reply from server=", string(reply))
}