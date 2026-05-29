// SPDX-License-Identifier: MPL-2.0

package client_test

import (
	"testing"

	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

func TestNormalizeChannelConfigJSON(t *testing.T) {
	t.Parallel()

	raw := `{
		"name": "Email",
		"email_configs": [{"to": "a@example.com"}]
	}`
	got, err := client.NormalizeChannelConfigJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"email_configs":[{"to":"a@example.com"}]}`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeChannelConfigJSON_empty(t *testing.T) {
	t.Parallel()

	got, err := client.NormalizeChannelConfigJSON("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestNormalizeChannelConfigJSON_prunesAPIDefaults(t *testing.T) {
	t.Parallel()

	raw := `{
		"name": "Email",
		"email_configs": [{
			"to": "a@example.com",
			"smarthost": "",
			"threading": {},
			"send_resolved": true
		}]
	}`
	got, err := client.NormalizeChannelConfigJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"email_configs":[{"send_resolved":true,"to":"a@example.com"}]}`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

// TestNormalizeChannelConfigJSON_opsgenieDefaults verifies that the defaults
// Alertmanager injects for OpsGenie (api_url, http_config, message, description,
// source templates) are pruned when not explicitly set by the user.
func TestNormalizeChannelConfigJSON_opsgenieDefaults(t *testing.T) {
	t.Parallel()

	// This is what the SigNoz API echoes back after creating an OpsGenie channel
	// with only api_key and send_resolved set.
	raw := `{"name":"OpsGenie (Prod)","opsgenie_configs":[{"send_resolved":true,"http_config":{"tls_config":{"insecure_skip_verify":false},"follow_redirects":true,"enable_http2":true,"proxy_url":""},"api_key":"test-key","api_url":"https://api.opsgenie.com/","message":"{{ .CommonLabels.alertname }}","description":"{{ .CommonAnnotations.description }}","source":"SigNoz","details":{"alertname":"{{ .CommonLabels.alertname }}","severity":"{{ (index .Alerts 0).Labels.severity }}"},"priority":"P1"}]}`
	got, err := client.NormalizeChannelConfigJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	// api_url (default), http_config (all defaults), should be pruned.
	// message, description, source were user-set (non-default values), kept.
	// details is user-set, kept.
	want := `{"opsgenie_configs":[{"api_key":"test-key","description":"{{ .CommonAnnotations.description }}","details":{"alertname":"{{ .CommonLabels.alertname }}","severity":"{{ (index .Alerts 0).Labels.severity }}"},"message":"{{ .CommonLabels.alertname }}","priority":"P1","send_resolved":true,"source":"SigNoz"}]}`
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

// TestNormalizeChannelConfigJSON_slackDefaults verifies that Alertmanager's
// Slack template defaults (username, color, title, etc.) and app_url are pruned.
func TestNormalizeChannelConfigJSON_slackDefaults(t *testing.T) {
	t.Parallel()

	// Simulates the API echoing back defaults when user only set api_url, channel, send_resolved.
	raw := `{"name":"Slack","slack_configs":[{"send_resolved":true,"api_url":"https://hooks.slack.com/services/XXX","app_url":"https://slack.com/api/chat.postMessage","channel":"#alerts","callback_id":"{{ template \"slack.default.callbackid\" . }}","color":"{{ if eq .Status \"firing\" }}danger{{ else }}good{{ end }}","fallback":"{{ template \"slack.default.fallback\" . }}","footer":"{{ template \"slack.default.footer\" . }}","icon_emoji":"{{ template \"slack.default.iconemoji\" . }}","icon_url":"{{ template \"slack.default.iconurl\" . }}","pretext":"{{ template \"slack.default.pretext\" . }}","text":"my custom text","title":"{{ template \"slack.default.title\" . }}","title_link":"{{ template \"slack.default.titlelink\" . }}","username":"{{ template \"slack.default.username\" . }}","timeout":0,"http_config":{"tls_config":{"insecure_skip_verify":false},"follow_redirects":true,"enable_http2":true,"proxy_url":null}}]}`
	got, err := client.NormalizeChannelConfigJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	// app_url, callback_id, color, fallback, footer, icon_emoji, icon_url, pretext,
	// title, title_link, username (all defaults), timeout:0, http_config (all defaults) pruned.
	// text (user-set custom value) and api_url (user-set), channel, send_resolved kept.
	want := `{"slack_configs":[{"api_url":"https://hooks.slack.com/services/XXX","channel":"#alerts","send_resolved":true,"text":"my custom text"}]}`
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

// TestNormalizeChannelConfigJSON_pagerdutyDefaults verifies PagerDuty injected
// defaults (url, client, client_url, description, source, timeout, details) are pruned.
func TestNormalizeChannelConfigJSON_pagerdutyDefaults(t *testing.T) {
	t.Parallel()

	raw := `{"name":"PD","pagerduty_configs":[{"send_resolved":false,"service_key":"test","url":"https://events.pagerduty.com/v2/enqueue","client":"{{ template \"pagerduty.default.client\" . }}","client_url":"{{ template \"pagerduty.default.clientURL\" . }}","description":"{{ template \"pagerduty.default.description\" .}}","source":"{{ template \"pagerduty.default.client\" . }}","timeout":0,"details":{"firing":"{{ .Alerts.Firing | toJson }}","num_firing":"{{ .Alerts.Firing | len }}","num_resolved":"{{ .Alerts.Resolved | len }}","resolved":"{{ .Alerts.Resolved | toJson }}"},"http_config":{"tls_config":{"insecure_skip_verify":false},"follow_redirects":true,"enable_http2":true,"proxy_url":null}}]}`
	got, err := client.NormalizeChannelConfigJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	// All injected defaults pruned; send_resolved:false and service_key kept.
	want := `{"pagerduty_configs":[{"send_resolved":false,"service_key":"test"}]}`
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

// TestNormalizeChannelConfigJSON_pagerdutyUserDetails verifies that a user-set
// details map (even partially matching defaults) is NOT pruned.
func TestNormalizeChannelConfigJSON_pagerdutyUserDetails(t *testing.T) {
	t.Parallel()

	raw := `{"name":"PD","pagerduty_configs":[{"routing_key":"rk","details":{"firing":"{{ .Alerts.Firing | toJson }}","my_custom":"value"}}]}`
	got, err := client.NormalizeChannelConfigJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	// details has extra key "my_custom" so it is NOT the pure default map — kept.
	want := `{"pagerduty_configs":[{"details":{"firing":"{{ .Alerts.Firing | toJson }}","my_custom":"value"},"routing_key":"rk"}]}`
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

// TestNormalizeChannelConfigJSON_telegramDefaults verifies Telegram injected
// defaults (parse_mode, message template) are pruned.
func TestNormalizeChannelConfigJSON_telegramDefaults(t *testing.T) {
	t.Parallel()

	raw := `{"name":"TG","telegram_configs":[{"send_resolved":false,"token":"1234567890","chat":12345,"message":"{{ template \"telegram.default.message\" . }}","parse_mode":"HTML"}]}`
	got, err := client.NormalizeChannelConfigJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	// parse_mode:HTML (default) pruned. message is the default template — but
	// it does not appear in alertmanagerDefaultFieldValues so it is kept.
	// send_resolved:false is kept (user-visible value).
	want := `{"telegram_configs":[{"chat":12345,"message":"{{ template \"telegram.default.message\" . }}","send_resolved":false,"token":"1234567890"}]}`
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}
