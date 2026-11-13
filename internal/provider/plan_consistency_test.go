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
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// fakeChannelAPI stores a channel as it's sent, and returns it as SigNoz does: the receiver JSON
// with its name and Alertmanager's send_resolved default in it, and the type taken from the
// receiver kind.
type fakeChannelAPI struct {
	mu   sync.Mutex
	body map[string]interface{}
}

func (f *fakeChannelAPI) channel() map[string]interface{} {
	data := map[string]interface{}{}
	b, _ := json.Marshal(f.body)
	_ = json.Unmarshal(b, &data)
	kind := ""
	for k, v := range data {
		if strings.HasSuffix(k, "_configs") {
			kind = strings.TrimSuffix(k, "_configs")
			for _, c := range v.([]interface{}) {
				if _, ok := c.(map[string]interface{})["send_resolved"]; !ok {
					c.(map[string]interface{})["send_resolved"] = true
				}
			}
		}
	}
	raw, _ := json.Marshal(data)
	return map[string]interface{}{"id": "ch-1", "name": data["name"], "type": kind, "data": string(raw), "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}
}

func (f *fakeChannelAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	reply := func(d interface{}) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": d})
	}
	switch {
	case (r.Method == http.MethodPost && r.URL.Path == "/api/v1/channels") || (r.Method == http.MethodPut && r.URL.Path == "/api/v1/channels/ch-1"):
		raw, _ := io.ReadAll(r.Body)
		f.body = map[string]interface{}{}
		_ = json.Unmarshal(raw, &f.body)
		reply(f.channel())
	case r.Method == http.MethodGet && r.URL.Path == "/api/v1/channels/ch-1":
		reply(f.channel())
	case r.Method == http.MethodDelete:
		w.WriteHeader(http.StatusNoContent)
	default:
		http.NotFound(w, r)
	}
}

func channelConfig(url, name, kind, channel string) string {
	return `
provider "signoz" {
  endpoint = "` + url + `"
  api_key  = "test"
}
resource "signoz_notification_channel" "test" {
  name   = "` + name + `"
  config = jsonencode({ ` + kind + `_configs = [{ api_url = "https://example.com/x", channel = "` + channel + `" }] })
}
`
}

// Changing a channel's config, name or receiver kind applies (issue #17), with data and type read
// back from the API as they are after the change.
func TestNotificationChannel_updates(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform isn't on PATH")
	}
	api := &fakeChannelAPI{}
	srv := httptest.NewServer(api)
	defer srv.Close()
	dataHas := func(s string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttrWith("signoz_notification_channel.test", "data", func(v string) error {
			if !strings.Contains(v, s) {
				return fmt.Errorf("data %s doesn't have %s", v, s)
			}
			return nil
		})
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"signoz": providerserver.NewProtocol6WithError(New("test")())},
		Steps: []resource.TestStep{
			{Config: channelConfig(srv.URL, "alerts", "slack", "#a")},
			{Config: channelConfig(srv.URL, "alerts", "slack", "#b"), Check: dataHas(`#b`)},
			{Config: channelConfig(srv.URL, "alerts-renamed", "slack", "#b"), Check: dataHas(`alerts-renamed`)},
			{
				Config: channelConfig(srv.URL, "alerts-renamed", "webhook", "#b"),
				Check:  resource.TestCheckResourceAttr("signoz_notification_channel.test", "type", "webhook"),
			},
		},
	})
}

// An empty description or labels map, which the API leaves out, applies and stays as configured.
func TestAlertRule_emptyValues(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform isn't on PATH")
	}
	api := &fakeRuleAPI{rules: map[string]map[string]interface{}{}}
	srv := httptest.NewServer(api)
	defer srv.Close()
	cfg := strings.Replace(alertRuleConfig(srv.URL, `{}`), `rule_type  = "threshold_rule"`, `rule_type  = "threshold_rule"
  description = ""
  labels      = {}`, 1)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"signoz": providerserver.NewProtocol6WithError(New("test")())},
		Steps:                    []resource.TestStep{{Config: cfg}},
	})
}
