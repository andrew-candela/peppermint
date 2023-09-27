package cmd

import (
	"github.com/andrew-candela/udpm/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCMD.AddCommand(writeCommand)
}

var writeCommand = &cobra.Command{
	Use:   "write",
	Short: "Send your messages to a group.",
	Long: `Instantiates a Writer and starts a readline loop.
	Each message is signed, encrypted and then sent to its
	intended recipient via the appropriate transport (WEB or UDP).
	`,
	Run: func(cmd *cobra.Command, args []string) {
		config := internal.ParseConfigWithViper(group)
		transport_type := internal.WEB
		if use_udp {
			transport_type = internal.UDP
		}
		internal.MessageEntrypoint(transport_type, internal.WRITE, config)
	},
}
