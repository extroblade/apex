// Package secretbox provides authenticated symmetric encryption (AES-256-GCM)
// for small secrets stored at rest — here, each user's iRacing encoded password.
//
// GCM is an AEAD cipher: it both encrypts and authenticates, so tampering with
// the ciphertext is detected on decrypt. Each message uses a fresh random nonce,
// which we prepend to the ciphertext.
package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	ErrKeySize   = errors.New("secretbox: key must decode to 32 bytes")
	ErrCiphertext = errors.New("secretbox: ciphertext too short or malformed")
)

// Box holds an initialized AEAD cipher.
type Box struct {
	gcm cipher.AEAD
}

// New builds a Box from a base64-encoded 32-byte (256-bit) key.
func New(base64Key string) (*Box, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, ErrKeySize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Box{gcm: gcm}, nil
}

// Encrypt returns base64(nonce || ciphertext), safe to store as text.
func (b *Box) Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, b.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	// Seal appends the ciphertext to its first argument (the nonce here), so the
	// result is nonce||ciphertext in one slice.
	sealed := b.gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt. It fails if the key is wrong or the data was tampered.
func (b *Box) Decrypt(encoded string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	ns := b.gcm.NonceSize()
	if len(raw) < ns {
		return nil, ErrCiphertext
	}
	nonce, ciphertext := raw[:ns], raw[ns:]
	return b.gcm.Open(nil, nonce, ciphertext, nil)
}
