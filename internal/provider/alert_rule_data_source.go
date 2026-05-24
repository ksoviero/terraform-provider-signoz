// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var _ datasource.DataSource = &AlertRuleDataSource{}

func NewAlertRuleDataSource() datasource.DataSource {
	return &AlertRuleDataSource{}
}

type AlertRuleDataSource struct {
	client *client.Client
}

type alertRuleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Alert       types.String `tfsdk:"alert"`
	AlertType   types.String `tfsdk:"alert_type"`
	RuleType    types.String `tfsdk:"rule_type"`
	Description types.String `tfsdk:"description"`
	Disabled    types.Bool   `tfsdk:"disabled"`
	Labels      types.Map    `tfsdk:"labels"`
	Annotations types.Map    `tfsdk:"annotations"`
	Spec        jsontypes.Normalized `tfsdk:"spec"`
	State       types.String `tfsdk:"state"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	CreatedBy   types.String `tfsdk:"created_by"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
}

func (d *AlertRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_rule"
}

func (d *AlertRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docAlertRuleDataSourceIntro,
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: docLookupID + " For alert rules, the other option is `alert` (title).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("alert"),
					),
				},
			},
			"alert": dataschema.StringAttribute{
				MarkdownDescription: docAlert + " Exactly one of `id` or `alert` must be set.",
				Optional:            true,
			},
			"alert_type":  dataschema.StringAttribute{MarkdownDescription: docAlertType, Computed: true},
			"rule_type":   dataschema.StringAttribute{MarkdownDescription: docRuleType, Computed: true},
			"description": dataschema.StringAttribute{MarkdownDescription: docDescription, Computed: true},
			"disabled":    dataschema.BoolAttribute{MarkdownDescription: docDisabled, Computed: true},
			"labels":      dataschema.MapAttribute{MarkdownDescription: docLabels, ElementType: types.StringType, Computed: true},
			"annotations": dataschema.MapAttribute{MarkdownDescription: docAnnotations, ElementType: types.StringType, Computed: true},
			"spec":        dataschema.StringAttribute{MarkdownDescription: docSpec, Computed: true},
			"state":       dataschema.StringAttribute{MarkdownDescription: docRuleState, Computed: true},
			"created_at":  dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
			"updated_at":  dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
			"created_by":  dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
			"updated_by":  dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
		},
	}
}

func (d *AlertRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AlertRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config alertRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}

	var rule alertRuleModel
	var err error
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		rule, err = readAlertRuleByID(ctx, d.client, config.ID.ValueString())
	} else {
		rule, err = readAlertRuleByTitle(ctx, d.client, config.Alert.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	state := alertRuleDataSourceModel(rule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
