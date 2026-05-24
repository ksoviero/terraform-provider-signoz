// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// normalizedJSONAttribute is required JSON with semantic equality (whitespace and key order ignored).
func normalizedJSONAttribute(markdown string, required bool) schema.StringAttribute {
	return normalizedJSONAttributeSensitive(markdown, required, false)
}

func normalizedJSONAttributeSensitive(markdown string, required bool, sensitive bool) schema.StringAttribute {
	attr := schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          jsontypes.NormalizedType{},
		Sensitive:           sensitive,
	}
	if required {
		attr.Required = true
	} else {
		attr.Optional = true
	}
	return attr
}

// normalizedJSONOptional is optional JSON (optionally sensitive) with semantic equality.
func normalizedJSONOptional(markdown string, sensitive bool, planModifiers ...planmodifier.String) schema.StringAttribute {
	attr := schema.StringAttribute{
		MarkdownDescription: markdown,
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
		Sensitive:           sensitive,
	}
	if len(planModifiers) > 0 {
		attr.PlanModifiers = planModifiers
	}
	return attr
}

// normalizedJSONOptionalWithUnknown copies prior state when credentials/config are not in config.
func normalizedJSONOptionalWithUnknown(markdown string, sensitive bool) schema.StringAttribute {
	return normalizedJSONOptional(markdown, sensitive, stringplanmodifier.UseStateForUnknown())
}

func newNormalizedJSON(value string) jsontypes.Normalized {
	return jsontypes.NewNormalizedValue(value)
}
