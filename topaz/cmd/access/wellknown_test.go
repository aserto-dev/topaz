package access

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestWellKnownURLUsesHTTPWhenPlaintext(t *testing.T) {
	t.Parallel()

	reqURL, err := url.Parse("https://example.com/.well-known/authzen-configuration")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	got, err := wellKnownURL(reqURL, true)
	if err != nil {
		t.Fatalf("failed to build URL: %v", err)
	}

	if got.String() != "http://example.com/.well-known/authzen-configuration" {
		t.Fatalf("expected plaintext URL, got %q", got.String())
	}
}

func TestWellKnownClientSupportsInsecureTLS(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"issuer":"topaz"}`)
	}))
	defer srv.Close()

	resp, err := wellKnownClient(true).Get(srv.URL)
	if err != nil {
		t.Fatalf("expected insecure client to connect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestWellKnownURLMatchesPlaintextServer(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"issuer":"topaz"}`)
	}))
	defer srv.Close()

	reqURL, err := url.Parse(strings.Replace(srv.URL, "http://", "https://", 1))
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	plaintextURL, err := wellKnownURL(reqURL, true)
	if err != nil {
		t.Fatalf("failed to build URL: %v", err)
	}

	resp, err := wellKnownClient(false).Get(plaintextURL.String())
	if err != nil {
		t.Fatalf("expected plaintext client to connect: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
