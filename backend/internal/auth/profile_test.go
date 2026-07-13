package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// These cases fail validation *before* any DB access, so a nil-DB Service is fine.
func TestUpdateProfileValidation(t *testing.T) {
	s := &Service{}
	ctx := context.Background()

	if _, err := s.UpdateProfile(ctx, 1, strings.Repeat("x", 51), "a@b.com"); !errors.Is(err, ErrNicknameTooLong) {
		t.Errorf("long nickname: want ErrNicknameTooLong, got %v", err)
	}
	if _, err := s.UpdateProfile(ctx, 1, "ok", "not-an-email"); !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("bad email: want ErrInvalidEmail, got %v", err)
	}
}

func TestUpdateAvatarValidation(t *testing.T) {
	s := &Service{}
	ctx := context.Background()

	if _, err := s.UpdateAvatar(ctx, 1, "https://evil.example/x.png"); !errors.Is(err, ErrInvalidAvatar) {
		t.Errorf("non-data URL: want ErrInvalidAvatar, got %v", err)
	}
	tooBig := "data:image/png;base64," + strings.Repeat("A", maxAvatarBytes)
	if _, err := s.UpdateAvatar(ctx, 1, tooBig); !errors.Is(err, ErrAvatarTooLarge) {
		t.Errorf("oversized avatar: want ErrAvatarTooLarge, got %v", err)
	}
}

func TestChangePasswordRejectsWeak(t *testing.T) {
	s := &Service{}
	if err := s.ChangePassword(context.Background(), 1, "whatever", "short"); !errors.Is(err, ErrWeakPassword) {
		t.Errorf("weak new password: want ErrWeakPassword, got %v", err)
	}
}
