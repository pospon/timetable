package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"timetable/internal/gtfs"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	db *sql.DB
}

func Open(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) IsEmpty() (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM stops").Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (s *Store) Import(feed *gtfs.Feed) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tables := []string{"stop_times", "trips", "routes", "stops", "calendar", "calendar_dates", "transfers", "feed_meta"}
	for _, t := range tables {
		if _, err := tx.Exec("DELETE FROM " + t); err != nil {
			return fmt.Errorf("clear %s: %w", t, err)
		}
	}

	if err := importStops(tx, feed.Stops); err != nil {
		return err
	}
	if err := importRoutes(tx, feed.Routes); err != nil {
		return err
	}
	if err := importTrips(tx, feed.Trips); err != nil {
		return err
	}
	if err := importCalendars(tx, feed.Calendars); err != nil {
		return err
	}
	if err := importCalendarDates(tx, feed.CalendarDates); err != nil {
		return err
	}
	if err := importStopTimes(tx, feed.StopTimes); err != nil {
		return err
	}
	if err := importTransfers(tx, feed.Transfers); err != nil {
		return err
	}

	return tx.Commit()
}

func importStops(tx *sql.Tx, stops []gtfs.Stop) error {
	stmt, err := tx.Prepare("INSERT INTO stops (stop_id, stop_code, stop_name, stop_lat, stop_lon, location_type, parent_station, wheelchair_boarding) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare stops: %w", err)
	}
	defer stmt.Close()
	for _, s := range stops {
		if _, err := stmt.Exec(s.ID, s.Code, s.Name, s.Lat, s.Lon, s.LocationType, s.ParentStation, s.WheelchairBoarding); err != nil {
			return fmt.Errorf("insert stop %s: %w", s.ID, err)
		}
	}
	return nil
}

func importRoutes(tx *sql.Tx, routes []gtfs.Route) error {
	stmt, err := tx.Prepare("INSERT INTO routes (route_id, agency_id, route_short_name, route_long_name, route_type) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare routes: %w", err)
	}
	defer stmt.Close()
	for _, r := range routes {
		if _, err := stmt.Exec(r.ID, r.AgencyID, r.ShortName, r.LongName, r.Type); err != nil {
			return fmt.Errorf("insert route %s: %w", r.ID, err)
		}
	}
	return nil
}

func importTrips(tx *sql.Tx, trips []gtfs.Trip) error {
	stmt, err := tx.Prepare("INSERT INTO trips (trip_id, route_id, service_id, trip_headsign, direction_id, shape_id, wheelchair_accessible) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare trips: %w", err)
	}
	defer stmt.Close()
	for _, t := range trips {
		if _, err := stmt.Exec(t.TripID, t.RouteID, t.ServiceID, t.Headsign, t.DirectionID, t.ShapeID, t.Wheelchair); err != nil {
			return fmt.Errorf("insert trip %s: %w", t.TripID, err)
		}
	}
	return nil
}

func importCalendars(tx *sql.Tx, cals []gtfs.Calendar) error {
	stmt, err := tx.Prepare("INSERT INTO calendar (service_id, monday, tuesday, wednesday, thursday, friday, saturday, sunday, start_date, end_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare calendar: %w", err)
	}
	defer stmt.Close()
	for _, c := range cals {
		if _, err := stmt.Exec(c.ServiceID, boolToInt(c.Monday), boolToInt(c.Tuesday), boolToInt(c.Wednesday), boolToInt(c.Thursday), boolToInt(c.Friday), boolToInt(c.Saturday), boolToInt(c.Sunday), c.StartDate, c.EndDate); err != nil {
			return fmt.Errorf("insert calendar %s: %w", c.ServiceID, err)
		}
	}
	return nil
}

func importCalendarDates(tx *sql.Tx, dates []gtfs.CalendarDate) error {
	stmt, err := tx.Prepare("INSERT INTO calendar_dates (service_id, date, exception_type) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare calendar_dates: %w", err)
	}
	defer stmt.Close()
	for _, d := range dates {
		if _, err := stmt.Exec(d.ServiceID, d.Date, d.ExceptionType); err != nil {
			return fmt.Errorf("insert calendar_date %s/%s: %w", d.ServiceID, d.Date, err)
		}
	}
	return nil
}

func importStopTimes(tx *sql.Tx, times []gtfs.StopTime) error {
	stmt, err := tx.Prepare("INSERT INTO stop_times (trip_id, arrival_time, departure_time, stop_id, stop_sequence) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare stop_times: %w", err)
	}
	defer stmt.Close()
	for _, st := range times {
		if _, err := stmt.Exec(st.TripID, st.ArrivalTime, st.DepartureTime, st.StopID, st.StopSequence); err != nil {
			return fmt.Errorf("insert stop_time %s/%d: %w", st.TripID, st.StopSequence, err)
		}
	}
	return nil
}

func importTransfers(tx *sql.Tx, transfers []gtfs.Transfer) error {
	stmt, err := tx.Prepare("INSERT INTO transfers (from_stop_id, to_stop_id, transfer_type, min_transfer_time) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare transfers: %w", err)
	}
	defer stmt.Close()
	for _, t := range transfers {
		if _, err := stmt.Exec(t.FromStopID, t.ToStopID, t.TransferType, t.MinTransferTime); err != nil {
			return fmt.Errorf("insert transfer: %w", err)
		}
	}
	return nil
}

func (s *Store) SetMeta(key, value string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO feed_meta (key, value) VALUES (?, ?)", key, value)
	return err
}

func (s *Store) GetMeta(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM feed_meta WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
