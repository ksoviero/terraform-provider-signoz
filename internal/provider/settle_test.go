// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settleModel struct {
	Name      types.String
	Labels    types.Map
	UpdatedAt types.String
}

func TestSettled(t *testing.T) {
	t.Parallel()

	emptyMap, _ := types.MapValue(types.StringType, nil)
	planned := settleModel{Name: types.StringValue(""), Labels: emptyMap, UpdatedAt: types.StringUnknown()}
	read := settleModel{Name: types.StringNull(), Labels: types.MapNull(types.StringType), UpdatedAt: types.StringValue("now")}
	got := settled(planned, read)
	if !got.Name.Equal(planned.Name) || !got.Labels.Equal(planned.Labels) {
		t.Errorf("planned values weren't kept: %+v", got)
	}
	if !got.UpdatedAt.Equal(read.UpdatedAt) {
		t.Errorf("an unknown planned value wasn't filled from the API: %+v", got)
	}
}

func TestRefreshed(t *testing.T) {
	t.Parallel()

	emptyMap, _ := types.MapValue(types.StringType, nil)
	prior := settleModel{Name: types.StringValue(""), Labels: emptyMap, UpdatedAt: types.StringValue("then")}
	read := settleModel{Name: types.StringNull(), Labels: types.MapNull(types.StringType), UpdatedAt: types.StringValue("now")}
	got := refreshed(prior, read)
	if !got.Name.Equal(prior.Name) || !got.Labels.Equal(prior.Labels) {
		t.Errorf("an empty value the API left out showed as drift: %+v", got)
	}
	if !got.UpdatedAt.Equal(read.UpdatedAt) {
		t.Errorf("a real change wasn't read: %+v", got)
	}
	changed := refreshed(settleModel{Name: types.StringValue("a")}, settleModel{Name: types.StringNull()})
	if !changed.Name.IsNull() {
		t.Errorf("a value the API dropped didn't show as drift: %+v", changed)
	}
}
