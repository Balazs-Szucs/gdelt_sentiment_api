package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"sentiment_dashboard_api/internal/database"
	"sentiment_dashboard_api/internal/gdelt"
)

type Server struct {
	port         int
	db           database.Service
	gdeltService *gdelt.Service
}

func NewServer(gdeltService *gdelt.Service) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port:         port,
		db:           database.New(),
		gdeltService: gdeltService,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

func (s *Server) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
