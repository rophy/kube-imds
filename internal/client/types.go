package client

import "time"

// TokenRequestResponse is the JSON shape returned by POST /api/v1/token.
type TokenRequestResponse struct {
	Kind       string             `json:"kind"`
	APIVersion string             `json:"apiVersion"`
	Status     TokenRequestStatus `json:"status"`
}

type TokenRequestStatus struct {
	Token               string    `json:"token"`
	ExpirationTimestamp time.Time `json:"expirationTimestamp"`
}

// StatusError is the JSON shape returned by the server on errors.
type StatusError struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Code       int    `json:"code"`
}
