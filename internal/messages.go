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
	"fmt"
	"net"
	"os"

	"google.golang.org/protobuf/proto"
)

const (
	// udpBuffer is 1024 bytes.
	// Marshalling as Protobuf adds 5 bytes.
	// We'll use 1000 bytes here to give some leeway
	GRAM_SIZE = 1000
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

// Produces a byte array by creating a PBMessage and then
// marshaling it to bytes.
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

// If a message is too long (encoded length is over 1024 bytes)
// then it will be split into one or more Grams
func SplitMessage(encoded_message []byte) [][]byte {
	message_length := len(encoded_message)
	expect_more := true
	var gram_list [][]byte
	for i := 0; i < message_length; i += GRAM_SIZE {
		end := i + GRAM_SIZE
		if end > message_length {
			end = message_length
		}
		if end == message_length {
			expect_more = false
		}
		gram_content := encoded_message[i:end]
		gram := Gram{
			content:     gram_content,
			expect_more: expect_more,
		}
		pb_gram := gram.Serialize()
		if len(pb_gram) > 1024 {
			fmt.Println("Encoded Gram is too long! Exiting the app...")
			os.Exit(1)
		}

		gram_list = append(gram_list, pb_gram)
	}
	return gram_list
}
