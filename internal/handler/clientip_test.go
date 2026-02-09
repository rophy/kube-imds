package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/rophy/kube-imds/internal/config"
)

func TestResolveClientIP_Default(t *testing.T) {
	cfg := &config.Config{}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.1.10:12345"

	got := resolveClientIP(req, cfg)
	if got != "10.0.1.10" {
		t.Errorf("expected 10.0.1.10, got %s", got)
	}
}

func TestResolveClientIP_HeaderIgnoredByDefault(t *testing.T) {
	cfg := &config.Config{}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.1.10:12345"
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	got := resolveClientIP(req, cfg)
	if got != "10.0.1.10" {
		t.Errorf("headers should be ignored by default; expected 10.0.1.10, got %s", got)
	}
}

func TestResolveClientIP_XForwardedFor(t *testing.T) {
	cfg := &config.Config{ClientIPHeader: "X-Forwarded-For"}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.96.1.5:12345"
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	got := resolveClientIP(req, cfg)
	if got != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", got)
	}
}

func TestResolveClientIP_XForwardedFor_MultipleIPs(t *testing.T) {
	cfg := &config.Config{ClientIPHeader: "X-Forwarded-For"}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.96.1.5:12345"
	req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.96.2.3")

	got := resolveClientIP(req, cfg)
	if got != "192.168.1.100" {
		t.Errorf("expected leftmost IP 192.168.1.100, got %s", got)
	}
}

func TestResolveClientIP_XEnvoyExternalAddress(t *testing.T) {
	cfg := &config.Config{ClientIPHeader: "X-Envoy-External-Address"}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.96.1.5:12345"
	req.Header.Set("X-Envoy-External-Address", "192.168.1.100")

	got := resolveClientIP(req, cfg)
	if got != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", got)
	}
}

func TestResolveClientIP_XRealIP(t *testing.T) {
	cfg := &config.Config{ClientIPHeader: "X-Real-IP"}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.96.1.5:12345"
	req.Header.Set("X-Real-IP", "192.168.1.100")

	got := resolveClientIP(req, cfg)
	if got != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", got)
	}
}

func TestResolveClientIP_HeaderMissing_FallbackToRemoteAddr(t *testing.T) {
	cfg := &config.Config{ClientIPHeader: "X-Forwarded-For"}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.96.1.5:12345"

	got := resolveClientIP(req, cfg)
	if got != "10.96.1.5" {
		t.Errorf("without header, should fall back to RemoteAddr; expected 10.96.1.5, got %s", got)
	}
}

func TestResolveClientIP_RemoteAddrWithoutPort(t *testing.T) {
	cfg := &config.Config{}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.1.10"

	got := resolveClientIP(req, cfg)
	if got != "10.0.1.10" {
		t.Errorf("expected 10.0.1.10, got %s", got)
	}
}
