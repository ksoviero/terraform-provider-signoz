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
