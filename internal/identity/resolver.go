package identity

import (
	"fmt"

	"github.com/rophy/kube-imds/internal/config"
)

type Resolver struct {
	byIP map[string]*config.Identity
}

func NewResolver(cfg *config.Config) *Resolver {
	byIP := make(map[string]*config.Identity, len(cfg.Identities))
	for i := range cfg.Identities {
		byIP[cfg.Identities[i].IP] = &cfg.Identities[i]
	}
	return &Resolver{byIP: byIP}
}

func (r *Resolver) Resolve(ip string) (*config.Identity, error) {
	id, ok := r.byIP[ip]
	if !ok {
		return nil, fmt.Errorf("no identity mapped for IP %s", ip)
	}
	return id, nil
}
