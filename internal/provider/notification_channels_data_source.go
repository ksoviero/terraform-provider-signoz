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

var _ datasource.DataSource = &NotificationChannelsDataSource{}

func NewNotificationChannelsDataSource() datasource.DataSource {
	return &NotificationChannelsDataSource{}
}

type NotificationChannelsDataSource struct {
	client *client.Client
}

type notificationChannelsDataSourceModel struct {
	Channels types.List `tfsdk:"channels"`
}

func channelListAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":         types.StringType,
		"name":       types.StringType,
		"config":     types.StringType,
		"type":       types.StringType,
		"data":       types.StringType,
		"created_at": types.StringType,
		"updated_at": types.StringType,
	}
}

func (d *NotificationChannelsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channels"
}

func (d *NotificationChannelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: "Lists all SigNoz notification channels (`GET /api/v1/channels`).",
		Attributes: map[string]dataschema.Attribute{
			"channels": dataschema.ListNestedAttribute{
				MarkdownDescription: "Notification channels in the organization.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: map[string]dataschema.Attribute{
						"id":         dataschema.StringAttribute{Computed: true},
						"name":       dataschema.StringAttribute{MarkdownDescription: docChannelName, Computed: true},
						"config":     dataschema.StringAttribute{MarkdownDescription: docChannelConfig, Computed: true, Sensitive: true},
						"type":       dataschema.StringAttribute{MarkdownDescription: docChannelType, Computed: true},
						"data":       dataschema.StringAttribute{Computed: true, Sensitive: true},
						"created_at": dataschema.StringAttribute{Computed: true},
						"updated_at": dataschema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *NotificationChannelsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationChannelsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}

	channels, err := d.client.ListChannels(ctx)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(channels))
	for i := range channels {
		state, err := modelFromChannel(&channels[i])
		if err != nil {
			resp.Diagnostics.AddError("Internal error", err.Error())
			return
		}
		obj, diags := types.ObjectValueFrom(ctx, channelListAttrTypes(), map[string]attr.Value{
			"id":         state.ID,
			"name":       state.Name,
			"config":     state.Config,
			"type":       state.Type,
			"data":       state.Data,
			"created_at": state.CreatedAt,
			"updated_at": state.UpdatedAt,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: channelListAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := notificationChannelsDataSourceModel{Channels: list}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
