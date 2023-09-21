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
	"net/http"
	"sync"

	"nhooyr.io/websocket"
)

type ChatServer struct {
	subscriber_mutex sync.Mutex
	serve_mux        http.ServeMux
	subscribers      map[string]struct{}
}

func NewChatServer() *ChatServer {
	cs := ChatServer{}
	cs.serve_mux.Handle("/", http.FileServer(http.Dir(".")))
	cs.serve_mux.HandleFunc("/subscribe", cs.subscribeHandler)
	cs.serve_mux.HandleFunc("/publish", cs.publishHandler)

	return &cs
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (cs *ChatServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		fmt.Printf("could not accept websocket connection %v, %v", err, r.Referer())
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	err = cs.subscribe(r.Context(), c)
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

func (cs *ChatServer) publishHandler(w http.ResponseWriter, r *http.Request) {

}

func (cs *ChatServer) subscribe(ctx context.Context, conn *websocket.Conn) error {
	return nil
}
