# Examples — agent notes

Terraform snippets under `examples/` are copied into the [Terraform Registry](https://registry.terraform.io/providers/ksoviero/signoz/latest/docs) by `tfplugindocs` and are meant to be **readable reference material**, not a deployable root module. Practitioners should be able to copy one resource block (or one data source block) and adapt it without hunting for shared `locals` or variables defined elsewhere in the file.

See also the repository root [AGENTS.md](../AGENTS.md) for provider implementation and schema documentation rules.

## Layout

| Path | Contents |
|------|----------|
| `provider/provider.tf` | Minimal `terraform` / `provider "signoz"` block; `signoz_api_key` variable for bootstrap only |
| `resources/<resource_name>/resource.tf` | One or more `resource` blocks for that resource type |
| `data-sources/signoz_<name>/data-source.tf` | Singular data source example (+ optional `output`) |
| `data-sources/signoz_<names>/data-source.tf` | Plural list data source example |
| `README.md` | Index table of all example directories (update when adding a resource) |

Directory names for resources use the Terraform resource name without the `signoz_` prefix (e.g. `alert_rule/`). Data source directories use the full data source name (e.g. `signoz_alert_rule/`).

When you add a new resource or data source to the provider, add matching example directories and a row in `examples/README.md`.

## Principles for good examples

### Be verbose and self-contained

- **Do not use `locals` or `merge()` to deduplicate** repeated JSON fragments across resources in the same file. Each `resource` block should show the full `spec`, `config`, `evaluation`, `notificationSettings`, `roleMapping`, etc., even when that repeats structure from a sibling example.
- **Prefer inlining** over `variable` blocks inside resource examples. Use obvious placeholder strings (`your-signoz-api-key`, `https://hooks.slack.com/services/XXX/YYY/ZZZ`) instead of `var.*` unless the example is specifically about provider configuration (`examples/provider/`).
- **Do not comment out** example resources to avoid apply-time conflicts. Multiple resources in one file are fine; practitioners copy the block they need. Use a short file-level comment when they should pick one variant (e.g. auth domain IdP type).

### Show realistic API shapes

- JSON attributes (`spec`, `config`, `credentials`, `schedule`, `data`, etc.) must use **`jsonencode({ ... })`** with **camelCase** keys matching SigNoz’s API (same as the UI export and OpenAPI), not snake_case inside the JSON object.
- Use values that match provider schema docs and validators in `internal/provider/schema_docs.go` (e.g. alert `alert_type`: `METRIC_BASED_ALERT`, `LOGS_BASED_ALERT`, `TRACES_BASED_ALERT`; threshold `op`: `above`, `below`; `matchType`: `at_least_once`, `all_the_times`).
- For complex resources, prefer **multiple named resources** in one file that illustrate distinct scenarios (e.g. metric vs log vs trace alert rules; SAML vs OIDC vs password-only auth domains) rather than one minimal stub.

### Cross-resource references are allowed when they teach wiring

It is OK to define a shared dependency when the example’s purpose is to show how resources link:

- `signoz_alert_rule` referencing `signoz_notification_channel.slack.name` in threshold `channels`
- `signoz_route_policy` referencing a channel by name

In that case, include the dependency resource in the same `resource.tf` and use `depends_on` when Terraform cannot infer the order. Still **inline** the full JSON for each alert/rule; only the channel name is shared by reference.

### Required / optional field comments (mandatory)

Every example file must document whether each field is **required** or **optional**:

1. **Terraform attributes** on `resource`, `data`, `provider`, and `variable` blocks — suffix or line-above comment using the provider schema:
   - `# required (Terraform)` — attribute is `Required: true` in `internal/provider/*.go`
   - `# optional` — attribute is `Optional: true` (state default if any)
   - `# required (lookup; exactly one of id or name)` — data sources with `ExactlyOneOf` (comment the attribute used in that example; note the alternate key in the file header)
   - `# read-only` — computed data source attributes (usually only on `output` examples, not set in config)

2. **Keys inside `jsonencode({ ... })`** — comment every key at the level shown in the example. Use:
   - `# required (Terraform)` / `# optional` when the JSON is stored in a provider attribute that is required/optional
   - `# required (SigNoz API)` / `# required (Alertmanager)` / `# optional` for inner API fields when OpenAPI or Alertmanager defines behavior (see `notification_channel/`)
   - For long template strings or heredocs, put `# optional` on the line **above** the assignment (HCL cannot trail-comment `<<-EOT`)

3. **Nested JSON** (alert `spec`, dashboard `widgets`, auth `config`) — at minimum comment each **top-level** key inside `jsonencode` and each **major section** (`condition`, `evaluation`, `samlConfig`, etc.). Deeper keys in the example should be commented when they are present.

4. **File header** — include one line: `# Required/optional comments follow provider schema and SigNoz OpenAPI / Alertmanager where applicable.`

Do not omit comments on “obvious” fields. Prefer `# optional` over leaving fields uncommented.

### Comments and naming

- Add a **1–3 line file header** when the file covers several variants or non-obvious API concepts (OpenAPI type name, link to generated `docs/resources/*.md`, signal type).
- Resource addresses should be **descriptive** (`metric_cpu`, `logs_panic`, `oidc_role_claim`), not generic (`example`, `test`).
- Use fictional hostnames (`signoz.example.com`, `saml.example.com`) and fake IDs/certs.

### Security

- Never commit real API keys, webhook URLs, client secrets, or certificates.
- Use placeholders; mark sensitive-looking fields in comments when the API expects a real secret in production.

### Data sources

- Singular examples: show lookup by the most natural key (`name`, `alert`, `id`) per the data source schema. Include a small **`output`** block when it clarifies what is returned (see `data-sources/signoz_alert_rule/`).
- Plural examples: minimal `data "signoz_*" "all" {}` or with optional filters documented in schema (e.g. `source_page` on saved views).
- Cloud integration: singular/plural examples must set `cloud_provider` where required.

### Keep examples out of provider logic

Examples are not acceptance tests. Do not add `provider` blocks to every resource directory—only `examples/provider/provider.tf` documents provider configuration. Resource examples assume a provider is configured in the user’s root module.

## Anti-patterns

| Avoid | Prefer |
|-------|--------|
| Fields without `# required` / `# optional` comments | Every Terraform and JSON key in examples annotated |
| `locals { shared_spec = ... }` | Repeat full `jsonencode` in each resource |
| `merge(local.base, { groupBy = [...] })` | Full `notificationSettings` object per resource |
| `variable` in resource examples for secrets | Inline `"your-api-key"` placeholders |
| Single-line `config = "{}"` | Expanded `jsonencode` showing required keys |
| Commented-out `resource` blocks | Active resources with distinct names |
| snake_case inside API JSON | camelCase per SigNoz API |
| Numeric UI-only threshold codes without context | String enums documented in `schema_docs.go` |

## Workflow when changing examples

1. Edit the relevant `resource.tf` or `data-source.tf`.
2. Run from repo root: `make generate` (runs `terraform fmt -recursive` on `examples/` and regenerates `docs/` from schema).
3. If you added a new resource type, update `examples/README.md`.
4. Skim the generated Registry page under `docs/resources/` or `docs/data-sources/` to confirm the example renders as expected.

## Reference implementations

Use these as style anchors when adding or extending examples:

| Directory | What it demonstrates |
|-----------|----------------------|
| `resources/alert_rule/` | Multiple signal types; full v2alpha1 `spec`; channel references; no locals |
| `resources/auth_domain/` | One resource per IdP / auth mode; full `config` and `roleMapping` inlined |
| `resources/notification_channel/` | Slack, email, webhook (bearer/basic auth), PagerDuty, Opsgenie; per-field required/optional comments |
| `resources/route_policy/` | Expression routing + channel reference |
| `data-sources/signoz_alert_rule/` | Lookup by `alert` + output |

For alert rule JSON field names and enums, cross-check SigNoz `ruler_examples.go` in upstream SigNoz and provider constants in `internal/provider/schema_docs.go` (`docSpec`, `docAlertType`, etc.).
