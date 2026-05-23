// SPDX-License-Identifier: MPL-2.0

package provider

// Markdown descriptions shared by resources and data sources (SigNoz OpenAPI enums).

const (
	docAlertRuleIntro = "Manages a SigNoz alert rule (`/api/v2/rules`). " +
		"Use `spec` for `condition`, `evaluation`, `notificationSettings`, and other fields not exposed as top-level attributes. " +
		"See the [SigNoz API spec](https://github.com/SigNoz/signoz/blob/main/docs/api/openapi.yml)."

	docAlert = "Human-readable alert rule title. Must be unique enough to identify the rule when using a data source lookup by `alert`."

	docAlertType = "Alert signal category. Valid values: `METRIC_BASED_ALERT`, `TRACES_BASED_ALERT`, `LOGS_BASED_ALERT`, `EXCEPTIONS_BASED_ALERT`."

	docRuleType = "Evaluation engine for the rule. Valid values: `threshold_rule`, `promql_rule`, `anomaly_rule` (anomaly rules may require UI-only fields; prefer `threshold_rule` or `promql_rule` for Terraform)."

	docDescription = "Optional longer description shown in the SigNoz UI. Default: omitted (empty)."

	docDisabled = "When `true`, the rule does not evaluate. Default: `false`."

	docLabels = "Optional key/value labels attached to the rule (Prometheus-style). Default: omitted (no labels)."

	docAnnotations = "Optional key/value annotations (e.g. `summary`, `description` for notifications). Default: omitted."

	docSpec = "JSON object with rule body fields: `condition`, `evaluation`, `notificationSettings`, `schemaVersion`, `version`, `source`, etc. " +
		"Common `spec` values: `schemaVersion` = `v2alpha1`, `version` = `v5`. " +
		"Inside `condition.thresholds.spec[]`, `op` may be `above`, `below`, `equal`, `not_equal`, `outside_bounds` (API also accepts legacy numeric codes from the UI). " +
		"`matchType`: `at_least_once`, `all_the_times`, `on_average`, `in_total`, `last`. " +
		"`evaluation.kind`: `rolling` or `cumulative`."

	docRuleState = "Computed evaluation state. Read-only values: `inactive`, `pending`, `recovering`, `firing`, `nodata`, `disabled`."

	docChannelIntro = "Manages a SigNoz notification channel (`/api/v1/channels`). " +
		"`config` is a JSON receiver object with exactly one `*_configs` array."

	docChannelName = "Channel name (unique per organization). Used for data source lookup when `id` is not set."

	docChannelConfig = "JSON receiver configuration. Include one of: `slack_configs`, `email_configs`, `webhook_configs`, `pagerduty_configs`, `opsgenie_configs`, `discord_configs`, `teams_configs`, `sns_configs`, `telegram_configs`, `pushover_configs`, `victorops_configs`, `wechat_configs`, `webex_configs`, `msteams_configs`, `msteamsv2_configs`, `jira_configs`, `rocketchat_configs`, `mattermost_configs`, `incidentio_configs`. " +
		"The top-level `name` field is set automatically from the resource `name` attribute on create/update."

	docChannelType = "Computed channel type derived from the receiver (e.g. `slack`, `email`, `webhook`)."

	docDashboardIntro = "Manages a SigNoz dashboard (`/api/v1/dashboards`). " +
		"`data` is the full dashboard JSON (layout, widgets, `title`, `version`, etc.). See `docs/DASHBOARD_API.md` in this provider repository."

	docDashboardData = "Dashboard JSON document. Typically includes `title`, `description`, `tags`, `layout`, `widgets`, and `version` (often `v5` for current SigNoz layouts)."

	docDashboardLocked = "Whether the dashboard is locked in the UI. Default: `false`."
)
