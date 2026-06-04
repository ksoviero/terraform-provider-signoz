// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

// normalizeAlertRuleSpecModifier normalizes the alert rule spec JSON in the plan,
// stripping API-default fields (disabled:false, empty strings, zero numbers) so the
// plan and post-Create state agree without requiring users to omit those fields.
type normalizeAlertRuleSpecModifier struct{}

func (m normalizeAlertRuleSpecModifier) Description(_ context.Context) string {
	return "Normalizes alert rule spec JSON by stripping API-default fields."
}

func (m normalizeAlertRuleSpecModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m normalizeAlertRuleSpecModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsUnknown() || req.PlanValue.IsNull() {
		return
	}
	normalized, err := client.NormalizeRuleSpecJSON(req.PlanValue.ValueString())
	if err != nil {
		// Leave the value as-is; validation will catch malformed JSON.
		return
	}
	resp.PlanValue = types.StringValue(normalized)
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
func normalizedJSONComputedSensitive(markdown string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          jsontypes.NormalizedType{},
		Computed:            true,
		Sensitive:           true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}
