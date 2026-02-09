package identity

import (
	"testing"

	"github.com/rophy/kube-imds/internal/config"
)

func TestResolve_Found(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs: []string{"10.0.1.10"},
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
				IPs: []string{"10.0.1.10"},
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
				IPs: []string{"10.0.1.10"},
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-1",
					Namespace: "ns-1",
				},
			},
			{
				IPs: []string{"10.0.1.11"},
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

func TestResolve_CIDR(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs: []string{"10.0.1.0/24"},
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-subnet",
					Namespace: "ns-1",
				},
			},
		},
	}
	r := NewResolver(cfg)

	id, err := r.Resolve("10.0.1.50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.ServiceAccount.Name != "vm-subnet" {
		t.Errorf("expected SA name 'vm-subnet', got %s", id.ServiceAccount.Name)
	}
}

func TestResolve_CIDRNoMatch(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs: []string{"10.0.1.0/24"},
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-subnet",
					Namespace: "ns-1",
				},
			},
		},
	}
	r := NewResolver(cfg)

	_, err := r.Resolve("10.0.2.50")
	if err == nil {
		t.Fatal("expected error for IP outside CIDR")
	}
}

func TestResolve_MultipleIPsSameIdentity(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs: []string{"10.0.1.10", "10.0.1.11", "10.0.1.12"},
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-pool",
					Namespace: "ns-1",
				},
			},
		},
	}
	r := NewResolver(cfg)

	for _, ip := range []string{"10.0.1.10", "10.0.1.11", "10.0.1.12"} {
		id, err := r.Resolve(ip)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", ip, err)
		}
		if id.ServiceAccount.Name != "vm-pool" {
			t.Errorf("for %s: expected SA name 'vm-pool', got %s", ip, id.ServiceAccount.Name)
		}
	}
}

func TestResolve_ExactIPAndCIDR(t *testing.T) {
	cfg := &config.Config{
		Identities: []config.Identity{
			{
				IPs: []string{"10.0.1.50", "10.0.1.0/24"},
				ServiceAccount: config.ServiceAccountRef{
					Name:      "vm-combined",
					Namespace: "ns-1",
				},
			},
		},
	}
	r := NewResolver(cfg)

	// Exact IP match
	id, err := r.Resolve("10.0.1.50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.ServiceAccount.Name != "vm-combined" {
		t.Errorf("expected 'vm-combined', got %s", id.ServiceAccount.Name)
	}

	// CIDR match
	id2, err := r.Resolve("10.0.1.99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id2.ServiceAccount.Name != "vm-combined" {
		t.Errorf("expected 'vm-combined', got %s", id2.ServiceAccount.Name)
	}
}
