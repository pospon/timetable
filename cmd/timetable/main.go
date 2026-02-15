package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"timetable/internal/updater"
	"timetable/internal/web"
)

func main() {
	dataDir := envOrDefault("GTFS_DATA_DIR", "gtfs")
	dbPath := envOrDefault("DB_PATH", filepath.Join(dataDir, "timetable.db"))
	sourceURL := os.Getenv("GTFS_SOURCE_URL")
	addr := envOrDefault("LISTEN_ADDR", ":8080")
	templateDir := envOrDefault("TEMPLATE_DIR", "web/templates")
	staticDir := envOrDefault("STATIC_DIR", "web/static")

	u := updater.New(dataDir, dbPath, sourceURL)
	defer u.Close()

	if _, err := u.LoadOrImport(); err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	u.StartBackgroundCheck()

	router, err := web.NewRouter(u, templateDir, staticDir)
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}

	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
