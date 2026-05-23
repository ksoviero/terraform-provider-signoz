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

var _ datasource.DataSource = &DashboardsDataSource{}

func NewDashboardsDataSource() datasource.DataSource {
	return &DashboardsDataSource{}
}

type DashboardsDataSource struct {
	client *client.Client
}

type dashboardsDataSourceModel struct {
	Dashboards types.List `tfsdk:"dashboards"`
}

func dashboardListAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":         types.StringType,
		"data":       types.StringType,
		"locked":     types.BoolType,
		"source":     types.StringType,
		"created_at": types.StringType,
		"updated_at": types.StringType,
		"created_by": types.StringType,
		"updated_by": types.StringType,
	}
}

func (d *DashboardsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboards"
}

func (d *DashboardsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docListAllIntro + " (`GET /api/v1/dashboards`).",
		Attributes: map[string]dataschema.Attribute{
			"dashboards": dataschema.ListNestedAttribute{
				MarkdownDescription: "Dashboards in the organization. Read-only.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: dashboardNestedDataSourceAttrs(),
				},
			},
		},
	}
}

func (d *DashboardsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DashboardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}

	dashboards, err := d.client.ListDashboards(ctx)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(dashboards))
	for i := range dashboards {
		state, err := modelFromDashboard(&dashboards[i])
		if err != nil {
			resp.Diagnostics.AddError("Internal error", err.Error())
			return
		}
		obj, diags := types.ObjectValueFrom(ctx, dashboardListAttrTypes(), map[string]attr.Value{
			"id":         state.ID,
			"data":       state.Data,
			"locked":     state.Locked,
			"source":     state.Source,
			"created_at": state.CreatedAt,
			"updated_at": state.UpdatedAt,
			"created_by": state.CreatedBy,
			"updated_by": state.UpdatedBy,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: dashboardListAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := dashboardsDataSourceModel{Dashboards: list}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
