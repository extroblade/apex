#!/usr/bin/env bash
#
# Orchestrates the backend, static, and frontend compose projects, which live
# in separate files so each can be split into its own repo later. They
# communicate over a shared external Docker network created here.
#
# Usage:
#   ./dev.sh up       # create network + build & start everything (default)
#   ./dev.sh down     # stop and remove containers
#   ./dev.sh logs     # tail logs from all services
#   ./dev.sh ps       # show running services

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NET="apex-net"
BACKEND="$ROOT/backend/docker-compose.yml"
STATIC="$ROOT/static/docker-compose.yml"
NAV="$ROOT/nav/docker-compose.yml"
BFF="$ROOT/bff/docker-compose.yml"
FRONTEND="$ROOT/frontend/docker-compose.yml"

ensure_network() {
  if ! docker network inspect "$NET" >/dev/null 2>&1; then
    echo "› creating network $NET"
    docker network create "$NET" >/dev/null
  fi
}

# One-time data migration from the pre-rename volume (vcapp-backend_db_data)
# into the pinned apex-db-data volume, so the project rename loses no users.
migrate_db_volume() {
  if docker volume inspect vcapp-backend_db_data >/dev/null 2>&1 \
     && ! docker volume inspect apex-db-data >/dev/null 2>&1; then
    echo "› migrating MySQL data: vcapp-backend_db_data → apex-db-data"
    docker volume create apex-db-data >/dev/null
    docker run --rm -v vcapp-backend_db_data:/from:ro -v apex-db-data:/to alpine \
      sh -c "cp -a /from/. /to/" >/dev/null
  fi
}

cmd="${1:-up}"
case "$cmd" in
  up)
    ensure_network
    migrate_db_volume
    echo "› starting backend (db + api)"
    docker compose -f "$BACKEND" up -d --build
    echo "› starting static (avatars)"
    docker compose -f "$STATIC" up -d --build
    echo "› starting nav (menu config)"
    docker compose -f "$NAV" up -d --build
    echo "› starting bff (mobile backend-for-frontend)"
    docker compose -f "$BFF" up -d --build
    echo "› starting frontend"
    docker compose -f "$FRONTEND" up -d --build
    echo
    echo "  frontend → http://localhost:3000"
    echo "  api      → http://localhost:8080/api/health"
    echo "  bff      → http://localhost:8083/bff/health"
    ;;
  down)
    docker compose -f "$FRONTEND" down
    docker compose -f "$BFF" down
    docker compose -f "$NAV" down
    docker compose -f "$STATIC" down
    docker compose -f "$BACKEND" down
    ;;
  logs)
    docker compose -f "$BACKEND" -f "$STATIC" -f "$NAV" -f "$BFF" -f "$FRONTEND" logs -f --tail=100
    ;;
  ps)
    docker compose -f "$BACKEND" ps
    docker compose -f "$STATIC" ps
    docker compose -f "$NAV" ps
    docker compose -f "$BFF" ps
    docker compose -f "$FRONTEND" ps
    ;;
  *)
    echo "usage: ./dev.sh [up|down|logs|ps]" >&2
    exit 1
    ;;
esac
