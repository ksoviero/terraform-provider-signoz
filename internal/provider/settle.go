// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"math/big"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// settled is the state a resource's Create or Update sets: for each attribute, the planned value
// wherever the plan knew it, and the value read back from the API (read) wherever the plan had it
// unknown ("known after apply"). That's exactly what Terraform checks after an apply, so an API
// that returns a value in another form (an empty map left out, "" as null, a computed field it
// rewrites) can't fail it with "Provider produced inconsistent result after apply". The next Read
// shows anything the API really changed.
//
// planned and read are the same resource model: a struct of attr.Value fields.
func settled[T any](planned, read T) T {
	out := read
	pv := reflect.ValueOf(&planned).Elem()
	ov := reflect.ValueOf(&out).Elem()
	for i := 0; i < pv.NumField(); i++ {
		v, ok := pv.Field(i).Interface().(attr.Value)
		if ok && !v.IsUnknown() {
			ov.Field(i).Set(pv.Field(i))
		}
	}
	return out
}

// refreshed is the state a resource's Read sets: the values read from the API, except that an
// attribute the API has as null and state as empty ("", 0, false, [] or {}), or the other way
// round, keeps state's value. The API often leaves out an empty value, and they mean the same,
// so neither shows as drift. JSON attributes are compared separately (shapedForState). After an
// import (imported) state has nothing to keep yet, so it's all the API's.
func refreshed[T any](prior, read T, imported bool) T {
	if imported {
		return read
	}
	out := read
	pv := reflect.ValueOf(&prior).Elem()
	ov := reflect.ValueOf(&out).Elem()
	for i := 0; i < pv.NumField(); i++ {
		p, ok1 := pv.Field(i).Interface().(attr.Value)
		r, ok2 := ov.Field(i).Interface().(attr.Value)
		if !ok1 || !ok2 || p.IsUnknown() || r.IsUnknown() {
			continue
		}
		if (r.IsNull() && isEmptyValue(p)) || (p.IsNull() && isEmptyValue(r)) {
			ov.Field(i).Set(pv.Field(i))
		}
	}
	return out
}

// isEmptyValue reports whether v is known, not null, and empty for its kind.
func isEmptyValue(v attr.Value) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	switch t := v.(type) {
	case basetypes.StringValuable:
		if s, ok := t.(interface{ ValueString() string }); ok {
			return s.ValueString() == ""
		}
	case basetypes.BoolValue:
		return !t.ValueBool()
	case basetypes.Int64Value:
		return t.ValueInt64() == 0
	case basetypes.Float64Value:
		return t.ValueFloat64() == 0
	case basetypes.NumberValue:
		return t.ValueBigFloat().Cmp(new(big.Float)) == 0
	case basetypes.ListValue:
		return len(t.Elements()) == 0
	case basetypes.SetValue:
		return len(t.Elements()) == 0
	case basetypes.MapValue:
		return len(t.Elements()) == 0
	}
	return false
}
