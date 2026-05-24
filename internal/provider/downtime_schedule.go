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

// --- resource ---

var (
	_ resource.Resource                = &DowntimeScheduleResource{}
	_ resource.ResourceWithImportState = &DowntimeScheduleResource{}
)

func NewDowntimeScheduleResource() resource.Resource { return &DowntimeScheduleResource{} }

type DowntimeScheduleResource struct{ client *client.Client }

type downtimeScheduleModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Schedule    types.String `tfsdk:"schedule"`
	AlertIDs    types.List   `tfsdk:"alert_ids"`
	Kind        types.String `tfsdk:"kind"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	CreatedBy   types.String `tfsdk:"created_by"`
	UpdatedBy   types.String `tfsdk:"updated_by"`
}

func downtimeScheduleAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id": types.StringType, "name": types.StringType, "description": types.StringType,
		"schedule": types.StringType, "alert_ids": types.ListType{ElemType: types.StringType},
		"kind": types.StringType, "status": types.StringType,
		"created_at": types.StringType, "updated_at": types.StringType,
		"created_by": types.StringType, "updated_by": types.StringType,
	}
}

func modelFromDowntimeScheduleMap(ctx context.Context, m map[string]interface{}) (downtimeScheduleModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	schedule, err := client.SubMapJSON(m, "schedule")
	if err != nil {
		diags.AddError("Internal error", err.Error())
		return downtimeScheduleModel{}, diags
	}
	alerts, d := typesStringListFromMap(ctx, m, "alertIds")
	diags.Append(d...)
	return downtimeScheduleModel{
		ID:          typesStringValueFromMap(m, "id"),
		Name:        typesStringValueFromMap(m, "name"),
		Description: typesStringValueFromMap(m, "description"),
		Schedule:    types.StringValue(schedule),
		AlertIDs:    alerts,
		Kind:        typesStringValueFromMap(m, "kind"),
		Status:      typesStringValueFromMap(m, "status"),
		CreatedAt:   typesStringValueFromMap(m, "createdAt"),
		UpdatedAt:   typesStringValueFromMap(m, "updatedAt"),
		CreatedBy:   typesStringValueFromMap(m, "createdBy"),
		UpdatedBy:   typesStringValueFromMap(m, "updatedBy"),
	}, diags
}

func (r *DowntimeScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_downtime_schedule"
}

func (r *DowntimeScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docDowntimeIntro,
		Attributes: map[string]schema.Attribute{
			"name":        schema.StringAttribute{MarkdownDescription: docDowntimeName, Required: true},
			"description": schema.StringAttribute{MarkdownDescription: docDowntimeDescription, Optional: true},
			"schedule": schema.StringAttribute{
				MarkdownDescription: docDowntimeSchedule,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					canonicalJSONPlanModifier{},
				},
			},
			"alert_ids": schema.ListAttribute{
				MarkdownDescription: docDowntimeAlertIDs,
				Optional:            true,
				ElementType:         types.StringType,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: docID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"kind":       schema.StringAttribute{MarkdownDescription: docDowntimeKind, Computed: true},
			"status":     schema.StringAttribute{MarkdownDescription: docDowntimeStatus, Computed: true},
			"created_at": schema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_at": schema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
			"created_by": schema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"updated_by": schema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
		},
	}
}

func (r *DowntimeScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureResource(context.Background(), req, resp)
}

func downtimeBodyFromModel(ctx context.Context, m downtimeScheduleModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	alerts, d := stringListFromTerraform(m.AlertIDs)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body, err := client.BuildDowntimeScheduleBody(
		m.Name.ValueString(),
		m.Description.ValueString(),
		m.Schedule.ValueString(),
		alerts,
	)
	if err != nil {
		diags.AddError("Invalid configuration", err.Error())
		return nil, diags
	}
	return body, diags
}

func (r *DowntimeScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan downtimeScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, diags := downtimeBodyFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.CreateDowntimeSchedule(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, diags := modelFromDowntimeScheduleMap(ctx, created)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DowntimeScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state downtimeScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	m, err := r.client.GetDowntimeSchedule(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, diags := modelFromDowntimeScheduleMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *DowntimeScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state downtimeScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	body, diags := downtimeBodyFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.UpdateDowntimeSchedule(ctx, state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	m, err := r.client.GetDowntimeSchedule(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	out, diags := modelFromDowntimeScheduleMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *DowntimeScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state downtimeScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteDowntimeSchedule(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *DowntimeScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// --- data source (singular) ---

var _ datasource.DataSource = &DowntimeScheduleDataSource{}

func NewDowntimeScheduleDataSource() datasource.DataSource { return &DowntimeScheduleDataSource{} }

type DowntimeScheduleDataSource struct{ client *client.Client }

func (d *DowntimeScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_downtime_schedule"
}

func (d *DowntimeScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docDowntimeDataSourceIntro,
		Attributes: map[string]dataschema.Attribute{
			"id": dataschema.StringAttribute{
				MarkdownDescription: docLookupID,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"name":        dataschema.StringAttribute{MarkdownDescription: docLookupName, Optional: true},
			"description": dataschema.StringAttribute{MarkdownDescription: docDowntimeDescription, Computed: true},
			"schedule":    dataschema.StringAttribute{MarkdownDescription: docDowntimeSchedule, Computed: true},
			"alert_ids":   dataschema.ListAttribute{MarkdownDescription: docDowntimeAlertIDs, ElementType: types.StringType, Computed: true},
			"kind":        dataschema.StringAttribute{MarkdownDescription: docDowntimeKind, Computed: true},
			"status":      dataschema.StringAttribute{MarkdownDescription: docDowntimeStatus, Computed: true},
			"created_at":  dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
			"updated_at":  dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
			"created_by":  dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
			"updated_by":  dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
		},
	}
}

func (d *DowntimeScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func readDowntimeByName(ctx context.Context, c *client.Client, name string) (map[string]interface{}, error) {
	list, err := c.ListDowntimeSchedules(ctx)
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
		return nil, fmt.Errorf("no downtime schedule named %q", name)
	case 1:
		return c.GetDowntimeSchedule(ctx, client.MapString(matches[0], "id"))
	default:
		return nil, fmt.Errorf("multiple downtime schedules named %q; use id", name)
	}
}

func (d *DowntimeScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config downtimeScheduleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || d.client == nil {
		return
	}
	var m map[string]interface{}
	var err error
	if !config.ID.IsNull() && config.ID.ValueString() != "" {
		m, err = d.client.GetDowntimeSchedule(ctx, config.ID.ValueString())
	} else {
		m, err = readDowntimeByName(ctx, d.client, config.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	state, diags := modelFromDowntimeScheduleMap(ctx, m)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// --- data source (plural) ---

var _ datasource.DataSource = &DowntimeSchedulesDataSource{}

func NewDowntimeSchedulesDataSource() datasource.DataSource { return &DowntimeSchedulesDataSource{} }

type DowntimeSchedulesDataSource struct{ client *client.Client }

type downtimeSchedulesDataSourceModel struct {
	Schedules types.List `tfsdk:"schedules"`
}

func (d *DowntimeSchedulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_downtime_schedules"
}

func (d *DowntimeSchedulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataschema.Schema{
		MarkdownDescription: docListAllIntro + " (`GET /api/v1/downtime_schedules`).",
		Attributes: map[string]dataschema.Attribute{
			"schedules": dataschema.ListNestedAttribute{
				MarkdownDescription: "Downtime schedules in the organization. Read-only.",
				Computed:            true,
				NestedObject: dataschema.NestedAttributeObject{
					Attributes: downtimeScheduleNestedDataSourceAttrs(),
				},
			},
		},
	}
}

func (d *DowntimeSchedulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureDataSource(context.Background(), req, resp)
}

func (d *DowntimeSchedulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		return
	}
	list, err := d.client.ListDowntimeSchedules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}
	elems := make([]attr.Value, 0, len(list))
	for _, m := range list {
		model, diags := modelFromDowntimeScheduleMap(ctx, m)
		resp.Diagnostics.Append(diags...)
		obj, diags := types.ObjectValueFrom(ctx, downtimeScheduleAttrTypes(), model)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}
	schedules, diags := types.ListValue(types.ObjectType{AttrTypes: downtimeScheduleAttrTypes()}, elems)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &downtimeSchedulesDataSourceModel{Schedules: schedules})...)
}
