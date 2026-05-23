// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

const authDomainsPath = "/api/v1/domains"

func (c *Client) ListAuthDomains(ctx context.Context) ([]map[string]interface{}, error) {
	return c.ListMap(ctx, authDomainsPath)
}

func (c *Client) GetAuthDomain(ctx context.Context, id string) (map[string]interface{}, error) {
	return c.GetMap(ctx, fmt.Sprintf("%s/%s", authDomainsPath, id))
}

func (c *Client) CreateAuthDomain(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	out, err := c.CreateMap(ctx, authDomainsPath, body)
	if err != nil {
		return nil, err
	}
	if id := MapString(out, "id"); id != "" {
		return c.GetAuthDomain(ctx, id)
	}
	return out, nil
}

func (c *Client) UpdateAuthDomain(ctx context.Context, id string, body map[string]interface{}) error {
	return c.UpdateMap(ctx, fmt.Sprintf("%s/%s", authDomainsPath, id), body)
}

func (c *Client) DeleteAuthDomain(ctx context.Context, id string) error {
	return c.DeleteMap(ctx, fmt.Sprintf("%s/%s", authDomainsPath, id))
}

func BuildAuthDomainBody(name, configJSON string) (map[string]interface{}, error) {
	body := map[string]interface{}{"name": name}
	if configJSON == "" {
		return body, nil
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("config must be valid JSON: %w", err)
	}
	body["config"] = cfg
	return body, nil
}
