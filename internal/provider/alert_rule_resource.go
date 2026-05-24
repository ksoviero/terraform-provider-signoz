// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var (
	_ resource.Resource                = &AlertRuleResource{}
	_ resource.ResourceWithImportState = &AlertRuleResource{}
)

func NewAlertRuleResource() resource.Resource {
	return &AlertRuleResource{}
}

type AlertRuleResource struct {
	client *client.Client
}

type alertRuleModel struct {
	ID          types.String `tfsdk:"id"`
	Alert       types.String `tfsdk:"alert"`
	AlertType   types.String `tfsdk:"alert_type"`
	RuleType    types.String `tfsdk:"rule_type"`
	Description types.String `tfsdk:"description"`
	Disabled    types.Bool   `tfsdk:"disabled"`
	Labels      types.Map    `tfsdk:"labels"`
	Annotations types.Map    `tfsdk:"annotations"`
	Spec        types.String `tfsdk:"spec"`
	State       types.String `tfsdk:"state"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	CreatedBy   types.String `tfsdk:"created_by"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
}

func (r *AlertRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_rule"
}

func (r *AlertRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docAlertRuleIntro,
		Attributes: map[string]schema.Attribute{
			"alert": schema.StringAttribute{
				MarkdownDescription: docAlert,
				Required:            true,
			},
			"alert_type": schema.StringAttribute{
				MarkdownDescription: docAlertType,
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"METRIC_BASED_ALERT",
						"TRACES_BASED_ALERT",
						"LOGS_BASED_ALERT",
						"EXCEPTIONS_BASED_ALERT",
					),
				},
			},
			"rule_type": schema.StringAttribute{
				MarkdownDescription: docRuleType,
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("threshold_rule", "promql_rule", "anomaly_rule"),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: docDescription,
				Optional:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: docDisabled,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: docLabels,
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"annotations": schema.MapAttribute{
				MarkdownDescription: docAnnotations,
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"spec": schema.StringAttribute{
				MarkdownDescription: docSpec,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					canonicalJSONPlanModifier{},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: docID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: docRuleState,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: docCreatedAt,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: docUpdatedAt,
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: docCreatedBy,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_by": schema.StringAttribute{
				MarkdownDescription: docUpdatedBy,
				Computed:            true,
			},
		},
	}
}

func (r *AlertRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func stringMapFromTerraform(m types.Map) map[string]string {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	elems := m.Elements()
	out := make(map[string]string, len(elems))
	for k, v := range elems {
		if s, ok := v.(types.String); ok && !s.IsNull() {
			out[k] = s.ValueString()
		}
	}
	return out
}

func terraformMapFromStrings(m map[string]string) types.Map {
	if len(m) == 0 {
		return types.MapNull(types.StringType)
	}
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	out, _ := types.MapValue(types.StringType, elems)
	return out
}

func (r *AlertRuleResource) ruleBodyFromModel(m alertRuleModel) ([]byte, error) {
	desc := ""
	if !m.Description.IsNull() {
		desc = m.Description.ValueString()
	}
	disabled := false
	if !m.Disabled.IsNull() {
		disabled = m.Disabled.ValueBool()
	}
	return client.BuildRuleBody(
		m.Alert.ValueString(),
		m.AlertType.ValueString(),
		m.RuleType.ValueString(),
		desc,
		disabled,
		stringMapFromTerraform(m.Labels),
		stringMapFromTerraform(m.Annotations),
		m.Spec.ValueString(),
	)
}

func modelFromRuleMap(full map[string]interface{}, spec string) alertRuleModel {
	rule := client.RuleFromMap(full)
	m := alertRuleModel{
		ID:          types.StringValue(rule.ID),
		Alert:       types.StringValue(rule.Alert),
		AlertType:   types.StringValue(rule.AlertType),
		RuleType:    types.StringValue(rule.RuleType),
		State:       types.StringValue(rule.State),
		Spec:        types.StringValue(spec),
		Labels:      terraformMapFromStrings(rule.Labels),
		Annotations: terraformMapFromStrings(rule.Annotations),
		CreatedAt:   types.StringValue(rule.CreatedAt),
		UpdatedAt:   types.StringValue(rule.UpdatedAt),
		CreatedBy:   types.StringValue(rule.CreatedBy),
		UpdatedBy:   types.StringValue(rule.UpdatedBy),
	}
	if rule.Description != "" {
		m.Description = types.StringValue(rule.Description)
	} else {
		m.Description = types.StringNull()
	}
	m.Disabled = types.BoolValue(rule.Disabled)
	return m
}

func (r *AlertRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}

	body, err := r.ruleBodyFromModel(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule", err.Error())
		return
	}

	created, err := r.client.CreateRule(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	full, err := r.client.GetRuleMap(ctx, created.ID)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	spec, err := client.RuleSpecFromMap(full)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}

	state := modelFromRuleMap(full, spec)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AlertRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}

	full, err := r.client.GetRuleMap(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	spec, err := client.RuleSpecFromMap(full)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}

	out := modelFromRuleMap(full, spec)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *AlertRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state alertRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}

	body, err := r.ruleBodyFromModel(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule", err.Error())
		return
	}

	if err := r.client.UpdateRule(ctx, state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	full, err := r.client.GetRuleMap(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	spec, err := client.RuleSpecFromMap(full)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}

	out := modelFromRuleMap(full, spec)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *AlertRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteRule(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *AlertRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
