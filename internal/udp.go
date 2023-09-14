package internal

import (
	"fmt"
	"net"
	"time"
)

// Get outbound ip of this machine. Doesn't quite work.
// Returns the internal IP of this machine.
func getOutboundIP() net.IP {
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

// Set up the UDP connection and listen for incoming messages.
// Doesn't care about the contents of the message,
// just passes the []byte on to the provided channel
func UDPListen(port string, message_channel chan<- RawUDPMessage) {
	udpAddr, err := net.ResolveUDPAddr(PROTOCOL, ":"+port)
	CheckErrFatal(err)

	fmt.Println("Listening at: ", getOutboundIP().String()+udpAddr.String())
	connection, err := net.ListenUDP(PROTOCOL, udpAddr)
	CheckErrFatal(err)
	defer connection.Close()

	// Listen loop
	for {
		buffer := make([]byte, 1024)
		n, respAddr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Could not read message from %v because of error: %v\n", respAddr.IP.To4(), err)
			continue
		}
		message_channel <- RawUDPMessage{
			content_buffer: buffer,
			length:         n,
			sender_address: respAddr,
		}
	}
}

// handles the UDP Specific stuff, and then calls GenericMessageHandler
func UDPMessageHandler(message_channel <-chan *[]byte) {
	for raw_message := range message_channel {
		fmt.Printf(
			// "Got message: '%v' from address: %v",
			// raw_message.content_buffer[:raw_message.length],
			// raw_message.sender_address.IP.To4(),
			"Got message: %v", string(*raw_message),
		)
	}
}
