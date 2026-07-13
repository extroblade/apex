package secretbox

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"
)

func newKey(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestRoundTrip(t *testing.T) {
	box, err := New(newKey(t))
	if err != nil {
		t.Fatal(err)
	}

	secret := []byte("my-iracing-encoded-password==")
	enc, err := box.Encrypt(secret)
	if err != nil {
		t.Fatal(err)
	}
	if enc == string(secret) {
		t.Fatal("ciphertext should not equal plaintext")
	}

	got, err := box.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(secret) {
		t.Errorf("round trip mismatch: got %q", got)
	}

	// Encrypting the same secret twice yields different ciphertext (random nonce).
	enc2, _ := box.Encrypt(secret)
	if enc == enc2 {
		t.Error("expected a fresh nonce per encryption")
	}
}

func TestWrongKeyFails(t *testing.T) {
	a, _ := New(newKey(t))
	b, _ := New(newKey(t))

	enc, err := a.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.Decrypt(enc); err == nil {
		t.Error("decrypt with the wrong key must fail")
	}
}

func TestBadKeySize(t *testing.T) {
	short := base64.StdEncoding.EncodeToString([]byte("too-short"))
	if _, err := New(short); !errors.Is(err, ErrKeySize) {
		t.Errorf("want ErrKeySize, got %v", err)
	}
}
