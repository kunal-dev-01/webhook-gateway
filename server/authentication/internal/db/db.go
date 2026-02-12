package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"shared/config" 
	"shared/logger"    

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
	cfg *config.DatabaseConfig
}

// Connect initializes the database connection with retry logic and custom logging
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	var db *sql.DB
	var err error

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, int(cfg.ConnectTimeout.Seconds()),
	)

	for attempt := 1; attempt <= cfg.RetryAttempts; attempt++ {
		// Logger use kar rahe hain instead of fmt.Printf
		logger.Info("Connecting to database...", "attempt", attempt, "max_attempts", cfg.RetryAttempts)

		db, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.Error("Failed to open database", err, "attempt", attempt)
			time.Sleep(cfg.RetryDelay)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			logger.Error("Failed to ping database", err, "attempt", attempt)
			db.Close()
			time.Sleep(cfg.RetryDelay)
			continue
		}

		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", cfg.RetryAttempts, err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	logger.Info("Database connected successfully",
		"pool_max_open", cfg.MaxOpenConns,
		"pool_max_idle", cfg.MaxIdleConns,
	)

	return &DB{DB: db, cfg: cfg}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	logger.Info("Closing database connection...")
	return db.DB.Close()
}

// ───────────────── Instrumented Wrappers ─────────────────

// Exec executes a query without returning any rows
func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	var err error

	// Defer logger execution to capture duration and error
	defer func() {
		logger.Get().Database("exec", time.Since(start), err, "query_snippet", getQuerySnippet(query))
	}()

	ctx, cancel := context.WithTimeout(ctx, db.cfg.QueryTimeout)
	defer cancel()

	res, execErr := db.DB.ExecContext(ctx, query, args...)
	err = execErr // Assign to outer err for the deferred logger
	return res, err
}

// Query executes a query that returns rows
func (db *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	var err error

	defer func() {
		logger.Get().Database("query", time.Since(start), err, "query_snippet", getQuerySnippet(query))
	}()

	ctx, cancel := context.WithTimeout(ctx, db.cfg.QueryTimeout)
	defer cancel()

	rows, queryErr := db.DB.QueryContext(ctx, query, args...)
	err = queryErr
	return rows, err
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	
	// QueryRow errors are returned when Scanning, but we can log the attempt
	defer func() {
		logger.Get().Database("query_row", time.Since(start), nil, "query_snippet", getQuerySnippet(query))
	}()

	ctx, cancel := context.WithTimeout(ctx, db.cfg.QueryTimeout)
	defer cancel()

	return db.DB.QueryRowContext(ctx, query, args...)
}

// Helper to keep logs clean
func getQuerySnippet(query string) string {
	if len(query) > 50 {
		return query[:50] + "..."
	}
	return query
}