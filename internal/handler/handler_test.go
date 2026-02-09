package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/rophy/kube-imds/internal/config"
	"github.com/rophy/kube-imds/internal/identity"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HealthHandler()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %s", w.Body.String())
	}
}

func TestTokenHandler_UnknownIP(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs:            []string{"10.0.1.10"},
				ServiceAccount: config.ServiceAccountRef{Name: "vm-1", Namespace: "ns-1"},
			},
		},
	}
	resolver := identity.NewResolver(cfg)
	clientset := fake.NewSimpleClientset()
	h := NewTokenHandler(resolver, clientset, cfg)

	req := httptest.NewRequest("POST", "/api/v1/token", nil)
	req.RemoteAddr = "10.0.1.99:12345"
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}

	var status metav1.Status
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.Status != metav1.StatusFailure {
		t.Errorf("expected status 'Failure', got %s", status.Status)
	}
}

func TestTokenHandler_MintToken(t *testing.T) {
	exp := int64(3600)
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs:            []string{"10.0.1.10"},
				ServiceAccount: config.ServiceAccountRef{Name: "vm-1", Namespace: "ns-1"},
			},
		},
		Defaults: config.Defaults{
			TokenSpec: config.TokenSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &exp,
			},
		},
	}
	resolver := identity.NewResolver(cfg)

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "vm-1", Namespace: "ns-1"},
	}
	clientset := fake.NewSimpleClientset(sa)

	// Add a reactor so CreateToken returns a valid response
	clientset.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "token" {
			return false, nil, nil
		}
		return true, &authv1.TokenRequest{
			Status: authv1.TokenRequestStatus{
				Token: "fake-token",
			},
		}, nil
	})

	h := NewTokenHandler(resolver, clientset, cfg)

	req := httptest.NewRequest("POST", "/api/v1/token", nil)
	req.RemoteAddr = "10.0.1.10:12345"
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp authv1.TokenRequest
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Kind != "TokenRequest" {
		t.Errorf("expected kind 'TokenRequest', got %s", resp.Kind)
	}
	if resp.APIVersion != "authentication.k8s.io/v1" {
		t.Errorf("expected apiVersion 'authentication.k8s.io/v1', got %s", resp.APIVersion)
	}
	if resp.Status.Token != "fake-token" {
		t.Errorf("expected token 'fake-token', got %s", resp.Status.Token)
	}
}

func TestTokenHandler_GetNotAllowed(t *testing.T) {
	cfg := &config.Config{}
	resolver := identity.NewResolver(cfg)
	clientset := fake.NewSimpleClientset()
	h := NewTokenHandler(resolver, clientset, cfg)

	req := httptest.NewRequest("GET", "/api/v1/token", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestTokenHandler_ClientIPHeader(t *testing.T) {
	exp := int64(3600)
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs:            []string{"192.168.1.100"},
				ServiceAccount: config.ServiceAccountRef{Name: "vm-1", Namespace: "ns-1"},
			},
		},
		Defaults: config.Defaults{
			TokenSpec: config.TokenSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &exp,
			},
		},
		ClientIPHeader: "X-Forwarded-For",
	}
	resolver := identity.NewResolver(cfg)

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "vm-1", Namespace: "ns-1"},
	}
	clientset := fake.NewSimpleClientset(sa)
	clientset.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "token" {
			return false, nil, nil
		}
		return true, &authv1.TokenRequest{
			Status: authv1.TokenRequestStatus{Token: "fake-token"},
		}, nil
	})

	h := NewTokenHandler(resolver, clientset, cfg)

	req := httptest.NewRequest("POST", "/api/v1/token", nil)
	req.RemoteAddr = "10.96.1.5:12345"
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp authv1.TokenRequest
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status.Token != "fake-token" {
		t.Errorf("expected token 'fake-token', got %s", resp.Status.Token)
	}
}

func TestTokenHandler_CIDRMatch(t *testing.T) {
	exp := int64(3600)
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs:            []string{"10.0.1.0/24"},
				ServiceAccount: config.ServiceAccountRef{Name: "vm-subnet", Namespace: "ns-1"},
			},
		},
		Defaults: config.Defaults{
			TokenSpec: config.TokenSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &exp,
			},
		},
	}
	resolver := identity.NewResolver(cfg)

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "vm-subnet", Namespace: "ns-1"},
	}
	clientset := fake.NewSimpleClientset(sa)
	clientset.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() != "token" {
			return false, nil, nil
		}
		return true, &authv1.TokenRequest{
			Status: authv1.TokenRequestStatus{Token: "cidr-token"},
		}, nil
	})

	h := NewTokenHandler(resolver, clientset, cfg)

	req := httptest.NewRequest("POST", "/api/v1/token", nil)
	req.RemoteAddr = "10.0.1.42:12345"
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp authv1.TokenRequest
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status.Token != "cidr-token" {
		t.Errorf("expected token 'cidr-token', got %s", resp.Status.Token)
	}
}

func TestWriteStatusError(t *testing.T) {
	w := httptest.NewRecorder()
	writeStatusError(w, http.StatusNotFound, "resource %s not found", "foo")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var status metav1.Status
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.Message != "resource foo not found" {
		t.Errorf("expected message 'resource foo not found', got %s", status.Message)
	}
	if status.Code != 404 {
		t.Errorf("expected code 404, got %d", status.Code)
	}


}
