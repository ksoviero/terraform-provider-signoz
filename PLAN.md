# SigNoz Terraform Provider — Roadmap

## MVP (implemented)

- Provider: `endpoint`, `api_key`, `insecure`; env `SIGNOZ_ENDPOINT`, `SIGNOZ_API_KEY`
- `signoz_notification_channel` — `/api/v1/channels`
- `signoz_alert_rule` — `/api/v2/rules` (top-level fields + `spec` JSON)
- `signoz_dashboard` — `/api/v1/dashboards` (see `docs/DASHBOARD_API.md`)

## Next

- Route policies, downtime schedules, ingestion keys
- `signoz_public_dashboard` (`/api/v1/dashboards/{id}/public`)
- Richer `signoz_alert_rule` schema (structured `evaluation`, `notification_settings`)
- Contribute dashboard CRUD to SigNoz OpenAPI

## Registry

See README **Publishing** section and [HashiCorp publishing docs](https://developer.hashicorp.com/terraform/registry/providers/publishing).
