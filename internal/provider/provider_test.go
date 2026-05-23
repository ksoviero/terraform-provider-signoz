// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"signoz": providerserver.NewProtocol6WithError(New("test")()),
}

func TestAccProvider_Configure(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC=1 to run acceptance tests")
	}
	if os.Getenv("SIGNOZ_ENDPOINT") == "" || os.Getenv("SIGNOZ_API_KEY") == "" {
		t.Skip("SIGNOZ_ENDPOINT and SIGNOZ_API_KEY required for acceptance tests")
	}
	// Full acceptance tests can be added when a stable test SigNoz instance is available.
}

// Example mock server pattern for future resource acceptance tests.
func TestMockServer_channelPaths(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/channels":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id": "11111111-1111-7111-8111-111111111111", "name": "test", "type": "slack",
					"data": `{}`, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	if srv.URL == "" {
		t.Fatal("expected test server url")
	}
}
