# terraform-provider-signoz — agent notes

Terraform Plugin Framework provider for the [SigNoz](https://signoz.io/) HTTP API. Registry address: `registry.terraform.io/ksoviero/signoz`.

## Before changing API behavior

Confirm request/response shapes against authoritative sources (do not guess field names):

- [SigNoz OpenAPI](https://github.com/SigNoz/signoz/blob/main/docs/api/openapi.yml) — most endpoints
- [`docs/DASHBOARD_API.md`](docs/DASHBOARD_API.md) — dashboards (`/api/v1/dashboards`, not in OpenAPI)
- Saved views — `/api/v1/explorer/views` (see `signoz/pkg/modules/savedview/` in upstream SigNoz)

Responses use `{ "status": "success", "data": ... }`. The client in `internal/client/client.go` unwraps `data`.

## Repository layout

| Path | Purpose |
|------|---------|
| `main.go` | Provider entrypoint; version via `-ldflags` / GoReleaser |
| `internal/client/` | HTTP client, envelope handling, per-API `*.go` files |
| `internal/client/maps.go` | Shared `ListMap`, `GetMap`, `CreateMap`, `PatchMap`, JSON helpers |
| `internal/provider/` | Resources, data sources, plan modifiers, schema doc constants |
| `internal/provider/configure.go` | `configureResource` / `configureDataSource` helpers |
| `internal/provider/map_attrs.go` | Map API objects → Terraform `types.*` |
| `docs/` | Generated Registry docs (`tfplugindocs`); edit via schema MarkdownDescription or regenerate |
| `examples/` | Copy-paste Terraform examples per resource/data source |
| `tools/tools.go` | `go:generate` for docs + `terraform fmt` on examples |

## Implemented surface

Ten resources and matching singular/plural data sources (see [README.md](README.md) tables). Singular data sources support lookup by `id` or unique `name` where the API allows; cloud accounts use `cloud_provider` + `id`; saved view list DS accepts optional `source_page`, `name`, `category` query filters.

## Adding or changing a resource

1. **Client** — Add `internal/client/<api>.go` with list/get/create/update/delete using `maps.go` helpers. POST may return only `{ "id" }`; fetch full object with GET when needed. Special cases:
   - Cloud account list: `data.accounts[]`, not a top-level array
   - Saved view create: response `data` is a bare UUID string
   - Role update: PATCH body is `{ "description" }` only
2. **Provider** — Prefer one file per domain (e.g. `downtime_schedule.go`) containing resource + singular DS + plural DS. Follow existing files (`notification_channel_resource.go`, `route_policy.go`).
3. **Schema** — API JSON uses **camelCase**; Terraform attributes use **snake_case**. Map with `MapString(m, "createdAt")` → `created_at`.
4. **JSON attributes** — Use `canonicalJSONPlanModifier` on user-supplied JSON strings to avoid plan drift (`internal/provider/jsonplanmodifier.go` + `client.CanonicalJSON`).
5. **Reserved names** — Never use `provider` as a root resource attribute (Terraform reserved). Example: `account_provider` for cloud account API `provider`.
6. **Register** — Add constructors to `Resources()` and `DataSources()` in `internal/provider/provider.go`.
7. **Examples** — Add `examples/resources/<name>/resource.tf` and `examples/data-sources/signoz_<name>/` (+ plural). Update `examples/README.md`.
8. **Docs** — Follow [Schema and documentation](#schema-and-documentation) below, then run `make generate` and commit updated `docs/` if the schema changed.
9. **Import** — Implement `ResourceWithImportState` with `ImportStatePassthroughID` on `id`, except cloud accounts: `<cloud_provider>/<id>`.

Shared schema descriptions for alert/channel/dashboard enums live in `internal/provider/schema_docs.go`.

## Schema and documentation

Registry docs are generated from `MarkdownDescription` on each schema attribute (`tfplugindocs` → `docs/resources/*.md`, `docs/data-sources/*.md`). **Every attribute must be documented completely enough that a practitioner never has to read SigNoz source to use the field.**

When adding or changing a resource or data source, update descriptions in Go (prefer shared constants in `internal/provider/schema_docs.go` for enums reused across resources and data sources).

### Required in every `MarkdownDescription`

| Situation | Document |
|-----------|----------|
| **Optional attribute** | Explicit **default** when omitted: e.g. `Default: false.`, `Default: omitted (empty string).`, `Default: omitted (no items).`, or `Default: null.` — match actual Terraform/API behavior. If the schema uses `Default:` in Framework (e.g. `booldefault.StaticBool(false)`), the description must state the same default. |
| **Discrete / enum input** | **All** allowed values, in backticks, from SigNoz OpenAPI or upstream types — e.g. `Valid values: \`rule\`, \`policy\`.` Do not write “one of …” without listing every value. |
| **Computed attribute with finite values** | All values the API may return (e.g. rule `state`, downtime `status`, channel `type`). Mark as read-only when applicable. |
| **JSON string attributes** | What the JSON object contains, required inner keys, common enum subfields, and link to OpenAPI schema name or provider guide when non-obvious. |
| **Sensitive attributes** | What is stored and that it may not be returned on read (if true). |
| **Immutable / replace-only fields** | That changing the value forces replacement (when using `RequiresReplace` or equivalent). |
| **Lookup data sources** | Which arguments are required vs optional for lookup (`id` vs `name`, `ExactlyOneOf` pairs, cloud `cloud_provider` + `id`). |

### Implementation expectations

- Use `stringvalidator.OneOf(...)` (or other validators) for discrete string inputs where the API enum is closed; keep validator values and `MarkdownDescription` in sync.
- For nested JSON (`spec`, `config`, `schedule`, `composite_query`), document top-level Terraform defaults and enumerate important **inner** enums (as in `docSpec` for alert rules).
- After editing descriptions, run `make generate` and skim the generated markdown under `docs/` to confirm defaults and value lists appear in the Registry output.
- Do not ship new resources with bare attribute names and no description, or with vague text like “optional string” / “type of alert”.

## Commands

```bash
make fmt          # gofmt
make lint         # golangci-lint (gofmt is under formatters.enable in .golangci.yml)
make test         # unit tests under ./internal/...
make generate     # tfplugindocs + terraform fmt examples
make install      # go install provider binary for dev_overrides
```

Acceptance tests (live SigNoz): `TF_ACC=1 SIGNOZ_ENDPOINT=... SIGNOZ_API_KEY=... make testacc`

Local Terraform dev: `dev_overrides` for `registry.terraform.io/ksoviero/signoz` → `$(go env GOPATH)/bin` (see README).

## Code style

- SPDX header `// SPDX-License-Identifier: MPL-2.0` on new Go files
- Match existing error messages: `SigNoz API error`, `Invalid configuration`, `Internal error`
- Keep diffs focused; do not refactor unrelated resources
- Do not commit API keys, `.env`, or real credentials in examples (use placeholders and variables)

## Testing

- Client tests: `internal/client/*_test.go` with `httptest` where practical
- Provider tests: `internal/provider/provider_test.go` and resource-specific tests as needed
- After schema changes, `make generate` must succeed (provider schema must load in Terraform)

## Commit messages and releases

Every commit on the default branch (`main`) should use a **release prefix** at the start of the subject line (case-insensitive). The [Release on merge](.github/workflows/tag-on-merge.yml) workflow scans commits since the latest `v*` tag and picks the **highest** semver bump, then pushes `vMAJOR.MINOR.PATCH`. That tag triggers [GoReleaser](.github/workflows/release.yml), which publishes a **GitHub Release** with grouped release notes (from the same commits) plus signed Registry artifacts.

### Required format

```text
<prefix>: <short description>
```

Examples:

```text
feat: add signoz_route_policy data source
patch: correct alert rule spec JSON canonicalization
bug: handle empty channel list on read
major: remove deprecated alert_rule fields
```

### Prefix → version bump

| Prefix | Semver bump | Notes |
|--------|-------------|--------|
| `major:` | **Major** (X.0.0) | Breaking changes |
| `minor:` | **Minor** (0.X.0) | New capability without breaking callers |
| `feat:` | **Minor** (0.X.0) | Same as `minor:` (feature work) |
| `patch:` | **Patch** (0.0.X) | Bugfixes, small safe changes |
| `bug:` | **Patch** (0.0.X) | Same as `patch:` |
| `fix:` | **Patch** (0.0.X) | Alias for `bug:` |

### No release (no new tag)

These prefixes are allowed but **do not** contribute to a version bump. If **only** these appear since the last tag, the tag workflow skips:

`chore:`, `docs:`, `test:`, `ci:`, `refactor:`, `style:`, `build:`, `skip:`

Merge commits (`Merge pull request …`) are ignored; the workflow uses each merged commit in the range.

### Rules for agents and contributors

1. **Always** start the subject with one of the release prefixes above when the change should ship to the Registry.
2. Use the **highest** applicable prefix in a PR (one breaking change → `major:` for the squash/merge commit).
3. Prefer **squash merge** subjects that include the prefix (e.g. `feat: add saved view resource`), not bare sentences like `Add saved view resource`.
4. Do not tag or publish manually for routine releases; merging to `main` creates the tag and GitHub Release. Manual `v*` tags are only for exceptional recovery.
5. Multiple commits since the last tag: the workflow takes the **maximum** bump (`major` > `minor` > `patch`).

### Publishing

1. Merge to `main` with release-prefixed commits → workflow tags `v*`.
2. Tag push runs GoReleaser: GitHub Release with notes from [generate-release-notes.sh](.github/scripts/generate-release-notes.sh) (Breaking / Features / Bug fixes / Other), plus signed zips and `terraform-registry-manifest.json`.
3. Terraform Registry picks up the new provider version from that release.

Do not replace assets on an already-published version—ship a new tag.
