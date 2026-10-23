package custom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// exchangeResult is the key exchange's answer: the same shape as the MCP key
// exchange.
type exchangeResult struct {
	APIKey           string `json:"apiKey"`
	OrganizationID   string `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
}

// errExchangeNotAvailable means the API has no CLI key exchange yet.
var errExchangeNotAvailable = errors.New("this Moderation API account cannot sign in from the CLI yet")

// exchangeForOrganizationKey trades a Clerk OAuth access token for an
// organization key of the organization the user picked on the consent screen.
// The API makes one key per user and device name.
func exchangeForOrganizationKey(ctx context.Context, client *http.Client, keyExchangeURL, accessToken, deviceName string) (*exchangeResult, error) {
	if err := requireHTTPS(keyExchangeURL); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(map[string]string{"deviceName": deviceName})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, keyExchangeURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach the Moderation API: %w", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))

	var result struct {
		exchangeResult
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &result)

	switch {
	case res.StatusCode == http.StatusOK && result.APIKey != "":
		return &result.exchangeResult, nil
	case res.StatusCode == http.StatusNotFound:
		return nil, errExchangeNotAvailable
	case result.Message != "":
		return nil, errors.New(result.Message)
	default:
		return nil, fmt.Errorf("the Moderation API returned %s for the CLI sign-in", res.Status)
	}
}
