package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var group string

var rootCMD = &cobra.Command{
	Use:   "udpm",
	Short: "udpm: send messages to your friends via UDP or websockets.",
	Long:  "udpm: Entrypoint to the messaging tool.",
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
	viper.AddConfigPath("/etc/udpm")
	viper.AddConfigPath("$HOME/.udpm")
	viper.AddConfigPath(".")
}
