package internal

import (
	"crypto/rsa"
	"fmt"
	"sync"
)

const (
	PROTOCOL   = "udp4"
	CHECK_MARK = "\u2705"
	X_MARK     = "\u274C"
)

type MessageTransport interface {
	Writer(*FriendDetail, []byte) error
	Reader(*[]FriendDetail)
}

type Messanger struct {
	recipients  []FriendDetail
	wait_group  *sync.WaitGroup
	private_key *rsa.PrivateKey
	port        string
	transport   *MessageTransport
	write_mutex *sync.Mutex
}

type WEBTransport struct {
}
type UDPTransport struct {
}

// Publish the message to the WEB recips
func (webt *WEBTransport) Writer(*FriendDetail, []byte) error {
	return fmt.Errorf("")
}

// Read incoming messages from the websocket connection
func (webt *WEBTransport) Reader(*[]FriendDetail) {
}

// Publish messages to the UDP recipients
func (udpt *UDPTransport) Writer(friend *FriendDetail, content []byte) error {
	return UDPSend(friend.host, friend.port, content)
}

// Read incoming messages from the open UDP port
func (udpt *UDPTransport) Reader(*[]FriendDetail) {}

// Holds details about who you will be sending/receiving messages from.
type FriendDetail struct {
	host            string
	port            string
	public_key      *rsa.PublicKey
	message_channel chan *Message
	name            string
}

// Publish a message by sending it to all the channels associated with recips
func (udpm *Messanger) Publish(message_text string) {
	// sign the message with your private key then pass along to the channels
	message := Message{
		content: []byte(message_text),
	}
	message.Sign(udpm.private_key)
	for _, friend := range udpm.recipients {
		friend.message_channel <- &message
		udpm.wait_group.Add(1)
	}
	udpm.wait_group.Wait()
}

// Listen on a UDP port and wait for messages to come in
func (udpm *Messanger) Listen() {
	LogWithFileName(fmt.Sprintf("Starting listener on port:%v", udpm.port))
	incoming_message_channel := make(chan RawUDPMessage)
	// start the message handler
	go UDPMessageHandler(incoming_message_channel)
	//start the 'server'
	UDPListen(udpm.port, incoming_message_channel)
}

// Instantiates a UDPMessanger.
func ConfigureMessanger(config *TransportConfig, transport_type TRANSPORT_TYPE) *Messanger {
	var friends []FriendDetail
	var transport MessageTransport
	write_mutex := &sync.Mutex{}
	if transport_type == WEB {
		transport = &WEBTransport{}
	} else if transport_type == UDP {
		transport = &UDPTransport{}
	} else {
		panic("Illegal transport value has been passed!")
	}
	for _, recip := range config.Users {
		friends = append(friends, FriendDetail{
			host:            recip.Host,
			port:            recip.Port,
			public_key:      ParsePublicKey(recip.Key),
			message_channel: make(chan *Message),
			name:            recip.Name,
		})
	}
	wg := sync.WaitGroup{}
	return &Messanger{
		recipients:  friends,
		wait_group:  &wg,
		private_key: config.PrivateKey,
		port:        config.Port,
		transport:   &transport,
		write_mutex: write_mutex,
	}
}

// Sets up goroutines for each recipient and then returns.
func (udpm *Messanger) OutboundConnect() {
	for i := range udpm.recipients {
		go sendAndReport(udpm.wait_group, &udpm.recipients[i], *udpm.transport, udpm.write_mutex)
	}
}

// Stand-in for actual functionality.
// Listens for data sent to a channel and pretends to send it
func sendAndReport(wg *sync.WaitGroup, friend *FriendDetail, transport MessageTransport, write_mutex *sync.Mutex) {

	for message := range friend.message_channel {
		message.Encrypt(friend.public_key)
		// I'm gonna have to split this up into packets later
		err := transport.Writer(friend, message.content)
		write_mutex.Lock()
		if err != nil {
			fmt.Println("Could not send message to", friend.name, "...", err, X_MARK)
		} else {
			fmt.Printf("Sent %v message: %v \u2705\n", friend.name, string(message.content))
		}
		write_mutex.Unlock()
		wg.Done()
	}
}
