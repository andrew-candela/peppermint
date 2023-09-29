package cmd

import (
	"github.com/andrew-candela/peppermint/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCMD.AddCommand(hostCommand)
}

var hostCommand = &cobra.Command{
	Use:   "host",
	Short: "Host a web server that relays messages",
	Long:  "Host a webserver that forwards messages via a websocket connection to group members.",
	Run: func(cmd *cobra.Command, args []string) {
		config := internal.ParseConfigWithViper(group)
		internal.MessageEntrypoint(internal.HOST, config)
	},
}
