# terraform-provider-signoz

Terraform provider for [SigNoz](https://signoz.io/) using the HTTP API documented in [signoz/docs/api/openapi.yml](https://github.com/SigNoz/signoz/blob/main/docs/api/openapi.yml). Built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) (protocol 6.0).

See `PLAN.md` for roadmap and `docs/DASHBOARD_API.md` for dashboard endpoints not yet in OpenAPI.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.25

## Resources

| Resource | API |
|----------|-----|
| `signoz_notification_channel` | `/api/v1/channels` |
| `signoz_alert_rule` | `/api/v2/rules` |
| `signoz_dashboard` | `/api/v1/dashboards` |
| `signoz_downtime_schedule` | `/api/v1/downtime_schedules` |
| `signoz_route_policy` | `/api/v1/route_policies` |
| `signoz_auth_domain` | `/api/v1/domains` |
| `signoz_role` | `/api/v1/roles` |
| `signoz_service_account` | `/api/v1/service_accounts` |
| `signoz_saved_view` | `/api/v1/explorer/views` |
| `signoz_cloud_integration_account` | `/api/v1/cloud_integrations/{cloud_provider}/accounts` |

## Data sources

| Data source | API | Lookup |
|-------------|-----|--------|
| `signoz_alert_rule` | `/api/v2/rules` | `id` or unique `alert` title |
| `signoz_alert_rules` | `/api/v2/rules` | Lists all rules |
| `signoz_notification_channel` | `/api/v1/channels` | `id` or unique `name` |
| `signoz_notification_channels` | `/api/v1/channels` | Lists all channels |
| `signoz_dashboard` | `/api/v1/dashboards` | `id` |
| `signoz_dashboards` | `/api/v1/dashboards` | Lists all dashboards |
| `signoz_downtime_schedule` | `/api/v1/downtime_schedules` | `id` or unique `name` |
| `signoz_downtime_schedules` | `/api/v1/downtime_schedules` | Lists all |
| `signoz_route_policy` | `/api/v1/route_policies` | `id` or unique `name` |
| `signoz_route_policies` | `/api/v1/route_policies` | Lists all |
| `signoz_auth_domain` | `/api/v1/domains` | `id` or unique `name` |
| `signoz_auth_domains` | `/api/v1/domains` | Lists all |
| `signoz_role` | `/api/v1/roles` | `id` or unique `name` |
| `signoz_roles` | `/api/v1/roles` | Lists all |
| `signoz_service_account` | `/api/v1/service_accounts` | `id` or unique `name` |
| `signoz_service_accounts` | `/api/v1/service_accounts` | Lists all |
| `signoz_saved_view` | `/api/v1/explorer/views` | `id` or `name` (+ optional filters) |
| `signoz_saved_views` | `/api/v1/explorer/views` | Lists with optional query filters |
| `signoz_cloud_integration_account` | cloud accounts | `cloud_provider` + `id` |
| `signoz_cloud_integration_accounts` | cloud accounts | `cloud_provider` required |

Field defaults and enumerated values: [docs/guides/field_reference.md](docs/guides/field_reference.md).

Authentication uses the `SigNoz-Api-Key` header. API keys need at least **VIEWER** for reads; **ADMIN** for channels, **EDITOR** for rules and dashboards (create/update/delete).

## Build

```bash
go build -o terraform-provider-signoz
```

Or:

```bash
make install
```

## Provider configuration

```hcl
terraform {
  required_providers {
    signoz = {
      source  = "ksoviero/signoz"
      version = "~> 0.1"
    }
  }
}

provider "signoz" {
  endpoint = "https://signoz.example.com"
  api_key  = var.signoz_api_key
}
```

Environment variables (used when attributes are omitted):

- `SIGNOZ_ENDPOINT`
- `SIGNOZ_API_KEY`

## Local development (private provider)

Provider address: `registry.terraform.io/ksoviero/signoz`. Use **dev_overrides** in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/ksoviero/signoz" = "/path/to/go/bin"
  }
  direct {}
}
```

Build with `go install` so `terraform-provider-signoz` is in `$(go env GOPATH)/bin`.

## Documentation

```bash
make generate
```

Preview Registry rendering: [Terraform Registry Doc Preview](https://registry.terraform.io/tools/doc-preview).

## Tests

```bash
make test
```

Acceptance tests against a live SigNoz instance (optional):

```bash
export TF_ACC=1
export SIGNOZ_ENDPOINT="https://signoz.example.com"
export SIGNOZ_API_KEY="..."
make testacc
```

## Publishing to the Terraform Registry

Follow [HashiCorp’s publishing guide](https://developer.hashicorp.com/terraform/registry/providers/publishing).

1. **Repository:** public GitHub repo named `terraform-provider-signoz` (lowercase).
2. **Manifest:** `terraform-registry-manifest.json` with `"protocol_versions": ["6.0"]` (included).
3. **Docs:** `docs/index.md` and `docs/resources/*.md` from `make generate`.
4. **GPG key:** RSA or DSA signing key; add the public key under Registry **Signing Keys**.
5. **GitHub secrets:** `GPG_PRIVATE_KEY`, `PASSPHRASE`, and `GPG_FINGERPRINT` for GoReleaser (see `.github/workflows/release.yml`).
6. **Release:** merge to `main` with a commit subject prefixed by `feat:`, `patch:`, `bug:`, `major:`, etc. (see [AGENTS.md](AGENTS.md)). CI tags the release, publishes a [GitHub Release](https://github.com/ksoviero/terraform-provider-signoz/releases) with notes, and GoReleaser attaches signed provider binaries. You can still push a `v*` tag manually if needed.
7. **Registry:** sign in with GitHub → **Publish** → **Provider** → `ksoviero/terraform-provider-signoz`.

Do not replace assets on an already-published release version; publish a new tag instead.

Tutorial: [Release and publish a provider](https://developer.hashicorp.com/terraform/tutorials/providers/provider-release-publish).

## License

MPL-2.0
