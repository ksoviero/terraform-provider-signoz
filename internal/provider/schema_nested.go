// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Nested list element schemas for plural data sources (documented attributes only).

func alertRuleNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":          dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"alert":       dataschema.StringAttribute{MarkdownDescription: docAlert, Computed: true},
		"alert_type":  dataschema.StringAttribute{MarkdownDescription: docAlertType, Computed: true},
		"rule_type":   dataschema.StringAttribute{MarkdownDescription: docRuleType, Computed: true},
		"description": dataschema.StringAttribute{MarkdownDescription: docDescription, Computed: true},
		"disabled":    dataschema.BoolAttribute{MarkdownDescription: docDisabled, Computed: true},
		"labels":      dataschema.MapAttribute{MarkdownDescription: docLabels, ElementType: types.StringType, Computed: true},
		"annotations": dataschema.MapAttribute{MarkdownDescription: docAnnotations, ElementType: types.StringType, Computed: true},
		"spec":        dataschema.StringAttribute{MarkdownDescription: docSpec, Computed: true},
		"state":       dataschema.StringAttribute{MarkdownDescription: docRuleState, Computed: true},
		"created_at":  dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at":  dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
		"created_by":  dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
		"updated_by":  dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
	}
}

func notificationChannelNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":         dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"name":       dataschema.StringAttribute{MarkdownDescription: docChannelName, Computed: true},
		"config":     dataschema.StringAttribute{MarkdownDescription: docChannelConfig, CustomType: jsontypes.NormalizedType{}, Computed: true, Sensitive: true},
		"type":       dataschema.StringAttribute{MarkdownDescription: docChannelType, Computed: true},
		"data":       dataschema.StringAttribute{MarkdownDescription: docChannelData, CustomType: jsontypes.NormalizedType{}, Computed: true, Sensitive: true},
		"created_at": dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at": dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
	}
}

func dashboardNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":         dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"data":       dataschema.StringAttribute{MarkdownDescription: docDashboardData, Computed: true},
		"locked":     dataschema.BoolAttribute{MarkdownDescription: docDashboardLocked, Computed: true},
		"source":     dataschema.StringAttribute{MarkdownDescription: docDashboardSource, Computed: true},
		"created_at": dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at": dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
		"created_by": dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
		"updated_by": dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
	}
}

func downtimeScheduleNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":          dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"name":        dataschema.StringAttribute{MarkdownDescription: docDowntimeName, Computed: true},
		"description": dataschema.StringAttribute{MarkdownDescription: docDowntimeDescription, Computed: true},
		"schedule":    dataschema.StringAttribute{MarkdownDescription: docDowntimeSchedule, Computed: true},
		"alert_ids":   dataschema.ListAttribute{MarkdownDescription: docDowntimeAlertIDs, ElementType: types.StringType, Computed: true},
		"kind":        dataschema.StringAttribute{MarkdownDescription: docDowntimeKind, Computed: true},
		"status":      dataschema.StringAttribute{MarkdownDescription: docDowntimeStatus, Computed: true},
		"created_at":  dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at":  dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
		"created_by":  dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
		"updated_by":  dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
	}
}

func routePolicyNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":          dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"name":        dataschema.StringAttribute{MarkdownDescription: docRoutePolicyName, Computed: true},
		"description": dataschema.StringAttribute{MarkdownDescription: docRoutePolicyDescription, Computed: true},
		"expression":  dataschema.StringAttribute{MarkdownDescription: docRoutePolicyExpression, Computed: true},
		"kind":        dataschema.StringAttribute{MarkdownDescription: docRoutePolicyKind, Computed: true},
		"channels":    dataschema.ListAttribute{MarkdownDescription: docRoutePolicyChannels, ElementType: types.StringType, Computed: true},
		"tags":        dataschema.ListAttribute{MarkdownDescription: docRoutePolicyTags, ElementType: types.StringType, Computed: true},
		"created_at":  dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at":  dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
		"created_by":  dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
		"updated_by":  dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
	}
}

func authDomainNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":         dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"name":       dataschema.StringAttribute{MarkdownDescription: docAuthDomainName, Computed: true},
		"config":     dataschema.StringAttribute{MarkdownDescription: docAuthDomainConfig, Computed: true, Sensitive: true},
		"org_id":     dataschema.StringAttribute{MarkdownDescription: docOrgID, Computed: true},
		"created_at": dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at": dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
	}
}

func roleNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":          dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"name":        dataschema.StringAttribute{MarkdownDescription: docRoleName, Computed: true},
		"description": dataschema.StringAttribute{MarkdownDescription: docRoleDescription, Computed: true},
		"type":        dataschema.StringAttribute{MarkdownDescription: docRoleType, Computed: true},
		"org_id":      dataschema.StringAttribute{MarkdownDescription: docOrgID, Computed: true},
		"created_at":  dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at":  dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
	}
}

func serviceAccountNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":         dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"name":       dataschema.StringAttribute{MarkdownDescription: docServiceAccountName, Computed: true},
		"email":      dataschema.StringAttribute{MarkdownDescription: docServiceAccountEmail, Computed: true},
		"org_id":     dataschema.StringAttribute{MarkdownDescription: docOrgID, Computed: true},
		"created_at": dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at": dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
	}
}

func savedViewNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":              dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"name":            dataschema.StringAttribute{MarkdownDescription: docSavedViewName, Computed: true},
		"source_page":     dataschema.StringAttribute{MarkdownDescription: docSavedViewSourcePage, Computed: true},
		"category":        dataschema.StringAttribute{MarkdownDescription: docSavedViewCategory, Computed: true},
		"tags":            dataschema.ListAttribute{MarkdownDescription: docSavedViewTags, ElementType: types.StringType, Computed: true},
		"composite_query": dataschema.StringAttribute{MarkdownDescription: docSavedViewCompositeQuery, Computed: true},
		"extra_data":      dataschema.StringAttribute{MarkdownDescription: docSavedViewExtraData, Computed: true},
		"created_at":      dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at":      dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
		"created_by":      dataschema.StringAttribute{MarkdownDescription: docCreatedBy, Computed: true},
		"updated_by":      dataschema.StringAttribute{MarkdownDescription: docUpdatedBy, Computed: true},
	}
}

func cloudAccountNestedDataSourceAttrs() map[string]dataschema.Attribute {
	return map[string]dataschema.Attribute{
		"id":                  dataschema.StringAttribute{MarkdownDescription: docID, Computed: true},
		"cloud_provider":      dataschema.StringAttribute{MarkdownDescription: docCloudProvider, Computed: true},
		"config":              dataschema.StringAttribute{MarkdownDescription: docCloudConfig, Computed: true},
		"account_provider":    dataschema.StringAttribute{MarkdownDescription: docCloudAccountProvider, Computed: true},
		"provider_account_id": dataschema.StringAttribute{MarkdownDescription: docCloudProviderAccountID, Computed: true},
		"org_id":              dataschema.StringAttribute{MarkdownDescription: docOrgID, Computed: true},
		"created_at":          dataschema.StringAttribute{MarkdownDescription: docCreatedAt, Computed: true},
		"updated_at":          dataschema.StringAttribute{MarkdownDescription: docUpdatedAt, Computed: true},
	}
}
