package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sentiment_dashboard_api/internal/models"
	"strconv"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", s.EventsHandler)
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/gdelt/events", s.gdeltEventsHandler)
	mux.HandleFunc("/refresh", s.gdeltRefreshHandler)

	return s.corsMiddleware(mux)
}

func (s *Server) gdeltEventsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filtered := s.filterGdeltEvents(query)

	if query.Get("all") == "true" {
		s.respondJSON(w, map[string]interface{}{
			"total":   len(filtered),
			"results": filtered,
		})
		return
	}

	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 50
	}

	start := page * limit
	end := start + limit
	if start >= len(filtered) {
		start = 0
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	s.respondJSON(w, map[string]interface{}{
		"total":   len(filtered),
		"page":    page,
		"results": filtered[start:end],
	})
}
func (s *Server) EventsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT * FROM events")
	if err != nil {
		http.Error(w, "Failed to query events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.GlobalEventID,
			&event.Date,
			&event.SourceActor.Code,
			&event.TargetActor.Code,
			&event.EventCode,
			&event.EventRootCode,
			&event.GoldsteinScale,
			&event.AvgTone,
			&event.NumMentions,
			&event.NumSources,
			&event.NumArticles,
			&event.SourceURL,
			&event.Lat,
			&event.Lng,
			&event.Country,
		); err != nil {
			http.Error(w, "Failed to scan event", http.StatusInternalServerError)
			return
		}
		events = append(events, event)
	}

	s.respondJSON(w, events)
}
func (s *Server) gdeltRefreshHandler(w http.ResponseWriter, r *http.Request) {
	go func() {
		if err := s.gdeltService.Refresh(); err != nil {
			log.Printf("GDELT manual refresh failed: %v", err)
		}
	}()
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) filterGdeltEvents(query map[string][]string) []models.Event {
	return s.gdeltService.GetEvents()
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
