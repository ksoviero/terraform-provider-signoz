// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settleModel struct {
	Name      types.String `tfsdk:"name"`
	Labels    types.Map    `tfsdk:"labels"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func TestSettled(t *testing.T) {
	t.Parallel()

	emptyMap, _ := types.MapValue(types.StringType, nil)
	planned := settleModel{Name: types.StringValue(""), Labels: emptyMap, UpdatedAt: types.StringUnknown()}
	read := settleModel{Name: types.StringNull(), Labels: types.MapNull(types.StringType), UpdatedAt: types.StringValue("now")}
	var diags diag.Diagnostics
	got := settled(planned, read, &diags)
	if !got.Name.Equal(planned.Name) || !got.Labels.Equal(planned.Labels) {
		t.Errorf("planned values weren't kept: %+v", got)
	}
	if !got.UpdatedAt.Equal(read.UpdatedAt) {
		t.Errorf("an unknown planned value wasn't filled from the API: %+v", got)
	}
	if diags.WarningsCount() != 0 {
		t.Errorf("empty values the API left out were warned about: %v", diags)
	}
	// A value stored differently is kept as planned, with a warning naming it
	nm := types.MapNull(types.StringType)
	changed := settled(settleModel{Name: types.StringValue("a"), Labels: nm, UpdatedAt: types.StringNull()},
		settleModel{Name: types.StringValue("A"), Labels: nm, UpdatedAt: types.StringNull()}, &diags)
	if !changed.Name.Equal(types.StringValue("a")) || diags.WarningsCount() != 1 || !strings.Contains(diags[0].Detail(), "name") {
		t.Errorf("got %+v, %v", changed, diags)
	}
	// A model that isn't a struct is an error, not a panic
	var bad diag.Diagnostics
	settled(&planned, &read, &bad)
	if !bad.HasError() {
		t.Error("a pointer model wasn't an error")
	}
}

func TestRefreshed(t *testing.T) {
	t.Parallel()
	var diags diag.Diagnostics

	emptyMap, _ := types.MapValue(types.StringType, nil)
	prior := settleModel{Name: types.StringValue(""), Labels: emptyMap, UpdatedAt: types.StringValue("then")}
	read := settleModel{Name: types.StringNull(), Labels: types.MapNull(types.StringType), UpdatedAt: types.StringValue("now")}
	got := refreshed(prior, read, false, &diags)
	if !got.Name.Equal(prior.Name) || !got.Labels.Equal(prior.Labels) {
		t.Errorf("an empty value the API left out showed as drift: %+v", got)
	}
	if !got.UpdatedAt.Equal(read.UpdatedAt) {
		t.Errorf("a real change wasn't read: %+v", got)
	}
	changed := refreshed(settleModel{Name: types.StringValue("a")}, settleModel{Name: types.StringNull()}, false, &diags)
	if !changed.Name.IsNull() {
		t.Errorf("a value the API dropped didn't show as drift: %+v", changed)
	}
	// After an import, the API's false isn't kept as state's null, and its "" is null
	type importModel struct {
		Name     types.String
		Disabled types.Bool
	}
	imported := refreshed(importModel{Name: types.StringNull(), Disabled: types.BoolNull()},
		importModel{Name: types.StringValue(""), Disabled: types.BoolValue(false)}, true, &diags)
	if !imported.Name.IsNull() || !imported.Disabled.Equal(types.BoolValue(false)) {
		t.Errorf("an import's values: %+v", imported)
	}
}

func TestLockedNotSupported(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		value   types.Bool
		wantErr bool
	}{{types.BoolValue(true), true}, {types.BoolValue(false), false}, {types.BoolNull(), false}} {
		resp := &validator.BoolResponse{}
		lockedNotSupported{}.ValidateBool(context.Background(), validator.BoolRequest{ConfigValue: tc.value}, resp)
		if resp.Diagnostics.HasError() != tc.wantErr {
			t.Errorf("locked = %s: error %v", tc.value, resp.Diagnostics)
		}
	}
}
