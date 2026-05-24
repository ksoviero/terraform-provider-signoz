// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

// normalizeChannelConfigPlanModifier strips top-level "name" and canonicalizes config JSON
// so plan matches the API round-trip stored in state.
type normalizeChannelConfigPlanModifier struct{}

func (normalizeChannelConfigPlanModifier) Description(_ context.Context) string {
	return "Normalize notification channel config JSON (omit top-level name, stable key order)."
}

func (normalizeChannelConfigPlanModifier) MarkdownDescription(ctx context.Context) string {
	return normalizeChannelConfigPlanModifier{}.Description(ctx)
}

func (m normalizeChannelConfigPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	normalized, err := client.NormalizeChannelConfigJSON(req.PlanValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid notification channel config", err.Error())
		return
	}
	resp.PlanValue = types.StringValue(normalized)
}
