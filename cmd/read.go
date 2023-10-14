package cmd

import (
	"github.com/andrew-candela/peppermint/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCMD.AddCommand(readCommand)
}

var readCommand = &cobra.Command{
	Use:   "read",
	Short: "Listen for messages the group sends.",
	Long: `
	Listens for messages sent to the specified group.
	Prints the group messages into stdOut.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		config := internal.ParseConfigWithViper(group)
		internal.MessageEntrypoint(internal.READ, config)
	},
}
