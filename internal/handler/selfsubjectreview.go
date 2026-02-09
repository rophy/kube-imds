package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rophy/kube-imds/internal/config"
	"github.com/rophy/kube-imds/internal/identity"
)

type SelfSubjectReviewHandler struct {
	resolver *identity.Resolver
	config   *config.Config
}

func NewSelfSubjectReviewHandler(resolver *identity.Resolver, cfg *config.Config) *SelfSubjectReviewHandler {
	return &SelfSubjectReviewHandler{
		resolver: resolver,
		config:   cfg,
	}
}

func (h *SelfSubjectReviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeStatusError(w, http.StatusMethodNotAllowed, "method %s not allowed, use POST", r.Method)
		return
	}

	clientIP := resolveClientIP(r, h.config)

	id, err := h.resolver.Resolve(clientIP)
	if err != nil {
		writeStatusError(w, http.StatusForbidden, "no identity mapped for IP %s", clientIP)
		return
	}

	resp := &authv1.SelfSubjectReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SelfSubjectReview",
			APIVersion: "kube-imds/v1",
		},
		Status: authv1.SelfSubjectReviewStatus{
			UserInfo: authv1.UserInfo{
				Username: fmt.Sprintf("system:serviceaccount:%s:%s", id.ServiceAccount.Namespace, id.ServiceAccount.Name),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
