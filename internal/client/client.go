// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const apiKeyHeader = "SigNoz-Api-Key"

// Client calls the SigNoz HTTP API (JSON envelope: status + data).
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	apiKey     string
}

// New parses endpoint (e.g. https://signoz.example.com) and returns a client.
func New(endpoint, apiKey string, insecureSkipVerify bool) (*Client, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("endpoint scheme must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("endpoint must include host")
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	} else {
		tr.TLSClientConfig.MinVersion = tls.VersionTLS12
	}
	tr.TLSClientConfig.InsecureSkipVerify = insecureSkipVerify

	return &Client{
		baseURL: u,
		httpClient: &http.Client{
			Transport: tr,
		},
		apiKey: strings.TrimSpace(apiKey),
	}, nil
}

func (c *Client) apiURL(path string) string {
	path = strings.TrimLeft(path, "/")
	return c.baseURL.String() + "/" + path
}

type envelope struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

type errorEnvelope struct {
	Status string `json:"status"`
	Error  struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// DoJSON performs an API request and decodes the success data payload into dest.
// dest may be nil for empty or 204 responses.
func (c *Client) DoJSON(ctx context.Context, method, path string, reqBody any, dest any) (int, error) {
	var bodyReader io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return 0, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), bodyReader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set(apiKeyHeader, c.apiKey)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode == http.StatusNoContent || len(bytes.TrimSpace(raw)) == 0 {
		if resp.StatusCode >= 400 {
			return resp.StatusCode, apiError(resp.StatusCode, raw)
		}
		return resp.StatusCode, nil
	}

	if resp.StatusCode >= 400 {
		return resp.StatusCode, apiError(resp.StatusCode, raw)
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return resp.StatusCode, fmt.Errorf("decode envelope: %w", err)
	}
	if strings.EqualFold(env.Status, "error") {
		var errEnv errorEnvelope
		if json.Unmarshal(raw, &errEnv) == nil && errEnv.Error.Message != "" {
			if errEnv.Error.Code != "" {
				return resp.StatusCode, fmt.Errorf("%s: %s", errEnv.Error.Code, errEnv.Error.Message)
			}
			return resp.StatusCode, fmt.Errorf("%s", errEnv.Error.Message)
		}
		return resp.StatusCode, fmt.Errorf("signoz api error status")
	}
	if dest != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, dest); err != nil {
			return resp.StatusCode, fmt.Errorf("decode data: %w", err)
		}
	}
	return resp.StatusCode, nil
}

// DoJSONRaw sends a raw JSON body (for rule create/update) and decodes data into dest.
func (c *Client) DoJSONRaw(ctx context.Context, method, path string, rawBody []byte, dest any) (int, error) {
	var bodyReader io.Reader
	if len(rawBody) > 0 {
		bodyReader = bytes.NewReader(rawBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), bodyReader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set(apiKeyHeader, c.apiKey)
	if len(rawBody) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode >= 400 {
		return resp.StatusCode, apiError(resp.StatusCode, raw)
	}

	if resp.StatusCode == http.StatusNoContent || len(bytes.TrimSpace(raw)) == 0 {
		return resp.StatusCode, nil
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return resp.StatusCode, fmt.Errorf("decode envelope: %w", err)
	}
	if dest != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, dest); err != nil {
			return resp.StatusCode, fmt.Errorf("decode data: %w", err)
		}
	}
	return resp.StatusCode, nil
}

func apiError(status int, body []byte) error {
	var errEnv errorEnvelope
	if err := json.Unmarshal(body, &errEnv); err == nil && errEnv.Error.Message != "" {
		if errEnv.Error.Code != "" {
			return fmt.Errorf("signoz api %d: %s: %s", status, errEnv.Error.Code, errEnv.Error.Message)
		}
		return fmt.Errorf("signoz api %d: %s", status, errEnv.Error.Message)
	}
	return fmt.Errorf("signoz api %d: %s", status, strings.TrimSpace(string(body)))
}
