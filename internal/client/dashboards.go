// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Dashboard is the GET model for /api/v1/dashboards/{id}.
type Dashboard struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	Locked    bool                   `json:"locked"`
	Source    string                 `json:"source"`
	CreatedAt string                 `json:"createdAt"`
	UpdatedAt string                 `json:"updatedAt"`
	CreatedBy string                 `json:"createdBy"`
	UpdatedBy string                 `json:"updatedBy"`
}

func (c *Client) ListDashboards(ctx context.Context) ([]Dashboard, error) {
	var out []Dashboard
	_, err := c.DoJSON(ctx, http.MethodGet, "/api/v1/dashboards", nil, &out)
	return out, err
}

func (c *Client) GetDashboard(ctx context.Context, id string) (*Dashboard, error) {
	var out Dashboard
	_, err := c.DoJSON(ctx, http.MethodGet, "/api/v1/dashboards/"+id, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateDashboard(ctx context.Context, data map[string]interface{}) (*Dashboard, error) {
	var out Dashboard
	_, err := c.DoJSON(ctx, http.MethodPost, "/api/v1/dashboards", data, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateDashboard(ctx context.Context, id string, data map[string]interface{}) (*Dashboard, error) {
	var out Dashboard
	_, err := c.DoJSON(ctx, http.MethodPut, "/api/v1/dashboards/"+id, data, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDashboard(ctx context.Context, id string) error {
	_, err := c.DoJSON(ctx, http.MethodDelete, "/api/v1/dashboards/"+id, nil, nil)
	return err
}

// ParseDashboardData unmarshals the Terraform data attribute into a map for the API.
func ParseDashboardData(dataJSON string) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &m); err != nil {
		return nil, fmt.Errorf("data must be valid JSON: %w", err)
	}
	return m, nil
}

// MarshalDashboardData encodes dashboard data for Terraform state.
func MarshalDashboardData(data map[string]interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return CanonicalJSON(string(b))
}
