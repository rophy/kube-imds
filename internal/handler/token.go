package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/rophy/kube-imds/internal/config"
	"github.com/rophy/kube-imds/internal/identity"
)

type TokenHandler struct {
	resolver  *identity.Resolver
	clientset kubernetes.Interface
	config    *config.Config
}

func NewTokenHandler(resolver *identity.Resolver, clientset kubernetes.Interface, cfg *config.Config) *TokenHandler {
	return &TokenHandler{
		resolver:  resolver,
		clientset: clientset,
		config:    cfg,
	}
}

func (h *TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeStatusError(w, http.StatusMethodNotAllowed, "method %s not allowed, use POST", r.Method)
		return
	}

	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		clientIP = r.RemoteAddr
	}

	id, err := h.resolver.Resolve(clientIP)
	if err != nil {
		writeStatusError(w, http.StatusForbidden, "no identity mapped for IP %s", clientIP)
		return
	}

	tokenSpec := h.config.ResolvedTokenSpec(id)

	tokenReq := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			Audiences: tokenSpec.Audiences,
		},
	}
	if tokenSpec.ExpirationSeconds != nil {
		tokenReq.Spec.ExpirationSeconds = tokenSpec.ExpirationSeconds
	}

	result, err := h.clientset.CoreV1().ServiceAccounts(id.ServiceAccount.Namespace).CreateToken(
		context.TODO(),
		id.ServiceAccount.Name,
		tokenReq,
		metav1.CreateOptions{},
	)
	if err != nil {
		log.Printf("error minting token for %s (SA %s/%s): %v", clientIP, id.ServiceAccount.Namespace, id.ServiceAccount.Name, err)
		writeStatusError(w, http.StatusInternalServerError, "failed to mint token")
		return
	}

	log.Printf("minted token for %s (SA %s/%s)", clientIP, id.ServiceAccount.Namespace, id.ServiceAccount.Name)

	resp := &authv1.TokenRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TokenRequest",
			APIVersion: "authentication.k8s.io/v1",
		},
		Status: authv1.TokenRequestStatus{
			Token:               result.Status.Token,
			ExpirationTimestamp: result.Status.ExpirationTimestamp,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeStatusError(w http.ResponseWriter, code int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	status := &metav1.Status{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Status",
			APIVersion: "v1",
		},
		Status:  metav1.StatusFailure,
		Message: msg,
		Code:    int32(code),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(status)
}
