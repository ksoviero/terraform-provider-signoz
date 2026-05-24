// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &RoutePolicyResource{}
	_ resource.ResourceWithImportState = &RoutePolicyResource{}
)

func NewRoutePolicyResource() resource.Resource { return &RoutePolicyResource{} }

type RoutePolicyResource struct{ client *client.Client }

type routePolicyModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Expression  types.String `tfsdk:"expression"`
	Kind        types.String `tfsdk:"kind"`
	Channels    types.List   `tfsdk:"channels"`
	Tags        types.List   `tfsdk:"tags"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	CreatedBy   types.String `tfsdk:"created_by"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
}

func routePolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id": types.StringType, "name": types.StringType, "description": types.StringType,
		"expression": types.StringType, "kind": types.StringType,
		"channels":   types.ListType{ElemType: types.StringType},
		"tags":       types.ListType{ElemType: types.StringType},
		"created_at": types.StringType, "updated_at": types.StringType,
		"created_by": types.StringType, "updated_by": types.StringType,
	}
}

func modelFromRoutePolicyMap(ctx context.Context, m map[string]interface{}) (routePolicyModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	channels, d := typesStringListFromMap(ctx, m, "channels")
	diags.Append(d...)
	tags, d := typesStringListFromMap(ctx, m, "tags")
	diags.Append(d...)
	return routePolicyModel{
		ID:          typesStringValueFromMap(m, "id"),
		Name:        typesStringValueFromMap(m, "name"),
		Description: typesStringValueFromMap(m, "description"),
		Expression:  typesStringValueFromMap(m, "expression"),
		Kind:        typesStringValueFromMap(m, "kind"),
		Channels:    channels,
		Tags:        tags,
		CreatedAt:   typesStringValueFromMap(m, "createdAt"),
		UpdatedAt:   typesStringValueFromMap(m, "updatedAt"),
		CreatedBy:   typesStringValueFromMap(m, "createdBy"),
		UpdatedBy:   typesStringValueFromMap(m, "updatedBy"),
	}, diags
}

func routePolicyBodyFromModel(ctx context.Context, m routePolicyModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	channels, d := stringListFromTerraform(m.Channels)
	diags.Append(d...)
	tags, d := stringListFromTerraform(m.Tags)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return client.BuildRoutePolicyBody(
		m.Name.ValueString(),
		m.Description.ValueString(),
		m.Expression.ValueString(),
		m.Kind.ValueString(),
		channels,
		tags,
	), diags
}

func (r *RoutePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_policy"
}

func (r *RoutePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docRoutePolicyIntro,
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{MarkdownDescription: docRoutePolicyName, Required: true},
			"description": schema.StringAttribute{MarkdownDescription: docRoutePolicyDescription, Optional: true},
			"expression":  schema.StringAttribute{MarkdownDescription: docRoutePolicyExpression, Required: true},
			"kind": schema.StringAttribute{
				MarkdownDescription: docRoutePolicyKind,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("rule", "policy"),
				},
			},
			"channels": schema.ListAttribute{MarkdownDescription: docRoutePolicyChannels, Required: true, ElementType: types.StringType},
			"tags":     schema.ListAttribute{MarkdownDescription: docRoutePolicyTags, Optional: true, ElementType: types.StringType},
			"id": schema.StringAttribute{
				MarkdownDescription: docID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_at": schema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
			"created_by": schema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_by": schema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
		},
	}
}

func (r *RoutePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResource(context.Background(), req, resp)
}

func (r *RoutePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, diags := routePolicyBodyFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateRoutePolicy(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, diags := modelFromRoutePolicyMap(ctx, created)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RoutePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	m, err := r.client.GetRoutePolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, diags := modelFromRoutePolicyMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *RoutePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state routePolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, diags := routePolicyBodyFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateRoutePolicy(ctx, state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	m, err := r.client.GetRoutePolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, diags := modelFromRoutePolicyMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *RoutePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routePolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteRoutePolicy(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *RoutePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &RoutePolicyDataSource{}

func NewRoutePolicyDataSource() datasource.DataSource { return &RoutePolicyDataSource{} }

type RoutePolicyDataSource struct{ client *client.Client }

func (d *RoutePolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_policy"
}

func (d *RoutePolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docRoutePolicyDataSourceIntro,
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: docLookupID,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"name":        dataschema.StringAttribute{MarkdownDescription: docLookupName, Optional: true},
			"description": dataschema.StringAttribute{MarkdownDescription: docRoutePolicyDescription, Computed: true},
			"expression":  dataschema.StringAttribute{MarkdownDescription: docRoutePolicyExpression, Computed: true},
			"kind":        dataschema.StringAttribute{MarkdownDescription: docRoutePolicyKind, Computed: true},
			"channels":    dataschema.ListAttribute{MarkdownDescription: docRoutePolicyChannels, ElementType: types.StringType, Computed: true},
			"tags":        dataschema.ListAttribute{MarkdownDescription: docRoutePolicyTags, ElementType: types.StringType, Computed: true},
			"created_at":  dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
			"updated_at":  dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
			"created_by":  dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
			"updated_by":  dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
		},
	}
}

func (d *RoutePolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func readRoutePolicyByName(ctx context.Context, c *client.Client, name string) (map[string]interface{}, error) {
	list, err := c.ListRoutePolicies(ctx)
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
		return nil, fmt.Errorf("no route policy named %q", name)
	case 1:
		return c.GetRoutePolicy(ctx, client.MapString(matches[0], "id"))
	default:
		return nil, fmt.Errorf("multiple route policies named %q; use id", name)
	}
}

func (d *RoutePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config routePolicyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}
	var m map[string]interface{}
	var err error
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		m, err = d.client.GetRoutePolicy(ctx, config.ID.ValueString())
	} else {
		m, err = readRoutePolicyByName(ctx, d.client, config.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, diags := modelFromRoutePolicyMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

var _ datasource.DataSource = &RoutePoliciesDataSource{}

func NewRoutePoliciesDataSource() datasource.DataSource { return &RoutePoliciesDataSource{} }

type RoutePoliciesDataSource struct{ client *client.Client }

type routePoliciesDataSourceModel struct {
	Policies types.List `tfsdk:"policies"`
}

func (d *RoutePoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route_policies"
}

func (d *RoutePoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docListAllIntro + " (`GET /api/v1/route_policies`).",
		Attributes: map[string]dataschema.Attribute{
			"policies": dataschema.ListNestedAttribute{
				MarkdownDescription: "Route policies in the organization. Read-only.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: routePolicyNestedDataSourceAttrs(),
				},
			},
		},
	}
}

func (d *RoutePoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func (d *RoutePoliciesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}
	list, err := d.client.ListRoutePolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(list))
	for _, m := range list {
		model, diags := modelFromRoutePolicyMap(ctx, m)
		resp.Diagnostics.Append(diags...)
		obj, diags := types.ObjectValueFrom(ctx, routePolicyAttrTypes(), model)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}
	policies, diags := types.ListValue(types.ObjectType{AttrTypes: routePolicyAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &routePoliciesDataSourceModel{Policies: policies})...)
}
