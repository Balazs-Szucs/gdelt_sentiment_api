package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"sentiment_dashboard_api/internal/database"
	"sentiment_dashboard_api/internal/gdelt"
	"sentiment_dashboard_api/internal/geography"
	"sentiment_dashboard_api/internal/server"
)

func main() {
	db := database.New()
	defer db.Close()

	geo := geography.NewProcessor()
	if err := geo.LoadCountryGeoJSON("./cmd/api/countries.geo.json"); err != nil {
		log.Fatalf("Failed to load GeoJSON: %v", err)
	}

	gdeltService := gdelt.NewService(geo, db.DB())
	if err := gdeltService.Refresh(); err != nil {
		log.Fatalf("Initial GDELT data load failed: %v", err)
	}
	gdeltService.StartAutoRefresh(15 * time.Minute)
	println("Auto refresh started")
	gdeltService.StartDailyReset()

	srv := server.NewServer(gdeltService)

	done := make(chan bool, 1)
	go gracefulShutdown(srv, done, gdeltService)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server error: %v", err)
	}

	<-done
	log.Println("Graceful shutdown complete.")
}

func gracefulShutdown(srv *http.Server, done chan bool, gdeltService *gdelt.Service) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	gdeltService.StopAutoRefresh()

	done <- true
}
