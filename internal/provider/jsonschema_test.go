// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func TestNormalizedJSON_semanticEqualsPrettyConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pretty := `{
  "condition": {
    "compositeQuery": {
      "queries": [{"type": "builder_query", "spec": {"filter": {"expression": "x = 1"}}}]
    }
  }
}`
	compact := `{"condition":{"compositeQuery":{"queries":[{"spec":{"filter":{"expression":"x = 1"}},"type":"builder_query"}]}}}`

	a := jsontypes.NewNormalizedValue(pretty)
	b := jsontypes.NewNormalizedValue(compact)
	equal, diags := a.StringSemanticEquals(ctx, b)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !equal {
		t.Fatal("pretty and compact JSON should be semantically equal")
	}
}
