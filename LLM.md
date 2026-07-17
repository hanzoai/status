# Hanzo Status

Gatus-shaped uptime monitor (fork). One Go binary serves BOTH the JSON API and
the embedded SPA. Multi-brand (Hanzo, Lux, Pars, Zoo, Adnexus) via per-brand config.

Module: `github.com/hanzoai/status`

## Architecture
- `main.go` — boots config → storage → `controller.Handle` (HTTP) + `watchdog.Monitor` (probes).
- `api/api.go` — Fiber router. **This is the API contract:**
  - `GET /v1/status/config` — UI config (unprotected) — the SPA's first call.
  - `GET /v1/status/endpoints/statuses` — per-endpoint results (the live data; `?page=&pageSize=`).
  - `GET /v1/status/endpoints/:key/...` — uptimes / response-times / badges.
  - `GET /v1/status/suites/...` — suite results.
  - `GET /health` — liveness `{"status":"UP"}`. `GET /metrics` — Prometheus (if `metrics: true`).
  - `GET /` + everything else — the embedded SPA (`web/static`, `//go:embed`).
  - NOTE: legacy Gatus used `/api/v1/*`; this fork is `/v1/status/*`. An old image
    serving `/api/v1/*` + a Next.js SPA will make `/v1/status/*` 404 → dead dashboard.
- `web/app/` — the SPA (Vite + React 19 + Tailwind, plain). `src/lib/api.ts` = the API client
  (calls `/v1/status/*`). `scripts/build.sh` → `vite build` → `web/static/` → `make-template.mjs`
  rewrites `index.html` into a Go `html/template` that injects `window.config` (brand). The built
  `web/static/` is committed AND rebuilt in the Docker frontend stage.
  - NOTE: `@hanzo/gui` was REMOVED (v1.1.5). `@hanzo/[email protected]` shipped an inconsistent
    tree (provider on `@hanzogui/core@4.4.0`, components on `@hanzogui/*@3.0.x`) → split theme
    React-context → uncaught `Missing theme.` → BLANK page. No good pin exists (4.7.3 leaks
    `workspace:*` and is uninstallable; 7.x drops `getDefaultGuiConfig`). Its 7 primitives now
    live as a tiny plain-React+Tailwind shim in `src/components/ui.tsx`. Do not re-add `@hanzo/gui`.
- `config/` — brand configs (`hanzo.yaml`, `lux.yaml`, …) in Gatus format
  (`web:`/`storage:`/`ui:`/`endpoints:`). Validated by `config.LoadConfiguration`.

## Config (monitors)
- Format: Gatus endpoints — `url`, `interval`, `conditions` (`[STATUS] == any(...)`,
  `[RESPONSE_TIME] < ms`, `[CERTIFICATE_EXPIRATION] > h`, `[BODY].x == y`).
- **Live source of truth is universe, not this repo:** the deployed `status-config`
  ConfigMap is `hanzoai/universe/infra/k8s/status/configmap-hanzo.yaml`. `config/hanzo.yaml`
  here mirrors it for the repo-native kustomize build.
- Hanzo design: the unified cloud binary serves the whole `/v1` surface at `api.hanzo.ai`.
  It is monitored **exhaustively** — one truthful probe per subsystem:
  - Each subsystem is probed at its OWN JSON health route `/v1/<product>/health`
    (`[STATUS] == 200`). The `/v1` surface returns a clean `404` for unmounted
    routes (no SPA catch-all), so `404`/`5xx`/timeout = DOWN. This is the whole
    `apps.Wire()` set (~82 live `/health` routes).
  - **SPA false-up:** bare `api.hanzo.ai/health` returns the `200` HTML SPA shell —
    a false-up. The core probe uses `/v1/health` (JSON, live-only). Never probe a
    `/v1` subsystem at a web host.
  - Tenancy-gated datastore primitives (`sql`/`kv`/`vector`/`docdb`/`datastore`/
    `search`) return `403 "X-Org-Id required"` = reachable; accepted via
    `any(200,401,403)`. (The old `500` mis-code is fixed, so a real `500` = DOWN.)
  - No `/health` route: `commerce` (`/v1/commerce/health`, auth-gated 401),
    `billing` (`/v1/billing/balance`), `pricing` (`/v1/pricing` catalog).
  - Web properties are edge SPAs: host-level `200` (+ `[CERTIFICATE_EXPIRATION]` on
    apex hosts) is the only meaningful liveness signal.
- Alerting: `alerting.slack` + `alerting.pagerduty` fire on transitions; webhooks
  come from env (`${ALERT_SLACK_WEBHOOK_URL}` / `${ALERT_PAGERDUTY_ROUTING_KEY}`,
  `os.ExpandEnv`) sourced from KMS via a `KMSSecret` CRD into the status Deployment.
  An unset webhook is validated-invalid and ignored (no crash), so the block ships
  inert until the secret is wired. Slack on every endpoint; pagerduty on the
  page-worthy core (api, engine, models, iam, kms, gateway, paas).

## Build & deploy — ONE way
- **Build**: tag `vX.Y.Z` → GHCR `ghcr.io/hanzoai/status:X.Y.Z` (multi-arch) via
  `.github/workflows/deploy.yml` (reusable `hanzoai/.github docker-build.yml@main`).
  Build-only; do NOT add per-brand `kubectl` deploy back (it pinned an arm64-only image
  and dropped the tuned Deployment). If Actions doesn't fire, build in-cluster with buildkit
  (see below).
- **Deploy**: GitOps via `hanzoai/universe/infra/k8s/status/` (`deployment.yaml` image pin +
  `configmap-hanzo.yaml`) and the hanzo operator (`infra/k8s/operator/crs/status*.yaml`).
  Roll with `kubectl set image deployment/status status=…:X.Y.Z -n hanzo` (preserves the
  `hostAliases` that route *.hanzo.ai to the internal ingress for probing). `status` + `status-lux`
  both live in ns `hanzo` on `do-sfo3-hanzo-k8s`.

### Dockerfile gotcha — pnpm
The frontend stage MUST pin pnpm via the `packageManager` field in
`web/app/package.json` (`[email protected]`). `corepack prepare pnpm@latest` pulls pnpm 10,
which hard-errors on esbuild's build script (`ERR_PNPM_IGNORED_BUILDS`) and ignores
`pnpm.onlyBuiltDependencies` — that had silently broken EVERY Vite image build, leaving
the cluster stuck on the pre-Vite image.

### In-cluster buildkit (when Actions is unavailable)
Job pattern (ns hanzo): `moby/buildkit` + `buildctl-daemonless.sh` (privileged),
`--opt=context=https://github.com/hanzoai/status.git#refs/heads/main`,
`--secret=id=GIT_AUTH_TOKEN,env=GIT_AUTH_TOKEN` (from `console-git-token`, org-broad),
`--output=…name=ghcr.io/hanzoai/status:X.Y.Z,push=true`, docker-config from `kaniko-ghcr`.

## status.lux.network — DNS shadow (KNOWN ISSUE)
Public `status.lux.network` resolves to a **Cloudflare Pages / OpenNext** deploy
(`x-opennext`, `server: cloudflare`), NOT the cluster. The in-cluster `status-lux` pod is
healthy but shadowed by DNS. `status.hanzo.ai` correctly points at the cluster ingress-lb
(`129.212.164.5`). To make Lux serve live data from the cluster, repoint `status.lux.network`
DNS to the ingress-lb (white-label: Lux branding only, never Hanzo).
