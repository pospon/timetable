package updater

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"timetable/internal/gtfs"
	"timetable/internal/search"
	"timetable/internal/store"
)

type Updater struct {
	dataDir   string
	dbPath    string
	sourceURL string
	index     atomic.Value
	store     *store.Store
}

func New(dataDir, dbPath, sourceURL string) *Updater {
	return &Updater{
		dataDir:   dataDir,
		dbPath:    dbPath,
		sourceURL: sourceURL,
	}
}

func (u *Updater) LoadOrImport() (*search.Index, error) {
	s, err := store.Open(u.dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	u.store = s

	empty, err := s.IsEmpty()
	if err != nil {
		return nil, fmt.Errorf("check empty: %w", err)
	}

	if empty {
		log.Println("Database empty, importing GTFS data...")
		if err := u.importFromDisk(); err != nil {
			return nil, fmt.Errorf("import: %w", err)
		}
	}

	log.Println("Building in-memory index...")
	feed, err := gtfs.ParseFeed(u.dataDir)
	if err != nil {
		return nil, fmt.Errorf("parse feed for index: %w", err)
	}

	idx := search.BuildIndex(feed)
	u.index.Store(idx)
	log.Printf("Index built: %d stations, %d trips", len(idx.Stations), len(idx.TripService))
	return idx, nil
}

func (u *Updater) Index() *search.Index {
	return u.index.Load().(*search.Index)
}

func (u *Updater) StartBackgroundCheck() {
	if u.sourceURL == "" {
		log.Println("No GTFS_SOURCE_URL configured, skipping periodic updates")
		return
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			u.checkAndUpdate()
		}
	}()
}

func (u *Updater) checkAndUpdate() {
	metaPath := filepath.Join(u.dataDir, "metadata.xml")
	meta, err := ParseMetadata(metaPath)
	if err != nil {
		log.Printf("Warning: cannot read metadata: %v", err)
		return
	}

	validTo, err := meta.ValidToTime()
	if err != nil {
		log.Printf("Warning: cannot parse ValidTo: %v", err)
		return
	}

	daysLeft := time.Until(validTo).Hours() / 24
	if daysLeft > 3 {
		log.Printf("Feed valid until %s (%.0f days left), no update needed", meta.ValidTo, daysLeft)
		return
	}

	log.Printf("Feed expires in %.0f days, downloading update...", daysLeft)
	if err := u.downloadAndReload(); err != nil {
		log.Printf("Update failed: %v", err)
	}
}

func (u *Updater) downloadAndReload() error {
	resp, err := http.Get(u.sourceURL)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	for _, f := range zr.File {
		outPath := filepath.Join(u.dataDir, f.Name)
		if err := extractFile(f, outPath); err != nil {
			return fmt.Errorf("extract %s: %w", f.Name, err)
		}
	}

	if err := u.importFromDisk(); err != nil {
		return fmt.Errorf("reimport: %w", err)
	}

	feed, err := gtfs.ParseFeed(u.dataDir)
	if err != nil {
		return fmt.Errorf("rebuild index: %w", err)
	}

	idx := search.BuildIndex(feed)
	u.index.Store(idx)
	log.Printf("Updated index: %d stations, %d trips", len(idx.Stations), len(idx.TripService))
	return nil
}

func (u *Updater) importFromDisk() error {
	feed, err := gtfs.ParseFeed(u.dataDir)
	if err != nil {
		return err
	}
	return u.store.Import(feed)
}

func extractFile(f *zip.File, outPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

func (u *Updater) Close() error {
	if u.store != nil {
		return u.store.Close()
	}
	return nil
}
