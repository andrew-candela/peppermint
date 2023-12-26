package cmd

import (
	"github.com/andrew-candela/peppermint/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCMD.AddCommand(writeCommand)
}

var writeCommand = &cobra.Command{
	Use:   "write",
	Short: "Send your messages to a group.",
	Long: `
	Instantiates a Writer and starts a readline loop.
	Each message is signed, encrypted and then sent to its
	intended recipient.
	`,
	PreRun: configureLogger,
	Run: func(cmd *cobra.Command, args []string) {
		config := internal.ParseConfigWithViper(group)
		internal.MessageEntrypoint(internal.WRITE, config)
	},
}
