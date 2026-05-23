// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func cloudAccountsPath(provider string) string {
	return fmt.Sprintf("/api/v1/cloud_integrations/%s/accounts", provider)
}

func (c *Client) ListCloudAccounts(ctx context.Context, provider string) ([]map[string]interface{}, error) {
	var wrapper map[string]interface{}
	if _, err := c.DoJSON(ctx, http.MethodGet, cloudAccountsPath(provider), nil, &wrapper); err != nil {
		return nil, err
	}
	raw, ok := wrapper["accounts"]
	if !ok || raw == nil {
		return nil, nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected cloud accounts list shape")
	}
	out := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if ok {
			out = append(out, m)
		}
	}
	return out, nil
}

func (c *Client) GetCloudAccount(ctx context.Context, provider, id string) (map[string]interface{}, error) {
	return c.GetMap(ctx, fmt.Sprintf("%s/%s", cloudAccountsPath(provider), id))
}

func (c *Client) CreateCloudAccount(ctx context.Context, provider string, body map[string]interface{}) (map[string]interface{}, error) {
	out, err := c.CreateMap(ctx, cloudAccountsPath(provider), body)
	if err != nil {
		return nil, err
	}
	if id := MapString(out, "id"); id != "" {
		return c.GetCloudAccount(ctx, provider, id)
	}
	return out, nil
}

func (c *Client) UpdateCloudAccount(ctx context.Context, provider, id string, body map[string]interface{}) error {
	return c.UpdateMap(ctx, fmt.Sprintf("%s/%s", cloudAccountsPath(provider), id), body)
}

func (c *Client) DeleteCloudAccount(ctx context.Context, provider, id string) error {
	return c.DeleteMap(ctx, fmt.Sprintf("%s/%s", cloudAccountsPath(provider), id))
}

func BuildCloudAccountUpdateBody(configJSON string) (map[string]interface{}, error) {
	if configJSON == "" {
		return map[string]interface{}{}, nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return map[string]interface{}{"config": cfg}, nil
}

func BuildCloudAccountBody(configJSON, credentialsJSON string) (map[string]interface{}, error) {
	body := map[string]interface{}{}
	if configJSON != "" {
		var cfg map[string]interface{}
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, fmt.Errorf("config: %w", err)
		}
		body["config"] = cfg
	}
	if credentialsJSON != "" {
		var cred map[string]interface{}
		if err := json.Unmarshal([]byte(credentialsJSON), &cred); err != nil {
			return nil, fmt.Errorf("credentials: %w", err)
		}
		body["credentials"] = cred
	}
	return body, nil
}
