/*
This package holds some helper functions to
load and parse the UDPM configuration file.
*/

package internal

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const SAMPLE_CONFIG_FILE string = "sample_config.toml"

func copySampleConfigFile(example_config_path string, out_path string) {
	config_file_contents, err := os.ReadFile(example_config_path)
	CheckErrFatal(err)
	err = os.WriteFile(out_path, config_file_contents, os.ModePerm)
	CheckErrFatal(err)
}

// Create the sample UDPM config file in the user's local
// filesystem or die trying.
func createUDPMConfig() (string, string) {
	home, err := os.UserHomeDir()
	CheckErrFatal(err)
	udpm_path := filepath.Join(home, ".udpm")
	udpm_config := filepath.Join(udpm_path, "config")
	err = os.MkdirAll(udpm_path, os.ModePerm)
	CheckErrFatal(err)
	copySampleConfigFile(SAMPLE_CONFIG_FILE, udpm_config)
	return udpm_path, udpm_config
}

// Write the config file and generate a new, random RSA keyfile
func InitUDPM() {
	udpm_path, _ := createUDPMConfig()
	WriteKeyToDisk(GenerateRandomKey(), filepath.Join(udpm_path, "udpm_id_rsa"))
	fmt.Println("Created UDPM config and private key files in dir:", udpm_path)
}

type TransportConfig struct {
	Name       string
	Users      []RecipientConfig
	PrivateKey *rsa.PrivateKey
	Port       string
}

type RecipientConfig struct {
	Host string
	Port string
	Key  string
	Name string
}

func ParseConfigWithViper(group string) *TransportConfig {
	var group_config TransportConfig
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	keyFile := viper.GetString("private_key_file")
	if keyFile == "" {
		fmt.Print("Nil value for keyfile!\n")
		os.Exit(1)
	}
	err = viper.UnmarshalKey(group, &group_config)
	CheckErrFatal(err)
	key := ReadExistingKey(keyFile)
	group_config.PrivateKey = key
	group_config.Port = viper.GetString("listen.internal_port")
	return &group_config
}
