CREATE TABLE IF NOT EXISTS stops (
    stop_id TEXT PRIMARY KEY,
    stop_code TEXT,
    stop_name TEXT NOT NULL,
    stop_lat REAL,
    stop_lon REAL,
    location_type INTEGER NOT NULL DEFAULT 0,
    parent_station TEXT,
    wheelchair_boarding INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS routes (
    route_id TEXT PRIMARY KEY,
    agency_id TEXT,
    route_short_name TEXT,
    route_long_name TEXT,
    route_type INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS trips (
    trip_id TEXT PRIMARY KEY,
    route_id TEXT NOT NULL,
    service_id TEXT NOT NULL,
    trip_headsign TEXT,
    direction_id INTEGER,
    shape_id TEXT,
    wheelchair_accessible INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stop_times (
    trip_id TEXT NOT NULL,
    arrival_time INTEGER NOT NULL,
    departure_time INTEGER NOT NULL,
    stop_id TEXT NOT NULL,
    stop_sequence INTEGER NOT NULL,
    PRIMARY KEY (trip_id, stop_sequence)
);

CREATE TABLE IF NOT EXISTS calendar (
    service_id TEXT PRIMARY KEY,
    monday INTEGER NOT NULL,
    tuesday INTEGER NOT NULL,
    wednesday INTEGER NOT NULL,
    thursday INTEGER NOT NULL,
    friday INTEGER NOT NULL,
    saturday INTEGER NOT NULL,
    sunday INTEGER NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS calendar_dates (
    service_id TEXT NOT NULL,
    date TEXT NOT NULL,
    exception_type INTEGER NOT NULL,
    PRIMARY KEY (service_id, date)
);

CREATE TABLE IF NOT EXISTS transfers (
    from_stop_id TEXT NOT NULL,
    to_stop_id TEXT NOT NULL,
    transfer_type INTEGER NOT NULL,
    min_transfer_time INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (from_stop_id, to_stop_id)
);

CREATE TABLE IF NOT EXISTS feed_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_stop_times_stop_dep ON stop_times(stop_id, departure_time);
CREATE INDEX IF NOT EXISTS idx_stops_parent ON stops(parent_station);
CREATE INDEX IF NOT EXISTS idx_trips_service ON trips(service_id);
CREATE INDEX IF NOT EXISTS idx_stop_times_trip ON stop_times(trip_id, stop_sequence);
