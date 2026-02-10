package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// FetchToken calls POST /api/v1/token on the kube-imds server.
func FetchToken(httpClient *http.Client, endpoint string) (*TokenRequestResponse, error) {
	url := endpoint + "/api/v1/token"

	resp, err := httpClient.Post(url, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var statusErr StatusError
		if err := json.NewDecoder(resp.Body).Decode(&statusErr); err != nil {
			return nil, fmt.Errorf("POST %s: status %d", url, resp.StatusCode)
		}
		return nil, fmt.Errorf("POST %s: status %d: %s", url, resp.StatusCode, statusErr.Message)
	}

	var tokenResp TokenRequestResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	if tokenResp.Status.Token == "" {
		return nil, fmt.Errorf("server returned empty token")
	}

	return &tokenResp, nil
}
