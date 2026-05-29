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

// pruneChannelConfigValue removes SigNoz API defaults that should not force Terraform drift:
// null, empty strings, empty objects, and empty arrays (nested maps are pruned recursively).
func pruneChannelConfigValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		pruneChannelConfigMap(x)
		return x
	case []interface{}:
		out := make([]interface{}, 0, len(x))
		for _, elem := range x {
			pruned := pruneChannelConfigValue(elem)
			if pruned == nil {
				continue
			}
			out = append(out, pruned)
		}
		return out
	default:
		return v
	}
}

// alertmanagerDefaultBools are http_config boolean fields that Alertmanager always
// echoes back at their default values. Pruning them prevents phantom drift when
// users omit http_config entirely.
var alertmanagerDefaultBools = map[string]bool{
	"follow_redirects": true,
	"enable_http2":     true,
	"insecure_skip_verify": false,
}

// alertmanagerDefaultStrings are fields the API echoes at known default values
// that users would never explicitly set.
var alertmanagerDefaultStrings = map[string]string{
	"api_url": "https://api.opsgenie.com/",
}

func pruneChannelConfigMap(m map[string]interface{}) {
	for k, v := range m {
		// Prune known Alertmanager default string values
		if defaultVal, isDefaultStr := alertmanagerDefaultStrings[k]; isDefaultStr {
			if s, ok := v.(string); ok && s == defaultVal {
				delete(m, k)
				continue
			}
		}
		// Prune known Alertmanager default boolean values
		if defaultVal, isDefaultBool := alertmanagerDefaultBools[k]; isDefaultBool {
			if b, ok := v.(bool); ok && b == defaultVal {
				delete(m, k)
				continue
			}
		}
		switch val := v.(type) {
		case nil:
			delete(m, k)
		case string:
			if val == "" {
				delete(m, k)
			}
		case map[string]interface{}:
			if len(val) == 0 {
				delete(m, k)
				continue
			}
			pruneChannelConfigMap(val)
			if len(val) == 0 {
				delete(m, k)
			}
		case []interface{}:
			if len(val) == 0 {
				delete(m, k)
				continue
			}
			pruned := pruneChannelConfigValue(val)
			if arr, ok := pruned.([]interface{}); ok {
				if len(arr) == 0 {
					delete(m, k)
				} else {
					m[k] = arr
				}
			}
		default:
			m[k] = pruneChannelConfigValue(val)
		}
	}
}

// NormalizeChannelConfigJSON returns canonical JSON for the Terraform config attribute:
// receiver configuration with top-level "name" removed and empty API defaults pruned.
func NormalizeChannelConfigJSON(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return "", fmt.Errorf("channel config must be valid JSON: %w", err)
	}
	delete(m, "name")
	pruneChannelConfigMap(m)
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
