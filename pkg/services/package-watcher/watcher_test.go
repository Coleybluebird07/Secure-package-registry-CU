package packagewatcher

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testHTTPClient(t *testing.T, handler func(*http.Request) (string, int)) *http.Client {
	t.Helper()

	return &http.Client{
		Timeout: 2 * time.Second,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, status := handler(req)
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}
}

func TestWatcherGetLatestPyPIVersion(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			if req.URL.Host != "pypi.org" {
				t.Fatalf("unexpected host: %s", req.URL.Host)
			}
			if req.URL.Path != "/pypi/requests/json" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}

			return `{"info":{"version":"2.33.1"}}`, http.StatusOK
		}),
	}

	version, err := w.getLatestPyPIVersion(context.Background(), "requests")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if version != "2.33.1" {
		t.Fatalf("expected 2.33.1, got %s", version)
	}
}

func TestWatcherGetLatestCargoVersionPrefersMaxStableVersion(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			if req.URL.Host != "crates.io" {
				t.Fatalf("unexpected host: %s", req.URL.Host)
			}
			if req.URL.Path != "/api/v1/crates/serde_plain" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}

			if got := req.Header.Get("User-Agent"); got == "" {
				t.Fatal("expected user-agent header to be set")
			}

			return `{"crate":{"newest_version":"1.0.3","max_stable_version":"1.0.2"}}`, http.StatusOK
		}),
	}

	version, err := w.getLatestCargoVersion(context.Background(), "serde_plain")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if version != "1.0.2" {
		t.Fatalf("expected max stable version 1.0.2, got %s", version)
	}
}

func TestWatcherGetLatestCargoVersionFallsBackToNewestVersion(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			return `{"crate":{"newest_version":"1.0.3","max_stable_version":""}}`, http.StatusOK
		}),
	}

	version, err := w.getLatestCargoVersion(context.Background(), "serde_plain")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if version != "1.0.3" {
		t.Fatalf("expected newest version 1.0.3, got %s", version)
	}
}

func TestWatcherGetLatestGoVersion(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			if req.URL.Host != "proxy.golang.org" {
				t.Fatalf("unexpected host: %s", req.URL.Host)
			}
			if req.URL.Path != "/github.com/pkg/errors/@latest" {
				t.Fatalf("unexpected path: %s", req.URL.Path)
			}

			return `{"Version":"v0.9.1","Time":"2020-01-01T00:00:00Z"}`, http.StatusOK
		}),
	}

	version, err := w.getLatestGoVersion(context.Background(), "github.com/pkg/errors")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if version != "v0.9.1" {
		t.Fatalf("expected v0.9.1, got %s", version)
	}
}

func TestWatcherGetLatestVersionUnsupportedEcosystem(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			t.Fatal("http client should not be called for unsupported ecosystem")
			return "", http.StatusInternalServerError
		}),
	}

	_, err := w.getLatestVersion(context.Background(), "rubygems", "rails")
	if err == nil {
		t.Fatal("expected unsupported ecosystem error")
	}
	if !strings.Contains(err.Error(), "unsupported ecosystem") {
		t.Fatalf("expected unsupported ecosystem error, got %v", err)
	}
}

func TestWatcherGetLatestPyPIVersionNotFound(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			return `{"message":"not found"}`, http.StatusNotFound
		}),
	}

	_, err := w.getLatestPyPIVersion(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !strings.Contains(err.Error(), "PyPI package not found") {
		t.Fatalf("expected PyPI package not found error, got %v", err)
	}
}

func TestWatcherGetLatestPyPIVersionMalformedJSON(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			return `{bad json`, http.StatusOK
		}),
	}

	_, err := w.getLatestPyPIVersion(context.Background(), "requests")
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decoding PyPI metadata") {
		t.Fatalf("expected decoding PyPI metadata error, got %v", err)
	}
}

func TestWatcherGetLatestCargoVersionNotFound(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			return `{"errors":[{"detail":"Not Found"}]}`, http.StatusNotFound
		}),
	}

	_, err := w.getLatestCargoVersion(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !strings.Contains(err.Error(), "crate not found") {
		t.Fatalf("expected crate not found error, got %v", err)
	}
}

func TestWatcherGetLatestGoVersionNotFound(t *testing.T) {
	w := &watcher{
		httpClient: testHTTPClient(t, func(req *http.Request) (string, int) {
			return `not found`, http.StatusNotFound
		}),
	}

	_, err := w.getLatestGoVersion(context.Background(), "example.com/does-not-exist")
	if err == nil {
		t.Fatal("expected Go module not found error")
	}
	if !strings.Contains(err.Error(), "go module not found") {
		t.Fatalf("expected go module not found error, got %v", err)
	}
}
