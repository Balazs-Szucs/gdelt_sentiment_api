package gdelt

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"sentiment_dashboard_api/internal/geography"
	"sentiment_dashboard_api/internal/models"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Service struct {
	mu              sync.RWMutex
	events          []models.Event
	geoProcessor    *geography.Processor
	refreshInterval time.Duration
	stopChan        chan struct{}
	db              *sql.DB
}

func NewService(geo *geography.Processor, db *sql.DB) *Service {
	return &Service{
		geoProcessor: geo,
		stopChan:     make(chan struct{}),
		db:           db,
	}
}

func (s *Service) StartAutoRefresh(interval time.Duration) {
	s.refreshInterval = interval
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := s.Refresh(); err != nil {
					log.Printf("GDELT auto-refresh failed: %v", err)
				}
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Service) StopAutoRefresh() {
	close(s.stopChan)
}

func (s *Service) Refresh() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.fetchLatestData()
	if err != nil {
		return fmt.Errorf("fetch data: %w", err)
	}

	newEvents, err := s.parseCSV(data)
	if err != nil {
		return fmt.Errorf("parse data: %w", err)
	}

	s.events = newEvents

	if err := s.storeEventsInDB(newEvents); err != nil {
		return fmt.Errorf("store events in db: %w", err)
	}
	println("Refreshed GDELT data at ", time.Now().Format(time.RFC3339))
	return nil
}

func (s *Service) storeEventsInDB(events []models.Event) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO events (GlobalEventID, Date, SourceActor, TargetActor, EventCode, EventRootCode, GoldsteinScale, AvgTone, NumMentions, NumSources, NumArticles, SourceURL, Lat, Lng, Country) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		_, err := stmt.Exec(
			event.GlobalEventID,
			event.Date,
			event.SourceActor.Code,
			event.TargetActor.Code,
			event.EventCode,
			event.EventRootCode,
			event.GoldsteinScale,
			event.AvgTone,
			event.NumMentions,
			event.NumSources,
			event.NumArticles,
			event.SourceURL,
			event.Lat,
			event.Lng,
			event.Country,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) StartDailyReset() {
	go func() {
		for {
			now := time.Now()
			next := now.AddDate(0, 0, 1)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			time.Sleep(time.Until(next))

			s.mu.Lock()
			if _, err := s.db.Exec("DELETE FROM events"); err != nil {
				log.Printf("Failed to reset events: %v", err)
			}
			s.mu.Unlock()
		}
	}()
}

func (s *Service) fetchLatestData() ([]byte, error) {
	resp, err := http.Get("http://data.gdeltproject.org/gdeltv2/lastupdate.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lastupdate.txt: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var csvURL string
	for _, line := range strings.Split(string(body), "\n") {
		if strings.Contains(line, "export.CSV.zip") {
			parts := strings.Fields(line)
			csvURL = parts[len(parts)-1]
			break
		}
	}

	resp2, err := http.Get(csvURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download CSV: %w", err)
	}
	defer resp2.Body.Close()

	zipData, _ := io.ReadAll(resp2.Body)
	zipReader, _ := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))

	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".export.CSV") {
			file, _ := f.Open()
			defer file.Close()
			return io.ReadAll(file)
		}
	}
	return nil, fmt.Errorf("no CSV file in ZIP")
}

func (s *Service) parseCSV(data []byte) ([]models.Event, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1

	records, _ := reader.ReadAll()
	var result []models.Event

	for _, rec := range records {
		if len(rec) < 61 {
			continue
		}

		lat := parseFloat(rec[56])
		lng := parseFloat(rec[57])

		result = append(result, models.Event{
			GlobalEventID:  rec[0],
			Date:           rec[1],
			SourceActor:    parseActor(rec[5], rec[6], rec[7]),
			TargetActor:    parseActor(rec[15], rec[16], rec[17]),
			EventCode:      rec[26],
			EventRootCode:  rec[28],
			GoldsteinScale: parseFloat(rec[30]),
			AvgTone:        parseFloat(rec[34]),
			NumMentions:    parseInt(rec[31]),
			NumSources:     parseInt(rec[32]),
			NumArticles:    parseInt(rec[33]),
			SourceURL:      rec[60],
			Lat:            lat,
			Lng:            lng,
			Country:        s.geoProcessor.GetCountry(lat, lng),
		})
	}
	return result, nil
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseActor(code, name, countryCode string) models.Actor {
	return models.Actor{
		Code:        code,
		Name:        name,
		CountryCode: countryCode,
	}
}

func (s *Service) GetEvents() []models.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.events
}
