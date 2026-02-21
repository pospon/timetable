# Timetable DPMLJ

Public transit timetable search for DPMLJ (Liberec & Jablonec nad Nisou, Czech Republic).

Search connections between stops, view departure boards, and check a live board for specific routes — all powered by GTFS data.

## Features

- **Connection search** — find direct connections between two stops within a time window
- **Departure board** — view all departures from a station
- **Live board** — real-time auto-refreshing view for Melantrichova → Fügnerova
- **Stop autocomplete** — Czech diacritics-aware search (e.g. "fug" matches "Fügnerova")
- **After-midnight handling** — trips with times >24:00 correctly appear in early morning searches
- **Automatic GTFS updates** — periodic check and reload when feed approaches expiration

## Stack

- Go (single binary, no runtime dependencies)
- SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- HTMX for interactive UI without JavaScript frameworks
- GTFS data from DPMLJ

## Running locally

```bash
# Build and run (GTFS .txt files must be in the working directory)
make run

# Or directly
go build -o timetable ./cmd/timetable
./timetable
```

Open http://localhost:8080

## Configuration

| Environment variable | Default | Description |
|---|---|---|
| `LISTEN_ADDR` | `:8080` | HTTP listen address |
| `GTFS_DATA_DIR` | `gtfs` | Directory containing GTFS .txt files |
| `DB_PATH` | `<GTFS_DATA_DIR>/timetable.db` | SQLite database path |
| `TEMPLATE_DIR` | `web/templates` | HTML template directory |
| `STATIC_DIR` | `web/static` | Static assets directory |
| `GTFS_SOURCE_URL` | `http://www.dpmlj.cz/gtfs.zip` | URL to download fresh GTFS zip for auto-update |
| `GTFS_SEED_DIR` | `/app/gtfs-seed` | (Docker only) Initial GTFS files baked into the image |

## Docker

```bash
make docker-build
docker compose up
```

## Deployment

Live at **https://odjezdy.poposkoc.cz**

The app runs on a Raspberry Pi 5 via Docker, deployed automatically through GitHub Actions:

1. Push to `main` triggers `.github/workflows/deploy.yml`
2. GitHub Actions builds a `linux/arm64` Docker image and pushes it to GHCR
3. Watchtower on the Pi polls GHCR every 5 minutes and auto-updates the container
4. Traefik handles HTTPS via Let's Encrypt

Infrastructure config (docker-compose.yml, Traefik) lives in the [homeserver](https://github.com/pospon/homeserver) repo.

The container uses a persistent Docker volume (`timetable-data`) mounted at `/data` for the SQLite database and GTFS files. GTFS seed data is baked into the image and copied on first run.

## Project structure

```
cmd/timetable/main.go        Entry point
internal/gtfs/                GTFS data model and CSV parser
internal/store/               SQLite persistence
internal/search/              In-memory indexes, connection search, departure board
internal/updater/             Periodic GTFS download and reload
internal/web/                 HTTP handlers and routing
web/templates/                HTML templates
web/static/                   CSS
```
