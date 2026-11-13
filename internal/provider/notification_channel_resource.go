// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

var (
	_ resource.Resource                = &NotificationChannelResource{}
	_ resource.ResourceWithImportState = &NotificationChannelResource{}
)

func NewNotificationChannelResource() resource.Resource {
	return &NotificationChannelResource{}
}

type NotificationChannelResource struct {
	client *client.Client
}

type notificationChannelModel struct {
	ID        types.String         `tfsdk:"id"`
	Name      types.String         `tfsdk:"name"`
	Config    jsontypes.Normalized `tfsdk:"config"`
	Type      types.String         `tfsdk:"type"`
	Data      jsontypes.Normalized `tfsdk:"data"`
	CreatedAt types.String         `tfsdk:"created_at"`
	UpdatedAt types.String         `tfsdk:"updated_at"`
}

func (r *NotificationChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (r *NotificationChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: docChannelIntro,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: docChannelName,
				Required:            true,
			},
			"config": normalizedJSONAttributeSensitive(docChannelConfig, true, true),
			"id": schema.StringAttribute{
				MarkdownDescription: docID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: docChannelType,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					sameReceiverKindKeepsType{},
				},
			},
			"data": normalizedJSONComputedSensitive(docChannelData),
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
		},
	}
}

func (r *NotificationChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func modelFromChannel(ch *client.Channel) (notificationChannelModel, error) {
	configJSON, err := client.NormalizeChannelConfigJSON(ch.Data)
	if err != nil {
		return notificationChannelModel{}, err
	}
	dataCanonical, err := client.CanonicalJSON(ch.Data)
	if err != nil {
		return notificationChannelModel{}, err
	}
	return notificationChannelModel{
		ID:        types.StringValue(ch.ID),
		Name:      types.StringValue(ch.Name),
		Config:    newNormalizedJSON(configJSON),
		Type:      types.StringValue(ch.Type),
		Data:      newNormalizedJSON(dataCanonical),
		CreatedAt: types.StringValue(ch.CreatedAt),
		UpdatedAt: types.StringValue(ch.UpdatedAt),
	}, nil
}

// notificationChannelStateAfterWrite builds state from the API but keeps the planned
// config value unchanged. SigNoz may canonicalize or mutate receiver JSON on readback;
// re-normalizing plan config here breaks Terraform sensitive-attribute apply checks.
func notificationChannelStateAfterWrite(plan notificationChannelModel, ch *client.Channel) (notificationChannelModel, error) {
	out, err := modelFromChannel(ch)
	if err != nil {
		return notificationChannelModel{}, err
	}
	if !plan.Config.IsNull() && !plan.Config.IsUnknown() {
		out.Config = plan.Config
	}
	return out, nil
}

func (r *NotificationChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}

	body, err := client.MergeChannelBody(plan.Name.ValueString(), plan.Config.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid config", err.Error())
		return
	}

	ch, err := r.client.CreateChannel(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	state, err := notificationChannelStateAfterWrite(plan, ch)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	state = settled(plan, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	imported := readImported(ctx, req, resp)

	ch, err := r.client.GetChannel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	// The config in state compared with the API's raw receiver JSON on the fields it sets
	// only, so the Alertmanager defaults the API adds never show as drift, but a change to
	// a field the config sets does (shapedForState). On import, the pruned config.
	out, err := modelFromChannel(ch)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	raw, err := client.ChannelConfigFromData(ch.Data)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	shaped, err := shapedForState(state.Config.StringValue, raw, out.Config.ValueString(), imported, client.SameDuration)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	out.Config = jsontypes.Normalized{StringValue: shaped}
	out = refreshed(state, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *NotificationChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state notificationChannelModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}

	body, err := client.MergeChannelBody(plan.Name.ValueString(), plan.Config.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid config", err.Error())
		return
	}

	if err := r.client.UpdateChannel(ctx, state.ID.ValueString(), body); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	ch, err := r.client.GetChannel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
		return
	}

	out, err := notificationChannelStateAfterWrite(plan, ch)
	if err != nil {
		resp.Diagnostics.AddError("Internal error", err.Error())
		return
	}
	out = settled(plan, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func (r *NotificationChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationChannelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() || r.client == nil {
		return
	}
	if err := r.client.DeleteChannel(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("SigNoz API error", err.Error())
	}
}

func (r *NotificationChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	markImported(ctx, resp)
}

// sameReceiverKindKeepsType plans a channel's type as it is unless the config's receiver kind
// (its *_configs key) changes, when it's "known after apply" (issue #17): the API derives the
// type from the kind, so promising the old one for a new kind failed the apply.
type sameReceiverKindKeepsType struct{}

func (sameReceiverKindKeepsType) Description(context.Context) string {
	return "Keeps the type in the plan unless the receiver kind (the *_configs key) changes."
}

func (m sameReceiverKindKeepsType) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (sameReceiverKindKeepsType) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || !req.PlanValue.IsUnknown() {
		return
	}
	var planned, prior jsontypes.Normalized
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("config"), &planned)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("config"), &prior)...)
	if resp.Diagnostics.HasError() || planned.IsUnknown() || planned.IsNull() || prior.IsNull() {
		return
	}
	if receiverKinds(planned.ValueString()) == receiverKinds(prior.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

// receiverKinds lists a receiver config's *_configs keys, sorted, or "" if it isn't JSON.
func receiverKinds(config string) string {
	var m map[string]interface{}
	if json.Unmarshal([]byte(config), &m) != nil {
		return ""
	}
	var kinds []string
	for k := range m {
		if strings.HasSuffix(k, "_configs") {
			kinds = append(kinds, k)
		}
	}
	sort.Strings(kinds)
	return strings.Join(kinds, ",")
}
