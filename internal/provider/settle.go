// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// settled is the state a resource's Create or Update sets: for each attribute, the planned value
// wherever the plan knew it, and the value read back from the API (read) wherever the plan had it
// unknown ("known after apply"). That's exactly what Terraform checks after an apply, so an API
// that returns a value in another form (an empty map left out, "" as null, a computed field it
// rewrites) can't fail it with "Provider produced inconsistent result after apply" (issue #17).
//
// Where the API stored a planned value differently, other than an empty one left out, diags gets
// a warning naming the attribute (not its values, which may be sensitive): the next plan shows the difference, and the warning says where
// it came from. JSON attributes are left to their own comparison (shapedForState).
//
// planned and read are the same resource model: a struct whose fields are attr.Values. Anything
// else is returned as read, with an error in diags.
func settled[T any](planned, read T, diags *diag.Diagnostics) T {
	out := read
	pv, ov, ok := modelValues(&planned, &out, diags)
	if !ok {
		return read
	}
	for i := 0; i < pv.NumField(); i++ {
		p, ok1 := fieldValue(pv, i)
		r, ok2 := fieldValue(ov, i)
		if !ok1 || !ok2 || p.IsUnknown() {
			continue
		}
		if !r.IsUnknown() && !p.Equal(r) && !isJSONValue(p) && !emptyAndNull(p, r) {
			diags.AddWarning(
				"SigNoz stored a value differently",
				// Without the values: an attribute may be sensitive
				fmt.Sprintf("SigNoz stored %s differently from the plan. The next plan will show the difference.",
					attributeName(pv.Type().Field(i))),
			)
		}
		ov.Field(i).Set(pv.Field(i))
	}
	return out
}

// refreshed is the state a resource's Read sets: the values read from the API, except that an
// attribute the API has as null and state as empty ("", 0, false, [] or {}), or the other way
// round, keeps state's value. The API often leaves out an empty value, and they mean the same,
// so neither shows as drift. JSON attributes are compared separately (shapedForState). After an
// import (imported) state has nothing to keep yet, so it's all the API's.
func refreshed[T any](prior, read T, imported bool, diags *diag.Diagnostics) T {
	if imported {
		return read
	}
	out := read
	pv, ov, ok := modelValues(&prior, &out, diags)
	if !ok {
		return read
	}
	for i := 0; i < pv.NumField(); i++ {
		p, ok1 := fieldValue(pv, i)
		r, ok2 := fieldValue(ov, i)
		if !ok1 || !ok2 || p.IsUnknown() || r.IsUnknown() {
			continue
		}
		if emptyAndNull(p, r) {
			ov.Field(i).Set(pv.Field(i))
		}
	}
	return out
}

// modelValues returns the structs a and b point to, or an error in diags if they aren't structs.
func modelValues(a, b interface{}, diags *diag.Diagnostics) (reflect.Value, reflect.Value, bool) {
	av, bv := reflect.ValueOf(a).Elem(), reflect.ValueOf(b).Elem()
	if av.Kind() != reflect.Struct {
		diags.AddError("Internal error", fmt.Sprintf("a resource model must be a struct, not %s", av.Type()))
		return av, bv, false
	}
	return av, bv, true
}

// fieldValue returns field i of struct v as an attr.Value, if it's an exported one.
func fieldValue(v reflect.Value, i int) (attr.Value, bool) {
	if !v.Type().Field(i).IsExported() {
		return nil, false
	}
	a, ok := v.Field(i).Interface().(attr.Value)
	return a, ok
}

// attributeName is the Terraform name of a model's field (its tfsdk tag).
func attributeName(f reflect.StructField) string {
	if name, _, _ := strings.Cut(f.Tag.Get("tfsdk"), ","); name != "" {
		return name
	}
	return f.Name
}

// isJSONValue reports whether v is one of the JSON attribute types, which are compared on their
// fields (shapedForState) rather than as strings.
func isJSONValue(v attr.Value) bool {
	switch v.(type) {
	case jsontypes.Normalized, alertRuleSpecValue:
		return true
	}
	return false
}

// emptyAndNull reports whether one of a and b is null and the other empty.
func emptyAndNull(a, b attr.Value) bool {
	return (a.IsNull() && isEmptyValue(b)) || (b.IsNull() && isEmptyValue(a))
}

// isEmptyValue reports whether v is known, not null, and empty for its kind: "", false, 0, or a
// list, set, map, object or tuple with nothing in it. Dynamic values are never empty.
func isEmptyValue(v attr.Value) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	switch t := v.(type) {
	case basetypes.BoolValue:
		return !t.ValueBool()
	case basetypes.Int64Value:
		return t.ValueInt64() == 0
	case basetypes.Int32Value:
		return t.ValueInt32() == 0
	case basetypes.Float64Value:
		return t.ValueFloat64() == 0
	case basetypes.Float32Value:
		return t.ValueFloat32() == 0
	case basetypes.NumberValue:
		return t.ValueBigFloat().Cmp(new(big.Float)) == 0
	case basetypes.ListValue:
		return len(t.Elements()) == 0
	case basetypes.SetValue:
		return len(t.Elements()) == 0
	case basetypes.MapValue:
		return len(t.Elements()) == 0
	case basetypes.ObjectValue:
		return len(t.Attributes()) == 0
	case basetypes.TupleValue:
		return len(t.Elements()) == 0
	case basetypes.DynamicValue:
		return false
	}
	// Strings, including custom string types (JSON)
	if s, ok := v.(interface{ ValueString() string }); ok {
		return s.ValueString() == ""
	}
	return false
}
