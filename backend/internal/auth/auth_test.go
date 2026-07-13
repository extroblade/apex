package auth

import "testing"

func TestPasswordHashing(t *testing.T) {
	hash, err := hashPassword("correct horse battery")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "correct horse battery" {
		t.Fatal("password was not hashed")
	}
	if !checkPassword(hash, "correct horse battery") {
		t.Error("correct password should verify")
	}
	if checkPassword(hash, "wrong password") {
		t.Error("wrong password should not verify")
	}
}

func TestSessionTokensAreUnique(t *testing.T) {
	t1, h1, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	t2, _, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if t1 == t2 {
		t.Error("tokens should be unique")
	}
	if h1 != hashToken(t1) {
		t.Error("stored hash must match hashToken of the raw token")
	}
	if h1 == t1 {
		t.Error("stored hash must differ from the raw token")
	}
}

func TestValidEmail(t *testing.T) {
	for _, ok := range []string{"a@b.com", "user.name@example.co"} {
		if !validEmail(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "nope", "@no.com", "a@", "a b@c.com"} {
		if validEmail(bad) {
			t.Errorf("%q should be invalid", bad)
		}
	}
}
