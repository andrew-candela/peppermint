package cmd

import (
	"github.com/andrew-candela/udpm/internal"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a config file for UDPM",
	Long: `Creates the ~/.udpm/ directory, with a few files:
	config: a toml file
	udpm_id_rsa: a randomly generated RSA private key file in PKCS #1, ASN.1 DER form.`,
	Run: func(cmd *cobra.Command, args []string) {
		internal.InitUDPM()
	},
}

func init() {
	rootCMD.AddCommand(initCmd)
}
