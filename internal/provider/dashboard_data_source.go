// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var _ datasource.DataSource = &DashboardDataSource{}

func NewDashboardDataSource() datasource.DataSource {
	return &DashboardDataSource{}
}

type DashboardDataSource struct {
	client *client.Client
}

type dashboardDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Data      types.String `tfsdk:"data"`
	Locked    types.Bool   `tfsdk:"locked"`
	Source    types.String `tfsdk:"source"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedBy types.String `tfsdk:"updated_by"`
}

func (d *DashboardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (d *DashboardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docDashboardDataSourceIntro,
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: docID + " Required for this data source.",
				Required:            true,
			},
			"data":       dataschema.StringAttribute{MarkdownDescription: docDashboardData, Computed: true},
			"locked":     dataschema.BoolAttribute{MarkdownDescription: docDashboardLocked, Computed: true},
			"source":     dataschema.StringAttribute{MarkdownDescription: docDashboardSource, Computed: true},
			"created_at": dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
			"updated_at": dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
			"created_by": dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
			"updated_by": dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
		},
	}
}

func (d *DashboardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DashboardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dashboardDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}

	dash, err := d.client.GetDashboard(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	state, err := modelFromDashboard(dash)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}

	out := dashboardDataSourceModel(state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
