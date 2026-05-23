// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

const routePoliciesPath = "/api/v1/route_policies"

func (c *Client) ListRoutePolicies(ctx context.Context) ([]map[string]interface{}, error) {
	return c.ListMap(ctx, routePoliciesPath)
}

func (c *Client) GetRoutePolicy(ctx context.Context, id string) (map[string]interface{}, error) {
	return c.GetMap(ctx, fmt.Sprintf("%s/%s", routePoliciesPath, id))
}

func (c *Client) CreateRoutePolicy(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	out, err := c.CreateMap(ctx, routePoliciesPath, body)
	if err != nil {
		return nil, err
	}
	if MapString(out, "id") == "" {
		return out, nil
	}
	return out, nil
}

func (c *Client) UpdateRoutePolicy(ctx context.Context, id string, body map[string]interface{}) error {
	return c.UpdateMap(ctx, fmt.Sprintf("%s/%s", routePoliciesPath, id), body)
}

func (c *Client) DeleteRoutePolicy(ctx context.Context, id string) error {
	return c.DeleteMap(ctx, fmt.Sprintf("%s/%s", routePoliciesPath, id))
}

func BuildRoutePolicyBody(name, description, expression, kind string, channels, tags []string) map[string]interface{} {
	body := map[string]interface{}{
		"name":       name,
		"expression": expression,
		"channels":   channels,
	}
	if description != "" {
		body["description"] = description
	}
	if kind != "" {
		body["kind"] = kind
	}
	if len(tags) > 0 {
		body["tags"] = tags
	}
	return body
}
