package internal

import (
	"github.com/chzyer/readline"
)

type TRANSPORT_TYPE int
type READ_OR_WRITE int

const (
	UDP TRANSPORT_TYPE = iota
	WEB
)

const (
	READ READ_OR_WRITE = iota
	WRITE
)

// Readline loop collecting input from user.
// Sends messages with messanger.Publish() for processing
func publisher(messanger *Messanger) {
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
		messanger.Publish(line)
	}
}

// Set up the transport and begin the Write or Read loop
func MessageEntrypoint(transport_type TRANSPORT_TYPE, action READ_OR_WRITE, config *TransportConfig) {
	transport := ConfigureMessanger(config, transport_type)
	if action == WRITE {
		transport.OutboundConnect()
		publisher(transport)
	} else if action == READ {
		transport.Listen()
	} else {
		panic("Illegal action type provided")
	}
}
