# Apex

**Apex** is an iRacing companion app: fuel & stint calculator, season planner,
garage, setups showroom, goal tracker, and driver stats. Go + MySQL backend,
React + TypeScript frontend. Everything is Dockerized.

- **Backend** — Go 1.25, [chi](https://github.com/go-chi/chi) router, `database/sql` + MySQL driver. HTTP API + a daily schedule-sync scheduler.
- **Frontend** — React + TypeScript, built with [rsbuild](https://rsbuild.dev). Routing via [wouter](https://github.com/molefrog/wouter), client state via [zustand](https://zustand.docs.pmnd.rs), server state via [TanStack Query](https://tanstack.com/query), UI via [shadcn/ui](https://ui.shadcn.com) + Tailwind CSS v4. Organized with [Feature-Sliced Design](https://feature-sliced.design).
- **Nav** — a small standalone Go service that owns the app's **menu configuration** (`nav_items`) and serves it at `/api/nav`, so navigation is data on the backend rather than hard-coded in the SPA. It needs only MySQL: it serves the menu plus each item's `requiresAuth`/`featureFlag` metadata and lets the client (which already holds the viewer and the flags) filter — navigation isn't a security boundary, since every API route enforces its own auth.
- **Static** — nginx serving generated default avatars and rehosted catalog images (`/media/catalog/*`) from the shared `apex-media-data` volume the scheduler writes to.
- **i18n** — backend-driven: only `en` is bundled in the app (instant fallback + the translation type); every other language is a row in the `locales` table, served by `GET /api/locales` + `GET /api/locales/{code}`, so a new language ships with **no frontend deploy**.
- **Database** — MySQL 8.4 (Dockerized).
- **Cache** — Redis 7 (Dockerized, no external port). A strictly **fail-open** wrapper (`internal/cache`): if Redis is unset or down, reads fall through to MySQL with no error and no stall (300ms op timeout, no retries). Caches feature flags today.
- **Docker** — the backend (`backend/docker-compose.yml`), frontend (`frontend/docker-compose.yml`), static (`static/docker-compose.yml`), and nav (`nav/docker-compose.yml`) are **separate compose projects** joined by a shared external network, so each can be split into its own repo later. `./dev.sh` runs them all with one command.

## Layout

```
.
├── backend/                  # Go API
│   ├── cmd/
│   │   ├── server/           # API entrypoint
│   │   └── scheduler/        # daily season-schedule PDF sync job
│   ├── internal/
│   │   ├── auth/             # session auth, profile
│   │   ├── cache/            # fail-open Redis wrapper (feature-flag cache)
│   │   ├── config/           # env-based config
│   │   ├── contentsync/      # weekly catalog sync: JSON lists + iRacing web
│   │   │                     # catalog + image rehost + description backfill
│   │   ├── db/               # MySQL connection pool
│   │   ├── features/         # feature flags
│   │   ├── fuel/             # fuel & stint strategy
│   │   ├── goals/            # goal tracker
│   │   ├── handler/          # HTTP handlers
│   │   ├── iracing/          # iRacing OAuth + Data API client
│   │   ├── middleware/       # CORS, auth
│   │   ├── migrate/          # migration runner
│   │   │   └── migrations/   # 0001_init.sql … append-only, applied on startup
│   │   ├── racing/           # planner, catalog, free/paid access model
│   │   ├── schedulepdf/      # season PDF parser
│   │   ├── secretbox/        # AES-256-GCM encryption at rest
│   │   ├── server/           # router wiring
│   │   └── setups/           # setups showroom + baseline generator
│   ├── scripts/              # gen-catalog-seed.py
│   └── Dockerfile            # multi-binary: server + scheduler
├── frontend/                 # React SPA (FSD)
│   ├── src/
│   │   ├── app/              # init: providers, router, styles, entry
│   │   ├── pages/            # home, fuel, planner, this-week, garage, setups,
│   │   │                     # goals, dashboard, drivers, driver-profile,
│   │   │                     # compare, profile, login, about
│   │   ├── widgets/          # header (minimal), side-nav, bottom-nav, user-menu
│   │   ├── features/         # auth, fuel-calculator, season-planner,
│   │   │                     # manage-content, setups-manager, goal-tracker,
│   │   │                     # link-iracing, profile, customize-theme
│   │   ├── entities/         # viewer, planner, setups, goals, driver,
│   │   │                     # iracing, features, nav
│   │   └── shared/           # ui (shadcn), lib (cn), api, config, i18n, theme
│   ├── e2e/                  # Playwright specs
│   ├── rsbuild.config.ts
│   ├── nginx.conf            # prod: serves SPA, proxies /api → backend
│   ├── docker-compose.yml    # frontend project
│   └── Dockerfile
├── static/                   # nginx: avatars + catalog media
├── nav/                      # menu service (own Go module + compose project)
│   ├── main.go               # GET /api/nav
│   ├── migrate.go            # owns + seeds the nav_items table
│   └── docker-compose.yml
├── dev.sh                    # runs all the projects together
└── .github/workflows/        # CI/CD (backend-ci, frontend-ci, e2e)
```

## Run everything (Docker)

```bash
./dev.sh up       # creates the shared network, builds & starts all three projects
./dev.sh down     # stop everything
./dev.sh logs     # tail all logs
./dev.sh ps       # show running services
```

The three stacks are separate compose projects on a shared external network
(`apex-net`, auto-created by the script). Run one on its own if you like:

```bash
docker network create apex-net           # once
docker compose -f backend/docker-compose.yml up --build
```

- Frontend → http://localhost:3000
- API      → http://localhost:8080/api/health
- MySQL    → localhost:3306 (user `app` / pass `app` / db `app`)

nginx in the frontend container proxies `/api/*` to the backend, so the SPA
calls the API on its own origin — no CORS needed in production.

## CI/CD

GitHub Actions (in `.github/workflows/`) runs on every push to `main` and PR:

| Workflow | What it does |
| -------- | ------------ |
| `backend-ci` | `gofmt`, `go vet`, `go test -race -cover`. On `main`/tags, builds & pushes `ghcr.io/<owner>/apex-backend`. |
| `nav-ci` | Same checks for the nav service. On `main`/tags, builds & pushes `ghcr.io/<owner>/apex-nav`. |
| `frontend-ci` | `tsc` typecheck, ESLint, Prettier check, Vitest, production build. On `main`/tags, builds & pushes `ghcr.io/<owner>/apex-frontend`. |
| `e2e` | Brings up the full stack (MySQL + backend + static + frontend) via the existing compose files and runs the Playwright suite. Uploads the HTML report as an artifact. |

Docker images are published to the **GitHub Container Registry (GHCR)** using the
auto-provided `GITHUB_TOKEN` — no extra secrets required. Each image is tagged
`latest` (on `main`), the short commit SHA, and the semver on `v*` tags.

Recommended branch-protection **required checks**: `backend-ci`, `frontend-ci`,
`e2e`.

## Local development

Backend:

```bash
cd backend
cp .env.example .env          # point DB_HOST at your MySQL
go run ./cmd/server
```

Frontend (rsbuild dev server proxies `/api` → `http://localhost:8080`):

```bash
cd frontend
npm install
npm run dev                   # http://localhost:3000
```

Tip: `docker compose up db` to run just MySQL while developing the apps natively.

## API

| Method | Path                  | Auth | Description                          |
| ------ | --------------------- | ---- | ------------------------------------ |
| GET    | `/api/health`         | —    | Liveness + DB connectivity           |
| GET    | `/api/features`       | —    | Public feature-flag map              |
| GET    | `/api/nav`            | —    | Menu config (nav service) — items + placement/gating metadata |
| GET    | `/api/locales`        | —    | Available languages (backend-driven i18n) |
| GET    | `/api/locales/{code}` | —    | One language's translation bundle (JSON) |
| POST   | `/api/fuel/plan`      | —    | Compute a fuel & stint strategy      |
| GET    | `/api/features/all`   | dev cookie | Cockpit: all flags (404 unless `developer` cookie = `DEVELOPER_KEY`) |
| PUT    | `/api/features/{key}` | dev cookie | Cockpit: toggle a flag (`{"enabled":bool}`) |
| GET    | `/api/health/cockpit` | dev cookie | Cockpit: DB + Redis health readout   |
| POST   | `/api/auth/register`  | —    | Create an account (auto-logs in)     |
| POST   | `/api/auth/login`     | —    | Log in, sets `session` cookie        |
| POST   | `/api/auth/logout`    | cookie | Log out, clears the session        |
| GET    | `/api/auth/me`        | cookie | Current user (401 if not logged in)|
| PATCH  | `/api/auth/profile`   | user | Update nickname + email              |
| PUT    | `/api/auth/avatar`    | user | Set/clear avatar (image data URL)    |
| POST   | `/api/auth/password`  | user | Change password (verifies current)   |
| GET    | `/api/iracing`        | user | iRacing link status                  |
| GET    | `/api/iracing/authorize` | user | Start OAuth (redirects to iRacing) |
| GET    | `/api/iracing/callback`  | — | OAuth redirect target (uses state) |
| DELETE | `/api/iracing`        | user | Unlink iRacing account               |
| GET    | `/api/iracing/stats`  | user | Live dashboard (licenses/career/recent) |
| POST   | `/api/iracing/sync`   | user | Sync recent races into MySQL         |
| GET    | `/api/compare/{categories,cars,tracks}` | user | Aggregated comparators |
| POST   | `/api/planner/catalog/sync` | user | Sync car/track/series catalog from iRacing |
| GET    | `/api/planner/{cars,tracks,series}` | user | Catalog with owned/favorite flags |
| PUT    | `/api/planner/{cars,tracks}/{id}` | user | Toggle owned (`{"owned":bool}`) |
| PUT    | `/api/planner/series/{id}` | user | Toggle favorite (`{"favorite":bool}`) |
| POST   | `/api/setups/generate` | user | Deterministic baseline setup for a car+track |
| POST   | `/api/setups/generate/pack` | user | A pack of setups: skill (safe/pro) × session (endurance/race/qual/rain) |
| GET/POST/DELETE | `/api/planner/plan[/{id}]` | user | Manual plan rows (series/track/car) |
| GET    | `/api/drivers/search?q=` | user | Driver search (via your own linked session) |
| GET    | `/api/drivers/{custId}` | user | Any driver's profile (cached 6h) |

App auth uses an httpOnly `session` cookie backed by a `sessions` table; only
the SHA-256 hash of each token is stored.

**iRacing auth is OAuth 2.x** (iRacing removed legacy email/password auth on
2025-12-09). Users click "Connect iRacing" → authorize on iRacing's own page →
we store their **refresh token** encrypted at rest (AES-256-GCM). To enable it,
register an OAuth client with iRacing and set `APP_ENCRYPTION_KEY` (base64 32
bytes), `IRACING_CLIENT_ID`, `IRACING_OAUTH_REDIRECT_URI` (and
`IRACING_CLIENT_SECRET` if issued). Without these, the `/api/iracing`,
`/api/drivers`, `/api/compare`, and `/api/planner` routes return 503.

Other env: `REDIS_ADDR` (e.g. `redis:6379`; empty = cache disabled, fail-open)
and `CATALOG_IMAGE_DIR` (scheduler; the shared media-volume root, `/media-data`
in Docker) — see `backend/.env.example`.

## Menu (nav service)

The navigation is **configuration, not code**. The `nav` service owns the
`nav_items` table (it creates and seeds it on startup) and serves it at
`/api/nav`; the SPA renders whatever it gets — a side menu on desktop, the
bottom bar on mobile (`placements`), ordered by `sort_order`.

Each row carries its own gating metadata, which the client applies:

| column | meaning |
| ------ | ------- |
| `placements` | `side` and/or `bottom` (comma-separated) |
| `sort_order` | order within a placement |
| `requires_auth` | hide from logged-out visitors |
| `feature_flag` | hide unless that flag is on (e.g. `iracing_oauth`); empty = always |
| `enabled` | soft on/off switch |
| `label_key` | an **i18n key** (`nav.planner`), not text |
| `icon` | a lucide **name**, resolved through a whitelist on the client |

Until the Cockpit gets an editor (roadmap), change it with SQL — no redeploy,
no frontend build:

```sql
-- hide an item, move another to the top
UPDATE nav_items SET enabled = 0    WHERE item_key = 'compare';
UPDATE nav_items SET sort_order = 5 WHERE item_key = 'goals';
```

The seed uses `INSERT IGNORE`, so restarts never clobber your edits.

## Cockpit (dev overlay)

A dev-only overlay to flip **feature flags at runtime** and read backend health,
without a redeploy or a DB query. It works in **any environment** — the only gate
is a `developer` cookie that matches the backend's `DEVELOPER_KEY`.

**How to open it:**

1. Start the stack — `./dev.sh up`.
2. Visit **<http://localhost:3000/?dev=3>** — the `?dev=` param stores the
   `developer` cookie and is then stripped from the URL.
3. A **wrench button** appears (bottom-right, floating). Click it: you get every
   feature flag with a live toggle, plus DB + Redis health.
4. Visit any page with **`?dev=off`** to clear the cookie and hide it again.

The key comes from the backend `DEVELOPER_KEY` env, which the Docker stack
defaults to `3` — hence `?dev=3`. Set your own to change it:

```bash
DEVELOPER_KEY=my-secret docker compose -f backend/docker-compose.yml up -d backend
# → then open http://localhost:3000/?dev=my-secret
```

Without a matching cookie, `GET /api/features/all`, `PUT /api/features/{key}`,
and `GET /api/health/cockpit` all return **404** (they don't exist as far as any
other client can tell), and the frontend renders no button at all. Setting
`DEVELOPER_KEY=` (empty) disables Cockpit entirely.

> **Security:** `3` is a convenience default for local development. Anyone who
> knows the key can toggle your feature flags, so set a long random
> `DEVELOPER_KEY` for any publicly reachable deploy.

### Phases

1. **Fuel & stint calculator** — `internal/fuel`, `/fuel`
2. **App auth + iRacing linking** — `internal/auth`, `internal/iracing` (Data API
   client), `internal/secretbox` (encryption), `internal/racing` (link + session
   cache), `/dashboard`
3. **Driver dashboard** — live licenses / career / recent races
4. **Race sync** — `POST /api/iracing/sync` persists recent races
5. **Comparators** — `/compare`: per category / car / track aggregates
6. **Racing Planner** — `/garage` (sync catalog, mark owned cars/tracks, favorite
   series) + `/planner` (build a manual series → track → car plan table)
7. **Driver browser** — `/drivers`: search any driver and view their stats.
   Each lookup runs through the **logged-in user's own** linked iRacing session
   (no shared account, so no single account carries all traffic), with a 6h
   MySQL cache (`driver_cache`, keyed by cust_id)

## Adding shadcn/ui components

`components.json` is preconfigured (aliases point into `src/shared`). Add
components with:

```bash
cd frontend
npx shadcn@latest add dialog
```
