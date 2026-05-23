# Examples

Copy a resource or data source example into your Terraform root module together with `examples/provider/provider.tf` (or your own `provider "signoz"` block).

## Resources (`resources/`)

| Directory | Resource |
|-----------|----------|
| `notification_channel/` | `signoz_notification_channel` |
| `alert_rule/` | `signoz_alert_rule` |
| `dashboard/` | `signoz_dashboard` |
| `downtime_schedule/` | `signoz_downtime_schedule` |
| `route_policy/` | `signoz_route_policy` |
| `auth_domain/` | `signoz_auth_domain` |
| `role/` | `signoz_role` |
| `service_account/` | `signoz_service_account` |
| `saved_view/` | `signoz_saved_view` |
| `cloud_integration_account/` | `signoz_cloud_integration_account` |

## Data sources (`data-sources/`)

Each resource has a singular data source (lookup by `id` or `name`, plus `source_page` for saved views) and a plural data source (list all). Cloud integration data sources require `cloud_provider`.

| Directory | Data source |
|-----------|-------------|
| `signoz_notification_channel/` | `signoz_notification_channel` |
| `signoz_notification_channels/` | `signoz_notification_channels` |
| `signoz_alert_rule/` | `signoz_alert_rule` |
| `signoz_alert_rules/` | `signoz_alert_rules` |
| `signoz_dashboard/` | `signoz_dashboard` |
| `signoz_dashboards/` | `signoz_dashboards` |
| `signoz_downtime_schedule/` | `signoz_downtime_schedule` |
| `signoz_downtime_schedules/` | `signoz_downtime_schedules` |
| `signoz_route_policy/` | `signoz_route_policy` |
| `signoz_route_policies/` | `signoz_route_policies` |
| `signoz_auth_domain/` | `signoz_auth_domain` |
| `signoz_auth_domains/` | `signoz_auth_domains` |
| `signoz_role/` | `signoz_role` |
| `signoz_roles/` | `signoz_roles` |
| `signoz_service_account/` | `signoz_service_account` |
| `signoz_service_accounts/` | `signoz_service_accounts` |
| `signoz_saved_view/` | `signoz_saved_view` |
| `signoz_saved_views/` | `signoz_saved_views` |
| `signoz_cloud_integration_account/` | `signoz_cloud_integration_account` |
| `signoz_cloud_integration_accounts/` | `signoz_cloud_integration_accounts` |

For local provider development, see the **Local development** section in the repository README.
