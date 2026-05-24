// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Channel is a SigNoz notification channel (GET model).
type Channel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Data      string `json:"data"`
	OrgID     string `json:"orgId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func (c *Client) ListChannels(ctx context.Context) ([]Channel, error) {
	var out []Channel
	_, err := c.DoJSON(ctx, http.MethodGet, "/api/v1/channels", nil, &out)
	return out, err
}

func (c *Client) GetChannel(ctx context.Context, id string) (*Channel, error) {
	var out Channel
	_, err := c.DoJSON(ctx, http.MethodGet, "/api/v1/channels/"+id, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateChannel posts a receiver JSON body (PostableChannel: name + *_configs).
func (c *Client) CreateChannel(ctx context.Context, body map[string]interface{}) (*Channel, error) {
	var out Channel
	_, err := c.DoJSON(ctx, http.MethodPost, "/api/v1/channels", body, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateChannel puts a receiver JSON body (ConfigReceiver) for the channel id.
func (c *Client) UpdateChannel(ctx context.Context, id string, body map[string]interface{}) error {
	_, err := c.DoJSON(ctx, http.MethodPut, "/api/v1/channels/"+id, body, nil)
	return err
}

func (c *Client) DeleteChannel(ctx context.Context, id string) error {
	_, err := c.DoJSON(ctx, http.MethodDelete, "/api/v1/channels/"+id, nil, nil)
	return err
}

// NormalizeChannelConfigJSON returns canonical JSON for the Terraform config attribute:
// receiver configuration with top-level "name" removed (name is set from the resource name).
func NormalizeChannelConfigJSON(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return "", fmt.Errorf("channel config must be valid JSON: %w", err)
	}
	delete(m, "name")
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return CanonicalJSON(string(b))
}

// MergeChannelBody combines name with config JSON into a POST/PUT receiver map.
func MergeChannelBody(name string, configJSON string) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &m); err != nil {
		return nil, fmt.Errorf("config must be valid JSON: %w", err)
	}
	m["name"] = name
	return m, nil
}
