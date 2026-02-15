package updater

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"
)

type Metadata struct {
	ValidFrom string `xml:"ValidFrom"`
	ValidTo   string `xml:"ValidTo"`
	Generated string `xml:"Generated"`
}

func ParseMetadata(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read metadata: %w", err)
	}
	var meta Metadata
	if err := xml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return &meta, nil
}

func (m *Metadata) ValidToTime() (time.Time, error) {
	return parseCzechDate(m.ValidTo)
}

func parseCzechDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	return time.Parse("02.01.2006", s)
}
