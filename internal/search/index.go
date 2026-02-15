package search

import (
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"timetable/internal/gtfs"
)

type Departure struct {
	TripID        string
	DepartureTime int
	StopSequence  int
}

type TripStop struct {
	StopID        string
	ArrivalTime   int
	DepartureTime int
	StopSequence  int
}

type Index struct {
	StationPlatforms map[string][]string
	StopDepartures   map[string][]Departure
	TripStops        map[string][]TripStop
	TripService      map[string]string
	TripRoute        map[string]string
	TripHeadsign     map[string]string
	RouteShortName   map[string]string
	RouteType        map[string]int
	StopName         map[string]string
	Stations         []Station
	Calendars        []gtfs.Calendar
	CalendarDates    []gtfs.CalendarDate
}

type Station struct {
	ID             string
	Name           string
	NormalizedName string
}

func BuildIndex(feed *gtfs.Feed) *Index {
	idx := &Index{
		StationPlatforms: make(map[string][]string),
		StopDepartures:   make(map[string][]Departure),
		TripStops:        make(map[string][]TripStop),
		TripService:      make(map[string]string),
		TripRoute:        make(map[string]string),
		TripHeadsign:     make(map[string]string),
		RouteShortName:   make(map[string]string),
		RouteType:        make(map[string]int),
		StopName:         make(map[string]string),
		Calendars:        feed.Calendars,
		CalendarDates:    feed.CalendarDates,
	}

	for _, r := range feed.Routes {
		idx.RouteShortName[r.ID] = r.ShortName
		idx.RouteType[r.ID] = r.Type
	}

	for _, t := range feed.Trips {
		idx.TripService[t.TripID] = t.ServiceID
		idx.TripRoute[t.TripID] = t.RouteID
		idx.TripHeadsign[t.TripID] = t.Headsign
	}

	for _, s := range feed.Stops {
		idx.StopName[s.ID] = s.Name
		if s.LocationType == 1 {
			idx.Stations = append(idx.Stations, Station{
				ID:             s.ID,
				Name:           s.Name,
				NormalizedName: NormalizeCzech(s.Name),
			})
		}
		if s.ParentStation != "" {
			idx.StationPlatforms[s.ParentStation] = append(idx.StationPlatforms[s.ParentStation], s.ID)
		}
	}

	sort.Slice(idx.Stations, func(i, j int) bool {
		return idx.Stations[i].Name < idx.Stations[j].Name
	})

	for _, st := range feed.StopTimes {
		idx.StopDepartures[st.StopID] = append(idx.StopDepartures[st.StopID], Departure{
			TripID:        st.TripID,
			DepartureTime: st.DepartureTime,
			StopSequence:  st.StopSequence,
		})
		idx.TripStops[st.TripID] = append(idx.TripStops[st.TripID], TripStop{
			StopID:        st.StopID,
			ArrivalTime:   st.ArrivalTime,
			DepartureTime: st.DepartureTime,
			StopSequence:  st.StopSequence,
		})
	}

	for stopID, deps := range idx.StopDepartures {
		sort.Slice(deps, func(i, j int) bool {
			return deps[i].DepartureTime < deps[j].DepartureTime
		})
		idx.StopDepartures[stopID] = deps
	}

	for tripID, stops := range idx.TripStops {
		sort.Slice(stops, func(i, j int) bool {
			return stops[i].StopSequence < stops[j].StopSequence
		})
		idx.TripStops[tripID] = stops
	}

	return idx
}

func NormalizeCzech(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (idx *Index) SearchStations(query string) []Station {
	if query == "" {
		return nil
	}
	normalized := NormalizeCzech(query)
	var results []Station
	for _, s := range idx.Stations {
		if strings.Contains(s.NormalizedName, normalized) {
			results = append(results, s)
			if len(results) >= 10 {
				break
			}
		}
	}
	return results
}
