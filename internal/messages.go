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

	"google.golang.org/protobuf/proto"
)

type RawUDPMessage struct {
	content_buffer []byte
	length         int
	sender_address *net.UDPAddr
}

// This is sent by the writer, and is unconcerned with the transport.
// The public key here is the public key of the
// message writer, not the message reader.
type Message struct {
	content    []byte
	signature  []byte
	aes_key    []byte
	public_key []byte
}

// A receiver gets a serialized version of this struct.
// The sender_pubkey is used as a verifiable identifier
// for the sender.
type IncomingMessage struct {
	message       Message
	sender_pubkey []byte
}

// Serialized messages (PBMessage) are split into chunks, or 'Grams'.
// When the receiver gets a Gram with expect_more, it will store it
// in an array. Once it gets a gram with expect_more == false,
// it will concatenate all the grams into a PBMessage.
// Then it can deserialize the PBMessage and decrypt the content.
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

func (message *Message) Serialize() []byte {
	new_pb := &PBMessage{
		Content:   message.content,
		Signature: message.signature,
		AesKey:    message.aes_key,
		PublicKey: message.aes_key,
	}
	data, err := proto.Marshal(new_pb)
	CheckErrFatal(err)
	return data
}

func MessageFromBytes(buffer []byte) (Message, error) {
	new_message := &PBMessage{}
	err := proto.Unmarshal(buffer, new_message)
	return Message{
		content:    new_message.Content,
		signature:  new_message.Signature,
		aes_key:    new_message.AesKey,
		public_key: new_message.PublicKey,
	}, err
}

func (gram *Gram) Serialize() []byte {
	new_pb := &PBGram{
		Content:    gram.content,
		ExpectMore: gram.expect_more,
	}
	data, err := proto.Marshal(new_pb)
	CheckErrFatal(err)
	return data
}

func GramFromBytes(buffer []byte) (Gram, error) {
	new_gram := &PBGram{}
	err := proto.Unmarshal(buffer, new_gram)
	return Gram{
		content:     new_gram.Content,
		expect_more: new_gram.ExpectMore,
	}, err
}
