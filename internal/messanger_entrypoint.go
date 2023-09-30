package internal

type TRANSPORT_TYPE int
type READ_OR_WRITE int

const (
	READ READ_OR_WRITE = iota
	WRITE
)

// Set up the transport and begin the Write or Read loop
func MessageEntrypoint(action READ_OR_WRITE, config *MessangerConfig) {
	messanger := ConfigureMessanger(config)
	if action == WRITE {
		messanger.OutboundConnect()
		messanger.WriteLoop()
	} else if action == READ {
		messanger.ReadLoop()
	} else {
		panic("Illegal action type provided")
	}
}
