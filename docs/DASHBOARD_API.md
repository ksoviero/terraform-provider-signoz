# Dashboard API (undocumented in OpenAPI)

The Terraform `signoz_dashboard` resource uses the legacy HTTP routes registered in SigNoz query-service:

| Method | Path | Body | Response `data` |
|--------|------|------|-----------------|
| GET | `/api/v1/dashboards` | — | Array of dashboards |
| POST | `/api/v1/dashboards` | Dashboard JSON (`PostableDashboard`) | Created dashboard |
| GET | `/api/v1/dashboards/{id}` | — | `GettableDashboard` |
| PUT | `/api/v1/dashboards/{id}` | Dashboard JSON (`UpdatableDashboard`) | Updated dashboard |
| DELETE | `/api/v1/dashboards/{id}` | — | 204 No Content |

Implementation references:

- [`signoz/pkg/query-service/app/http_handler.go`](https://github.com/SigNoz/signoz/blob/main/pkg/query-service/app/http_handler.go) — route registration
- [`signoz/pkg/modules/dashboard/impldashboard/handler.go`](https://github.com/SigNoz/signoz/blob/main/pkg/modules/dashboard/impldashboard/handler.go) — handlers
- [`signoz/pkg/types/dashboardtypes/dashboard.go`](https://github.com/SigNoz/signoz/blob/main/pkg/types/dashboardtypes/dashboard.go) — types

All responses use the standard SigNoz envelope: `{ "status": "success", "data": ... }`.

The provider maps Terraform attribute `data` to the dashboard document (the inner layout JSON), not the full `GettableDashboard` wrapper.
