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
| `GTFS_DATA_DIR` | `.` | Directory containing GTFS .txt files |
| `DB_PATH` | `<GTFS_DATA_DIR>/timetable.db` | SQLite database path |
| `TEMPLATE_DIR` | `web/templates` | HTML template directory |
| `STATIC_DIR` | `web/static` | Static assets directory |
| `GTFS_SOURCE_URL` | _(empty)_ | URL to download fresh GTFS zip (enables auto-update) |

## Docker

```bash
# Copy GTFS files into data/ directory first
make docker-build
docker compose up
```

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
