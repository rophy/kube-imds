package identity

import (
	"fmt"
	"net"
	"strings"

	"github.com/rophy/kube-imds/internal/config"
)

type cidrEntry struct {
	net      *net.IPNet
	identity *config.Identity
}

type Resolver struct {
	byIP  map[string]*config.Identity
	cidrs []cidrEntry
}

func NewResolver(cfg *config.Config) *Resolver {
	byIP := make(map[string]*config.Identity)
	var cidrs []cidrEntry

	for i := range cfg.Identities {
		id := &cfg.Identities[i]
		for _, entry := range id.IPs {
			if strings.Contains(entry, "/") {
				_, ipNet, err := net.ParseCIDR(entry)
				if err != nil {
					continue
				}
				cidrs = append(cidrs, cidrEntry{net: ipNet, identity: id})
			} else {
				ip := net.ParseIP(entry)
				if ip != nil {
					byIP[ip.String()] = id
				}
			}
		}
	}

	return &Resolver{byIP: byIP, cidrs: cidrs}
}

func (r *Resolver) Resolve(ip string) (*config.Identity, error) {
	parsed := net.ParseIP(ip)
	if parsed != nil {
		if id, ok := r.byIP[parsed.String()]; ok {
			return id, nil
		}
	}

	if parsed != nil {
		for _, entry := range r.cidrs {
			if entry.net.Contains(parsed) {
				return entry.identity, nil
			}
		}
	}

	return nil, fmt.Errorf("no identity mapped for IP %s", ip)
}
