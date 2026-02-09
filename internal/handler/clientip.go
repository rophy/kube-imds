package handler

import (
	"net"
	"net/http"
	"strings"

	"github.com/rophy/kube-imds/internal/config"
)

// resolveClientIP determines the client IP from the request.
// When UseXForwardedFor is enabled, the leftmost IP from the
// X-Forwarded-For header is used. Otherwise, RemoteAddr is used.
func resolveClientIP(r *http.Request, cfg *config.Config) string {
	if cfg.UseXForwardedFor {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
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
