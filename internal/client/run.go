package client

import (
	"context"
	"log"
	"net/http"
	"time"
)

const (
	renewalFraction   = 0.8
	initialBackoff    = 1 * time.Second
	maxBackoff        = 60 * time.Second
	backoffMultiplier = 2.0
)

// Run enters the token renewal loop. It fetches a token immediately, writes it
// to disk, then sleeps until 80% of the token's lifetime before renewing.
// It returns when ctx is cancelled.
func Run(ctx context.Context, cfg *Config) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	backoff := initialBackoff

	for {
		tokenResp, err := FetchToken(httpClient, cfg.Endpoint)
		if err != nil {
			log.Printf("error fetching token: %v (retrying in %s)", err, backoff)
			if !sleep(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		backoff = initialBackoff

		if err := WriteFileAtomic(cfg.TokenPath, []byte(tokenResp.Status.Token), 0600); err != nil {
			log.Printf("error writing token file: %v (retrying in %s)", err, backoff)
			if !sleep(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}

		log.Printf("wrote token to %s (expires %s)", cfg.TokenPath, tokenResp.Status.ExpirationTimestamp.Format(time.RFC3339))

		wait := renewalDuration(tokenResp.Status.ExpirationTimestamp)
		log.Printf("next renewal in %s", wait)

		if !sleep(ctx, wait) {
			return
		}
	}
}

// renewalDuration returns 80% of the remaining lifetime, with a 1s floor.
func renewalDuration(expiration time.Time) time.Duration {
	remaining := time.Until(expiration)
	renewal := time.Duration(float64(remaining) * renewalFraction)
	if renewal < 1*time.Second {
		renewal = 1 * time.Second
	}
	return renewal
}

func nextBackoff(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * backoffMultiplier)
	if next > maxBackoff {
		next = maxBackoff
	}
	return next
}

// sleep waits for d or until ctx is cancelled. Returns true if d elapsed.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
