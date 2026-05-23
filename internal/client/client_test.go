// SPDX-License-Identifier: MPL-2.0

package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

func TestDoJSON_successEnvelope(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("SigNoz-Api-Key") != "test-key" {
			t.Fatalf("missing api key header")
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data": map[string]string{
				"id":   "abc",
				"name": "slack-alerts",
			},
		})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, "test-key", true)
	if err != nil {
		t.Fatal(err)
	}

	var ch client.Channel
	_, err = c.DoJSON(context.Background(), http.MethodGet, "/api/v1/channels/abc", nil, &ch)
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != "abc" || ch.Name != "slack-alerts" {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}

func TestDoJSON_errorEnvelope(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error": map[string]string{
				"code":    "not_found",
				"message": "channel not found",
			},
		})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, "key", true)
	if err != nil {
		t.Fatal(err)
	}

	var ch client.Channel
	_, err = c.DoJSON(context.Background(), http.MethodGet, "/api/v1/channels/x", nil, &ch)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMergeChannelBody(t *testing.T) {
	t.Parallel()

	body, err := client.MergeChannelBody("my-slack", `{"slack_configs":[{"api_url":"https://example.com","channel":"#alerts"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if body["name"] != "my-slack" {
		t.Fatalf("name not merged: %+v", body)
	}
}

func TestBuildRuleBody(t *testing.T) {
	t.Parallel()

	raw, err := client.BuildRuleBody("CPU high", "METRIC_BASED_ALERT", "threshold_rule", "", false, nil, nil, `{"condition":{"compositeQuery":{}},"version":"v5"}`)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["alert"] != "CPU high" || m["version"] != "v5" {
		t.Fatalf("unexpected body: %+v", m)
	}
}
