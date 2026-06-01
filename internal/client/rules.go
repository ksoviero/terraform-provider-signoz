// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
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

// inListSpaceRe matches bracket-enclosed IN list values in filter expressions.
var inListSpaceRe = regexp.MustCompile(`IN\s*\[([^\]]+)\]`)

// normalizeFilterExpression normalizes filter expressions returned by the API:
//   - Replaces double-quoted IN list values with single-quoted ones
//   - Ensures consistent ", " spacing between list items
func normalizeFilterExpression(expr string) string {
	return inListSpaceRe.ReplaceAllStringFunc(expr, func(match string) string {
		// Extract inner content between [ and ]
		inner := match[strings.Index(match, "[")+1 : len(match)-1]
		parts := strings.Split(inner, ",")
		normalized := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			// Normalize double-quoted values to single-quoted
			if len(p) >= 2 && p[0] == '"' && p[len(p)-1] == '"' {
				p = "'" + p[1:len(p)-1] + "'"
			}
			normalized = append(normalized, p)
		}
		return "IN [" + strings.Join(normalized, ", ") + "]"
	})
}

// normalizeRuleSpec recursively walks the decoded spec and removes zero-value
// fields that the API always echoes back but users never set in config.
func normalizeRuleSpec(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			switch k {
			case "source":
				// Remove empty source (zero value, not meaningful)
				if s, ok := val.(string); ok && s == "" {
					delete(t, k)
				}
			case "stepInterval":
				// Remove zero stepInterval (default for log/trace queries)
				if isZeroNumber(val) {
					delete(t, k)
				}
			case "step":
				// Remove zero step (default for promql queries)
				if isZeroNumber(val) {
					delete(t, k)
				}
			case "reduceTo", "temporality":
				// Remove empty aggregation defaults injected by the UI/API.
				if s, ok := val.(string); ok && s == "" {
					delete(t, k)
				}
			case "having":
				// Remove having:{expression:""} (zero value)
				if m, ok := val.(map[string]interface{}); ok {
					if expr, ok := m["expression"].(string); ok && expr == "" && len(m) == 1 {
						delete(t, k)
					}
				}
			case "recoveryTarget":
				// Remove null recoveryTarget
				if val == nil {
					delete(t, k)
				}
			case "targetUnit":
				// Remove empty targetUnit
				if s, ok := val.(string); ok && s == "" {
					delete(t, k)
				}
			case "expression":
				// Normalize filter expressions
				if s, ok := val.(string); ok {
					t[k] = normalizeFilterExpression(s)
				}
			default:
				t[k] = normalizeRuleSpec(val)
			}
		}
	case []interface{}:
		for i, item := range t {
			t[i] = normalizeRuleSpec(item)
		}
	}
	return v
}

func isZeroNumber(v interface{}) bool {
	switch n := v.(type) {
	case float64:
		return n == 0
	case int:
		return n == 0
	case int64:
		return n == 0
	}
	return false
}

// NormalizeRuleSpecJSON normalizes spec JSON using the same drift-pruning rules
// applied to API read payloads before storing Terraform state.
func NormalizeRuleSpecJSON(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		return "", fmt.Errorf("spec must be valid JSON: %w", err)
	}
	stripTopLevelDuplicates(spec)
	normalizeRuleSpec(spec)
	b, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	return CanonicalJSON(string(b))
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

	// The API echoes evalWindow/frequency at the top level as duplicates of
	// evaluation.spec values. Strip them only when they match, so users who
	// intentionally set top-level values still see a diff.
	stripTopLevelDuplicates(copy)

	normalizeRuleSpec(copy)
	b, err := json.Marshal(copy)
	if err != nil {
		return "", err
	}
	return CanonicalJSON(string(b))
}

// stripTopLevelDuplicates removes top-level evalWindow/frequency. The API always
// echoes these as legacy fields regardless of the canonical values in
// evaluation.spec, so they are stripped unconditionally.
func stripTopLevelDuplicates(spec map[string]interface{}) {
	for _, k := range []string{"evalWindow", "frequency"} {
		delete(spec, k)
	}
}
