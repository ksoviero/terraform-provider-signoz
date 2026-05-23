// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ListMap decodes a JSON array from a list endpoint.
func (c *Client) ListMap(ctx context.Context, path string) ([]map[string]interface{}, error) {
	var out []map[string]interface{}
	_, err := c.DoJSON(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

// GetMap decodes a single object from GET by id.
func (c *Client) GetMap(ctx context.Context, path string) (map[string]interface{}, error) {
	var out map[string]interface{}
	_, err := c.DoJSON(ctx, http.MethodGet, path, nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CreateMap posts body and decodes the created object from the API data envelope.
func (c *Client) CreateMap(ctx context.Context, path string, body map[string]interface{}) (map[string]interface{}, error) {
	var out map[string]interface{}
	_, err := c.DoJSON(ctx, http.MethodPost, path, body, &out)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return map[string]interface{}{}, nil
	}
	return out, nil
}

// CreateID posts body and returns the created resource id.
func (c *Client) CreateID(ctx context.Context, path string, body map[string]interface{}) (string, error) {
	out, err := c.CreateMap(ctx, path, body)
	if err != nil {
		return "", err
	}
	return MapString(out, "id"), nil
}

// CreateMapRaw posts and decodes data into dest (string id, map, etc.).
func (c *Client) CreateMapRaw(ctx context.Context, path string, body map[string]interface{}, dest any) error {
	_, err := c.DoJSON(ctx, http.MethodPost, path, body, dest)
	return err
}

// UpdateMap puts body to path (204/no body).
func (c *Client) UpdateMap(ctx context.Context, path string, body map[string]interface{}) error {
	_, err := c.DoJSON(ctx, http.MethodPut, path, body, nil)
	return err
}

// PatchMap patches body to path (204/no body).
func (c *Client) PatchMap(ctx context.Context, path string, body map[string]interface{}) error {
	_, err := c.DoJSON(ctx, http.MethodPatch, path, body, nil)
	return err
}

// DeleteMap deletes by path.
func (c *Client) DeleteMap(ctx context.Context, path string) error {
	_, err := c.DoJSON(ctx, http.MethodDelete, path, nil, nil)
	return err
}

// MapString returns a string field from an API map (camelCase keys).
func MapString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

// MapBool returns a bool field from an API map.
func MapBool(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}

// MapStringSlice returns a string slice from JSON array values.
func MapStringSlice(m map[string]interface{}, key string) []string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case []string:
		return t
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// SubMapJSON extracts a nested object as canonical JSON.
func SubMapJSON(m map[string]interface{}, key string) (string, error) {
	if m == nil {
		return "{}", nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return "{}", nil
	}
	sub, ok := v.(map[string]interface{})
	if !ok {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return CanonicalJSON(string(b))
	}
	b, err := json.Marshal(sub)
	if err != nil {
		return "", err
	}
	return CanonicalJSON(string(b))
}

// MergeJSONIntoMap parses jsonStr and merges into dst.
func MergeJSONIntoMap(dst map[string]interface{}, jsonStr string) error {
	if strings.TrimSpace(jsonStr) == "" {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	for k, v := range m {
		dst[k] = v
	}
	return nil
}

// StripKeys removes keys from a map (for spec/body round-trip).
func StripKeys(m map[string]interface{}, keys ...string) {
	for _, k := range keys {
		delete(m, k)
	}
}

// CloneMap shallow-copies a map.
func CloneMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
