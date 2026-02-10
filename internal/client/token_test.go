package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchToken_Success(t *testing.T) {
	expiration := time.Now().Add(1 * time.Hour).Truncate(time.Second)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/token" {
			t.Errorf("expected /api/v1/token, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenRequestResponse{
			Kind:       "TokenRequest",
			APIVersion: "kube-imds/v1",
			Status: TokenRequestStatus{
				Token:               "test-token-abc",
				ExpirationTimestamp: expiration,
			},
		})
	}))
	defer srv.Close()

	resp, err := FetchToken(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status.Token != "test-token-abc" {
		t.Errorf("expected token 'test-token-abc', got %s", resp.Status.Token)
	}
	if !resp.Status.ExpirationTimestamp.Equal(expiration) {
		t.Errorf("expected expiration %v, got %v", expiration, resp.Status.ExpirationTimestamp)
	}
}

func TestFetchToken_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(StatusError{
			Kind:    "Status",
			Status:  "Failure",
			Message: "no identity mapped for IP 10.0.1.99",
			Code:    403,
		})
	}))
	defer srv.Close()

	_, err := FetchToken(srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); !contains(got, "403") || !contains(got, "no identity") {
		t.Errorf("expected error to contain '403' and 'no identity', got: %s", got)
	}
}

func TestFetchToken_EmptyToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenRequestResponse{
			Kind:   "TokenRequest",
			Status: TokenRequestStatus{Token: ""},
		})
	}))
	defer srv.Close()

	_, err := FetchToken(srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
