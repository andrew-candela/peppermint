/*
Implements the message server.
I'm taking heavy inspiration from
https://github.com/nhooyr/websocket/blob/master/examples/chat/chat.go
except mine's probably going to be worse.
I want to make some mistakes and learn on my own.
*/

package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"nhooyr.io/websocket"
)

const (
	HEADER_PUBLIC_KEY        = "PUBLIC_KEY"
	HEADER_TARGET_PUBLIC_KEY = "TARGET_KEY"
	HEADER_SIGNATURE_TOKEN   = "SIGNATURE_TOKEN"
	HEADER_SIGNATURE_VALUE   = "SIGNATURE_VALUE"
)

type ChatServer struct {
	subscriber_mutex sync.Mutex
	serve_mux        http.ServeMux
	subscribers      map[string]Subscriber
}

type ChatClient struct {
}

type Subscriber struct {
	msgs chan []byte
}

func NewChatServer() *ChatServer {
	cs := ChatServer{
		subscribers: map[string]Subscriber{},
	}
	cs.serve_mux.HandleFunc("/subscribe", cs.subscribeHandler)
	cs.serve_mux.HandleFunc("/publish", cs.publishHandler)
	cs.serve_mux.HandleFunc("/", cs.echoHandler)

	return &cs
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (cs *ChatServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got a request!")
	pub_key := r.Header.Get(HEADER_PUBLIC_KEY)
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		fmt.Printf("could not accept websocket connection %v, %v", err, r.UserAgent())
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	err = cs.subscribe(r.Context(), c, pub_key)
	// Cleanup
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
}

func (cs *ChatServer) echoHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Println("Got Something!")
	fmt.Fprintln(w, "Got something!")
}

// A message is published by posting the body to the /publish endpoint
// The payload of the request will have the public key of the recip along
// with the message itself.
// The webserver doesn't look at the payload though, it looks for the recipient
// in the request headers
func (cs *ChatServer) publishHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	pub_key := r.Header.Get(HEADER_TARGET_PUBLIC_KEY)
	message := body
	err = cs.publish(pub_key, message)
	if err != nil {
		err_string := fmt.Sprintf("unable to publish message... %v", err)
		http.Error(w, err_string, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

// Creates a new subscriber object and adds it to the map.
// Then listens for incoming messages to write to the websocket connection.
func (cs *ChatServer) subscribe(ctx context.Context, conn *websocket.Conn, pub_key string) error {
	sub := Subscriber{
		msgs: make(chan []byte),
	}
	cs.addSubscriber(pub_key, sub)
	defer cs.deleteSubscriber(pub_key)

	// the listen loop
	// TODO: handle timeout and cancel
	for {
		select {
		case msg := <-sub.msgs:
			err := conn.Write(ctx, websocket.MessageBinary, msg)
			if err != nil {
				return err
			}
		}
	}

}

func (cs *ChatServer) publish(pub_key string, message []byte) error {
	cs.subscriber_mutex.Lock()
	defer cs.subscriber_mutex.Unlock()
	sub, ok := cs.subscribers[pub_key]
	if !ok {
		return fmt.Errorf("given public key is not in subscriber map")
	}
	sub.msgs <- message
	return nil
}

// Adds the given subscriber to the server's map of subscribers.
func (cs *ChatServer) addSubscriber(pub_key string, sub Subscriber) {
	cs.subscriber_mutex.Lock()
	cs.subscribers[pub_key] = sub
	cs.subscriber_mutex.Unlock()
}

func (cs *ChatServer) deleteSubscriber(pub_key string) {
	cs.subscriber_mutex.Lock()
	delete(cs.subscribers, pub_key)
	cs.subscriber_mutex.Unlock()
}

func (cs *ChatServer) Run(port string) {
	fmt.Println("Listening on port: ", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), &cs.serve_mux)
	if err != nil {
		fmt.Printf("Error serving app: %v\n", err)
	}
	os.Exit(1)
}

// run the webserver to accept websocket connections
func HostWeb(port string) {
	if port == "" {
		port = "80"
	}
	server := NewChatServer()
	server.Run(port)
}
