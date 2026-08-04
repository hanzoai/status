# Hanzo Status

Gatus-shaped uptime monitor (fork). One Go binary serves BOTH the JSON API and
the embedded SPA. Multi-brand (Hanzo, Lux, Pars, Zoo, Adnexus) via per-brand config.

Module: `github.com/hanzoai/status`

## Architecture
- `main.go` — boots config → storage → `controller.Handle` (HTTP) + `watchdog.Monitor` (probes).
- `api/api.go` — `zip` router (`github.com/zap-proto/zip`, the fleet-wide framework;
  a fiber v3 fork). Handlers are `func(c *zip.Ctx) error`. **This is the API contract:**
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
    React-context → uncaught `Missing theme.` → BLANK page. No good pin existed then (4.7.3 leaks
    `workspace:*` and is uninstallable; 7.x drops `getDefaultGuiConfig`). Its 7 primitives now
    live as a tiny plain-React+Tailwind shim in `src/components/ui.tsx`.
  - `@hanzo/gui@8.0.1` FIXES that blocker — verified in a clean Vite 8 + React 19 spike:
    the whole tree is `@hanzogui/*@8.0.x` on ONE `core@8.0.1`, `GuiProvider` + `createGui(defaultConfig)`
    (from `@hanzogui/config/v4`) mounts with no `Missing theme.`, themed dark values apply, Button
    is 44px with a real border-style. So the standard-stack migration (@hanzo/ui 8 on @hanzo/gui 8,
    Tailwind deleted) is UNBLOCKED. It is still a full rewrite of all 19 `src/components/*` files
    (className → style props) + `@hanzogui/vite-plugin` extraction; copy the integration recipe from
    `~/work/hanzo/hanzo.one` (vite knobs, `.npmrc` `public-hoist-pattern[]=@hanzogui/*`, `outputCSS`
    + `disableInjectCSS`, @hanzo/design tokens bound over the v4 preset — the raw preset renders
    Button text black on dark). Do it in ONE landing, never half-on-Tailwind.
  - Mobile is first-class: every interactive target ≥44px (`min-h-11`, icon buttons `h-11 w-11`,
    card-name buttons `py-3 -my-3`), form fields ≥16px text (iOS zoom), `viewport-fit=cover` +
    `env(safe-area-inset-*)` on body/header/footer/refresh chip, chart canvas redraws via
    ResizeObserver. Keep these invariants across any restyle.
- `config/` — brand configs (`hanzo.yaml`, `lux.yaml`, …) in Gatus format
  (`web:`/`storage:`/`ui:`/`endpoints:`). Validated by `config.LoadConfiguration`.

## Config (monitors)
- Format: Gatus endpoints — `url`, `interval`, `conditions` (`[STATUS] == any(...)`,
  `[RESPONSE_TIME] < ms`, `[CERTIFICATE_EXPIRATION] > h`, `[BODY].x == y`).
- **Live source of truth is universe, not this repo:** the deployed `status-config`
  ConfigMap is `hanzoai/universe/infra/k8s/status/configmap-hanzo.yaml`. `config/hanzo.yaml`
  here mirrors it for the repo-native kustomize build.
- Hanzo design: the unified cloud binary serves the whole `/v1` surface at `api.hanzo.ai`;
  probe each product's own `/v1/<product>` route (200, or auth-gate 401/403 = UP).
  Tenancy-gated data routes return `500 "X-Org-Id required"` (a gateway mis-code) —
  accepted via `any(200,401,403,500)`; a real outage is 502/503/504/timeout.

## Routing on zip — what to know before editing `api/api.go`
- The route table is **pinned** by `declaredRoutes` in `api/routes_test.go`. Changing the
  published surface means editing that list deliberately; a refactor must never move it.
- `zipx.Wrap` is the ONE bridge from a fiber middleware (`fiber.Handler`) to a
  `zip.Handler`. Use `zip.AdaptNetHTTP*` for net/http handlers and middleware.
- Route precedence in zip's fiber fork is **specificity-based**, not registration order
  (`static ≻ :param ≻ *`), and two distinct patterns of equal specificity panic at boot
  rather than silently shadow. `Use` middleware are precedence *barriers*, which is what
  keeps the "unprotected routes registered before the security middleware" split working.
- `api.writeJSON` exists because `zip.Ctx.JSON` tags bodies
  `application/json; charset=utf-8`; this API has always sent bare `application/json`.
- `controller.Handle` listens via `app.Fiber().Listen(addr, fiber.ListenConfig{…})` — the
  one escape from the zip surface, because `zip.Listen` carries no TLS or dual-stack knob.

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
