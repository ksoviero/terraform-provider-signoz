// SPDX-License-Identifier: MPL-2.0

package client_test

import (
	"context"
	"strings"
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

func TestNormalizeRuleSpecJSON_builderQueryDisabledFalse(t *testing.T) {
	t.Parallel()

	// Config has disabled:false inside the builder query spec; API omits it on read.
	config := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"stepInterval": 60,
						"disabled": false,
						"aggregations": [{
							"metricName": "amazonaws.com/AWS/RDS/DatabaseConnections.quantile",
							"temporality": "unspecified",
							"timeAggregation": "min",
							"spaceAggregation": "max",
							"reduceTo": "min"
						}],
						"filter": {"expression": "DBInstanceIdentifier = 'zengrc-prod-4'"},
						"groupBy": [{"name": "DBInstanceIdentifier", "fieldContext": "attribute"}]
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 6871, "matchType": "all_the_times", "op": "above", "channels": ["Slack (Prod)"]}]
			}
		},
		"disabled": false,
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"notificationSettings": {"groupBy": ["__all__"], "renotify": {"enabled": false, "interval": "30m"}},
		"schemaVersion": "v2alpha1",
		"source": "https://signoz.management.zengrc.net/alerts",
		"version": "v5"
	}`

	apiResponse := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"stepInterval": 60,
						"aggregations": [{
							"metricName": "amazonaws.com/AWS/RDS/DatabaseConnections.quantile",
							"temporality": "unspecified",
							"timeAggregation": "min",
							"spaceAggregation": "max",
							"reduceTo": "min"
						}],
						"filter": {"expression": "DBInstanceIdentifier = 'zengrc-prod-4'"},
						"groupBy": [{"name": "DBInstanceIdentifier", "fieldContext": "attribute"}]
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 6871, "matchType": "all_the_times", "op": "above", "channels": ["Slack (Prod)"]}]
			}
		},
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"notificationSettings": {"groupBy": ["__all__"], "renotify": {"enabled": false, "interval": "30m"}},
		"schemaVersion": "v2alpha1",
		"source": "https://signoz.management.zengrc.net/alerts",
		"version": "v5"
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

func TestNormalizeRuleSpecJSON_legendAndUsePolicyDefaults(t *testing.T) {
	t.Parallel()

	// Config omits legend (builder_query spec) and notificationSettings.usePolicy;
	// the API echoes both back with their zero values on read.
	config := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "logs",
						"aggregations": [{"expression": "count()"}],
						"filter": {"expression": "k8s.deployment.name = 'gateway-failover'"}
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 0, "matchType": "1", "op": "1", "channels": ["Webhook"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {
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
						"signal": "logs",
						"legend": "",
						"aggregations": [{"expression": "count()"}],
						"filter": {"expression": "k8s.deployment.name = 'gateway-failover'"}
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 0, "matchType": "1", "op": "1", "channels": ["Webhook"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {
			"usePolicy": false,
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
}

func TestNormalizeRuleSpecJSON_groupByEmptySignal(t *testing.T) {
	t.Parallel()

	// Config omits signal on the groupBy entry; the API echoes it back as "".
	config := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"aggregations": [{"metricName": "k8s.pod.phase", "timeAggregation": "latest", "spaceAggregation": "avg"}],
						"filter": {"expression": ""},
						"groupBy": [{"name": "k8s.pod.name", "fieldContext": "attribute"}]
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 0, "matchType": "1", "op": "1", "channels": ["Webhook"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {"renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing"]}}
	}`

	apiResponse := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"aggregations": [{"metricName": "k8s.pod.phase", "timeAggregation": "latest", "spaceAggregation": "avg"}],
						"filter": {"expression": ""},
						"groupBy": [{"name": "k8s.pod.name", "fieldContext": "attribute", "signal": ""}]
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 0, "matchType": "1", "op": "1", "channels": ["Webhook"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {"renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing"]}}
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

func TestNormalizeRuleSpecJSON_functionNameCase(t *testing.T) {
	t.Parallel()

	// Config names functions in camelCase; the API echoes them back in lower case.
	spec := func(cumulativeSum, clampMax string) string {
		return `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"aggregations": [{"metricName": "signoz.meter.log.size", "timeAggregation": "rate", "spaceAggregation": "sum"}],
						"functions": [{"name": "` + clampMax + `", "args": [{"value": 1}]}, {"name": "` + cumulativeSum + `"}]
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"thresholds": {
				"kind": "basic",
				"spec": [{"name": "critical", "target": 44.1, "matchType": "5", "op": "1", "channels": ["Webhook"]}]
			}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {"renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing"]}}
	}`
	}

	configNorm, err := client.NormalizeRuleSpecJSON(spec("cumulativeSum", "clampMax"))
	if err != nil {
		t.Fatal(err)
	}
	apiNorm, err := client.NormalizeRuleSpecJSON(spec("cumulativesum", "clampmax"))
	if err != nil {
		t.Fatal(err)
	}
	if configNorm != apiNorm {
		t.Fatalf("normalized config and API response differ:\nconfig: %s\napi:    %s", configNorm, apiNorm)
	}
}

func TestNormalizeRuleSpecJSON_nullsAndEvaluationDefaults(t *testing.T) {
	t.Parallel()

	// The config leaves unset fields out; the API (as of November 2026) writes
	// them all back: unset ones as null, and its defaults for nodata alerts,
	// seasonality and the new group grace period.
	config := `{
		"condition": {
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"aggregations": [{"metricName": "signoz.meter.log.size", "timeAggregation": "rate", "spaceAggregation": "sum"}],
						"groupBy": []
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"selectedQueryName": "A",
			"thresholds": {"kind": "basic", "spec": [{"name": "critical", "target": 44.1, "matchType": "3", "op": "1", "channels": ["Cliq"]}]}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "24h0m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {"groupBy": [], "renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing"]}}
	}`
	api := `{
		"condition": {
			"absentFor": 0,
			"alertOnAbsent": false,
			"compositeQuery": {
				"queries": [{
					"type": "builder_query",
					"spec": {
						"name": "A",
						"signal": "metrics",
						"aggregations": [{"metricName": "signoz.meter.log.size", "timeAggregation": "rate", "spaceAggregation": "sum", "reduceTo": "", "temporality": ""}],
						"functions": null,
						"groupBy": [],
						"legend": "",
						"order": null,
						"secondaryAggregations": null,
						"selectFields": null
					}
				}],
				"panelType": "graph",
				"queryType": "builder"
			},
			"seasonality": "",
			"selectedQueryName": "A",
			"thresholds": {"kind": "basic", "spec": [{"name": "critical", "target": 44.1, "matchType": "3", "op": "1", "channels": ["Cliq"], "recoveryTarget": null, "targetUnit": ""}]}
		},
		"version": "v5",
		"evaluation": {"kind": "rolling", "spec": {"evalWindow": "24h0m0s", "frequency": "1m"}},
		"schemaVersion": "v2alpha1",
		"notificationSettings": {"groupBy": [], "newGroupEvalDelay": "0s", "usePolicy": false, "renotify": {"enabled": true, "interval": "30m", "alertStates": ["firing"]}}
	}`

	configNorm, err := client.NormalizeRuleSpecJSON(config)
	if err != nil {
		t.Fatal(err)
	}
	apiNorm, err := client.NormalizeRuleSpecJSON(api)
	if err != nil {
		t.Fatal(err)
	}
	if configNorm != apiNorm {
		t.Fatalf("normalized config and API response differ:\nconfig: %s\napi:    %s", configNorm, apiNorm)
	}

	// Values that aren't the defaults still count
	for from, to := range map[string]string{
		`"absentFor": 0`:            `"absentFor": 10`,
		`"alertOnAbsent": false`:    `"alertOnAbsent": true`,
		`"seasonality": ""`:         `"seasonality": "daily"`,
		`"newGroupEvalDelay": "0s"`: `"newGroupEvalDelay": "2m"`,
	} {
		changed, err := client.NormalizeRuleSpecJSON(strings.Replace(api, from, to, 1))
		if err != nil {
			t.Fatal(err)
		}
		if changed == apiNorm {
			t.Errorf("%s was dropped as a default", to)
		}
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
