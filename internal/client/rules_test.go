// SPDX-License-Identifier: MPL-2.0

package client_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

func TestNormalizeRuleSpecJSON_promqlStepZero(t *testing.T) {
	t.Parallel()

	config := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "promql",
					"spec": {
						"name": "A",
						"query": "count(longhorn_volume_capacity_bytes) - count(longhorn_volume_robustness==1)",
						"disabled": false,
						"stats": false,
						"legend": "Unhealthy Volumes"
					}
				}],
				"panelType": "graph",
				"queryType": "promql"
			},
			"alertOnAbsent": true,
			"absentFor": 15,
			"selectedQueryName": "A",
			"requireMinPoints": true,
			"requiredNumPoints": 3,
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 0, "matchType": "2", "op": "1", "channels": ["Email"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {
			"renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing", "nodata"]}
		}
	}`

	apiResponse := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "promql",
					"spec": {
						"name": "A",
						"query": "count(longhorn_volume_capacity_bytes) - count(longhorn_volume_robustness==1)",
						"disabled": false,
						"stats": false,
						"legend": "Unhealthy Volumes",
						"step": 0
					}
				}],
				"panelType": "graph",
				"queryType": "promql"
			},
			"alertOnAbsent": true,
			"absentFor": 15,
			"selectedQueryName": "A",
			"requireMinPoints": true,
			"requiredNumPoints": 3,
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 0, "matchType": "2", "op": "1", "channels": ["Email"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {
			"renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing", "nodata"]}
		}
	}`

	configNorm, err := client.NormalizeRuleSpecJSON(config)
	if err != nil {
		t.Fatal(err)
	}
	apiNorm, err := client.NormalizeRuleSpecJSON(apiResponse)
	if err != nil {
		t.Fatal(err)
	}
	if configNorm != apiNorm {
		t.Fatalf("normalized config and API response differ:\nconfig: %s\napi:    %s", configNorm, apiNorm)
	}
}

func TestNormalizeRuleSpecJSON_builderEmptyAggregationFields(t *testing.T) {
	t.Parallel()

	config := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"aggregations": [{
							"metricName": "k8s.pod.phase",
							"timeAggregation": "latest",
							"spaceAggregation": "avg"
						}],
						"groupBy": [{
							"name": "k8s.pod.name",
							"fieldContext": "attribute",
							"fieldDataType": "string"
						}],
						"having": {"expression": "avg(k8s.pod.phase) IN (1, 4, 5)"}
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"requireMinPoints": true,
			"requiredNumPoints": 3,
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 0, "matchType": "2", "op": "1", "channels": ["Email"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {
			"groupBy": ["k8s.pod.name"],
			"renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing"]}
		}
	}`

	apiResponse := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"aggregations": [{
							"metricName": "k8s.pod.phase",
							"reduceTo": "",
							"spaceAggregation": "avg",
							"temporality": "",
							"timeAggregation": "latest"
						}],
						"groupBy": [{
							"name": "k8s.pod.name",
							"fieldContext": "attribute",
							"fieldDataType": "string"
						}],
						"having": {"expression": "avg(k8s.pod.phase) IN (1, 4, 5)"}
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"requireMinPoints": true,
			"requiredNumPoints": 3,
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 0, "matchType": "2", "op": "1", "channels": ["Email"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {
			"groupBy": ["k8s.pod.name"],
			"renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing"]}
		}
	}`

	configNorm, err := client.NormalizeRuleSpecJSON(config)
	if err != nil {
		t.Fatal(err)
	}
	apiNorm, err := client.NormalizeRuleSpecJSON(apiResponse)
	if err != nil {
		t.Fatal(err)
	}
	if configNorm != apiNorm {
		t.Fatalf("normalized config and API response differ:\nconfig: %s\napi:    %s", configNorm, apiNorm)
	}

	ctx := context.Background()
	a := jsontypes.NewNormalizedValue(config)
	b := jsontypes.NewNormalizedValue(apiNorm)
	equal, diags := a.StringSemanticEquals(ctx, b)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !equal {
		t.Fatal("user config should semantically equal normalized API spec")
	}
}

func TestRuleSpecFromMap(t *testing.T) {
	t.Parallel()

	full := map[string]interface{}{
		"id":            "abc",
		"alert":         "Test",
		"alertType":     "METRIC_BASED_ALERT",
		"ruleType":      "threshold_rule",
		"schemaVersion": "v2alpha1",
		"version":       "v5",
		"condition": map[string]interface{}{
			"compositeQuery": map[string]interface{}{
				"queries": []interface{}{
					map[string]interface{}{
						"type": "promql",
						"spec": map[string]interface{}{
							"name":  "A",
							"query": "up",
							"step":  float64(0),
						},
					},
				},
			},
		},
	}

	got, err := client.RuleSpecFromMap(full)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"condition":{"compositeQuery":{"queries":[{"spec":{"name":"A","query":"up"},"type":"promql"}]}},"schemaVersion":"v2alpha1","version":"v5"}`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
