package mail

import (
	"context"
	"strings"
	"testing"
)

func TestStripAngle(t *testing.T) {
	cases := map[string]string{
		"ContentPilot <no-reply@contentpilot.app>": "no-reply@contentpilot.app",
		"plain@contentpilot.app":                    "plain@contentpilot.app",
		"<bare@contentpilot.app>":                   "bare@contentpilot.app",
		"No angle here": "No angle here",
	}
	for in, want := range cases {
		if got := stripAngle(in); got != want {
			t.Errorf("stripAngle(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDisabledMailerNoOps(t *testing.T) {
	m := New(Config{}) // no Host → disabled
	if m.Enabled() {
		t.Error("Enabled() = true, want false when Host is empty")
	}
	// Send must not error and must not block — it's a no-op.
	if err := m.Send(context.Background(), "x@y.com", "subj", "body"); err != nil {
		t.Errorf("disabled Send returned %v, want nil", err)
	}
}

func TestBuildMessageHasHeadersAndBody(t *testing.T) {
	msg := buildMessage("ContentPilot <no-reply@contentpilot.app>", "user@x.com", "Reset your password", "Click here:\nhttps://app/reset?token=abc")
	s := string(msg)
	// Headers are RFC 5322 (\r\n-terminated); the body keeps its own newlines.
	for _, want := range []string{
		"From: ContentPilot <no-reply@contentpilot.app>\r\n",
		"To: user@x.com\r\n",
		"Subject: Reset your password\r\n",
		"Content-Type: text/plain; charset=utf-8\r\n",
		"Auto-Submitted: auto-generated\r\n",
		"\r\nClick here:\nhttps://app/reset?token=abc",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("message missing %q\n--- got ---\n%s", want, s)
		}
	}
}
