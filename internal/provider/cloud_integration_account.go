// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var (
	_ resource.Resource                = &CloudIntegrationAccountResource{}
	_ resource.ResourceWithImportState = &CloudIntegrationAccountResource{}
)

func NewCloudIntegrationAccountResource() resource.Resource {
	return &CloudIntegrationAccountResource{}
}

type CloudIntegrationAccountResource struct{ client *client.Client }

type cloudIntegrationAccountModel struct {
	ID                types.String `tfsdk:"id"`
	CloudProvider     types.String `tfsdk:"cloud_provider"`
	Config            types.String `tfsdk:"config"`
	Credentials       types.String `tfsdk:"credentials"`
	AccountProvider   types.String `tfsdk:"account_provider"`
	ProviderAccountID types.String `tfsdk:"provider_account_id"`
	OrgID             types.String `tfsdk:"org_id"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func cloudIntegrationAccountAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id": types.StringType, "cloud_provider": types.StringType, "config": types.StringType,
		"credentials": types.StringType, "account_provider": types.StringType,
		"provider_account_id": types.StringType, "org_id": types.StringType,
		"created_at": types.StringType, "updated_at": types.StringType,
	}
}

func modelFromCloudAccountMap(m map[string]interface{}, cloudProvider string) (cloudIntegrationAccountModel, error) {
	cfg, err := client.SubMapJSON(m, "config")
	if err != nil {
		return cloudIntegrationAccountModel{}, err
	}
	return cloudIntegrationAccountModel{
		ID:                typesStringValueFromMap(m, "id"),
		CloudProvider:     types.StringValue(cloudProvider),
		Config:            types.StringValue(cfg),
		Credentials:       types.StringNull(),
		AccountProvider:   typesStringValueFromMap(m, "provider"),
		ProviderAccountID: typesStringValueFromMap(m, "providerAccountId"),
		OrgID:             typesStringValueFromMap(m, "orgId"),
		CreatedAt:         typesStringValueFromMap(m, "createdAt"),
		UpdatedAt:         typesStringValueFromMap(m, "updatedAt"),
	}, nil
}

func (r *CloudIntegrationAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_integration_account"
}

func (r *CloudIntegrationAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docCloudIntro,
		Attributes: map[string]schema.Attribute{
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: docCloudProvider,
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("aws", "azure"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.StringAttribute{
				MarkdownDescription: docCloudConfig,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					canonicalJSONPlanModifier{},
				},
			},
			"credentials": schema.StringAttribute{
				MarkdownDescription: docCloudCredentials,
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					canonicalJSONPlanModifier{},
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: docID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_provider": schema.StringAttribute{
				MarkdownDescription: docCloudAccountProvider,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"provider_account_id": schema.StringAttribute{
				MarkdownDescription: docCloudProviderAccountID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"org_id":     schema.StringAttribute{MarkdownDescription: docOrgID, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"created_at": schema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_at": schema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func (r *CloudIntegrationAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResource(context.Background(), req, resp)
}

func (r *CloudIntegrationAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudIntegrationAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if plan.Credentials.IsNull() || plan.Credentials.ValueString() == "" {
		resp.Diagnostics.AddError("Missing credentials", "credentials is required on create.")
		return
	}
	body, err := client.BuildCloudAccountBody(plan.Config.ValueString(), plan.Credentials.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}
	created, err := r.client.CreateCloudAccount(ctx, plan.CloudProvider.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, err := modelFromCloudAccountMap(created, plan.CloudProvider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	state.Credentials = plan.Credentials
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CloudIntegrationAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudIntegrationAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	m, err := r.client.GetCloudAccount(ctx, state.CloudProvider.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, err := modelFromCloudAccountMap(m, state.CloudProvider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	out.Credentials = state.Credentials
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *CloudIntegrationAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state cloudIntegrationAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, err := client.BuildCloudAccountUpdateBody(plan.Config.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid config", err.Error())
		return
	}
	if err := r.client.UpdateCloudAccount(ctx, state.CloudProvider.ValueString(), state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	m, err := r.client.GetCloudAccount(ctx, state.CloudProvider.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, err := modelFromCloudAccountMap(m, state.CloudProvider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	out.Credentials = state.Credentials
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *CloudIntegrationAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudIntegrationAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteCloudAccount(ctx, state.CloudProvider.ValueString(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *CloudIntegrationAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID, 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Use format `<cloud_provider>/<id>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cloud_provider"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func splitImportID(id string, n int) []string {
	parts := make([]string, 0, n)
	start := 0
	for i := 0; i < len(id) && len(parts) < n-1; i++ {
		if id[i] == '/' {
			parts = append(parts, id[start:i])
			start = i + 1
		}
	}
	if start <= len(id) {
		parts = append(parts, id[start:])
	}
	return parts
}

var _ datasource.DataSource = &CloudIntegrationAccountDataSource{}

func NewCloudIntegrationAccountDataSource() datasource.DataSource {
	return &CloudIntegrationAccountDataSource{}
}

type CloudIntegrationAccountDataSource struct{ client *client.Client }

type cloudIntegrationAccountDataSourceConfig struct {
	CloudProvider types.String `tfsdk:"cloud_provider"`
	ID            types.String `tfsdk:"id"`
}

func (d *CloudIntegrationAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_integration_account"
}

func (d *CloudIntegrationAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docCloudDataSourceIntro,
		Attributes: map[string]dataschema.Attribute{
			"cloud_provider":      dataschema.StringAttribute{MarkdownDescription: docCloudProvider, Required: true},
			"id":                  dataschema.StringAttribute{MarkdownDescription: docID + " Required together with `cloud_provider`.", Required: true},
			"config":              dataschema.StringAttribute{MarkdownDescription: docCloudConfig, Computed: true},
			"account_provider":    dataschema.StringAttribute{MarkdownDescription: docCloudAccountProvider, Computed: true},
			"provider_account_id": dataschema.StringAttribute{MarkdownDescription: docCloudProviderAccountID, Computed: true},
			"org_id":              dataschema.StringAttribute{MarkdownDescription: docOrgID, Computed: true},
			"created_at":          dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
			"updated_at":          dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
		},
	}
}

func (d *CloudIntegrationAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func (d *CloudIntegrationAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config cloudIntegrationAccountDataSourceConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}
	m, err := d.client.GetCloudAccount(ctx, config.CloudProvider.ValueString(), config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, err := modelFromCloudAccountMap(m, config.CloudProvider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

var _ datasource.DataSource = &CloudIntegrationAccountsDataSource{}

func NewCloudIntegrationAccountsDataSource() datasource.DataSource {
	return &CloudIntegrationAccountsDataSource{}
}

type CloudIntegrationAccountsDataSource struct{ client *client.Client }

type cloudIntegrationAccountsDataSourceModel struct {
	CloudProvider types.String `tfsdk:"cloud_provider"`
	Accounts      types.List   `tfsdk:"accounts"`
}

func (d *CloudIntegrationAccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_integration_accounts"
}

func (d *CloudIntegrationAccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docCloudListIntro,
		Attributes: map[string]dataschema.Attribute{
			"cloud_provider": dataschema.StringAttribute{MarkdownDescription: docCloudProvider, Required: true},
			"accounts": dataschema.ListNestedAttribute{
				MarkdownDescription: "Cloud integration accounts for the provider. Read-only.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: cloudAccountNestedDataSourceAttrs(),
				},
			},
		},
	}
}

func (d *CloudIntegrationAccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func (d *CloudIntegrationAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config cloudIntegrationAccountsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}
	list, err := d.client.ListCloudAccounts(ctx, config.CloudProvider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(list))
	for _, m := range list {
		model, err := modelFromCloudAccountMap(m, config.CloudProvider.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Internal error", err.Error())
			return
		}
		obj, diags := types.ObjectValueFrom(ctx, cloudIntegrationAccountAttrTypes(), model)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	accounts, diags := types.ListValue(types.ObjectType{AttrTypes: cloudIntegrationAccountAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &cloudIntegrationAccountsDataSourceModel{
		CloudProvider: config.CloudProvider,
		Accounts:      accounts,
	})...)
}
