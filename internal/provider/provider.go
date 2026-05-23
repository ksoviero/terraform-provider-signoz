// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var _ provider.Provider = &SignozProvider{}

type SignozProvider struct {
	version string
}

type SignozProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (p *SignozProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "signoz"
	resp.Version = p.version
}

func (p *SignozProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Configure access to a [SigNoz](https://signoz.io) instance via its HTTP API. Authentication uses the `SigNoz-Api-Key` header.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base URL of the SigNoz instance, e.g. `https://signoz.example.com` (no trailing `/api`). May be set with the `SIGNOZ_ENDPOINT` environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "SigNoz API key value sent as the `SigNoz-Api-Key` header. May be set with the `SIGNOZ_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "If true, TLS certificate verification is skipped. Use only in lab environments.",
				Optional:            true,
			},
		},
	}
}

func (p *SignozProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SignozProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("SIGNOZ_ENDPOINT")
	if !data.Endpoint.IsNull() && !data.Endpoint.IsUnknown() {
		endpoint = data.Endpoint.ValueString()
	}
	apiKey := os.Getenv("SIGNOZ_API_KEY")
	if !data.APIKey.IsNull() && !data.APIKey.IsUnknown() {
		apiKey = data.APIKey.ValueString()
	}

	if endpoint == "" {
		resp.Diagnostics.AddError("Missing endpoint", "Set provider endpoint or SIGNOZ_ENDPOINT.")
		return
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("Missing api_key", "Set provider api_key or SIGNOZ_API_KEY.")
		return
	}

	insecure := false
	if !data.Insecure.IsNull() && !data.Insecure.IsUnknown() {
		insecure = data.Insecure.ValueBool()
	}

	c, err := client.New(endpoint, apiKey, insecure)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API client", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *SignozProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNotificationChannelResource,
		NewAlertRuleResource,
		NewDashboardResource,
	}
}

func (p *SignozProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAlertRuleDataSource,
		NewAlertRulesDataSource,
		NewNotificationChannelDataSource,
		NewNotificationChannelsDataSource,
		NewDashboardDataSource,
		NewDashboardsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SignozProvider{version: version}
	}
}
