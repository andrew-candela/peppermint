/*
This package holds some helper functions to
load and parse the PPMT configuration file.
*/

package internal

import (
	"crypto/rsa"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const SAMPLE_CONFIG_FILE string = "sample_config.toml"

//go:embed sample_config.toml
var sample_config embed.FS

func copySampleConfigFile(out_path string) {
	config_file_contents, err := sample_config.ReadFile(SAMPLE_CONFIG_FILE)
	CheckErrFatal(err)
	err = os.WriteFile(out_path, config_file_contents, os.ModePerm)
	CheckErrFatal(err)
}

// Create the sample PPMT config file in the user's local
// filesystem or die trying.
func createPPMTConfig() (string, string) {
	home, err := os.UserHomeDir()
	CheckErrFatal(err)
	ppmt_path := filepath.Join(home, ".peppermint")
	ppmt_config := filepath.Join(ppmt_path, "config")
	err = os.MkdirAll(ppmt_path, os.ModePerm)
	CheckErrFatal(err)
	copySampleConfigFile(ppmt_config)
	return ppmt_path, ppmt_config
}

// Write the config file and generate a new, random RSA keyfile
func InitPPMT() {
	ppmt_path, _ := createPPMTConfig()
	WriteKeyToDisk(GenerateRandomKey(), filepath.Join(ppmt_path, "id_rsa"))
	fmt.Println("Created peppermint config and private key files in dir:", ppmt_path)
}

type MessangerConfig struct {
	Users      []RecipientConfig
	PrivateKey *rsa.PrivateKey
	URL        string
	Port       string
}

type RecipientConfig struct {
	Key  string
	Name string
}

// Parse the config with Viper and handle errors
func ParseConfig() {
	if err := viper.ReadInConfig(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func ParseConfigWithViper(group string) *MessangerConfig {
	var group_config MessangerConfig
	ParseConfig()
	keyFile := viper.GetString("private_key_file")
	if keyFile == "" {
		fmt.Print("Nil value for keyfile!\n")
		os.Exit(1)
	}
	err := viper.UnmarshalKey(group, &group_config)
	CheckErrFatal(err)
	key := ReadExistingKey(keyFile)
	group_config.PrivateKey = key
	return &group_config
}

// Checks to see if --verbose is set by the user
// by checking the 'verbose' viper setting
func CheckDebug() bool {
	return viper.GetBool("verbose")
}
