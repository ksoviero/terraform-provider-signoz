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

// alertmanagerHTTPConfigDefaultBools are boolean fields inside http_config that
// Alertmanager always echoes at their default values. Pruning them prevents
// phantom drift when users omit http_config entirely.
var alertmanagerHTTPConfigDefaultBools = map[string]bool{
	"follow_redirects":     true,
	"enable_http2":         true,
	"insecure_skip_verify": false,
}

// alertmanagerDefaultFieldValues maps field names to the set of string values
// that Alertmanager injects as defaults. A field is pruned only when its value
// exactly matches one of these known defaults, so user-set values are preserved.
//
// Sources: Prometheus Alertmanager DefaultXxxConfig structs and global config:
//   - OpsGenie api_url: GlobalConfig.OpsGenieAPIURL
//   - PagerDuty url: hardcoded in pagerduty notifier
//   - Slack app_url: always injected (internal Slack app URL)
//   - Telegram parse_mode: DefaultTelegramConfig
//   - Email html: DefaultEmailConfig
//   - Template strings: DefaultXxxConfig per receiver
var alertmanagerDefaultFieldValues = map[string][]string{
	// OpsGenie: global default API URL
	"api_url": {"https://api.opsgenie.com/"},

	// Slack: internal app URL, always injected
	"app_url": {"https://slack.com/api/chat.postMessage"},

	// Telegram
	"parse_mode": {"HTML"},

	// Slack template defaults
	"callback_id": {`{{ template "slack.default.callbackid" . }}`},
	"color":       {`{{ if eq .Status "firing" }}danger{{ else }}good{{ end }}`},
	"fallback":    {`{{ template "slack.default.fallback" . }}`},
	"footer":      {`{{ template "slack.default.footer" . }}`},
	"icon_emoji":  {`{{ template "slack.default.iconemoji" . }}`},
	"icon_url":    {`{{ template "slack.default.iconurl" . }}`},
	"pretext":     {`{{ template "slack.default.pretext" . }}`},
	"title":       {`{{ template "slack.default.title" . }}`},
	"title_link":  {`{{ template "slack.default.titlelink" . }}`},
	"username":    {`{{ template "slack.default.username" . }}`},

	// PagerDuty template defaults and default API URL
	"client":     {`{{ template "pagerduty.default.client" . }}`},
	"client_url": {`{{ template "pagerduty.default.clientURL" . }}`},
	// PagerDuty url default (hardcoded in notifier)
	"url": {"https://events.pagerduty.com/v2/enqueue"},

	// description and source are shared between PagerDuty and OpsGenie with
	// different default values — list both so either is pruned.
	"description": {
		`{{ template "pagerduty.default.description" .}}`,
		`{{ template "opsgenie.default.description" . }}`,
	},
	"source": {
		`{{ template "pagerduty.default.client" . }}`,
		`{{ template "opsgenie.default.source" . }}`,
	},

	// Email template default
	"html": {`{{ template "email.default.html" . }}`},
}

// alertmanagerDefaultNumericZeroFields are fields where a value of 0
// (integer or float) is always an Alertmanager default and never
// meaningful when explicitly set to 0 by the user.
var alertmanagerDefaultNumericZeroFields = map[string]bool{
	"timeout": true,
}

// alertmanagerDefaultPagerDutyDetails is the map of details fields PagerDuty
// injects by default. If the user did not set details, the API echoes these;
// we prune them to prevent drift.
var alertmanagerDefaultPagerDutyDetails = map[string]string{
	"firing":       `{{ .Alerts.Firing | toJson }}`,
	"num_firing":   `{{ .Alerts.Firing | len }}`,
	"num_resolved": `{{ .Alerts.Resolved | len }}`,
	"resolved":     `{{ .Alerts.Resolved | toJson }}`,
}

// pruneChannelConfigValue removes SigNoz/Alertmanager API defaults that should
// not force Terraform drift. Pruned values: null, empty strings, empty objects,
// empty arrays, known Alertmanager default booleans/strings/numerics.
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

func pruneChannelConfigMap(m map[string]interface{}) {
	for k, v := range m {
		// Prune known Alertmanager default string values (only when the value
		// exactly matches one of the injected defaults — preserves user-set values).
		if defaults, ok := alertmanagerDefaultFieldValues[k]; ok {
			if s, ok := v.(string); ok {
				for _, defaultVal := range defaults {
					if s == defaultVal {
						delete(m, k)
						break
					}
				}
				if _, stillThere := m[k]; !stillThere {
					continue
				}
			}
		}

		// Prune known http_config boolean defaults.
		if defaultVal, ok := alertmanagerHTTPConfigDefaultBools[k]; ok {
			if b, ok := v.(bool); ok && b == defaultVal {
				delete(m, k)
				continue
			}
		}

		// Prune numeric zero for fields that are never meaningfully zero.
		if alertmanagerDefaultNumericZeroFields[k] {
			switch n := v.(type) {
			case float64:
				if n == 0 {
					delete(m, k)
					continue
				}
			case int:
				if n == 0 {
					delete(m, k)
					continue
				}
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
			// Special case: prune PagerDuty default details map.
			if k == "details" && isPagerDutyDefaultDetails(val) {
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

// isPagerDutyDefaultDetails returns true when the details map contains only
// the four default PagerDuty template strings and nothing else.
func isPagerDutyDefaultDetails(m map[string]interface{}) bool {
	if len(m) != len(alertmanagerDefaultPagerDutyDetails) {
		return false
	}
	for k, want := range alertmanagerDefaultPagerDutyDetails {
		got, ok := m[k].(string)
		if !ok || got != want {
			return false
		}
	}
	return true
}
// NormalizeChannelConfigJSON returns canonical JSON for the Terraform config attribute:
// receiver configuration with top-level "name" removed and Alertmanager-injected
// defaults pruned to prevent phantom drift.
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
