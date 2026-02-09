package server

import (
	"log"
	"net/http"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/rophy/kube-imds/internal/config"
	"github.com/rophy/kube-imds/internal/handler"
	"github.com/rophy/kube-imds/internal/identity"
)

func New(cfg *config.Config, clientset kubernetes.Interface) *http.Server {
	resolver := identity.NewResolver(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.HealthHandler())
	mux.Handle("/api/v1/token", handler.NewTokenHandler(resolver, clientset, cfg))
	mux.Handle("/api/v1/selfsubjectreviews", handler.NewSelfSubjectReviewHandler(resolver, cfg))

	return &http.Server{
		Addr:         ":8080",
		Handler:      logMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}
