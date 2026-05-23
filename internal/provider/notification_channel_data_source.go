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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var _ datasource.DataSource = &NotificationChannelDataSource{}

func NewNotificationChannelDataSource() datasource.DataSource {
	return &NotificationChannelDataSource{}
}

type NotificationChannelDataSource struct {
	client *client.Client
}

type notificationChannelDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Config    types.String `tfsdk:"config"`
	Type      types.String `tfsdk:"type"`
	Data      types.String `tfsdk:"data"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *NotificationChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (d *NotificationChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: "Reads a single SigNoz notification channel by `id` or by unique `name`.",
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: "Channel UUID. Exactly one of `id` or `name` must be set.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("name"),
					),
				},
			},
			"name": dataschema.StringAttribute{
				MarkdownDescription: docChannelName + " Exactly one of `id` or `name` must be set.",
				Optional:            true,
			},
			"config": dataschema.StringAttribute{
				MarkdownDescription: docChannelConfig,
				Computed:            true,
				Sensitive:           true,
			},
			"type": dataschema.StringAttribute{
				MarkdownDescription: docChannelType,
				Computed:            true,
			},
			"data": dataschema.StringAttribute{
				MarkdownDescription: "Stored receiver JSON from the API.",
				Computed:            true,
				Sensitive:           true,
			},
			"created_at": dataschema.StringAttribute{Computed: true},
			"updated_at": dataschema.StringAttribute{Computed: true},
		},
	}
}

func (d *NotificationChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func readChannelByName(ctx context.Context, c *client.Client, name string) (*client.Channel, error) {
	channels, err := c.ListChannels(ctx)
	if err != nil {
		return nil, err
	}
	var matches []client.Channel
	for _, ch := range channels {
		if ch.Name == name {
			matches = append(matches, ch)
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no notification channel named %q", name)
	case 1:
		return c.GetChannel(ctx, matches[0].ID)
	default:
		return nil, fmt.Errorf("multiple notification channels named %q (%d matches); use id instead", name, len(matches))
	}
}

func (d *NotificationChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config notificationChannelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}

	var ch *client.Channel
	var err error
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		ch, err = d.client.GetChannel(ctx, config.ID.ValueString())
	} else {
		ch, err = readChannelByName(ctx, d.client, config.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	state, err := modelFromChannel(ch)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}

	out := notificationChannelDataSourceModel(state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}
