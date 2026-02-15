package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"timetable/internal/search"
	"timetable/internal/updater"
)

type Handler struct {
	updater   *updater.Updater
	templates *template.Template
}

func NewHandler(u *updater.Updater, templateDir string) (*Handler, error) {
	funcMap := template.FuncMap{
		"formatTime": search.FormatTime,
		"formatDuration": func(seconds int) string {
			return strconv.Itoa(seconds/60) + " min"
		},
		"routeTypeIcon": func(rt int) string {
			switch rt {
			case 0:
				return "tram"
			case 3:
				return "bus"
			default:
				return "bus"
			}
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &Handler{updater: u, templates: tmpl}, nil
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *Handler) HandleStopAutocomplete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	idx := h.updater.Index()
	stations := idx.SearchStations(query)

	type stationResult struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	results := make([]stationResult, len(stations))
	for i, s := range stations {
		results[i] = stationResult{ID: s.ID, Name: s.Name}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	fromID := r.URL.Query().Get("from")
	toID := r.URL.Query().Get("to")
	windowStr := r.URL.Query().Get("window")
	timeStr := r.URL.Query().Get("time")

	window := 60
	if w, err := strconv.Atoi(windowStr); err == nil && w > 0 {
		window = w
	}

	now := time.Now()
	currentTime := now.Hour()*3600 + now.Minute()*60 + now.Second()
	date := now

	if timeStr != "" {
		if t, err := time.Parse("15:04", timeStr); err == nil {
			currentTime = t.Hour()*3600 + t.Minute()*60
		}
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = d
		}
	}

	idx := h.updater.Index()
	connections := idx.FindConnections(fromID, toID, currentTime, window, date)

	fromName := ""
	toName := ""
	if name, ok := idx.StopName[fromID]; ok {
		fromName = name
	}
	if name, ok := idx.StopName[toID]; ok {
		toName = name
	}

	data := struct {
		Connections []search.Connection
		FromName    string
		ToName      string
		Count       int
	}{
		Connections: connections,
		FromName:    fromName,
		ToName:      toName,
		Count:       len(connections),
	}

	h.templates.ExecuteTemplate(w, "results.html", data)
}

func (h *Handler) HandleDepartures(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station")
	windowStr := r.URL.Query().Get("window")
	timeStr := r.URL.Query().Get("time")

	window := 60
	if w, err := strconv.Atoi(windowStr); err == nil && w > 0 {
		window = w
	}

	now := time.Now()
	currentTime := now.Hour()*3600 + now.Minute()*60 + now.Second()
	date := now

	if timeStr != "" {
		if t, err := time.Parse("15:04", timeStr); err == nil {
			currentTime = t.Hour()*3600 + t.Minute()*60
		}
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			date = d
		}
	}

	idx := h.updater.Index()
	departures := idx.DepartureBoard(stationID, currentTime, window, date)

	stationName := ""
	if name, ok := idx.StopName[stationID]; ok {
		stationName = name
	}

	data := struct {
		Departures  []search.DepartureInfo
		StationName string
		Count       int
	}{
		Departures:  departures,
		StationName: stationName,
		Count:       len(departures),
	}

	h.templates.ExecuteTemplate(w, "departures.html", data)
}

func (h *Handler) HandleLiveBoard(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "liveboard.html", nil)
}

func (h *Handler) HandleLiveBoardData(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	currentTime := now.Hour()*3600 + now.Minute()*60 + now.Second()

	idx := h.updater.Index()
	connections := idx.FindConnections("11311", "911", currentTime, 60, now)

	data := struct {
		Connections []search.Connection
		UpdatedAt   string
		Count       int
	}{
		Connections: connections,
		UpdatedAt:   now.Format("15:04:05"),
		Count:       len(connections),
	}

	h.templates.ExecuteTemplate(w, "liveboard_data.html", data)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	idx := h.updater.Index()
	data := map[string]interface{}{
		"status":   "ok",
		"stations": len(idx.Stations),
		"trips":    len(idx.TripService),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
