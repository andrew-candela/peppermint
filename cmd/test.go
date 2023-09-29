package cmd

import (
	"github.com/andrew-candela/peppermint/internal"
	"github.com/spf13/cobra"
)

func init() {
	rootCMD.AddCommand(testCommand)
}

var testCommand = &cobra.Command{
	Use:   "test",
	Short: "do something custom - runs test.go",
	Run: func(cmd *cobra.Command, args []string) {
		internal.RunTestCommand()
	},
}
