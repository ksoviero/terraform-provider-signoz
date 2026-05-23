// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &ServiceAccountResource{}
	_ resource.ResourceWithImportState = &ServiceAccountResource{}
)

func NewServiceAccountResource() resource.Resource { return &ServiceAccountResource{} }

type ServiceAccountResource struct{ client *client.Client }

type serviceAccountModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Email     types.String `tfsdk:"email"`
	OrgID     types.String `tfsdk:"org_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func serviceAccountAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id": types.StringType, "name": types.StringType, "email": types.StringType,
		"org_id": types.StringType, "created_at": types.StringType, "updated_at": types.StringType,
	}
}

func modelFromServiceAccountMap(m map[string]interface{}) serviceAccountModel {
	return serviceAccountModel{
		ID:        typesStringValueFromMap(m, "id"),
		Name:      typesStringValueFromMap(m, "name"),
		Email:     typesStringValueFromMap(m, "email"),
		OrgID:     typesStringValueFromMap(m, "orgId"),
		CreatedAt: typesStringValueFromMap(m, "createdAt"),
		UpdatedAt: typesStringValueFromMap(m, "updatedAt"),
	}
}

func (r *ServiceAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *ServiceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docServiceAccountIntro,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{MarkdownDescription: docServiceAccountName, Required: true},
			"id": schema.StringAttribute{
				MarkdownDescription: docID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email":      schema.StringAttribute{MarkdownDescription: docServiceAccountEmail, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"org_id":     schema.StringAttribute{MarkdownDescription: docOrgID, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"created_at": schema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_at": schema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func (r *ServiceAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResource(context.Background(), req, resp)
}

func (r *ServiceAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	created, err := r.client.CreateServiceAccount(ctx, map[string]interface{}{"name": plan.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state := modelFromServiceAccountMap(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServiceAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	m, err := r.client.GetServiceAccount(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out := modelFromServiceAccountMap(m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *ServiceAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceAccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.UpdateServiceAccount(ctx, state.ID.ValueString(), map[string]interface{}{"name": plan.Name.ValueString()}); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	m, err := r.client.GetServiceAccount(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out := modelFromServiceAccountMap(m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *ServiceAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceAccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteServiceAccount(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *ServiceAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &ServiceAccountDataSource{}

func NewServiceAccountDataSource() datasource.DataSource { return &ServiceAccountDataSource{} }

type ServiceAccountDataSource struct{ client *client.Client }

func (d *ServiceAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *ServiceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docServiceAccountDataSourceIntro,
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: docLookupID,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"name":       dataschema.StringAttribute{MarkdownDescription: docLookupName, Optional: true},
			"email":      dataschema.StringAttribute{MarkdownDescription: docServiceAccountEmail, Computed: true},
			"org_id":     dataschema.StringAttribute{MarkdownDescription: docOrgID, Computed: true},
			"created_at": dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
			"updated_at": dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
		},
	}
}

func (d *ServiceAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func readServiceAccountByName(ctx context.Context, c *client.Client, name string) (map[string]interface{}, error) {
	list, err := c.ListServiceAccounts(ctx)
	if err != nil {
		return nil, err
	}
	var matches []map[string]interface{}
	for _, item := range list {
		if client.MapString(item, "name") == name {
			matches = append(matches, item)
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no service account named %q", name)
	case 1:
		return c.GetServiceAccount(ctx, client.MapString(matches[0], "id"))
	default:
		return nil, fmt.Errorf("multiple service accounts named %q; use id", name)
	}
}

func (d *ServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serviceAccountModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}
	var m map[string]interface{}
	var err error
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		m, err = d.client.GetServiceAccount(ctx, config.ID.ValueString())
	} else {
		m, err = readServiceAccountByName(ctx, d.client, config.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state := modelFromServiceAccountMap(m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

var _ datasource.DataSource = &ServiceAccountsDataSource{}

func NewServiceAccountsDataSource() datasource.DataSource { return &ServiceAccountsDataSource{} }

type ServiceAccountsDataSource struct{ client *client.Client }

type serviceAccountsDataSourceModel struct {
	Accounts types.List `tfsdk:"accounts"`
}

func (d *ServiceAccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_accounts"
}

func (d *ServiceAccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docListAllIntro + " (`GET /api/v1/service_accounts`).",
		Attributes: map[string]dataschema.Attribute{
			"accounts": dataschema.ListNestedAttribute{
				MarkdownDescription: "Service accounts in the organization. Read-only.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: serviceAccountNestedDataSourceAttrs(),
				},
			},
		},
	}
}

func (d *ServiceAccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func (d *ServiceAccountsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}
	list, err := d.client.ListServiceAccounts(ctx)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(list))
	for _, m := range list {
		obj, diags := types.ObjectValueFrom(ctx, serviceAccountAttrTypes(), modelFromServiceAccountMap(m))
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	accounts, diags := types.ListValue(types.ObjectType{AttrTypes: serviceAccountAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &serviceAccountsDataSourceModel{Accounts: accounts})...)
}
