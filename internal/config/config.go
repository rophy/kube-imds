package config

import (
	"fmt"
	"net"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Identities       []Identity `yaml:"identities"`
	Defaults         Defaults   `yaml:"defaults"`
	ClientIPHeader   string     `yaml:"clientIPHeader,omitempty"`
}

type Identity struct {
	IPs            []string          `yaml:"ips"`
	ServiceAccount ServiceAccountRef `yaml:"serviceAccount"`
	TokenSpec      *TokenSpec        `yaml:"tokenSpec,omitempty"`
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
	seenIPs := make(map[string]int)   // canonical IP -> identity index
	seenCIDRs := make(map[string]int) // normalized CIDR -> identity index

	type cidrRecord struct {
		net   *net.IPNet
		index int
	}
	var allCIDRs []cidrRecord
	var allExactIPs []struct {
		ip    net.IP
		index int
	}

	for i, id := range c.Identities {
		if len(id.IPs) == 0 {
			return fmt.Errorf("identities[%d]: ips is required", i)
		}
		if id.ServiceAccount.Name == "" {
			return fmt.Errorf("identities[%d]: serviceAccount.name is required", i)
		}
		if id.ServiceAccount.Namespace == "" {
			return fmt.Errorf("identities[%d]: serviceAccount.namespace is required", i)
		}

		for _, entry := range id.IPs {
			if strings.Contains(entry, "/") {
				_, ipNet, err := net.ParseCIDR(entry)
				if err != nil {
					return fmt.Errorf("identities[%d]: invalid CIDR %q: %w", i, entry, err)
				}
				normalized := ipNet.String()
				if prev, ok := seenCIDRs[normalized]; ok {
					return fmt.Errorf("identities[%d]: duplicate CIDR %q (also in identities[%d])", i, normalized, prev)
				}
				seenCIDRs[normalized] = i
				allCIDRs = append(allCIDRs, cidrRecord{net: ipNet, index: i})
			} else {
				ip := net.ParseIP(entry)
				if ip == nil {
					return fmt.Errorf("identities[%d]: invalid IP %q", i, entry)
				}
				canonical := ip.String()
				if prev, ok := seenIPs[canonical]; ok {
					return fmt.Errorf("identities[%d]: duplicate IP %q (also in identities[%d])", i, canonical, prev)
				}
				seenIPs[canonical] = i
				allExactIPs = append(allExactIPs, struct {
					ip    net.IP
					index int
				}{ip: ip, index: i})
			}
		}
	}

	// Cross-check: exact IPs falling in CIDRs from other identities
	for _, ipRec := range allExactIPs {
		for _, cidrRec := range allCIDRs {
			if cidrRec.index != ipRec.index && cidrRec.net.Contains(ipRec.ip) {
				return fmt.Errorf("identities[%d]: IP %s overlaps with CIDR %s in identities[%d]",
					ipRec.index, ipRec.ip.String(), cidrRec.net.String(), cidrRec.index)
			}
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
