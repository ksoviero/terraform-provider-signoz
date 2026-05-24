// SPDX-License-Identifier: MPL-2.0

package provider

// Markdown descriptions shared by resources and data sources.
// See AGENTS.md — every attribute needs defaults, enum values, and read-only notes where applicable.

const (
	// Provider
	docProviderEndpoint = "Base URL of the SigNoz instance (scheme + host), e.g. `https://signoz.example.com`. Do not include `/api`. " +
		"Default: read from `SIGNOZ_ENDPOINT` when this attribute is omitted."
	docProviderAPIKey   = "API key sent as the `SigNoz-Api-Key` header. Default: read from `SIGNOZ_API_KEY` when this attribute is omitted. Sensitive."
	docProviderInsecure = "When `true`, TLS certificate verification is disabled. Default: `false`. Use only in lab environments."

	// Common read-only / lookup
	docID        = "SigNoz-assigned UUID. Read-only."
	docCreatedAt = "RFC3339 timestamp when the object was created. Read-only."
	docUpdatedAt = "RFC3339 timestamp when the object was last updated. Read-only."
	docCreatedBy = "Identifier of the user or service that created the object. Read-only."
	docUpdatedBy = "Identifier of the user or service that last updated the object. Read-only."
	docOrgID     = "Organization UUID. Read-only."

	docLookupID   = "Object UUID. Exactly one of `id` or `name` must be set for lookup."
	docLookupName = "Object name (unique within the organization for that resource type). Exactly one of `id` or `name` must be set for lookup."

	docListAllIntro = "Lists all objects returned by the list API endpoint for this resource type."

	// Alert rule
	docAlertRuleIntro = "Manages a SigNoz alert rule (`/api/v2/rules`). " +
		"Use `spec` for `condition`, `evaluation`, `notificationSettings`, and other fields not exposed as top-level attributes. " +
		"See the [SigNoz API spec](https://github.com/SigNoz/signoz/blob/main/docs/api/openapi.yml)."

	docAlert = "Human-readable alert rule title. Must be unique enough to identify the rule when using a data source lookup by `alert`."

	docAlertType = "Alert signal category. Valid values: `METRIC_BASED_ALERT`, `TRACES_BASED_ALERT`, `LOGS_BASED_ALERT`, `EXCEPTIONS_BASED_ALERT`."

	docRuleType = "Evaluation engine for the rule. Valid values: `threshold_rule`, `promql_rule`, `anomaly_rule` (anomaly rules may require UI-only fields; prefer `threshold_rule` or `promql_rule` for Terraform)."

	docDescription = "Optional longer description shown in the SigNoz UI. Default: omitted (empty string)."

	docDisabled = "When `true`, the rule does not evaluate. Default: `false`."

	docLabels = "Optional key/value labels attached to the rule (Prometheus-style). Default: omitted (no labels)."

	docAnnotations = "Optional key/value annotations (e.g. `summary`, `description` for notifications). Default: omitted (no annotations)."

	docSpec = "JSON object with rule body fields: `condition`, `evaluation`, `notificationSettings`, `schemaVersion`, `version`, `source`, etc. " +
		"Common `spec` values: `schemaVersion` = `v2alpha1`, `version` = `v5`. " +
		"Inside `condition.thresholds.spec[]`, `op` may be `above`, `below`, `equal`, `not_equal`, `outside_bounds` (API also accepts legacy numeric codes from the UI). " +
		"`matchType`: `at_least_once`, `all_the_times`, `on_average`, `in_total`, `last`. " +
		"`evaluation.kind`: `rolling` or `cumulative`."

	docRuleState = "Computed evaluation state. Read-only. Valid values: `inactive`, `pending`, `recovering`, `firing`, `nodata`, `disabled`."

	docAlertRuleDataSourceIntro = "Reads a single SigNoz alert rule by `id` or by unique `alert` title."

	// Notification channel
	docChannelIntro = "Manages a SigNoz notification channel (`/api/v1/channels`). " +
		"`config` is a JSON receiver object with exactly one `*_configs` array. " +
		"Set the channel display name with the resource `name` attribute, not inside `config`."

	docChannelName = "Channel name (unique per organization). Used for data source lookup when `id` is not set."

	docChannelConfig = "JSON receiver configuration (`AlertmanagertypesPostableChannel`). Include exactly one of: `slack_configs`, `email_configs`, `webhook_configs`, `pagerduty_configs`, `opsgenie_configs`, `discord_configs`, `teams_configs`, `sns_configs`, `telegram_configs`, `pushover_configs`, `victorops_configs`, `wechat_configs`, `webex_configs`, `msteams_configs`, `msteamsv2_configs`, `jira_configs`, `rocketchat_configs`, `mattermost_configs`, `incidentio_configs`. " +
		"Do not include a top-level `name` key; the provider sets it from the resource `name` attribute on create/update and omits it from stored `config`. Sensitive."

	docChannelType = "Computed channel type derived from the receiver configuration (e.g. `slack`, `email`, `webhook`). Read-only."

	docChannelData = "Canonical JSON of the stored receiver object returned by the API. Read-only. Sensitive."

	docChannelDataSourceIntro = "Reads a single SigNoz notification channel by `id` or by unique `name`."

	// Dashboard
	docDashboardIntro = "Manages a SigNoz dashboard (`/api/v1/dashboards`). " +
		"`data` is the full dashboard JSON (layout, widgets, `title`, `version`, etc.). See `docs/DASHBOARD_API.md` in this provider repository."

	docDashboardData = "Dashboard JSON document (`PostableDashboard` / `UpdatableDashboard`). Typically includes `title`, `description`, `tags`, `layout`, `widgets`, and `version` (often `v5` for current SigNoz layouts)."

	docDashboardLocked = "Whether the dashboard is locked in the UI. Default: `false`."

	docDashboardSource = "Dashboard source identifier from the API (e.g. how the dashboard was created). Read-only."

	docDashboardDataSourceIntro = "Reads a single SigNoz dashboard by `id`."

	// Downtime schedule
	docDowntimeIntro = "Manages a SigNoz planned maintenance / downtime schedule (`/api/v1/downtime_schedules`)."

	docDowntimeName = "Schedule name (unique per organization). Used for data source lookup when `id` is not set."

	docDowntimeDescription = "Optional human-readable description. Default: omitted (empty string)."

	docDowntimeSchedule = "JSON schedule object (`AlertmanagertypesSchedule`). Required key: `timezone` (IANA name, e.g. `UTC`, `America/Chicago`). " +
		"For a one-off window use `startTime` and `endTime` (RFC3339). For recurrence include `recurrence` with `startTime`, `duration`, and `repeatType` (`daily`, `weekly`, `monthly`). " +
		"Optional `repeatOn` for weekly schedules: `sunday`, `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday`."

	docDowntimeAlertIDs = "List of alert rule UUIDs affected by this schedule. Default: omitted (applies per SigNoz API semantics when unset)."

	docDowntimeKind = "Schedule kind from the API. Read-only. Valid values: `fixed`, `recurring`."

	docDowntimeStatus = "Current schedule status from the API. Read-only. Valid values: `active`, `upcoming`, `expired`."

	docDowntimeDataSourceIntro = "Reads a single downtime schedule by `id` or by unique `name`."

	// Route policy
	docRoutePolicyIntro = "Manages a SigNoz alert route policy (`/api/v1/route_policies`). Requires ADMIN API access."

	docRoutePolicyName = "Route policy name (unique per organization). Used for data source lookup when `id` is not set."

	docRoutePolicyDescription = "Optional description shown in the SigNoz UI. Default: omitted (empty string)."

	docRoutePolicyExpression = "Routing expression evaluated against alert labels (SigNoz route matcher syntax)."

	docRoutePolicyKind = "Expression kind. Valid values: `rule`, `policy`. Default: omitted (API default applies)."

	docRoutePolicyChannels = "List of notification channel **names** (not UUIDs) to notify when the expression matches."

	docRoutePolicyTags = "Optional string tags attached to the policy. Default: omitted (no tags)."

	docRoutePolicyDataSourceIntro = "Reads a single route policy by `id` or by unique `name`."

	// Auth domain
	docAuthDomainIntro = "Manages a SigNoz authentication domain (`/api/v1/domains`). Requires ADMIN API access."

	docAuthDomainName = "Auth domain name. Used for data source lookup when `id` is not set."

	docAuthDomainConfig = "JSON auth domain configuration (`AuthtypesAuthDomainConfig`): `ssoEnabled`, and one of `samlConfig`, `oidcConfig`, or `googleAuthConfig`, plus optional `roleMapping`. " +
		"Default: omitted (empty object sent when unset). Sensitive."

	docAuthDomainDataSourceIntro = "Reads a single auth domain by `id` or by unique `name`."

	// Role
	docRoleIntro = "Manages a SigNoz custom role (`/api/v1/roles`). Only `description` is mutable after create (PATCH). `name` changes force resource replacement."

	docRoleName = "Role name. Changing this value replaces the resource. Managed roles cannot be created via this resource."

	docRoleDescription = "Role description. Default: omitted (empty string). This is the only field updated on apply after create."

	docRoleType = "Role type from the API. Read-only. Valid values: `custom`, `managed`."

	docRoleDataSourceIntro = "Reads a single role by `id` or by unique `name`."

	// Service account
	docServiceAccountIntro = "Manages a SigNoz service account (`/api/v1/service_accounts`)."

	docServiceAccountName = "Service account display name. Used for data source lookup when `id` is not set."

	docServiceAccountEmail = "Email address assigned by SigNoz for the service account. Read-only."

	docServiceAccountDataSourceIntro = "Reads a single service account by `id` or by unique `name`."

	// Saved view
	docSavedViewIntro = "Manages a SigNoz explorer saved view (`/api/v1/explorer/views`). Not documented in SigNoz OpenAPI."

	docSavedViewName = "Saved view name. Used for data source lookup when `id` is not set (with optional `source_page` filter)."

	docSavedViewSourcePage = "Explorer page this view belongs to. Valid values: `logs`, `traces`, `metrics` (SigNoz explorer tabs)."

	docSavedViewCategory = "Optional category label in the UI. Default: omitted (empty string)."

	docSavedViewTags = "Optional list of tag strings. Default: omitted (no tags)."

	docSavedViewCompositeQuery = "JSON explorer query (`CompositeQuery` in query-service). Required. Must include `queryType` (typically `builder`) and `queries` with valid builder specs for the `source_page` signal."

	docSavedViewExtraData = "Optional opaque JSON string used by the SigNoz UI for extra state. Default: omitted (empty string)."

	docSavedViewDataSourceIntro = "Reads a single saved view by `id`, or by `name` with optional `source_page` and `category` query filters."

	docSavedViewListSourcePage = "Optional filter passed to the list API as `sourcePage`. Default: omitted (no filter)."

	docSavedViewListName = "Optional filter passed to the list API as `name`. Default: omitted (no filter)."

	docSavedViewListCategory = "Optional filter passed to the list API as `category`. Default: omitted (no filter)."

	docSavedViewListIntro = "Lists explorer saved views (`GET /api/v1/explorer/views`) with optional query filters."

	// Cloud integration account
	docCloudIntro = "Manages a SigNoz cloud integration account (`/api/v1/cloud_integrations/{cloud_provider}/accounts`). Requires ADMIN API access."

	docCloudProvider = "Cloud provider slug in the API path. Valid values: `aws`, `azure`. Changing this value replaces the resource."

	docCloudConfig = "JSON account configuration (`CloudintegrationtypesPostableAccountConfig`). For AWS include `aws.deploymentRegion` and `aws.regions`; for Azure include `azure` per OpenAPI."

	docCloudCredentials = "JSON credentials (`CloudintegrationtypesCredentials`) required on create: `sigNozApiUrl`, `sigNozApiKey`, `ingestionUrl`, `ingestionKey`. " +
		"Default: omitted on update (not sent; value retained in Terraform state only). Sensitive."

	docCloudAccountProvider = "Provider slug on the account object returned by the API (e.g. `aws`, `azure`). Read-only."

	docCloudProviderAccountID = "Cloud provider account identifier (e.g. AWS account ID). Read-only."

	docCloudDataSourceIntro = "Reads a single cloud integration account by `cloud_provider` and `id`."

	docCloudListIntro = "Lists cloud integration accounts for the given `cloud_provider`."
)
