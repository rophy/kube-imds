package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Identities       []Identity `yaml:"identities"`
	Defaults         Defaults   `yaml:"defaults"`
	ClientIPHeader   string     `yaml:"clientIPHeader,omitempty"`
}

type Identity struct {
	IP             string          `yaml:"ip"`
	ServiceAccount ServiceAccountRef `yaml:"serviceAccount"`
	TokenSpec      *TokenSpec      `yaml:"tokenSpec,omitempty"`
}

type ServiceAccountRef struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type TokenSpec struct {
	Audiences         []string `yaml:"audiences,omitempty"`
	ExpirationSeconds *int64   `yaml:"expirationSeconds,omitempty"`
}

type Defaults struct {
	TokenSpec TokenSpec `yaml:"tokenSpec"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	seen := make(map[string]bool)
	for i, id := range c.Identities {
		if id.IP == "" {
			return fmt.Errorf("identities[%d]: ip is required", i)
		}
		if seen[id.IP] {
			return fmt.Errorf("identities[%d]: duplicate ip %q", i, id.IP)
		}
		seen[id.IP] = true
		if id.ServiceAccount.Name == "" {
			return fmt.Errorf("identities[%d]: serviceAccount.name is required", i)
		}
		if id.ServiceAccount.Namespace == "" {
			return fmt.Errorf("identities[%d]: serviceAccount.namespace is required", i)
		}
	}
	return nil
}

// ResolvedTokenSpec returns the effective TokenSpec for an identity,
// merging per-identity overrides with defaults.
func (c *Config) ResolvedTokenSpec(id *Identity) TokenSpec {
	spec := c.Defaults.TokenSpec
	if id.TokenSpec != nil {
		if len(id.TokenSpec.Audiences) > 0 {
			spec.Audiences = id.TokenSpec.Audiences
		}
		if id.TokenSpec.ExpirationSeconds != nil {
			spec.ExpirationSeconds = id.TokenSpec.ExpirationSeconds
		}
	}
	return spec
}
