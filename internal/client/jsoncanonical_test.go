// SPDX-License-Identifier: MPL-2.0

package client_test

import (
	"testing"

	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

func TestCanonicalJSON_whitespace(t *testing.T) {
	t.Parallel()
	pretty := `{
  "a": 1,
  "b": {"c": 2}
}`
	compact, err := client.CanonicalJSON(pretty)
	if err != nil {
		t.Fatal(err)
	}
	again, err := client.CanonicalJSON(compact)
	if err != nil {
		t.Fatal(err)
	}
	if compact != again {
		t.Fatalf("not stable: %q vs %q", compact, again)
	}
}

func TestCanonicalJSON_keyOrder(t *testing.T) {
	t.Parallel()
	a, err := client.CanonicalJSON(`{"z":1,"a":2}`)
	if err != nil {
		t.Fatal(err)
	}
	b, err := client.CanonicalJSON(`{"a":2,"z":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("key order should not matter: %q vs %q", a, b)
	}
}
