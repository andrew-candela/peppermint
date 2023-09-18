/*
Holds all of the functions meant to deal with encryption.
Provides utilities to encrypt strings, create RSA keys etc.
*/

package internal

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh"
)

const LABEL = "myCoolMessagingApp"

// Generates a random 32 byte key
func GenerateRandomAESKey() (key []byte) {
	key = make([]byte, 32)
	_, err := rand.Read(key)
	CheckErrFatal(err)
	return key
}

// Encrypts a message with the given AES key
func AESEncrypt(message []byte, key []byte) ([]byte, error) {
	c_block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("unable to create new CipherBlock...%w", err)
	}
	gcm_cipher, err := cipher.NewGCM(c_block)
	if err != nil {
		return nil, fmt.Errorf("unable to create new GCM Cipher...%w", err)
	}
	nonce := make([]byte, gcm_cipher.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("unable to populate the Nonce...%w", err)
	}
	return gcm_cipher.Seal(nonce, nonce, message, nil), nil
}

func AESDecrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c_block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to create new CipherBlock...%w", err)
	}
	gcm_cipher, err := cipher.NewGCM(c_block)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to create new GCM Cipher...%w", err)
	}
	nonce_size := gcm_cipher.NonceSize()
	if len(ciphertext) < nonce_size {
		return []byte{}, fmt.Errorf("nonce size is greater than length of ciphertext...%w", err)
	}
	nonce, ciphertext := ciphertext[:nonce_size], ciphertext[nonce_size:]
	decrypted_bytes, err := gcm_cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("unable to decrypt ciphertext...%w", err)
	}
	return decrypted_bytes, nil
}

// Convert a byte slice to a hex string
func BytesToString(sig []byte) string {
	return hex.EncodeToString(sig)
}

// Take a JSON string (an array of ints) and build a []byte
func StringToBytes(sig_str string) ([]byte, error) {
	return hex.DecodeString(sig_str)
}

// Encrypts a message with the given public key.
func RSAEncrypt(publicKey *rsa.PublicKey, message []byte) ([]byte, error) {
	encrypted, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		[]byte(message),
		[]byte(LABEL),
	)
	if err != nil {
		new_err := fmt.Errorf("trouble encrypting string...%w", err)
		return nil, new_err
	}
	return encrypted, nil
}

func RSADecrypt(privateKey *rsa.PrivateKey, message []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptOAEP(
		sha256.New(),
		nil,
		privateKey,
		message,
		[]byte(LABEL),
	)
	if err != nil {
		new_err := fmt.Errorf("unable to decrypt message... %w", err)
		return nil, new_err
	}
	return decrypted, nil
}

func RSASign(key *rsa.PrivateKey, message []byte) (sig []byte, err error) {
	hashed := sha256.Sum256(message)
	sig, err = rsa.SignPKCS1v15(nil, key, crypto.SHA256, hashed[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return
	}
	return
}

// Verify that the given message was signed by the private key
// corresponding to the public key we have.
func RSAVerify(pub *rsa.PublicKey, message []byte, sig []byte) bool {
	digest := sha256.Sum256(message)
	err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], sig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error verifying signature: %s\n", err)
		return false
	}
	return true
}

// Reads an existing .pem or rsa keyfile and returns a
// reference to it.
func ReadExistingKey(keyFile string) *rsa.PrivateKey {
	keyfile, err := os.ReadFile(keyFile)
	if err != nil {
		fmt.Printf("Could not read file %v :%v\n", keyFile, err)
		os.Exit(1)
	}
	key, err := ssh.ParseRawPrivateKey(keyfile)
	CheckErrFatal(err)
	return key.(*rsa.PrivateKey)
}

// Returns a public RSA key from bytes.
// This is the inverse of PublicKeyToBytes
func ParsePublicKey(keyString []byte) *rsa.PublicKey {
	pKeyBlock, _ := pem.Decode(keyString)
	if pKeyBlock == nil {
		fmt.Println("Error in pem.Decode, keyblock is nil...")
		panic("Oops")
	}
	pubKeyInterface, err_two := x509.ParsePKIXPublicKey(pKeyBlock.Bytes)
	if err_two != nil {
		fmt.Println("Error in x509.ParsePKIXPublicKey...", err_two)
		panic(err_two)
	}
	pubKey := pubKeyInterface.(*rsa.PublicKey)
	return pubKey
}

func WriteKeyToDisk(key *rsa.PrivateKey, fileName string) {
	pemData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	err := os.WriteFile(fileName, pemData, 0600)
	CheckErrFatal(err)
}

// Produces the public key bytearray from the given private key
func EncodePublicKey(key *rsa.PrivateKey) []byte {
	pubKey := key.PublicKey
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&pubKey)
	if err != nil {
		panic(err)
	}
	pemData := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubKeyBytes,
		},
	)
	return pemData
}

func WritePublicKey(key *rsa.PrivateKey, fileName string) {
	bytes := EncodePublicKey(key)
	err := os.WriteFile(fileName, bytes, 0600)
	CheckErrFatal(err)
}

// Prints the public key byte array from a given private key
func DisplayPublicKey(key *rsa.PrivateKey) {
	pem_key := EncodePublicKey(key)
	fmt.Println(string(pem_key))
}

// Generates a new, random RSA private key
func GenerateRandomKey() *rsa.PrivateKey {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	CheckErrFatal(err)
	return k
}

// Converts a public RSA key into bytes.
// This is the inverse of ParsePublicKey.
func PublicKeyToBytes(key *rsa.PublicKey) []byte {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(key)
	CheckErrFatal(err)
	return pubKeyBytes
}
