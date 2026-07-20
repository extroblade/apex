// Package auth handles application accounts: registration, password hashing,
// and opaque login sessions. (This is *our* app's login — linking a user's
// iRacing account comes later and is a separate concern.)
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	sessionTTL     = 30 * 24 * time.Hour
	minPasswordLen = 8
)

// User is the app account, safe to serialize to the client (no password hash).
type User struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	PendingEmail  string    `json:"pendingEmail,omitempty"`
	Nickname      string    `json:"nickname"`
	AvatarURL     string    `json:"avatarUrl"` // a data: URL, or "" if unset
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
}

// maxAvatarBytes caps the stored avatar data URL (~0.5 MB image, base64-inflated).
const maxAvatarBytes = 900_000

// Sentinel errors let handlers map failures to HTTP status codes via errors.Is.
var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidEmail       = errors.New("a valid email address is required")
	ErrNicknameTooLong    = errors.New("nickname must be 50 characters or fewer")
	ErrInvalidAvatar      = errors.New("avatar must be an image data URL")
	ErrAvatarTooLarge     = errors.New("avatar image is too large")
	ErrUnauthorized       = errors.New("unauthorized")
)

// Service groups the auth operations. It holds the DB handle its methods need —
// a common Go pattern (dependencies as struct fields, behavior as methods).
type Service struct {
	db     *sql.DB
	mailer Mailer
	baseURL string
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Mailer is the slice of the mail package the auth service needs. Depending on
// an interface (not the concrete *mail.Mailer) keeps the package loosely coupled
// and trivial to test with a fake.
type Mailer interface {
	Enabled() bool
	Send(ctx context.Context, to, subject, body string) error
}

// WithMailer injects the transactional-email sender. Without it, password reset
// and email-verification still run but no mail is delivered (dev/test).
func (s *Service) WithMailer(m Mailer, baseURL string) *Service {
	s.mailer = m
	s.baseURL = baseURL
	return s
}

// hashPassword returns a bcrypt hash; checkPassword verifies one. bcrypt is
// deliberately slow and salts each hash, which is what you want for passwords.
func hashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func checkPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// newSessionToken returns a random token to hand to the client, plus its hash
// to store in the DB. 32 random bytes = 256 bits of entropy, unguessable.
func newSessionToken() (token, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
