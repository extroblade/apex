package racing

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"apex/internal/iracing"
	"apex/internal/secretbox"
)

func testBox(t *testing.T) *secretbox.Box {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	box, err := secretbox.New(base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatal(err)
	}
	return box
}

// fakeClient implements APIClient with no network.
type fakeClient struct{ tokenSet string }

func (f *fakeClient) SetToken(t string) { f.tokenSet = t }
func (f *fakeClient) Info(context.Context) (iracing.Member, error) {
	return iracing.Member{CustID: 1}, nil
}
func (f *fakeClient) Member(_ context.Context, id int) (iracing.Member, error) {
	return iracing.Member{CustID: id}, nil
}
func (f *fakeClient) Career(context.Context, int) ([]iracing.CareerStat, error)      { return nil, nil }
func (f *fakeClient) RecentRaces(context.Context, int) ([]iracing.RecentRace, error) { return nil, nil }
func (f *fakeClient) SearchDrivers(context.Context, string) ([]iracing.DriverSearchResult, error) {
	return nil, nil
}
func (f *fakeClient) Cars(context.Context) ([]iracing.Car, error)             { return nil, nil }
func (f *fakeClient) Tracks(context.Context) ([]iracing.CatalogTrack, error)  { return nil, nil }
func (f *fakeClient) Series(context.Context) ([]iracing.CatalogSeries, error) { return nil, nil }

func newFakeService() (*Service, *fakeClient) {
	fake := &fakeClient{}
	s := NewService(nil, nil, func() APIClient { return fake }, iracing.OAuthConfig{})
	return s, fake
}

// A cached, unexpired access token should be used directly — no DB, no refresh —
// and applied to the client the factory produces.
func TestClientForReusesCachedToken(t *testing.T) {
	s, fake := newFakeService()
	s.sessions[7] = &cachedToken{accessToken: "tok", custID: 99, expires: time.Now().Add(time.Hour)}

	client, custID, err := s.clientFor(context.Background(), 7)
	if err != nil {
		t.Fatalf("clientFor: %v", err)
	}
	if custID != 99 {
		t.Errorf("custID: want 99, got %d", custID)
	}
	if client != fake {
		t.Error("expected the factory's client")
	}
	if fake.tokenSet != "tok" {
		t.Errorf("access token should be applied to the client, got %q", fake.tokenSet)
	}
}

func TestBeginLinkRemembersState(t *testing.T) {
	s, _ := newFakeService()
	s.box = testBox(t)
	s.oauth = iracing.OAuthConfig{ClientID: "cid", RedirectURI: "https://app/cb", BaseURL: "https://oauth.example/oauth2"}

	url, err := s.BeginLink(42)
	if err != nil {
		t.Fatalf("BeginLink: %v", err)
	}
	if url == "" {
		t.Fatal("expected an authorize URL")
	}
	if len(s.pending) != 1 {
		t.Fatalf("expected 1 pending request, got %d", len(s.pending))
	}
	for _, p := range s.pending {
		if p.userID != 42 || p.verifier == "" {
			t.Errorf("bad pending entry: %+v", p)
		}
	}
}

func TestForgetSession(t *testing.T) {
	s, _ := newFakeService()
	s.sessions[7] = &cachedToken{expires: time.Now().Add(time.Hour)}
	s.forgetSession(7)
	if _, ok := s.sessions[7]; ok {
		t.Error("session should be removed")
	}
}
