// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const savedViewsPath = "/api/v1/explorer/views"

func (c *Client) ListSavedViews(ctx context.Context, sourcePage, name, category string) ([]map[string]interface{}, error) {
	path := savedViewsPath
	q := url.Values{}
	if sourcePage != "" {
		q.Set("sourcePage", sourcePage)
	}
	if name != "" {
		q.Set("name", name)
	}
	if category != "" {
		q.Set("category", category)
	}
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	return c.ListMap(ctx, path)
}

func (c *Client) GetSavedView(ctx context.Context, id string) (map[string]interface{}, error) {
	return c.GetMap(ctx, fmt.Sprintf("%s/%s", savedViewsPath, id))
}

func (c *Client) CreateSavedView(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	var id string
	if _, err := c.DoJSON(ctx, http.MethodPost, savedViewsPath, body, &id); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, fmt.Errorf("create saved view: empty id in response")
	}
	return c.GetSavedView(ctx, id)
}

func (c *Client) UpdateSavedView(ctx context.Context, id string, body map[string]interface{}) error {
	_, err := c.DoJSON(ctx, http.MethodPut, fmt.Sprintf("%s/%s", savedViewsPath, id), body, nil)
	return err
}

func (c *Client) DeleteSavedView(ctx context.Context, id string) error {
	_, err := c.DoJSON(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", savedViewsPath, id), nil, nil)
	return err
}

func ParseSavedViewData(dataJSON string) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &m); err != nil {
		return nil, fmt.Errorf("data must be valid JSON: %w", err)
	}
	return m, nil
}

func MarshalSavedViewData(m map[string]interface{}) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return CanonicalJSON(string(b))
}

// BuildSavedViewBody builds POST/PUT body for explorer saved views.
func BuildSavedViewBody(name, sourcePage, category, extraData, compositeQueryJSON string, tags []string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"name":       name,
		"sourcePage": sourcePage,
	}
	if category != "" {
		body["category"] = category
	}
	if extraData != "" {
		body["extraData"] = extraData
	}
	if len(tags) > 0 {
		body["tags"] = tags
	}
	if compositeQueryJSON == "" {
		return nil, fmt.Errorf("composite_query is required")
	}
	var cq map[string]interface{}
	if err := json.Unmarshal([]byte(compositeQueryJSON), &cq); err != nil {
		return nil, fmt.Errorf("composite_query must be valid JSON: %w", err)
	}
	body["compositeQuery"] = cq
	return body, nil
}
