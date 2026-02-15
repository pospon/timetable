package search

import (
	"sort"
	"time"
)

type DepartureInfo struct {
	Line          string
	RouteType     int
	Headsign      string
	DepartureTime int
	StopName      string
}

func (idx *Index) DepartureBoard(stationID string, currentTime int, windowMinutes int, date time.Time) []DepartureInfo {
	activeServices := ActiveServices(idx.Calendars, idx.CalendarDates, date)

	var prevDayServices map[string]bool
	if currentTime < 4*3600 {
		prevDayServices = ActiveServices(idx.Calendars, idx.CalendarDates, date.AddDate(0, 0, -1))
	}

	platforms := idx.StationPlatforms[stationID]
	endTime := currentTime + windowMinutes*60
	var results []DepartureInfo

	for _, platformID := range platforms {
		departures := idx.StopDepartures[platformID]
		startIdx := sort.Search(len(departures), func(i int) bool {
			return departures[i].DepartureTime >= currentTime
		})

		for i := startIdx; i < len(departures); i++ {
			dep := departures[i]
			if dep.DepartureTime > endTime {
				break
			}
			serviceID := idx.TripService[dep.TripID]
			if !activeServices[serviceID] {
				continue
			}
			routeID := idx.TripRoute[dep.TripID]
			results = append(results, DepartureInfo{
				Line:          idx.RouteShortName[routeID],
				RouteType:     idx.RouteType[routeID],
				Headsign:      idx.TripHeadsign[dep.TripID],
				DepartureTime: dep.DepartureTime,
				StopName:      idx.StopName[platformID],
			})
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
				serviceID := idx.TripService[dep.TripID]
				if !prevDayServices[serviceID] {
					continue
				}
				routeID := idx.TripRoute[dep.TripID]
				results = append(results, DepartureInfo{
					Line:          idx.RouteShortName[routeID],
					RouteType:     idx.RouteType[routeID],
					Headsign:      idx.TripHeadsign[dep.TripID],
					DepartureTime: dep.DepartureTime - 24*3600,
					StopName:      idx.StopName[platformID],
				})
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].DepartureTime < results[j].DepartureTime
	})

	return results
}
