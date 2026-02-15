package web

import (
	"net/http"

	"timetable/internal/updater"
)

func NewRouter(u *updater.Updater, templateDir, staticDir string) (http.Handler, error) {
	h, err := NewHandler(u, templateDir)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.HandleIndex)
	mux.HandleFunc("GET /api/stops", h.HandleStopAutocomplete)
	mux.HandleFunc("GET /search", h.HandleSearch)
	mux.HandleFunc("GET /departures", h.HandleDepartures)
	mux.HandleFunc("GET /live", h.HandleLiveBoard)
	mux.HandleFunc("GET /live/data", h.HandleLiveBoardData)
	mux.HandleFunc("GET /health", h.HandleHealth)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	return mux, nil
}
