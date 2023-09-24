package internal

type TRANSPORT_TYPE int
type READ_OR_WRITE int

const (
	UDP TRANSPORT_TYPE = iota
	WEB
)

const (
	READ READ_OR_WRITE = iota
	WRITE
	HOST
)

// Set up the transport and begin the Write or Read loop
func MessageEntrypoint(transport_type TRANSPORT_TYPE, action READ_OR_WRITE, config *MessangerConfig) {
	messanger := ConfigureMessanger(config, transport_type)
	if action == WRITE {
		messanger.OutboundConnect()
		messanger.WriteLoop()
	} else if action == READ {
		messanger.ReadLoop()
	} else if action == HOST {
		messanger.HostWeb()
	} else {
		panic("Illegal action type provided")
	}
}
