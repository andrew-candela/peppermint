package internal

import (
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/chzyer/readline"
)

const (
	PROTOCOL          = "udp4"
	CHECK_MARK        = "\u2705"
	X_MARK            = "\u274C"
	MEDIUM_CHECK_MARK = "\u2713"
)

type MessageTransport interface {
	Writer(*FriendDetail, []byte) error
	Reader()
}

type Messanger struct {
	recipients  []FriendDetail
	wait_group  *sync.WaitGroup
	private_key *rsa.PrivateKey
	public_key  []byte
	port        string
	transport   MessageTransport
	write_mutex *sync.Mutex
}

type WEBTransport struct {
	friends []FriendDetail
	port    string
}
type UDPTransport struct {
	friends []FriendDetail
	port    string
}

// Publish the message to the WEB recips
func (webt *WEBTransport) Writer(*FriendDetail, []byte) error {
	return fmt.Errorf("not Implemented")
}

// Read incoming messages from the websocket connection
func (webt *WEBTransport) Reader() {
	_ = createFriendPubKeyMap(webt.friends)
	panic("not implemented")
}

// Publish messages to the UDP recipients
func (udpt *UDPTransport) Writer(friend *FriendDetail, content []byte) error {
	return UDPSend(friend.host, friend.port, content)
}

/*
Open a UDP port and listen for incoming messages.
If the message is from a known host, then pass to the associated
channel, and send a success ack.
*/
func (udpt *UDPTransport) Reader() {
	friend_map := createFriendHostMap(udpt.friends)
	udp_addr, err := net.ResolveUDPAddr(PROTOCOL, ":"+udpt.port)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Listening on address %v:%v ...\n", GetOutboundIP().String(), udp_addr.Port)
	connection, err := net.ListenUDP(PROTOCOL, udp_addr)
	if err != nil {
		err = fmt.Errorf("could not open UDP port to listen for messages... %w", err)
		panic(err)
	}
	defer connection.Close()
	for {
		buffer := make([]byte, 1024)
		n, resp_addr, _ := connection.ReadFromUDP(buffer)
		sender_address := resp_addr.IP.String()
		friend, ok := friend_map[sender_address]
		if !ok {
			fmt.Println("Could not find user located at host: ", sender_address)
			continue
		}
		friend.inbound_messages <- buffer[:n]
		_, err := connection.WriteToUDP([]byte(MEDIUM_CHECK_MARK), resp_addr)
		if err != nil {
			fmt.Printf("Unable to send ack byte to %s:%v %s", resp_addr.IP.String(), resp_addr.Port, err)
		}
	}

}

// Holds details about who you will be sending/receiving messages from.
type FriendDetail struct {
	host             string
	port             string
	public_key       *rsa.PublicKey
	message_channel  chan *Message
	inbound_messages chan []byte
	name             string
}

// Use the friend public key as an identifier for each friend in map
func createFriendPubKeyMap(friend_list []FriendDetail) map[string]FriendDetail {
	friend_map := make(map[string]FriendDetail, len(friend_list))
	for _, friend := range friend_list {
		friend_map[string(PublicKeyToBytes(friend.public_key))] = friend
	}
	return friend_map
}

func createFriendHostMap(friend_list []FriendDetail) map[string]FriendDetail {
	friend_map := make(map[string]FriendDetail, len(friend_list))
	for _, friend := range friend_list {
		friend_map[friend.host] = friend
	}
	return friend_map
}

// Publish a message by sending it to all the channels associated with recips
func (udpm *Messanger) Publish(message_text string) {
	// sign the message with your private key then pass along to the channels
	message := Message{
		content:    []byte(message_text),
		public_key: udpm.public_key,
	}
	message.Sign(udpm.private_key)
	for _, friend := range udpm.recipients {
		friend.message_channel <- &message
		udpm.wait_group.Add(1)
	}
	udpm.wait_group.Wait()
}

// Readline loop collecting input from user.
// Sends messages with messanger.Publish() for processing
func (udpm *Messanger) WriteLoop() {
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		udpm.Publish(line)
	}
}

/*
Handle incoming serialized grams.
Each handler instance is associated to a particular friend.
The ReadLoop assigns incoming Grams to the appropriate handler.
Incoming byte arrays are handled thusly:
  - unmarshaled to PBGrams and then converted to Grams
  - added to the Gram Buffer, until a Gram with expect_more == false
  - the grams in the buffer are concatenated
  - the content is unmarshaled into a PBMessage and converted to a Message
  - the message is decrypted
  - the message signature is verified
  - the message content is written to stdout
*/
func IncomingMessageHandler(friend FriendDetail, write_mutex *sync.Mutex, private_key *rsa.PrivateKey) {
	var gram_content_buffer []byte
	for raw_gram := range friend.inbound_messages {
		gram, err := GramFromBytes(raw_gram)
		gram_content_buffer = append(gram_content_buffer, gram.content...)
		CheckErrFatal(err)
		if !gram.expect_more {
			message, err := MessageFromBytes(gram_content_buffer)
			CheckErrFatal(err)
			err = message.Decrypt(private_key)
			CheckErrFatal(err)
			verified := message.VerifySignature()
			if !verified {
				fmt.Println("Could not verify message came from ", friend.name)
				os.Exit(1)
			}
			write_mutex.Lock()
			fmt.Printf(
				"%v\n%v\n\n", friend.name, string(message.content),
			)
			write_mutex.Unlock()
		}
	}
}

// Listen on a UDP port and assign messages
func (udpm *Messanger) ReadLoop() {
	// start the message handler
	fmt.Println("Listening for messages...")
	for _, friend := range udpm.recipients {
		go IncomingMessageHandler(friend, udpm.write_mutex, udpm.private_key)
	}
	udpm.transport.Reader()
}

/*
Instantiates a UDPMessanger by
  - selecting a transport (UDP or WEB)
  - building the recipients object
  - instantiating a waitgroup and a write mutex
*/
func ConfigureMessanger(config *MessangerConfig, transport_type TRANSPORT_TYPE) *Messanger {
	var friends []FriendDetail
	var transport MessageTransport
	write_mutex := &sync.Mutex{}
	for _, recip := range config.Users {
		friends = append(friends, FriendDetail{
			host:             recip.Host,
			port:             recip.Port,
			public_key:       ParsePublicKey([]byte(recip.Key)),
			message_channel:  make(chan *Message),
			inbound_messages: make(chan []byte),
			name:             recip.Name,
		})
	}
	if transport_type == WEB {
		transport = &WEBTransport{
			friends: friends,
			port:    config.Port,
		}
	} else if transport_type == UDP {
		transport = &UDPTransport{
			friends: friends,
			port:    config.Port,
		}
	} else {
		panic("Illegal transport value has been passed!")
	}
	wg := sync.WaitGroup{}
	return &Messanger{
		recipients:  friends,
		wait_group:  &wg,
		private_key: config.PrivateKey,
		public_key:  EncodePublicKey(config.PrivateKey),
		port:        config.Port,
		transport:   transport,
		write_mutex: write_mutex,
	}
}

// Sets up goroutines for each recipient and then returns.
func (udpm *Messanger) OutboundConnect() {
	for i := range udpm.recipients {
		go sendAndReport(udpm.wait_group, &udpm.recipients[i], udpm.transport, udpm.write_mutex)
	}
}

// Listens for data sent to a channel, prep and send it via the transport.
// Blocks the main thread until done.
func sendAndReport(wg *sync.WaitGroup, friend *FriendDetail, transport MessageTransport, write_mutex *sync.Mutex) {

	for message := range friend.message_channel {
		message.Encrypt(friend.public_key)
		serialized_message := message.Serialize()
		err := transport.Writer(friend, serialized_message)
		write_mutex.Lock()
		if err != nil {
			fmt.Println("Could not send message to", friend.name, "...", err, X_MARK)
		} else {
			fmt.Printf("%v:\u2705\n", friend.name)
		}
		write_mutex.Unlock()
		wg.Done()
	}
}
