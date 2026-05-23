// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

const serviceAccountsPath = "/api/v1/service_accounts"

func (c *Client) ListServiceAccounts(ctx context.Context) ([]map[string]interface{}, error) {
	return c.ListMap(ctx, serviceAccountsPath)
}

func (c *Client) GetServiceAccount(ctx context.Context, id string) (map[string]interface{}, error) {
	return c.GetMap(ctx, fmt.Sprintf("%s/%s", serviceAccountsPath, id))
}

func (c *Client) CreateServiceAccount(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	out, err := c.CreateMap(ctx, serviceAccountsPath, body)
	if err != nil {
		return nil, err
	}
	if id := MapString(out, "id"); id != "" {
		return c.GetServiceAccount(ctx, id)
	}
	return out, nil
}

func (c *Client) UpdateServiceAccount(ctx context.Context, id string, body map[string]interface{}) error {
	return c.UpdateMap(ctx, fmt.Sprintf("%s/%s", serviceAccountsPath, id), body)
}

func (c *Client) DeleteServiceAccount(ctx context.Context, id string) error {
	return c.DeleteMap(ctx, fmt.Sprintf("%s/%s", serviceAccountsPath, id))
}
