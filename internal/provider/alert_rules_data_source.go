// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var _ datasource.DataSource = &AlertRulesDataSource{}

func NewAlertRulesDataSource() datasource.DataSource {
	return &AlertRulesDataSource{}
}

type AlertRulesDataSource struct {
	client *client.Client
}

type alertRulesDataSourceModel struct {
	Rules types.List `tfsdk:"rules"`
}

type alertRuleListElementModel struct {
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

func alertRuleListAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"alert":       types.StringType,
		"alert_type":  types.StringType,
		"rule_type":   types.StringType,
		"description": types.StringType,
		"disabled":    types.BoolType,
		"labels":      types.MapType{ElemType: types.StringType},
		"annotations": types.MapType{ElemType: types.StringType},
		"spec":        types.StringType,
		"state":       types.StringType,
		"created_at":  types.StringType,
		"updated_at":  types.StringType,
		"created_by":  types.StringType,
		"updated_by":  types.StringType,
	}
}

func (d *AlertRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_rules"
}

func (d *AlertRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: "Lists all SigNoz alert rules (`GET /api/v2/rules`).",
		Attributes: map[string]dataschema.Attribute{
			"rules": dataschema.ListNestedAttribute{
				MarkdownDescription: "Alert rules in the organization.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: map[string]dataschema.Attribute{
						"id":          dataschema.StringAttribute{MarkdownDescription: "Rule UUID.", Computed: true},
						"alert":       dataschema.StringAttribute{MarkdownDescription: docAlert, Computed: true},
						"alert_type":  dataschema.StringAttribute{MarkdownDescription: docAlertType, Computed: true},
						"rule_type":   dataschema.StringAttribute{MarkdownDescription: docRuleType, Computed: true},
						"description": dataschema.StringAttribute{MarkdownDescription: docDescription, Computed: true},
						"disabled":    dataschema.BoolAttribute{MarkdownDescription: docDisabled, Computed: true},
						"labels":      dataschema.MapAttribute{MarkdownDescription: docLabels, ElementType: types.StringType, Computed: true},
						"annotations": dataschema.MapAttribute{MarkdownDescription: docAnnotations, ElementType: types.StringType, Computed: true},
						"spec":        dataschema.StringAttribute{MarkdownDescription: docSpec, Computed: true},
						"state":       dataschema.StringAttribute{MarkdownDescription: docRuleState, Computed: true},
						"created_at":  dataschema.StringAttribute{Computed: true},
						"updated_at":  dataschema.StringAttribute{Computed: true},
						"created_by":  dataschema.StringAttribute{Computed: true},
						"updated_by":  dataschema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *AlertRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *AlertRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}

	rules, err := readAllAlertRules(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(rules))
	for _, r := range rules {
		obj, diags := types.ObjectValueFrom(ctx, alertRuleListAttrTypes(), alertRuleListElementModel(r))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: alertRuleListAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := alertRulesDataSourceModel{Rules: list}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
