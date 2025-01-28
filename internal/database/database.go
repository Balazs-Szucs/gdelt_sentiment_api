package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

// Service represents a service that interacts with a database.
type Service interface {
	Health() map[string]string
	Close() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Begin() (*sql.Tx, error)
	DB() *sql.DB
}

type service struct {
	db *sql.DB
}

var (
	dburl      = os.Getenv("DB_URL")
	dbInstance *service
)

// New creates a new database connection and returns a Service instance.
func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}

	if err := dbInstance.createEventsTable(); err != nil {
		log.Fatalf("Failed to create events table: %v", err)
	}

	return dbInstance
}

// createEventsTable creates the events table if it doesn't exist.
func (s *service) createEventsTable() error {
	query := `
 CREATE TABLE IF NOT EXISTS events (
  GlobalEventID TEXT PRIMARY KEY,
  Date TEXT,
  SourceActor TEXT,
  TargetActor TEXT,
  EventCode TEXT,
  EventRootCode TEXT,
  GoldsteinScale REAL,
  AvgTone REAL,
  NumMentions INTEGER,
  NumSources INTEGER,
  NumArticles INTEGER,
  SourceURL TEXT,
  Lat REAL,
  Lng REAL,
  Country TEXT
 );`
	_, err := s.db.Exec(query)
	return err
}

// Health checks the health of the database connection by pinging the database.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	return s.db.Close()
}

// Exec executes a query without returning any rows.
func (s *service) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

// Query executes a query that returns rows.
func (s *service) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (s *service) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(query, args...)
}

// Begin starts a transaction.
func (s *service) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}

// DB returns the underlying *sql.DB instance.
func (s *service) DB() *sql.DB {
	return s.db
}
