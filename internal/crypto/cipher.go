package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// activeKey holds the 32-byte SHA-256 hash for the current session
var activeKey []byte

func SetPassword(password string) {
	hash := sha256.Sum256([]byte(password))
	activeKey = hash[:]
}

// Encrypt secures the raw packet using ChaCha20-Poly1305
func Encrypt(rawPacket []byte) ([]byte, error) {
	if activeKey == nil {
		return nil, errors.New("encryption key is not set")
	}

	aead, err := chacha20poly1305.New(activeKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encryptedPacket := aead.Seal(nonce, nonce, rawPacket, nil)
	return encryptedPacket, nil
}

// Decrypt authenticates and extracts the raw packet
func Decrypt(encryptedPacket []byte) ([]byte, error) {
	if activeKey == nil {
		return nil, errors.New("encryption key is not set")
	}

	aead, err := chacha20poly1305.New(activeKey)
	if err != nil {
		return nil, err
	}

	if len(encryptedPacket) < aead.NonceSize() {
		return nil, errors.New("packet is too short")
	}

	nonce := encryptedPacket[:aead.NonceSize()]
	cipherText := encryptedPacket[aead.NonceSize():]

	decrypted, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
