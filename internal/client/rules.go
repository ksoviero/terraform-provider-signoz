// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Rule is the v2 API read model for an alert rule.
type Rule struct {
	ID          string            `json:"id"`
	State       string            `json:"state"`
	Alert       string            `json:"alert"`
	AlertType   string            `json:"alertType"`
	Description string            `json:"description"`
	RuleType    string            `json:"ruleType"`
	Disabled    bool              `json:"disabled"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
	CreatedBy   string            `json:"createdBy"`
	UpdatedBy   string            `json:"updatedBy"`
}

func (c *Client) ListRules(ctx context.Context) ([]Rule, error) {
	var out []Rule
	_, err := c.DoJSON(ctx, http.MethodGet, "/api/v2/rules", nil, &out)
	return out, err
}

func (c *Client) GetRule(ctx context.Context, id string) (*Rule, error) {
	m, err := c.GetRuleMap(ctx, id)
	if err != nil {
		return nil, err
	}
	return RuleFromMap(m), nil
}

// GetRuleMap returns the full rule JSON object from the API (for spec round-tripping).
func (c *Client) GetRuleMap(ctx context.Context, id string) (map[string]interface{}, error) {
	var out map[string]interface{}
	_, err := c.DoJSON(ctx, http.MethodGet, "/api/v2/rules/"+id, nil, &out)
	return out, err
}

// RuleFromMap builds a Rule from a generic API map.
func RuleFromMap(m map[string]interface{}) *Rule {
	r := &Rule{}
	if v, ok := m["id"].(string); ok {
		r.ID = v
	}
	if v, ok := m["state"].(string); ok {
		r.State = v
	}
	if v, ok := m["alert"].(string); ok {
		r.Alert = v
	}
	if v, ok := m["alertType"].(string); ok {
		r.AlertType = v
	}
	if v, ok := m["description"].(string); ok {
		r.Description = v
	}
	if v, ok := m["ruleType"].(string); ok {
		r.RuleType = v
	}
	if v, ok := m["disabled"].(bool); ok {
		r.Disabled = v
	}
	if v, ok := m["labels"].(map[string]interface{}); ok {
		r.Labels = stringMapFromAPI(v)
	}
	if v, ok := m["annotations"].(map[string]interface{}); ok {
		r.Annotations = stringMapFromAPI(v)
	}
	if v, ok := m["createdAt"].(string); ok {
		r.CreatedAt = v
	}
	if v, ok := m["updatedAt"].(string); ok {
		r.UpdatedAt = v
	}
	if v, ok := m["createdBy"].(string); ok {
		r.CreatedBy = v
	}
	if v, ok := m["updatedBy"].(string); ok {
		r.UpdatedBy = v
	}
	return r
}

func stringMapFromAPI(m map[string]interface{}) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func (c *Client) CreateRule(ctx context.Context, body []byte) (*Rule, error) {
	var out Rule
	_, err := c.DoJSONRaw(ctx, http.MethodPost, "/api/v2/rules", body, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateRule(ctx context.Context, id string, body []byte) error {
	_, err := c.DoJSONRaw(ctx, http.MethodPut, "/api/v2/rules/"+id, body, nil)
	return err
}

func (c *Client) DeleteRule(ctx context.Context, id string) error {
	_, err := c.DoJSON(ctx, http.MethodDelete, "/api/v2/rules/"+id, nil, nil)
	return err
}

// BuildRuleBody merges top-level Terraform fields with spec JSON for POST/PUT.
func BuildRuleBody(alert, alertType, ruleType, description string, disabled bool, labels, annotations map[string]string, specJSON string) ([]byte, error) {
	body := map[string]interface{}{
		"alert":     alert,
		"alertType": alertType,
		"ruleType":  ruleType,
		"disabled":  disabled,
	}
	if description != "" {
		body["description"] = description
	}
	if len(labels) > 0 {
		body["labels"] = labels
	}
	if len(annotations) > 0 {
		body["annotations"] = annotations
	}
	if specJSON != "" {
		var spec map[string]interface{}
		if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
			return nil, fmt.Errorf("spec must be valid JSON: %w", err)
		}
		for k, v := range spec {
			body[k] = v
		}
	}
	return json.Marshal(body)
}

// RuleSpecFromMap returns JSON for attributes stored in spec (everything except top-level fields).
func RuleSpecFromMap(full map[string]interface{}) (string, error) {
	copy := make(map[string]interface{}, len(full))
	for k, v := range full {
		copy[k] = v
	}
	for _, k := range []string{
		"id", "state", "alert", "alertType", "description", "ruleType", "disabled",
		"labels", "annotations", "createdAt", "updatedAt", "createdBy", "updatedBy",
	} {
		delete(copy, k)
	}
	b, err := json.Marshal(copy)
	if err != nil {
		return "", err
	}
	return CanonicalJSON(string(b))
}
