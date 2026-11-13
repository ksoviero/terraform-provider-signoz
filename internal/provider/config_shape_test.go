// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// fakeRuleAPI stores alert rules as they're sent, and returns them as SigNoz does: with its
// defaults added (nulls, absentFor:0, alertOnAbsent:false, seasonality:"", newGroupEvalDelay:"0s",
// a field no one has heard of yet), function names in lower case, and the evaluation echoed at
// the top level.
type fakeRuleAPI struct {
	mu    sync.Mutex
	rules map[string]map[string]interface{}
}

func (f *fakeRuleAPI) echo(stored map[string]interface{}) map[string]interface{} {
	var out map[string]interface{}
	b, _ := json.Marshal(stored)
	_ = json.Unmarshal(b, &out)
	out["id"] = "rule-1"
	out["state"] = "inactive"
	out["createdBy"] = "terraform@example.com"
	cond := out["condition"].(map[string]interface{})
	for k, v := range map[string]interface{}{"absentFor": 0.0, "alertOnAbsent": false, "seasonality": "", "aBrandNewDefault": "x"} {
		if _, ok := cond[k]; !ok {
			cond[k] = v
		}
	}
	for _, q := range cond["compositeQuery"].(map[string]interface{})["queries"].([]interface{}) {
		spec := q.(map[string]interface{})["spec"].(map[string]interface{})
		for _, k := range []string{"functions", "order", "secondaryAggregations", "selectFields"} {
			if _, ok := spec[k]; !ok {
				spec[k] = nil
			}
		}
		if fns, ok := spec["functions"].([]interface{}); ok {
			for _, fn := range fns {
				m := fn.(map[string]interface{})
				m["name"] = strings.ToLower(m["name"].(string))
			}
		}
	}
	ns := out["notificationSettings"].(map[string]interface{})
	if _, ok := ns["newGroupEvalDelay"]; !ok {
		ns["newGroupEvalDelay"] = "0s"
	}
	eval := out["evaluation"].(map[string]interface{})["spec"].(map[string]interface{})
	out["evalWindow"], out["frequency"] = eval["evalWindow"], eval["frequency"]
	return out
}

func (f *fakeRuleAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	reply := func(data interface{}) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": data})
	}
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/v2/rules":
		var body map[string]interface{}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		f.rules["rule-1"] = body
		w.WriteHeader(http.StatusCreated)
		reply(f.echo(body))
	case r.Method == http.MethodPut && r.URL.Path == "/api/v2/rules/rule-1":
		var body map[string]interface{}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		f.rules["rule-1"] = body
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && r.URL.Path == "/api/v2/rules/rule-1":
		stored, ok := f.rules["rule-1"]
		if !ok {
			http.NotFound(w, r)
			return
		}
		reply(f.echo(stored))
	case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/rules/rule-1":
		delete(f.rules, "rule-1")
		w.WriteHeader(http.StatusNoContent)
	default:
		http.NotFound(w, r)
	}
}

// editTarget changes the stored rule's threshold, as an edit in SigNoz's UI would.
func (f *fakeRuleAPI) editTarget(target float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cond := f.rules["rule-1"]["condition"].(map[string]interface{})
	tier := cond["thresholds"].(map[string]interface{})["spec"].([]interface{})[0].(map[string]interface{})
	tier["target"] = target
}

func alertRuleConfig(endpoint, extraCondition string) string {
	return `
provider "signoz" {
  endpoint = "` + endpoint + `"
  api_key  = "test"
}

resource "signoz_alert_rule" "test" {
  alert      = "Bill on pace"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "threshold_rule"
  spec = jsonencode({
    schemaVersion = "v2alpha1"
    version       = "v5"
    condition = merge({
      compositeQuery = {
        queryType = "builder"
        panelType = "graph"
        queries = [{
          type = "builder_query"
          spec = {
            name         = "A"
            signal       = "metrics"
            aggregations = [{ metricName = "signoz.meter.log.size", timeAggregation = "rate", spaceAggregation = "sum" }]
            functions    = [{ name = "cumulativeSum" }]
            groupBy      = []
          }
        }]
      }
      selectedQueryName = "A"
      thresholds = {
        kind = "basic"
        spec = [{ name = "critical", target = 44.1, op = "1", matchType = "5", channels = ["Cliq"] }]
      }
    }, ` + extraCondition + `)
    evaluation = { kind = "rolling", spec = { evalWindow = "24h0m0s", frequency = "1m" } }
    notificationSettings = {
      groupBy  = []
      renotify = { enabled = true, interval = "30m", alertStates = ["firing"] }
    }
  })
}
`
}

// The API adding defaults and rewriting function names leaves the plan empty; an edit in SigNoz
// to a field the config sets, or a field removed from the config, still shows as a change.
// Runs Terraform against a fake API, so it needs terraform on PATH (CI's unit tests set it up);
// it's skipped without one rather than downloading it.
func TestAlertRuleSpec_onlyConfiguredFieldsCount(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform isn't on PATH")
	}
	api := &fakeRuleAPI{rules: map[string]map[string]interface{}{}}
	srv := httptest.NewServer(api)
	defer srv.Close()

	factories := map[string]func() (tfprotov6.ProviderServer, error){
		"signoz": providerserver.NewProtocol6WithError(New("test")()),
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		Steps: []resource.TestStep{
			// Applied, then planned again: no changes, whatever the API added
			{Config: alertRuleConfig(srv.URL, `{ alertOnAbsent = true, absentFor = 10 }`)},
			// Someone changes the threshold in SigNoz: the plan puts it back
			{
				PreConfig:          func() { api.editTarget(50) },
				Config:             alertRuleConfig(srv.URL, `{ alertOnAbsent = true, absentFor = 10 }`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{Config: alertRuleConfig(srv.URL, `{ alertOnAbsent = true, absentFor = 10 }`)},
			// Fields taken out of the config are updated away, and the API's defaults for them
			// don't come back as drift
			{Config: alertRuleConfig(srv.URL, `{}`)},
			// An import keeps the API's spec, tidied of its defaults
			{
				Config:        alertRuleConfig(srv.URL, `{}`),
				ResourceName:  "signoz_alert_rule.test",
				ImportState:   true,
				ImportStateId: "rule-1",
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					spec := states[0].Attributes["spec"]
					if !strings.Contains(spec, `"selectedQueryName":"A"`) || strings.Contains(spec, `"order":null`) {
						return fmt.Errorf("imported spec isn't the API's: %s", spec)
					}
					return nil
				},
			},
		},
	})
}

func TestShapedForState(t *testing.T) {
	t.Parallel()

	null := basetypes.NewStringNull()
	set := basetypes.NewStringValue(`{"a":1}`)
	api, tidied := `{"a":1,"b":null}`, `{"a":1}`
	for _, tc := range []struct {
		name     string
		prior    basetypes.StringValue
		imported bool
		want     basetypes.StringValue
	}{
		{"an optional attribute left out stays null", null, false, null},
		{"an import keeps the API's, tidied", null, true, basetypes.NewStringValue(tidied)},
		{"a configured value is compared on its fields", set, false, set},
	} {
		got, err := shapedForState(tc.prior, api, tidied, tc.imported, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equal(tc.want) {
			t.Errorf("%s: got %s, want %s", tc.name, got, tc.want)
		}
	}
}
