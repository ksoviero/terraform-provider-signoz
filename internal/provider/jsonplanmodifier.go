// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var _ planmodifier.String = canonicalJSONPlanModifier{}

type canonicalJSONPlanModifier struct{}

func (canonicalJSONPlanModifier) Description(context.Context) string {
	return "Normalizes JSON string attributes to a stable compact form so plan matches API read results."
}

func (m canonicalJSONPlanModifier) MarkdownDescription(_ context.Context) string {
	return m.Description(context.Background())
}

func (m canonicalJSONPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	normalized, err := client.CanonicalJSON(req.PlanValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid JSON", err.Error())
		return
	}
	resp.PlanValue = types.StringValue(normalized)
}
