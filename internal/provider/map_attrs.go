// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ksoviero/terraform-provider-signoz/internal/client"
)

func typesStringValueFromMap(m map[string]interface{}, key string) types.String {
	return types.StringValue(client.MapString(m, key))
}

func typesStringListFromMap(ctx context.Context, m map[string]interface{}, key string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	slice := client.MapStringSlice(m, key)
	if len(slice) == 0 {
		return types.ListNull(types.StringType), diags
	}
	elems := make([]attr.Value, len(slice))
	for i, s := range slice {
		elems[i] = types.StringValue(s)
	}
	list, d := types.ListValue(types.StringType, elems)
	diags.Append(d...)
	return list, diags
}

func stringListFromTerraform(list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var out []string
	diags := list.ElementsAs(context.Background(), &out, false)
	return out, diags
}
