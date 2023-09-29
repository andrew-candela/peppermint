package internal

import (
	"testing"

	"github.com/spf13/viper"
)

// Test if the transport is correctly configured from the sample_config
func TestConfigTransport(t *testing.T) {
	viper.SetConfigName("sample_config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("..")
	err := viper.ReadInConfig()
	if err != nil {
		t.Errorf("could not load viper config: %v", err)
	}
	viper.Set("private_key_file", "../sample_private_key_rsa")
	config := ParseConfigWithViper("group_two")
	// test that the config was parsed successfully.
	if config.Users[1].Name != "Bill" {
		t.Errorf("2nd User name is wrong, %v", config.Users[1].Name)
	}
	messanger := ConfigureMessanger(config)
	recip_private_key := GenerateRandomKey()
	if err != nil {
		t.Errorf("could not sign message, %v", err)
	}
	message := Message{
		content:    []byte("Hello"),
		public_key: messanger.public_key,
	}
	message.Sign(messanger.private_key)
	message.Encrypt(&recip_private_key.PublicKey)
	err = message.Decrypt(recip_private_key)
	if err != nil {
		t.Error(err)
	}
	message.VerifySignature()
	if string(message.content) != "Hello" {
		t.Errorf("decrypted message content is not as expected: %v", string(message.content))
	}
}
