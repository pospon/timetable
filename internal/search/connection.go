package search

import (
	"fmt"
	"sort"
	"time"
)

type Connection struct {
	Line          string
	RouteType     int
	Headsign      string
	DepartureTime int
	ArrivalTime   int
	Duration      int
	FromStop      string
	ToStop        string
}

func FormatTime(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if h >= 24 {
		h -= 24
	}
	return fmt.Sprintf("%02d:%02d", h, m)
}

func (idx *Index) FindConnections(fromStationID, toStationID string, currentTime int, windowMinutes int, date time.Time) []Connection {
	activeServices := ActiveServices(idx.Calendars, idx.CalendarDates, date)

	var prevDayServices map[string]bool
	if currentTime < 4*3600 {
		prevDayServices = ActiveServices(idx.Calendars, idx.CalendarDates, date.AddDate(0, 0, -1))
	}

	fromPlatforms := idx.StationPlatforms[fromStationID]
	toPlatformSet := make(map[string]bool)
	for _, p := range idx.StationPlatforms[toStationID] {
		toPlatformSet[p] = true
	}

	endTime := currentTime + windowMinutes*60

	var connections []Connection

	for _, platformID := range fromPlatforms {
		departures := idx.StopDepartures[platformID]
		startIdx := sort.Search(len(departures), func(i int) bool {
			return departures[i].DepartureTime >= currentTime
		})

		for i := startIdx; i < len(departures); i++ {
			dep := departures[i]
			if dep.DepartureTime > endTime {
				break
			}
			if c, ok := idx.checkTrip(dep, toPlatformSet, activeServices); ok {
				connections = append(connections, c)
			}
		}

		if prevDayServices != nil {
			searchFrom := currentTime + 24*3600
			searchTo := endTime + 24*3600

			startIdx := sort.Search(len(departures), func(i int) bool {
				return departures[i].DepartureTime >= searchFrom
			})

			for i := startIdx; i < len(departures); i++ {
				dep := departures[i]
				if dep.DepartureTime > searchTo {
					break
				}
				if c, ok := idx.checkTrip(dep, toPlatformSet, prevDayServices); ok {
					c.DepartureTime -= 24 * 3600
					c.ArrivalTime -= 24 * 3600
					connections = append(connections, c)
				}
			}
		}
	}

	sort.Slice(connections, func(i, j int) bool {
		return connections[i].DepartureTime < connections[j].DepartureTime
	})

	return connections
}

func (idx *Index) checkTrip(dep Departure, toPlatformSet map[string]bool, activeServices map[string]bool) (Connection, bool) {
	serviceID := idx.TripService[dep.TripID]
	if !activeServices[serviceID] {
		return Connection{}, false
	}

	tripStops := idx.TripStops[dep.TripID]
	for _, ts := range tripStops {
		if ts.StopSequence <= dep.StopSequence {
			continue
		}
		if toPlatformSet[ts.StopID] {
			routeID := idx.TripRoute[dep.TripID]
			return Connection{
				Line:          idx.RouteShortName[routeID],
				RouteType:     idx.RouteType[routeID],
				Headsign:      idx.TripHeadsign[dep.TripID],
				DepartureTime: dep.DepartureTime,
				ArrivalTime:   ts.ArrivalTime,
				Duration:      ts.ArrivalTime - dep.DepartureTime,
				FromStop:      idx.StopName[dep.TripID],
				ToStop:        idx.StopName[ts.StopID],
			}, true
		}
	}
	return Connection{}, false
}
