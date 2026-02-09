package identity

import (
	"testing"

	"github.com/rophy/kube-imds/internal/config"
)

func TestResolve_Found(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IP: "10.0.1.10",
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-1",
					Namespace: "ns-1",
				},
			},
		},
	}
	r := NewResolver(cfg)
	id, err := r.Resolve("10.0.1.10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.ServiceAccount.Name != "vm-1" {
		t.Errorf("expected SA name 'vm-1', got %s", id.ServiceAccount.Name)
	}
}

func TestResolve_NotFound(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IP: "10.0.1.10",
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-1",
					Namespace: "ns-1",
				},
			},
		},
	}
	r := NewResolver(cfg)
	_, err := r.Resolve("10.0.1.99")
	if err == nil {
		t.Fatal("expected error for unknown IP")
	}
}

func TestResolve_MultipleIdentities(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IP: "10.0.1.10",
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-1",
					Namespace: "ns-1",
				},
			},
			{
				IP: "10.0.1.11",
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-2",
					Namespace: "ns-2",
				},
			},
		},
	}
	r := NewResolver(cfg)

	id, err := r.Resolve("10.0.1.11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.ServiceAccount.Name != "vm-2" {
		t.Errorf("expected SA name 'vm-2', got %s", id.ServiceAccount.Name)
	}
}
