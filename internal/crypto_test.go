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
