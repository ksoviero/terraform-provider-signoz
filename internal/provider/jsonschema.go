// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

// alertRuleSpecType is a custom string type that normalizes alert rule spec JSON
// (stripping API-default fields like disabled:false) before comparing for semantic
// equality. This prevents phantom diffs when users include default-value fields
// in their config that the API omits on read.
type alertRuleSpecType struct {
	basetypes.StringType
}

var alertRuleSpecTypeInstance = alertRuleSpecType{}

func (t alertRuleSpecType) String() string {
	return "alertRuleSpecType"
}

func (t alertRuleSpecType) Equal(o attr.Type) bool {
	_, ok := o.(alertRuleSpecType)
	return ok
}

func (t alertRuleSpecType) ValueType(_ context.Context) attr.Value {
	return alertRuleSpecValue{}
}

func (t alertRuleSpecType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return alertRuleSpecValue{StringValue: in}, nil
}

func (t alertRuleSpecType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}
	v, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue: %v", diags)
	}
	return v, nil
}

// alertRuleSpecValue is the value type for alertRuleSpecType.
type alertRuleSpecValue struct {
	basetypes.StringValue
}

func (v alertRuleSpecValue) Type(_ context.Context) attr.Type {
	return alertRuleSpecTypeInstance
}

func (v alertRuleSpecValue) Equal(o attr.Value) bool {
	other, ok := o.(alertRuleSpecValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals normalizes both JSON strings via NormalizeRuleSpecJSON
// before comparing. This means config values with disabled:false (and other
// API-default fields) are considered equal to the normalized form the API returns.
func (v alertRuleSpecValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(alertRuleSpecValue)
	if !ok {
		// Fall back to raw string equality for unexpected types.
		return v.StringValue.Equal(newValuable.(basetypes.StringValue)), diags
	}

	normA, err := client.NormalizeRuleSpecJSON(v.ValueString())
	if err != nil {
		diags.AddError("Semantic Equality Check Error", "Failed to normalize alert rule spec: "+err.Error())
		return false, diags
	}
	normB, err := client.NormalizeRuleSpecJSON(newValue.ValueString())
	if err != nil {
		diags.AddError("Semantic Equality Check Error", "Failed to normalize alert rule spec: "+err.Error())
		return false, diags
	}
	return normA == normB, diags
}

// ToStringValue implements basetypes.StringValuable.
func (v alertRuleSpecValue) ToStringValue(_ context.Context) (basetypes.StringValue, diag.Diagnostics) {
	return v.StringValue, nil
}

// normalizedJSONAttribute is required JSON with semantic equality (whitespace and key order ignored).
func normalizedJSONAttribute(markdown string, required bool) schema.StringAttribute {
	return normalizedJSONAttributeSensitive(markdown, required, false)
}

func normalizedJSONAttributeSensitive(markdown string, required bool, sensitive bool) schema.StringAttribute {
	attr := schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          jsontypes.NormalizedType{},
		Sensitive:           sensitive,
	}
	if required {
		attr.Required = true
	} else {
		attr.Optional = true
	}
	return attr
}

// normalizedJSONOptional is optional JSON (optionally sensitive) with semantic equality.
func normalizedJSONOptional(markdown string, sensitive bool, planModifiers ...planmodifier.String) schema.StringAttribute {
	attr := schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
		Sensitive:           sensitive,
	}
	if len(planModifiers) > 0 {
		attr.PlanModifiers = planModifiers
	}
	return attr
}

// normalizedJSONOptionalWithUnknown copies prior state when credentials/config are not in config.
func normalizedJSONOptionalWithUnknown(markdown string, sensitive bool) schema.StringAttribute {
	return normalizedJSONOptional(markdown, sensitive, stringplanmodifier.UseStateForUnknown())
}

func newNormalizedJSON(value string) jsontypes.Normalized {
	return jsontypes.NewNormalizedValue(value)
}

// normalizedJSONComputedSensitive is computed JSON with semantic equality (for API payloads).
// Without UseStateForUnknown: the API's payload changes when the resource does, so it's "known
// after apply" on any change, and the old one isn't promised in the plan (issue #17).
func normalizedJSONComputedSensitive(markdown string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          jsontypes.NormalizedType{},
		Computed:            true,
		Sensitive:           true,
	}
}

// importedKey is the private state key ImportState sets: the Read that follows an import has
// no configured value in state to compare the API's JSON with, so it keeps the API's.
const importedKey = "imported"

// markImported is for ImportState (see importedKey).
func markImported(ctx context.Context, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, importedKey, []byte("true"))...)
}

// readImported reports whether this Read follows an import, and clears the flag so later Reads
// compare as usual.
func readImported(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) bool {
	v, diags := req.Private.GetKey(ctx, importedKey)
	resp.Diagnostics.Append(diags...)
	if len(v) == 0 {
		return false
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, importedKey, nil)...)
	return true
}

// shapedForState is what a resource's Read keeps in state for a JSON attribute, given what's in
// state (prior) and the API's value (api): the API's value compared with state on the fields
// state sets only, so fields the API adds are never drift (client.ShapeLike). Create and Update
// keep the planned value itself. After an import (imported), tidied: the API's value as data
// sources show it. An optional attribute left out of the configuration stays null.
func shapedForState(prior basetypes.StringValue, api, tidied string, imported bool, eq client.Equivalent) (basetypes.StringValue, error) {
	if imported {
		return basetypes.NewStringValue(tidied), nil
	}
	if prior.IsNull() || prior.IsUnknown() {
		return prior, nil
	}
	shaped, err := client.ShapeLike(prior.ValueString(), api, eq)
	if err != nil {
		return prior, err
	}
	return basetypes.NewStringValue(shaped), nil
}

// shapeReadJSON sets read, a JSON attribute as Read got it from the API, to what to keep in
// state (shapedForState), for attributes whose API value is the one compared.
func shapeReadJSON(prior jsontypes.Normalized, read *jsontypes.Normalized, imported bool, diags *diag.Diagnostics) {
	v, err := shapedForState(prior.StringValue, read.ValueString(), read.ValueString(), imported, nil)
	if err != nil {
		diags.AddError("Internal error", err.Error())
		return
	}
	*read = jsontypes.Normalized{StringValue: v}
}
