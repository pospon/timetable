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

## Render deployment

The app is configured for Render via `render.yaml`.

### Free plan (default)

- GTFS data is baked into the Docker image and copied to `/tmp/gtfs` on startup
- SQLite DB is rebuilt on every cold start (the free tier spins down after inactivity)
- Auto-update checks run but won't survive restarts

### Paid plan with persistent storage

To switch to a paid plan with a persistent disk (data survives restarts and deploys):

1. In `render.yaml`, change `plan: free` to `plan: starter`
2. Uncomment the `disk` and env override sections in `render.yaml`
3. Set the env vars in Render dashboard or via the blueprint:
   - `GTFS_DATA_DIR=/data/gtfs`
   - `DB_PATH=/data/timetable.db`
4. The 1 GB disk mounted at `/data` will persist the SQLite DB and downloaded GTFS updates

With persistent storage, the DB is built once and reused across restarts. The updater downloads fresh GTFS data from DPMLJ when the feed approaches its expiration date (within 3 days of `ValidTo`).

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
