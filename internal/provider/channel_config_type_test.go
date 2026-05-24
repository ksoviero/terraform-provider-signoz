// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestChannelConfigNormalized_semanticEqualsIgnoresName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	typ := channelConfigNormalizedType{}

	withName, diags := typ.ValueFromString(ctx, basetypes.NewStringValue(`{"name":"Email","email_configs":[{"to":"a@example.com"}]}`))
	if diags.HasError() {
		t.Fatal(diags)
	}
	withoutName, diags := typ.ValueFromString(ctx, basetypes.NewStringValue(`{"email_configs":[{"to":"a@example.com"}]}`))
	if diags.HasError() {
		t.Fatal(diags)
	}

	equal, diags := withName.(channelConfigNormalized).StringSemanticEquals(ctx, withoutName)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !equal {
		t.Fatal("config with top-level name should equal config without name")
	}
}

func TestChannelConfigNormalized_valueFromStringStripsName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	typ := channelConfigNormalizedType{}

	val, diags := typ.ValueFromString(ctx, basetypes.NewStringValue(`{"name":"Email","email_configs":[{"to":"a@example.com"}]}`))
	if diags.HasError() {
		t.Fatal(diags)
	}
	got := val.(channelConfigNormalized).ValueString()
	if got != `{"email_configs":[{"to":"a@example.com"}]}` {
		t.Fatalf("got %q", got)
	}
}
