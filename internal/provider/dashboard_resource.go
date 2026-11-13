// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var (
	_ resource.Resource                = &DashboardResource{}
	_ resource.ResourceWithImportState = &DashboardResource{}
)

func NewDashboardResource() resource.Resource {
	return &DashboardResource{}
}

type DashboardResource struct {
	client *client.Client
}

type dashboardModel struct {
	ID        types.String         `tfsdk:"id"`
	Data      jsontypes.Normalized `tfsdk:"data"`
	Locked    types.Bool           `tfsdk:"locked"`
	Source    types.String         `tfsdk:"source"`
	CreatedAt types.String         `tfsdk:"created_at"`
	UpdatedAt types.String         `tfsdk:"updated_at"`
	CreatedBy types.String         `tfsdk:"created_by"`
	UpdatedBy types.String         `tfsdk:"updated_by"`
}

func (r *DashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboard"
}

func (r *DashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docDashboardIntro,
		Attributes: map[string]schema.Attribute{
			"data": normalizedJSONAttribute(docDashboardData, true),
			"locked": schema.BoolAttribute{
				MarkdownDescription: docDashboardLocked,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Validators:          []validator.Bool{lockedNotSupported{}},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: docID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// Not kept in the plan: nothing says an update leaves it as it is
			"source": schema.StringAttribute{
				MarkdownDescription: docDashboardSource,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: docCreatedAt,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: docUpdatedAt,
				Computed:            true,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: docCreatedBy,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_by": schema.StringAttribute{
				MarkdownDescription: docUpdatedBy,
				Computed:            true,
			},
		},
	}
}

func (r *DashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func modelFromDashboard(d *client.Dashboard) (dashboardModel, error) {
	dataJSON, err := client.MarshalDashboardData(d.Data)
	if err != nil {
		return dashboardModel{}, err
	}
	return dashboardModel{
		ID:        types.StringValue(d.ID),
		Data:      newNormalizedJSON(dataJSON),
		Locked:    types.BoolValue(d.Locked),
		Source:    types.StringValue(d.Source),
		CreatedAt: types.StringValue(d.CreatedAt),
		UpdatedAt: types.StringValue(d.UpdatedAt),
		CreatedBy: types.StringValue(d.CreatedBy),
		UpdatedBy: types.StringValue(d.UpdatedBy),
	}, nil
}

func (r *DashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}

	data, err := client.ParseDashboardData(plan.Data.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid data", err.Error())
		return
	}

	created, err := r.client.CreateDashboard(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	state, err := modelFromDashboard(created)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	state = settled(plan, state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	imported := readImported(ctx, req, resp)

	d, err := r.client.GetDashboard(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	out, err := modelFromDashboard(d)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	shapeReadJSON(state.Data, &out.Data, imported, &resp.Diagnostics)
	out = refreshed(state, out, imported, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *DashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}

	data, err := client.ParseDashboardData(plan.Data.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid data", err.Error())
		return
	}

	updated, err := r.client.UpdateDashboard(ctx, state.ID.ValueString(), data)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	out, err := modelFromDashboard(updated)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	out = settled(plan, out, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *DashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteDashboard(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *DashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	markImported(ctx, resp)
}

// lockedNotSupported rejects locked = true: the dashboard API this provider uses (v1) doesn't take
// it, so it was never sent, and the plan would show it on every run. SigNoz locks dashboards with
// its v2 API's /lock, which rejects updates while locked.
type lockedNotSupported struct{}

func (lockedNotSupported) Description(context.Context) string {
	return "locked = true isn't supported yet."
}

func (v lockedNotSupported) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (lockedNotSupported) ValidateBool(_ context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() && req.ConfigValue.ValueBool() {
		resp.Diagnostics.AddAttributeError(req.Path, "locked can't be set yet",
			"This provider can't lock a dashboard yet: the dashboard API it uses doesn't take locked, so it would never be sent. Lock the dashboard in SigNoz's UI, and leave locked out of the configuration.")
	}
}
