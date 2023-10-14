/*
This file contains the logic used to set up the user interface.
The UI for this will be inside a terminal, so it's mostly just
going to be about prompting the user for input and displaying
output to them in reasonable ways.
*/

package internal

import (
	"bytes"
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/chzyer/readline"
	"nhooyr.io/websocket"
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
	port        string
	transport   MessageTransport
	write_mutex *sync.Mutex
}

type WEBTransport struct {
	friends     []FriendDetail
	host_url    string
	private_key *rsa.PrivateKey
}

// Publish the message to the WEB recips
func (webt *WEBTransport) Writer(friend *FriendDetail, content []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webt.host_url+"/publish", bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("problem constructing publish request... %w", err)
	}
	sig, token := CreateSignature(webt.private_key)
	req.Header.Set(HEADER_TARGET_PUBLIC_KEY, PublicKeyToString(friend.public_key))
	req.Header.Set(HEADER_SIGNATURE_TOKEN, token)
	req.Header.Set(HEADER_SIGNATURE_VALUE, sig)
	req.Header.Set(HEADER_PUBLIC_KEY, PublicKeyToString(&webt.private_key.PublicKey))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("problem performing publish request... %w", err)
	}
	if resp.StatusCode == 200 {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unable to publish message to server... %s", string(body))
}

// Read incoming messages from the websocket connection
func (webt *WEBTransport) Reader() {
	friend_map := createFriendPubKeyMap(webt.friends)
	headers := GenerateRequestAuthHeaders(webt.private_key)
	options := websocket.DialOptions{HTTPHeader: *headers}
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	connection, _, err := websocket.Dial(ctx, webt.host_url+"/subscribe", &options)
	if err != nil {
		err = fmt.Errorf("could not create websocket connection to host: %v, %w", webt.host_url, err)
		fmt.Println(err)
		os.Exit(1)
	}
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled by server... Exiting read loop")
			return
		default:
			message_type, message_bytes, err := connection.Read(ctx)
			if err != nil {
				fmt.Println("error: could not read message from websocket conn: ", err)
				return
			}
			if message_type != websocket.MessageBinary {
				fmt.Println("error: could not read message of type: ", message_type.String())
				continue
			}
			message, err := MessageFromBytes(message_bytes)
			if err != nil {
				fmt.Println("could not deserialize message...", err)
				continue
			}
			err = message.Decrypt(webt.private_key)
			if err != nil {
				fmt.Println("Could not decrypt message: ", err)
				continue
			}
			// if !message.VerifySignature() {
			// 	fmt.Println("Could not verify signature of message. Skipping...")
			// 	continue
			// }
			pub_key, err := ParsePublicKey(message.public_key)
			if err != nil {
				fmt.Println("Could not parse public key: ", err)
				continue
			}
			pub_key_string := PublicKeyToString(pub_key)
			friend, ok := friend_map[pub_key_string]
			if !ok {
				fmt.Println("Could not find friend associated with public key: ", pub_key_string)
				continue
			}
			PrintLeftJustifiedMessage(friend.name)
			PrintLeftJustifiedMessage(string(message.content))
			fmt.Println()
		}
	}

}

// Holds details about who you will be sending/receiving messages from.
type FriendDetail struct {
	public_key       *rsa.PublicKey
	message_channel  chan Message
	inbound_messages chan []byte
	name             string
}

type FriendDetailMap map[string]FriendDetail

// Use the friend public key as an identifier for each friend in map
func createFriendPubKeyMap(friend_list []FriendDetail) FriendDetailMap {
	friend_map := make(map[string]FriendDetail, len(friend_list))
	for _, friend := range friend_list {
		friend_map[PublicKeyToString(friend.public_key)] = friend
	}
	return friend_map
}

// Publish a message by sending it to all the channels associated with recips
func (ppmt *Messanger) Publish(message_text string) {
	pub_key := EncodePublicKey(ppmt.private_key)
	message := Message{
		content:    []byte(message_text),
		public_key: pub_key,
	}
	// sign the message with your private key then pass along to the channels
	message.Sign(ppmt.private_key)
	for _, friend := range ppmt.recipients {
		friend.message_channel <- message
		ppmt.wait_group.Add(1)
	}
	ppmt.wait_group.Wait()
}

// Readline loop collecting input from user.
// Sends messages with messanger.Publish() for processing
func (ppmt *Messanger) WriteLoop() {
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
		if line != "" {
			ppmt.Publish(line)
		}
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

func (ppmt *Messanger) ReadLoop() {
	// start the message handler
	fmt.Println("Listening for messages...")
	ppmt.transport.Reader()
}

/*
Instantiates a PPMTessanger by
  - building the recipients object
  - instantiating a waitgroup and a write mutex
*/
func ConfigureMessanger(config *MessangerConfig) *Messanger {
	var friends []FriendDetail
	var transport MessageTransport
	write_mutex := &sync.Mutex{}
	for _, recip := range config.Users {
		pub_key, err := ParsePublicKey([]byte(recip.Key))
		CheckErrFatal(err)
		friends = append(friends, FriendDetail{
			public_key:       pub_key,
			message_channel:  make(chan Message),
			inbound_messages: make(chan []byte),
			name:             recip.Name,
		})
	}
	transport = &WEBTransport{
		friends:     friends,
		host_url:    config.URL,
		private_key: config.PrivateKey,
	}
	wg := sync.WaitGroup{}
	return &Messanger{
		recipients:  friends,
		wait_group:  &wg,
		private_key: config.PrivateKey,
		transport:   transport,
		write_mutex: write_mutex,
		port:        config.Port,
	}
}

// Sets up goroutines for each recipient and then returns.
func (ppmt *Messanger) OutboundConnect() {
	for i := range ppmt.recipients {
		go sendAndReport(ppmt.wait_group, &ppmt.recipients[i], ppmt.transport, ppmt.write_mutex)
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
