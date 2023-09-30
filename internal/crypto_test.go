package internal

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateRandomAESKey(t *testing.T) {
	k := GenerateRandomAESKey()
	fmt.Println(k)
	if len(k) != 32 {
		t.Errorf("key generated is %v bytes. Should be 32!", len(k))
	}
}

func TestAESEncrypt(t *testing.T) {
	plainText := []byte("Hello Andrew!")
	key := GenerateRandomAESKey()
	cipherText, err := AESEncrypt(plainText, key)
	if err != nil {
		t.Error(err)
	}
	recoveredText, err := AESDecrypt(cipherText, key)
	if err != nil {
		t.Error(err)
	}
	if string(plainText) != string(recoveredText) {
		t.Errorf("unexpected results!\nExpecting:%v\nGot:%v", plainText, recoveredText)
	}
}

func TestAESEncryptLong(t *testing.T) {
	plainText := []byte(strings.Repeat("Hello Andrew! ", 50))
	key := GenerateRandomAESKey()
	cipherText, _ := AESEncrypt(plainText, key)
	recoveredText, _ := AESDecrypt(cipherText, key)
	if string(plainText) != string(recoveredText) {
		t.Error("unexpected results!")
	}
}

func TestRSAEncrypt(t *testing.T) {
	key := GenerateRandomKey()
	pub_key := key.PublicKey
	clear_text := []byte("Hello world!")
	ciphertext, err := RSAEncrypt(&pub_key, clear_text)
	if err != nil {
		t.Error(err)
	}
	decrypted_text, err := RSADecrypt(key, ciphertext)
	if err != nil {
		t.Error(err)
	}
	if string(decrypted_text) != string(clear_text) {
		t.Errorf("%v != %v", decrypted_text, clear_text)
	}
}

func TestPublicKeyEncode(t *testing.T) {
	key := GenerateRandomKey()
	pub_key_rsa := key.PublicKey
	pub_bytes := PublicKeyToBytes(&pub_key_rsa)
	parsed_pub_key, err := BytesToPublicKey(pub_bytes)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(parsed_pub_key)
}
