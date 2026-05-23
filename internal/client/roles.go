// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

const rolesPath = "/api/v1/roles"

func (c *Client) ListRoles(ctx context.Context) ([]map[string]interface{}, error) {
	return c.ListMap(ctx, rolesPath)
}

func (c *Client) GetRole(ctx context.Context, id string) (map[string]interface{}, error) {
	return c.GetMap(ctx, fmt.Sprintf("%s/%s", rolesPath, id))
}

func (c *Client) CreateRole(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	out, err := c.CreateMap(ctx, rolesPath, body)
	if err != nil {
		return nil, err
	}
	if id := MapString(out, "id"); id != "" {
		return c.GetRole(ctx, id)
	}
	return out, nil
}

func (c *Client) UpdateRole(ctx context.Context, id string, description string) error {
	return c.PatchMap(ctx, fmt.Sprintf("%s/%s", rolesPath, id), map[string]interface{}{
		"description": description,
	})
}

func (c *Client) DeleteRole(ctx context.Context, id string) error {
	return c.DeleteMap(ctx, fmt.Sprintf("%s/%s", rolesPath, id))
}
