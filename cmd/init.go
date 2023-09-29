package cmd

import (
	"github.com/andrew-candela/peppermint/internal"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a config file for peppermint",
	Long: `Creates the ~/.peppermint/ directory, with a few files:
	config: a toml file
	peppermint_id_rsa: a randomly generated RSA private key file in PKCS #1, ASN.1 DER form.`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.InitPPMT()
	},
}

func init() {
	rootCMD.AddCommand(initCmd)
}
