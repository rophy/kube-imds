package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	yaml := `
identities:
  - ip: "10.0.1.10"
    serviceAccount:
      name: "vm-1"
      namespace: "ns-1"
  - ip: "10.0.1.11"
    serviceAccount:
      name: "vm-2"
      namespace: "ns-2"
    tokenSpec:
      audiences: ["custom"]
      expirationSeconds: 7200
defaults:
  tokenSpec:
    audiences: ["api"]
    expirationSeconds: 3600
`
	path := writeTemp(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Identities) != 2 {
		t.Fatalf("expected 2 identities, got %d", len(cfg.Identities))
	}
	if cfg.Identities[0].IP != "10.0.1.10" {
		t.Errorf("expected IP 10.0.1.10, got %s", cfg.Identities[0].IP)
	}
	if cfg.Defaults.TokenSpec.Audiences[0] != "api" {
		t.Errorf("expected default audience 'api', got %s", cfg.Defaults.TokenSpec.Audiences[0])
	}
}

func TestValidation_MissingIP(t *testing.T) {
	yaml := `
identities:
  - serviceAccount:
      name: "vm-1"
      namespace: "ns-1"
`
	path := writeTemp(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing IP")
	}
}

func TestValidation_DuplicateIP(t *testing.T) {
	yaml := `
identities:
  - ip: "10.0.1.10"
    serviceAccount:
      name: "vm-1"
      namespace: "ns-1"
  - ip: "10.0.1.10"
    serviceAccount:
      name: "vm-2"
      namespace: "ns-2"
`
	path := writeTemp(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate IP")
	}
}

func TestValidation_MissingSAName(t *testing.T) {
	yaml := `
identities:
  - ip: "10.0.1.10"
    serviceAccount:
      namespace: "ns-1"
`
	path := writeTemp(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing serviceAccount.name")
	}
}

func TestValidation_MissingSANamespace(t *testing.T) {
	yaml := `
identities:
  - ip: "10.0.1.10"
    serviceAccount:
      name: "vm-1"
`
	path := writeTemp(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing serviceAccount.namespace")
	}
}

func TestResolvedTokenSpec_DefaultsOnly(t *testing.T) {
	exp := int64(3600)
	cfg := &Config{
		Defaults: Defaults{
			TokenSpec: TokenSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &exp,
			},
		},
	}
	id := &Identity{}
	spec := cfg.ResolvedTokenSpec(id)
	if spec.Audiences[0] != "api" {
		t.Errorf("expected audience 'api', got %s", spec.Audiences[0])
	}
	if *spec.ExpirationSeconds != 3600 {
		t.Errorf("expected expiration 3600, got %d", *spec.ExpirationSeconds)
	}
}

func TestResolvedTokenSpec_PerIdentityOverride(t *testing.T) {
	defaultExp := int64(3600)
	overrideExp := int64(7200)
	cfg := &Config{
		Defaults: Defaults{
			TokenSpec: TokenSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &defaultExp,
			},
		},
	}
	id := &Identity{
		TokenSpec: &TokenSpec{
			Audiences:         []string{"custom"},
			ExpirationSeconds: &overrideExp,
		},
	}
	spec := cfg.ResolvedTokenSpec(id)
	if spec.Audiences[0] != "custom" {
		t.Errorf("expected audience 'custom', got %s", spec.Audiences[0])
	}
	if *spec.ExpirationSeconds != 7200 {
		t.Errorf("expected expiration 7200, got %d", *spec.ExpirationSeconds)
	}
}

func TestResolvedTokenSpec_PartialOverride(t *testing.T) {
	defaultExp := int64(3600)
	cfg := &Config{
		Defaults: Defaults{
			TokenSpec: TokenSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &defaultExp,
			},
		},
	}
	id := &Identity{
		TokenSpec: &TokenSpec{
			Audiences: []string{"custom"},
		},
	}
	spec := cfg.ResolvedTokenSpec(id)
	if spec.Audiences[0] != "custom" {
		t.Errorf("expected audience 'custom', got %s", spec.Audiences[0])
	}
	if *spec.ExpirationSeconds != 3600 {
		t.Errorf("expected default expiration 3600, got %d", *spec.ExpirationSeconds)
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}
