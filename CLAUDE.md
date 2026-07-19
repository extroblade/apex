# Apex (repo: apex) — project guide

**Apex** is an iRacing companion app (fuel calculator, season planner, garage,
setups showroom, goal tracker, driver stats). Go + MySQL backend, React +
TypeScript frontend (rsbuild, Feature-Sliced Design). Everything is Dockerized. The brand is "Apex" (apex-arc
logo in `frontend/src/shared/ui/logo.tsx`, favicon in `frontend/public/`);
"Racing Planner" is the old name — don't reintroduce it.

## Structure

- `backend/` — Go API (`cmd/server`, `internal/...`). See `backend/CLAUDE.md`.
- `frontend/` — React SPA (FSD under `src/`). See `frontend/CLAUDE.md`.
- `static/` — nginx serving generated assets (8 default avatars under
  `/avatars`; regenerate with `node static/generate-avatars.mjs`).
- `nav/` — **menu service** (own Go module, own compose project): owns the
  `nav_items` table and serves `GET /api/nav`. See "Backend-driven navigation".
- `bff/` — **mobile Backend-for-Frontend** (NestJS, own compose project): reshapes
  the Go API + nav for a future mobile app. Owns no data/auth — forwards the
  caller's auth. `GET /bff/home` (aggregated+gated), `/bff/health`, `/metrics`.
  Web app does NOT use it; the mobile app hits it directly on :8083.
- The compose projects are **separate**, joined by the external `apex-net`
  network. Frontend nginx proxies `/api/nav` → nav, `/api` → backend, and
  `/media` → static, and caches (`media_cache` 7d, `/api/features` 30s).
  `/api/nav` is a LONGER prefix than `/api/`, which is how nginx splits them —
  keep that ordering. `/static` is rsbuild's OWN bundle path — never proxy it.
- `./dev.sh up|down|logs|ps` runs the whole stack (frontend :3000, api :8080,
  bff :8083, mysql :3306).

## Feature flags & the no-iRacing planner

- Flags live in the `feature_flags` table (`internal/features`, cached 30s);
  the frontend reads `GET /api/features` (`entities/features`, `useFeature`).
- **`iracing_oauth` is OFF** (iRacing paused OAuth client registration; our
  OAuth code is ready but can't be exercised). Gated behind it: linking, driver
  lookup, live stats, race sync, comparators (routes 404 via `requireFeature`,
  nav hidden, pages show `IRacingUnavailable`).
- **Works WITHOUT iRacing**: auth/profile, fuel calc, garage, the **season
  planner**, the **setups showroom**, and the **goal tracker**.
- **Season planner** (`GET /api/planner/season`): a racingplanner.com-style grid
  of series × 13 weeks, colored by **track access** (`trackAccess`): FREE/included
  = green, OWNED/purchased = aquamarine (teal), MISSING = red; a **planned** race
  glows amber. Every cell toggles a planned race (`PUT /api/planner/season/plan`,
  `planned_races`, any number per week). The grid breaks out to ~full screen
  width and has a transpose button + favorites filter. The current week's races
  live on their **own page** (`/this-week`, `features/season-planner` → `ThisWeek`),
  not in the grid.
- **Free/paid content model** (`internal/racing/access.go`): `cars.is_free` /
  `tracks.is_free` mark included content; a purchase unlocks every config sharing
  a `tracks.sku_group`; combined layouts in `track_requirements` are owned when
  all their component tracks are (e.g. Nürburgring Combined = GP + Nordschleife).
  `trackAccess()` resolves free/owned/missing per user. Tracks are stored per
  config but **displayed deduped by base track** — the garage shows one row per
  track with its layouts in the info dialog; buying toggles all configs together.
- The schedule in `season_schedule` is **REAL** (never fabricate it): the seed
  `backend/internal/racing/catalog_seed.json` is GENERATED from
  my-racing-planner's data exports (real iRacing ids for 186 cars / 425 track
  configs / 150 series, prices, free flags, series→car mapping in `series_cars`,
  and each series' actual weekly tracks with `race_date`, windowed to the
  current 13-week season) — regenerate with `backend/scripts/gen-catalog-seed.py`.
  `racing.SeedCatalog` upserts it on startup and reconciles any pre-real-id rows
  by name (moving ownership/favorites/plans). The current week comes from
  `race_date`, not ISO math. The **scheduler service** (`cmd/scheduler`) checks
  **daily** for a new season PDF (`internal/schedulepdf`; drop folder
  `backend/schedules/`; iracing.com URLs 403 anonymously) and **weekly** runs
  `internal/contentsync`: JSON lists for cars/tracks/series+schedule, plus the
  iRacing web catalog for artwork/free flags. The season PDF has only a header
  banner (no per-series logos), extracted to `SCHEDULE_IMAGE_DIR`. The old
  manual plan rows + `.ics` calendar were REMOVED (`plan_entries` remains, unused).

## Run & verify (ALWAYS before calling a change done)

- Backend: `cd backend && go build ./... && go vet ./... && go test ./...`
- Frontend: `cd frontend && pnpm run typecheck && pnpm run lint && pnpm test && pnpm run build`
- E2E (stack up; `pnpm exec playwright install chromium` once): `cd frontend && pnpm run e2e`
  — 8 specs, incl. a Pixel-7 mobile spec. Use Chromium device profiles only.
- Storybook: `pnpm run storybook` / `pnpm run build-storybook`.

## Non-negotiable standards

- **Migrations are append-only and additive.** Never edit an applied file in
  `backend/internal/migrate/migrations/`; add a new numbered one. User data must
  survive every deploy — verify with a rebuild WITHOUT `down -v`.
- **Match surrounding style**; tests for key scenarios (table-driven + httptest
  on Go; Vitest + Testing Library on React; Playwright for flows).
- **Theming**: only semantic tokens (`bg-background`, `text-foreground`, …) —
  light/dark/custom themes swap CSS variables.
- **i18n**: user-facing strings via `useTranslation()`; add keys to BOTH
  `locales/en.ts` and `locales/ru.ts` (ru is typed against en).
- **Responsive + a11y**: 320px→1920px; mobile uses the **bottom nav bar**
  (`widgets/bottom-nav`), desktop the top header. Keep aria-labels/aria-current,
  the skip-link in `app/App.tsx`, and cursor-pointer on interactive elements.
- **No secrets in the repo** (env via gitignored `.env`).

## Domain specifics worth remembering

- Fuel calculator supports **strategies** (`balanced` default | `undercut` |
  `overcut`) and always balances stints — never 20+1 splits
  (`backend/internal/fuel`). Optional **race rules**: `rules.mandatoryStops`
  forces extra stops; `rules.windows[i]` bounds stop i (`from`/`to` in `laps`
  or `minutes` — minutes need `lapTimeSec`). Infeasible rules → 422. Stints
  carry `pitOnLap`.
- Series carry `license_needed` (R/D/C/B/A) shown as "Lic. X" badges.
- **Setups showroom** (`internal/setups`, `/api/setups`, `pages/setups`): users
  save car setups (plain text — an exported `.sto` or a values dump) privately
  and optionally publish them; browse public + own, download (bumps a counter).
- **Goal tracker** (`internal/goals`, `/api/goals`, `pages/goals`): personal
  numeric goals (target/current/unit, optional due date) with a progress bar;
  auto-completes at target, +/- quick-adjust, manual done toggle.
- **Setup generator** (`internal/setups/generator.go`): deterministic, NOT
  telemetry — car-discipline baseline + track-character tweak + optional
  skill/session deltas, rendered to a plain-text `.sto`-ish file.
  `POST /api/setups/generate` → one balanced baseline ("Generate baseline"
  button). `POST /api/setups/generate/pack` → the **2×4 pack**: skill
  (`safe`/`pro`) × session (`endurance`/`race`/`qual`/`rain`) = 8 variants
  ("Generate pack" → review panel with per-variant "Use" + "Save all"). Both
  share `computeSetup`/`render`; deltas live in `skillDeltas`/`sessionDeltas`
  and only apply where the discipline has the field (an oval never grows a
  wing/diff). `race` session = zero delta (the balanced centre).
- **Forms** use `react-hook-form` + `zod` (`@hookform/resolvers/zod`) with a
  shared Radix `Select` (`shared/ui/select.tsx`), `Textarea`, and a
  `DatePicker` (`shared/ui/date-picker.tsx`, react-day-picker + Popover —
  never native `<input type="date">`); validation messages are i18n keys
  resolved through `t()`. Prefer this over ad-hoc state + native controls.
- **fx components** (`shared/ui/fx/`): `Aurora` animated background,
  `ShinyText`, `SpotlightCard` — reactbits.dev-style, token-driven (theme-safe),
  `prefers-reduced-motion` aware. Used on Home hero + Login.
- **Backend-driven i18n** (`internal/locales`, `handler/locales.go`,
  `shared/i18n`): `en` is bundled in the frontend as the instant/offline fallback
  AND the source of the `Translation` type; every other language (incl. `ru`) is
  a row in the `locales` table (migration 0023), served by `GET /api/locales`
  (list) + `GET /api/locales/{code}` (bundle JSON). A new language is a DB row —
  no frontend deploy (proven: `INSERT` a locale, it appears in the menu). `en.ts`
  + `ru.ts` stay the authored, type-checked source; `pnpm run gen:locales`
  (part of `pnpm run build`) exports them to `backend/internal/locales/data/*.json`
  for the backend to `go:embed` + seed on startup (built-ins re-seed via
  `ON DUPLICATE KEY UPDATE`; runtime-added locales are untouched). A key-parity
  test (`internal/locales`) guards the generated bundles against drift.
  `setLanguage` fetches non-`en` bundles on demand; missing keys fall back to en.
- **Catalog image rehosting** (`internal/contentsync/images.go`): the scheduler
  downloads each catalog image once (per base name — configs share art) into the
  shared **`apex-media-data`** volume and rewrites `image_path` to a relative
  `/media/catalog/<table>/<file>`. The volume ROOT is the catalog dir:
  `CATALOG_IMAGE_DIR=/media-data` in the scheduler; static mounts the same volume
  at `/usr/share/nginx/html/catalog` (do NOT nest an extra `/catalog` in
  `CATALOG_IMAGE_DIR` — it double-nests). The Dockerfile pre-creates
  `/media-data/{cars,tracks}` owned by `nonroot` so the fresh volume is writable
  (distroless runs as nonroot). Descriptions are backfilled from each detail page
  (`detail_url`, migration 0020) — meta/og/first-`<p>`, capped to 1000 chars,
  column widened to `VARCHAR(2000)` (migration 0022). Rehost + descriptions run
  as their OWN DB-scanning steps every sync (image_path LIKE 'http%' /
  description='' AND detail_url<>''), INDEPENDENT of the content-hash guard.
- **Cockpit dev overlay** (`internal/handler/cockpit.go`, `features/cockpit`):
  `?dev=KEY` sets a `developer` cookie (`?dev=off` clears it; handled in
  `app/index.tsx`). Backend `DEVELOPER_KEY` env gates it — empty = all off;
  `backend/docker-compose.yml` defaults it to **`3`**, so the overlay works in
  any environment the stack starts via **`?dev=3`** (documented in the README's
  "Cockpit" section; override it for a public deploy).
  `GET /api/features/all`, `PUT /api/features/{key}`, `GET /api/health/cockpit`
  return 404 unless the cookie matches (cookie is the ONLY gate — no feature-flag
  gate, that'd be chicken-and-egg). The `cockpit` flag (migration 0021, seeded
  off) is just a normal togglable flag now. Frontend: floating wrench button
  (visible only when `isDev()`) opens a Radix Dialog listing flags with toggles +
  a health readout; `shared/lib/dev.ts` (`isDev`, `devlog` no-op without cookie).
- **Backend-driven navigation** (`nav/` service, `entities/nav`,
  `widgets/side-nav` + `widgets/bottom-nav`): the menu is DATA, not hard-coded.
  The nav service owns `nav_items` (`item_key, label_key, href, icon,
  placements, sort_order, requires_auth, feature_flag, enabled`), creates+seeds
  it itself on startup (`INSERT IGNORE`, so Cockpit edits are never clobbered),
  and serves `GET /api/nav`. It deliberately does NOT read sessions or feature
  flags: it ships the whole menu plus gating metadata, and the **client filters**
  (`visibleNav()` by placement + `requiresAuth` + `featureFlag`) since it already
  has the viewer and flags. Nav is not a security boundary — routes enforce their
  own auth — which is what keeps this service dependency-free (MySQL only).
  Labels are **i18n keys** (`nav.planner`), never text; icons are **names**
  mapped through a whitelist (`entities/nav/ui/NavIcon.tsx`, unknown → dot).
  Layout: desktop = side menu (`hidden md:block`), mobile = bottom bar
  (`md:hidden`, max 5 slots, overflow folds into "More"); the **header is
  minimal** (brand + user menu only, on every viewport).
- **Redis cache** (`internal/cache`, `redis:7-alpine`, no external port): a
  strictly **fail-open** go-redis v9 wrapper — nil client / downed Redis / miss
  all fall through to MySQL with NO error surfaced, and every op has a 300ms
  timeout + no retries so a dead Redis never stalls a request. Feature flags read
  through it (30s TTL, key `features:flags`); the Cockpit toggle invalidates it.
  `REDIS_ADDR` empty disables it. Catalog reads are NOT cached yet (roadmap).
- **Metrics**: provider-agnostic. FRONTEND — `shared/metrics`: `useCounter()`
  returns the curried `counter('event')(params)`; default provider is Yandex
  Metrica (`reachGoal`, id from `PUBLIC_YM_ID` — EMPTY for now, so it no-ops and
  just `devlog`s). Known events in `MetricEvents`, but any string works. BACKEND
  & BFF — Prometheus (`internal/metrics` in Go, `src/metrics` in the BFF): request
  middleware (labeled by route PATTERN, not raw path), a `/metrics` exposition
  endpoint (root, unproxied → internal scrape only), and a generic
  `Count(name, help, labels)` domain counter mirroring the frontend helper.
- **BFF** (`bff/`, NestJS): the mobile edge. `UpstreamService` forwards the
  caller's cookie/bearer to the Go API (`API_BASE_URL`) and nav (`NAV_BASE_URL`)
  — it knows the topology since nginx isn't in front of it. `HomeService.home()`
  fans out to me+nav+features and does the nav gating server-side (so the app
  doesn't reimplement `visibleNav`), returning one mobile-shaped payload. Jest
  unit (mocked upstream) + supertest e2e; own `bff-ci`.

## Roadmap / staged next (agreed with the user)

Done recently: planner redesign (grid + this-week page + access colors), free/paid
content model + track dedup, content sync (JSON list + iRacing web catalog),
setups showroom, goal tracker, forms on zod + react-hook-form, precise
series→car mapping (`series_cars` in the seed), **catalog image rehosting +
description backfill**, **Cockpit dev overlay**, **Redis cache (fail-open)**,
**backend-driven navigation** (nav service + side menu + minimal header),
**setup pack generator** (2×4 skill×session), **backend-driven i18n** (locales
service + DB-served bundles), **metrics** (frontend counterHelper + Prometheus on
backend & BFF), **BFF** (NestJS mobile Backend-for-Frontend), **legal/IP hygiene
slice 1** (MIT LICENSE, iRacing non-affiliation disclaimer + footer, `IRACING_SCRAPE`
gate default-OFF, committed AES key removed, Terms/Privacy pages, product-ized
About page), **security hardening slice 1** (auth rate limiting `AUTH_RATE_LIMIT`
+ nginx `limit_req`, allowlist CORS, nginx security headers, loopback-only host
ports, constant-time Cockpit compare, session revocation on password change,
React ErrorBoundary).

## Product / commercialization (decided with the user)

Turning this from a personal project into a **commercial product**. Decisions:
- **Model**: freemium subscription (free tier + Pro; Pro = pack generator, full
  season planner, garage sync, private cloud setups, unlimited goals).
- **License**: MIT (open repo) — the value is the hosted service, not the code.
- **iRacing**: an official partnership is NOT expected to be attainable, so we
  **monetize WITHOUT it**. This means: never sell access to iRacing's
  copyrighted assets/data. Sell the ORIGINAL tools (fuel calc, setup generator,
  planner UX). `IRACING_SCRAPE` is off; the scraped artwork/descriptions must be
  replaced with our own before those catalog features are part of a paid tier.
- **Do not reintroduce** "learning/pet project" tells in user-facing surfaces.

Productization backlog (ordered; slices 1 done — see "Done recently"):
1. Finish IP hygiene: replace scraped car/track artwork with **own track-map SVGs
   + original descriptions** (Terms/Privacy pages are DONE). Rotate the leaked AES
   key (still in git history) before any real deploy.
2. Finish security/ops: **TLS termination + HSTS** (deployment), **error tracking
   SDK** (wire Sentry into the ErrorBoundary hook + backend Recoverer), full
   **session rotation** + expired-row purge, drop the loopback port maps entirely
   in a prod compose override. (Rate limiting, CORS allowlist, security headers,
   constant-time compare, session revocation-on-password-change — DONE.)
3. **Account lifecycle**: transactional email (SMTP config + mailer pkg) →
   password reset, email verification, account deletion + data export (GDPR),
   password-confirmed email change.
4. **Billing**: Stripe subscriptions + Pro feature-gating + upgrade page.
5. Edit `nav_items` from the Cockpit (menu is DB-driven; the editor UI is missing).
6. Microfrontend split (module federation) — design first.
7. Extend the Redis cache to catalog reads; grow the BFF as the mobile app needs.
