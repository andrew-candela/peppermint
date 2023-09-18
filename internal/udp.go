package internal

import (
	"fmt"
	"net"
	"time"
)

// Get outbound ip of this machine. Doesn't quite work.
// Returns the internal IP of this machine.
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	CheckErrFatal(err)
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func createUDPConnection(ip_port string) (conn *net.UDPConn, err error) {
	udpAddr, err := net.ResolveUDPAddr(PROTOCOL, ip_port)
	if err != nil {
		return nil, err
	}
	conn, err = net.DialUDP(PROTOCOL, nil, udpAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Attempt to send a byte array to the given connection.
func UDPSend(host string, port string, content []byte) error {
	conn, _ := createUDPConnection(host + ":" + port)
	resp_buffer := make([]byte, 1024)
	grams := SplitMessage(content)
	for _, gram := range grams {
		_, err := conn.Write(gram)
		if err != nil {
			return fmt.Errorf("unable to write to conection...%w", err)
		}
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err = conn.ReadFromUDP(resp_buffer)
		if err != nil {
			return fmt.Errorf("did not get success ack from connection after write...%w", err)
		}
	}
	return nil
}
