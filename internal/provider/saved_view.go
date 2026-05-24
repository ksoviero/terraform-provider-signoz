// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
	_ resource.Resource                = &SavedViewResource{}
	_ resource.ResourceWithImportState = &SavedViewResource{}
)

func NewSavedViewResource() resource.Resource { return &SavedViewResource{} }

type SavedViewResource struct{ client *client.Client }

type savedViewModel struct {
	ID             types.String         `tfsdk:"id"`
	Name           types.String         `tfsdk:"name"`
	SourcePage     types.String         `tfsdk:"source_page"`
	Category       types.String         `tfsdk:"category"`
	Tags           types.List           `tfsdk:"tags"`
	CompositeQuery jsontypes.Normalized `tfsdk:"composite_query"`
	ExtraData      types.String         `tfsdk:"extra_data"`
	CreatedAt      types.String         `tfsdk:"created_at"`
	UpdatedAt      types.String         `tfsdk:"updated_at"`
	CreatedBy      types.String         `tfsdk:"created_by"`
	UpdatedBy      types.String         `tfsdk:"updated_by"`
}

func savedViewAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id": types.StringType, "name": types.StringType, "source_page": types.StringType,
		"category": types.StringType, "tags": types.ListType{ElemType: types.StringType},
		"composite_query": jsontypes.NormalizedType{}, "extra_data": types.StringType,
		"created_at": types.StringType, "updated_at": types.StringType,
		"created_by": types.StringType, "updated_by": types.StringType,
	}
}

func modelFromSavedViewMap(ctx context.Context, m map[string]interface{}) (savedViewModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	cq, err := client.SubMapJSON(m, "compositeQuery")
	if err != nil {
		diags.AddError("Internal error", err.Error())
		return savedViewModel{}, diags
	}
	tags, d := typesStringListFromMap(ctx, m, "tags")
	diags.Append(d...)
	return savedViewModel{
		ID:             typesStringValueFromMap(m, "id"),
		Name:           typesStringValueFromMap(m, "name"),
		SourcePage:     typesStringValueFromMap(m, "sourcePage"),
		Category:       typesStringValueFromMap(m, "category"),
		Tags:           tags,
		CompositeQuery: newNormalizedJSON(cq),
		ExtraData:      typesStringValueFromMap(m, "extraData"),
		CreatedAt:      typesStringValueFromMap(m, "createdAt"),
		UpdatedAt:      typesStringValueFromMap(m, "updatedAt"),
		CreatedBy:      typesStringValueFromMap(m, "createdBy"),
		UpdatedBy:      typesStringValueFromMap(m, "updatedBy"),
	}, diags
}

func savedViewBodyFromModel(ctx context.Context, m savedViewModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	tags, d := stringListFromTerraform(m.Tags)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body, err := client.BuildSavedViewBody(
		m.Name.ValueString(),
		m.SourcePage.ValueString(),
		m.Category.ValueString(),
		m.ExtraData.ValueString(),
		m.CompositeQuery.ValueString(),
		tags,
	)
	if err != nil {
		diags.AddError("Invalid configuration", err.Error())
		return nil, diags
	}
	return body, diags
}

func (r *SavedViewResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saved_view"
}

func (r *SavedViewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docSavedViewIntro,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{MarkdownDescription: docSavedViewName, Required: true},
			"source_page": schema.StringAttribute{
				MarkdownDescription: docSavedViewSourcePage,
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("logs", "traces", "metrics"),
				},
			},
			"category":        schema.StringAttribute{MarkdownDescription: docSavedViewCategory, Optional: true},
			"tags":            schema.ListAttribute{MarkdownDescription: docSavedViewTags, Optional: true, ElementType: types.StringType},
			"composite_query": normalizedJSONAttribute(docSavedViewCompositeQuery, true),
			"extra_data":      schema.StringAttribute{MarkdownDescription: docSavedViewExtraData, Optional: true},
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

func (r *SavedViewResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResource(context.Background(), req, resp)
}

func (r *SavedViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan savedViewModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, diags := savedViewBodyFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateSavedView(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, diags := modelFromSavedViewMap(ctx, created)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SavedViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state savedViewModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	m, err := r.client.GetSavedView(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, diags := modelFromSavedViewMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *SavedViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state savedViewModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, diags := savedViewBodyFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateSavedView(ctx, state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	m, err := r.client.GetSavedView(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, diags := modelFromSavedViewMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *SavedViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state savedViewModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteSavedView(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *SavedViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

var _ datasource.DataSource = &SavedViewDataSource{}

func NewSavedViewDataSource() datasource.DataSource { return &SavedViewDataSource{} }

type SavedViewDataSource struct{ client *client.Client }

type savedViewDataSourceConfig struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	SourcePage types.String `tfsdk:"source_page"`
	Category   types.String `tfsdk:"category"`
}

func (d *SavedViewDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saved_view"
}

func (d *SavedViewDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docSavedViewDataSourceIntro,
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: docLookupID + " When using `name` lookup, set optional `source_page` and `category` to match list API filters.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"name":            dataschema.StringAttribute{MarkdownDescription: docLookupName, Optional: true},
			"source_page":     dataschema.StringAttribute{MarkdownDescription: docSavedViewSourcePage + " Optional filter when looking up by `name`. Default: omitted.", Optional: true},
			"category":        dataschema.StringAttribute{MarkdownDescription: docSavedViewCategory + " Optional filter when looking up by `name`. Default: omitted.", Optional: true},
			"tags":            dataschema.ListAttribute{MarkdownDescription: docSavedViewTags, ElementType: types.StringType, Computed: true},
			"composite_query": dataschema.StringAttribute{MarkdownDescription: docSavedViewCompositeQuery, Computed: true},
			"extra_data":      dataschema.StringAttribute{MarkdownDescription: docSavedViewExtraData, Computed: true},
			"created_at":      dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
			"updated_at":      dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
			"created_by":      dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
			"updated_by":      dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
		},
	}
}

func (d *SavedViewDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func readSavedViewByName(ctx context.Context, c *client.Client, sourcePage, name, category string) (map[string]interface{}, error) {
	list, err := c.ListSavedViews(ctx, sourcePage, name, category)
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
		return nil, fmt.Errorf("no saved view named %q", name)
	case 1:
		return c.GetSavedView(ctx, client.MapString(matches[0], "id"))
	default:
		return nil, fmt.Errorf("multiple saved views named %q; use id", name)
	}
}

func (d *SavedViewDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config savedViewDataSourceConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}
	var m map[string]interface{}
	var err error
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		m, err = d.client.GetSavedView(ctx, config.ID.ValueString())
	} else {
		m, err = readSavedViewByName(ctx, d.client, config.SourcePage.ValueString(), config.Name.ValueString(), config.Category.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, diags := modelFromSavedViewMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

var _ datasource.DataSource = &SavedViewsDataSource{}

func NewSavedViewsDataSource() datasource.DataSource { return &SavedViewsDataSource{} }

type SavedViewsDataSource struct{ client *client.Client }

type savedViewsDataSourceModel struct {
	SourcePage types.String `tfsdk:"source_page"`
	Name       types.String `tfsdk:"name"`
	Category   types.String `tfsdk:"category"`
	Views      types.List   `tfsdk:"views"`
}

func (d *SavedViewsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_saved_views"
}

func (d *SavedViewsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docSavedViewListIntro,
		Attributes: map[string]dataschema.Attribute{
			"source_page": dataschema.StringAttribute{MarkdownDescription: docSavedViewListSourcePage, Optional: true},
			"name":        dataschema.StringAttribute{MarkdownDescription: docSavedViewListName, Optional: true},
			"category":    dataschema.StringAttribute{MarkdownDescription: docSavedViewListCategory, Optional: true},
			"views": dataschema.ListNestedAttribute{
				MarkdownDescription: "Saved views matching the optional filters. Read-only.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: savedViewNestedDataSourceAttrs(),
				},
			},
		},
	}
}

func (d *SavedViewsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func (d *SavedViewsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config savedViewsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}
	list, err := d.client.ListSavedViews(ctx, config.SourcePage.ValueString(), config.Name.ValueString(), config.Category.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(list))
	for _, m := range list {
		model, diags := modelFromSavedViewMap(ctx, m)
		resp.Diagnostics.Append(diags...)
		obj, diags := types.ObjectValueFrom(ctx, savedViewAttrTypes(), model)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}
	views, diags := types.ListValue(types.ObjectType{AttrTypes: savedViewAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &savedViewsDataSourceModel{
		SourcePage: config.SourcePage,
		Name:       config.Name,
		Category:   config.Category,
		Views:      views,
	})...)
}
