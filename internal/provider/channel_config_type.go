// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var _ basetypes.StringTypable = (*channelConfigNormalizedType)(nil)

// channelConfigNormalizedType is like jsontypes.NormalizedType but strips top-level "name"
// before semantic equality (name comes from the resource name attribute).
type channelConfigNormalizedType struct {
	jsontypes.NormalizedType
}

func (channelConfigNormalizedType) String() string {
	return "provider.channelConfigNormalizedType"
}

func (t channelConfigNormalizedType) Equal(o attr.Type) bool {
	other, ok := o.(channelConfigNormalizedType)
	if !ok {
		return false
	}
	return t.NormalizedType.Equal(other.NormalizedType)
}

func (t channelConfigNormalizedType) ValueType(ctx context.Context) attr.Value {
	return channelConfigNormalized{}
}

func (t channelConfigNormalizedType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	if in.IsNull() || in.IsUnknown() {
		return channelConfigNormalized{Normalized: jsontypes.Normalized{StringValue: in}}, nil
	}
	normalized, err := client.NormalizeChannelConfigJSON(in.ValueString())
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid notification channel config", err.Error())
		return nil, diags
	}
	return channelConfigNormalized{Normalized: jsontypes.NewNormalizedValue(normalized)}, nil
}

func (t channelConfigNormalizedType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}
	valuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}
	return valuable, nil
}

type channelConfigNormalized struct {
	jsontypes.Normalized
}

func (v channelConfigNormalized) Type(_ context.Context) attr.Type {
	return channelConfigNormalizedType{}
}

func (v channelConfigNormalized) Equal(o attr.Value) bool {
	other, ok := o.(channelConfigNormalized)
	if !ok {
		return false
	}
	return v.Normalized.Equal(other.Normalized)
}

func (v channelConfigNormalized) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	other, ok := newValuable.(channelConfigNormalized)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("expected channelConfigNormalized, got %T", newValuable),
		)
		return false, diags
	}
	left, err := client.NormalizeChannelConfigJSON(v.ValueString())
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid notification channel config", err.Error())
		return false, diags
	}
	right, err := client.NormalizeChannelConfigJSON(other.ValueString())
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Invalid notification channel config", err.Error())
		return false, diags
	}
	return jsontypes.NewNormalizedValue(left).StringSemanticEquals(ctx, jsontypes.NewNormalizedValue(right))
}

func channelConfigJSONAttribute(markdown string, required bool, sensitive bool) schema.StringAttribute {
	attr := schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          channelConfigNormalizedType{},
		Sensitive:           sensitive,
	}
	if required {
		attr.Required = true
	} else {
		attr.Optional = true
	}
	return attr
}
