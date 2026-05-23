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
	_ resource.Resource                = &AuthDomainResource{}
	_ resource.ResourceWithImportState = &AuthDomainResource{}
)

func NewAuthDomainResource() resource.Resource { return &AuthDomainResource{} }

type AuthDomainResource struct{ client *client.Client }

type authDomainModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Config    types.String `tfsdk:"config"`
	OrgID     types.String `tfsdk:"org_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func authDomainAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id": types.StringType, "name": types.StringType, "config": types.StringType,
		"org_id": types.StringType, "created_at": types.StringType, "updated_at": types.StringType,
	}
}

func modelFromAuthDomainMap(m map[string]interface{}) (authDomainModel, error) {
	cfg, err := client.SubMapJSON(m, "config")
	if err != nil {
		return authDomainModel{}, err
	}
	return authDomainModel{
		ID:        typesStringValueFromMap(m, "id"),
		Name:      typesStringValueFromMap(m, "name"),
		Config:    types.StringValue(cfg),
		OrgID:     typesStringValueFromMap(m, "orgId"),
		CreatedAt: typesStringValueFromMap(m, "createdAt"),
		UpdatedAt: typesStringValueFromMap(m, "updatedAt"),
	}, nil
}

func (r *AuthDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_domain"
}

func (r *AuthDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docAuthDomainIntro,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{MarkdownDescription: docAuthDomainName, Required: true},
			"config": schema.StringAttribute{
				MarkdownDescription: docAuthDomainConfig,
				Optional:            true,
				Sensitive:           true,
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
			"org_id":     schema.StringAttribute{MarkdownDescription: docOrgID, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"created_at": schema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_at": schema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

func (r *AuthDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResource(context.Background(), req, resp)
}

func (r *AuthDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan authDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, err := client.BuildAuthDomainBody(plan.Name.ValueString(), plan.Config.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid config", err.Error())
		return
	}
	created, err := r.client.CreateAuthDomain(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, err := modelFromAuthDomainMap(created)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AuthDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state authDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	m, err := r.client.GetAuthDomain(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, err := modelFromAuthDomainMap(m)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *AuthDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state authDomainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, err := client.BuildAuthDomainBody(plan.Name.ValueString(), plan.Config.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid config", err.Error())
		return
	}
	if err := r.client.UpdateAuthDomain(ctx, state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	m, err := r.client.GetAuthDomain(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, err := modelFromAuthDomainMap(m)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *AuthDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state authDomainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteAuthDomain(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *AuthDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &AuthDomainDataSource{}

func NewAuthDomainDataSource() datasource.DataSource { return &AuthDomainDataSource{} }

type AuthDomainDataSource struct{ client *client.Client }

func (d *AuthDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_domain"
}

func (d *AuthDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docAuthDomainDataSourceIntro,
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: docLookupID,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"name":       dataschema.StringAttribute{MarkdownDescription: docLookupName, Optional: true},
			"config":     dataschema.StringAttribute{MarkdownDescription: docAuthDomainConfig, Computed: true, Sensitive: true},
			"org_id":     dataschema.StringAttribute{MarkdownDescription: docOrgID, Computed: true},
			"created_at": dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
			"updated_at": dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
		},
	}
}

func (d *AuthDomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func readAuthDomainByName(ctx context.Context, c *client.Client, name string) (map[string]interface{}, error) {
	list, err := c.ListAuthDomains(ctx)
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
		return nil, fmt.Errorf("no auth domain named %q", name)
	case 1:
		return c.GetAuthDomain(ctx, client.MapString(matches[0], "id"))
	default:
		return nil, fmt.Errorf("multiple auth domains named %q; use id", name)
	}
}

func (d *AuthDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config authDomainModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}
	var m map[string]interface{}
	var err error
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		m, err = d.client.GetAuthDomain(ctx, config.ID.ValueString())
	} else {
		m, err = readAuthDomainByName(ctx, d.client, config.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, err := modelFromAuthDomainMap(m)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

var _ datasource.DataSource = &AuthDomainsDataSource{}

func NewAuthDomainsDataSource() datasource.DataSource { return &AuthDomainsDataSource{} }

type AuthDomainsDataSource struct{ client *client.Client }

type authDomainsDataSourceModel struct {
	Domains types.List `tfsdk:"domains"`
}

func (d *AuthDomainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_domains"
}

func (d *AuthDomainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docListAllIntro + " (`GET /api/v1/domains`).",
		Attributes: map[string]dataschema.Attribute{
			"domains": dataschema.ListNestedAttribute{
				MarkdownDescription: "Auth domains in the organization. Read-only.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: authDomainNestedDataSourceAttrs(),
				},
			},
		},
	}
}

func (d *AuthDomainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func (d *AuthDomainsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}
	list, err := d.client.ListAuthDomains(ctx)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(list))
	for _, m := range list {
		model, err := modelFromAuthDomainMap(m)
		if err != nil {
			resp.Diagnostics.AddError("Internal error", err.Error())
			return
		}
		obj, diags := types.ObjectValueFrom(ctx, authDomainAttrTypes(), model)
		resp.Diagnostics.Append(diags...)
		elems = append(elems, obj)
	}
	domains, diags := types.ListValue(types.ObjectType{AttrTypes: authDomainAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &authDomainsDataSourceModel{Domains: domains})...)
}
