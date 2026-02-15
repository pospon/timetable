package gtfs

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseFeed(dir string) (*Feed, error) {
	feed := &Feed{}

	var err error
	if feed.Stops, err = parseStops(filepath.Join(dir, "stops.txt")); err != nil {
		return nil, fmt.Errorf("stops: %w", err)
	}
	if feed.Routes, err = parseRoutes(filepath.Join(dir, "routes.txt")); err != nil {
		return nil, fmt.Errorf("routes: %w", err)
	}
	if feed.Trips, err = parseTrips(filepath.Join(dir, "trips.txt")); err != nil {
		return nil, fmt.Errorf("trips: %w", err)
	}
	if feed.Calendars, err = parseCalendars(filepath.Join(dir, "calendar.txt")); err != nil {
		return nil, fmt.Errorf("calendar: %w", err)
	}
	if feed.CalendarDates, err = parseCalendarDates(filepath.Join(dir, "calendar_dates.txt")); err != nil {
		return nil, fmt.Errorf("calendar_dates: %w", err)
	}
	if feed.StopTimes, err = parseStopTimes(filepath.Join(dir, "stop_times.txt")); err != nil {
		return nil, fmt.Errorf("stop_times: %w", err)
	}
	if feed.Transfers, err = parseTransfers(filepath.Join(dir, "transfers.txt")); err != nil {
		return nil, fmt.Errorf("transfers: %w", err)
	}

	return feed, nil
}

func openCSV(path string) (*csv.Reader, io.Closer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	return r, f, nil
}

func readHeader(r *csv.Reader) (map[string]int, error) {
	row, err := r.Read()
	if err != nil {
		return nil, err
	}
	idx := make(map[string]int, len(row))
	for i, col := range row {
		idx[strings.TrimSpace(col)] = i
	}
	return idx, nil
}

func col(row []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func colInt(row []string, idx map[string]int, name string) int {
	v, _ := strconv.Atoi(col(row, idx, name))
	return v
}

func colFloat(row []string, idx map[string]int, name string) float64 {
	v, _ := strconv.ParseFloat(col(row, idx, name), 64)
	return v
}

func ParseTimeToSeconds(s string) int {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	sec, _ := strconv.Atoi(parts[2])
	return h*3600 + m*60 + sec
}

func parseStops(path string) ([]Stop, error) {
	r, closer, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	idx, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	var stops []Stop
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		stops = append(stops, Stop{
			ID:                 col(row, idx, "stop_id"),
			Code:               col(row, idx, "stop_code"),
			Name:               col(row, idx, "stop_name"),
			Lat:                colFloat(row, idx, "stop_lat"),
			Lon:                colFloat(row, idx, "stop_lon"),
			LocationType:       colInt(row, idx, "location_type"),
			ParentStation:      col(row, idx, "parent_station"),
			WheelchairBoarding: colInt(row, idx, "wheelchair_boarding"),
		})
	}
	return stops, nil
}

func parseRoutes(path string) ([]Route, error) {
	r, closer, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	idx, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	var routes []Route
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		routes = append(routes, Route{
			ID:        col(row, idx, "route_id"),
			AgencyID:  col(row, idx, "agency_id"),
			ShortName: col(row, idx, "route_short_name"),
			LongName:  col(row, idx, "route_long_name"),
			Type:      colInt(row, idx, "route_type"),
		})
	}
	return routes, nil
}

func parseTrips(path string) ([]Trip, error) {
	r, closer, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	idx, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	var trips []Trip
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		trips = append(trips, Trip{
			RouteID:     col(row, idx, "route_id"),
			ServiceID:   col(row, idx, "service_id"),
			TripID:      col(row, idx, "trip_id"),
			Headsign:    col(row, idx, "trip_headsign"),
			DirectionID: colInt(row, idx, "direction_id"),
			ShapeID:     col(row, idx, "shape_id"),
			Wheelchair:  colInt(row, idx, "wheelchair_accessible"),
		})
	}
	return trips, nil
}

func parseCalendars(path string) ([]Calendar, error) {
	r, closer, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	idx, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	var cals []Calendar
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		cals = append(cals, Calendar{
			ServiceID: col(row, idx, "service_id"),
			Monday:    col(row, idx, "monday") == "1",
			Tuesday:   col(row, idx, "tuesday") == "1",
			Wednesday: col(row, idx, "wednesday") == "1",
			Thursday:  col(row, idx, "thursday") == "1",
			Friday:    col(row, idx, "friday") == "1",
			Saturday:  col(row, idx, "saturday") == "1",
			Sunday:    col(row, idx, "sunday") == "1",
			StartDate: col(row, idx, "start_date"),
			EndDate:   col(row, idx, "end_date"),
		})
	}
	return cals, nil
}

func parseCalendarDates(path string) ([]CalendarDate, error) {
	r, closer, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	idx, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	var dates []CalendarDate
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		dates = append(dates, CalendarDate{
			ServiceID:     col(row, idx, "service_id"),
			Date:          col(row, idx, "date"),
			ExceptionType: colInt(row, idx, "exception_type"),
		})
	}
	return dates, nil
}

func parseStopTimes(path string) ([]StopTime, error) {
	r, closer, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	idx, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	var times []StopTime
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		times = append(times, StopTime{
			TripID:        col(row, idx, "trip_id"),
			ArrivalTime:   ParseTimeToSeconds(col(row, idx, "arrival_time")),
			DepartureTime: ParseTimeToSeconds(col(row, idx, "departure_time")),
			StopID:        col(row, idx, "stop_id"),
			StopSequence:  colInt(row, idx, "stop_sequence"),
		})
	}
	return times, nil
}

func parseTransfers(path string) ([]Transfer, error) {
	r, closer, err := openCSV(path)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	idx, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	var transfers []Transfer
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, Transfer{
			FromStopID:      col(row, idx, "from_stop_id"),
			ToStopID:        col(row, idx, "to_stop_id"),
			TransferType:    colInt(row, idx, "transfer_type"),
			MinTransferTime: colInt(row, idx, "min_transfer_time"),
		})
	}
	return transfers, nil
}
