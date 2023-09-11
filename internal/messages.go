/*
This file will house the logice required for message SerDe.

The types defined here will handle encryption/decryption as well.

A user will prepare a Message, that is an encrypted byte array of arbitrary length.
The Message is serialized to bytes with protpbuf (PBMessage).
The bytes from PBMessage are then broken up into an array of Grams,
which are []byte of length <= 1024
*/
package internal

import (
	"crypto/rsa"
	"net"
)

type RawUDPMessage struct {
	content_buffer []byte
	length         int
	sender_address *net.UDPAddr
}

type Message struct {
	content   []byte
	signature []byte
	aes_key   []byte
}

type IncomingMessage struct {
	message       Message
	sender_pubkey []byte
}

type Gram struct {
	content     []byte
	expect_more bool
}

// Encrypts the Message content, modifying the Message in place
func (message *Message) Encrypt(pub_key *rsa.PublicKey) {
	new_aes_key := GenerateRandomAESKey()
	ciphertext, err := AESEncrypt(message.content, new_aes_key)
	CheckErrFatal(err)
	encrypted_aes_key, err := RSAEncrypt(pub_key, new_aes_key)
	CheckErrFatal(err)
	message.content = ciphertext
	message.aes_key = encrypted_aes_key

}

func (message *Message) Sign(private_key *rsa.PrivateKey) {
	signature, err := RSASign(private_key, message.content)
	CheckErrFatal(err)
	message.signature = signature
}
