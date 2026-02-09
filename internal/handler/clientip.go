package handler

import (
	"net"
	"net/http"
	"strings"

	"github.com/rophy/kube-imds/internal/config"
)

// resolveClientIP determines the client IP from the request.
// When ClientIPHeader is set, the value of that header is used.
// For multi-value headers like X-Forwarded-For, the leftmost IP is used.
// Otherwise, RemoteAddr is used.
func resolveClientIP(r *http.Request, cfg *config.Config) string {
	if cfg.ClientIPHeader != "" {
		if val := r.Header.Get(cfg.ClientIPHeader); val != "" {
			// Handle comma-separated values (e.g. X-Forwarded-For: client, proxy1)
			ip := strings.TrimSpace(strings.SplitN(val, ",", 2)[0])
			if ip != "" {
				return ip
			}
		}
	}

	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return clientIP
}
