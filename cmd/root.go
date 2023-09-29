package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var group string

var rootCMD = &cobra.Command{
	Use:   "peppermint",
	Short: "",
	Long: `
	Peppermint: Peer to Peer Messaging in a Terminal 🍬

	Host a Peppermint server or subscribe and publish to a group.`,
}

func Execute() {
	rootCMD.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCMD.PersistentFlags().StringVarP(&group, "group", "g", "", "Group Name to listen or write to")
}

// Sets up where Viper will look for the config file
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/peppermint")
	viper.AddConfigPath("$HOME/.peppermint")
	viper.AddConfigPath(".")
}
