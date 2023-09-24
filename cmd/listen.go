package cmd

import (
	"github.com/andrew-candela/udpm/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCMD.AddCommand(readCommand)
}

var readCommand = &cobra.Command{
	Use:   "read",
	Short: "Listen for messages the group sends.",
	Long:  `Prints the messages sent to the group into stdOut.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := internal.ParseConfigWithViper(group)
		transport_type := internal.WEB
		if use_udp {
			transport_type = internal.UDP
		}
		internal.MessageEntrypoint(transport_type, internal.READ, config)
	},
}
