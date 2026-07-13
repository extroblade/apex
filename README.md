# full-stack-vc-app

Full-stack starter.

- **Backend** — Go 1.23, [chi](https://github.com/go-chi/chi) router, `database/sql` + MySQL driver.
- **Frontend** — React + TypeScript, built with [rsbuild](https://rsbuild.dev). Routing via [wouter](https://github.com/molefrog/wouter), client state via [zustand](https://zustand.docs.pmnd.rs), server state via [TanStack Query](https://tanstack.com/query), UI via [shadcn/ui](https://ui.shadcn.com) + Tailwind CSS v4 + clsx. Organized with [Feature-Sliced Design](https://feature-sliced.design).
- **Database** — MySQL 8.4 (Dockerized).
- **Docker** — every service containerized. The backend (`backend/docker-compose.yml`) and frontend (`frontend/docker-compose.yml`) are **separate compose projects** joined by a shared external network, so either can be split into its own repo later. `./dev.sh` runs both with one command.

## Layout

```
.
├── backend/                  # Go API
│   ├── cmd/server/           # main entrypoint
│   ├── internal/
│   │   ├── config/           # env-based config
│   │   ├── db/               # MySQL connection pool
│   │   ├── handler/          # HTTP handlers (health, users)
│   │   ├── middleware/       # CORS, etc.
│   │   └── server/           # router wiring
│   ├── migrations/           # *.sql, applied on first DB init
│   └── Dockerfile
├── frontend/                 # React SPA (FSD)
│   ├── src/
│   │   ├── app/              # init: providers, router, styles, entry
│   │   ├── pages/            # home, about
│   │   ├── widgets/          # header
│   │   ├── features/         # (empty — add user interactions here)
│   │   ├── entities/         # user (api + model)
│   │   └── shared/           # ui (shadcn), lib (cn), api, config, store
│   ├── rsbuild.config.ts
│   ├── nginx.conf            # prod: serves SPA, proxies /api → backend
│   ├── docker-compose.yml    # frontend project
│   └── Dockerfile
├── backend/docker-compose.yml  # db + api project
└── dev.sh                      # runs both projects together
```

## Run everything (Docker)

```bash
./dev.sh up       # creates the shared network, builds & starts both projects
./dev.sh down     # stop everything
./dev.sh logs     # tail all logs
```

The two stacks are separate compose projects on a shared external network
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
| POST   | `/api/fuel/plan`      | —    | Compute a fuel & stint strategy      |
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
