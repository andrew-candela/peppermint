package internal

import (
	"fmt"
	"log"
	"net"
	"runtime"
)

func CheckErrFatal(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

func LogWithFileName(message string) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("[%s:%d] %s\n", file, line, message)
}

// Get outbound ip of this machine.
// This is the internal IP, and only available to other machines
// on the internal network.
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
