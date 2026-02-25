package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"billing/internal/config"
	"shared/logger"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
	cfg *config.DatabaseConfig
}

// Connect initializes the database connection
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	// 1. Validate DSN (Lazy Connection)
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, int(cfg.ConnectTimeout.Seconds()),
	)

	dbPool, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database driver: %w", err)
	}

	// 2. Set Pool Settings
	dbPool.SetMaxOpenConns(cfg.MaxOpenConns)
	dbPool.SetMaxIdleConns(cfg.MaxIdleConns)
	dbPool.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	dbPool.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 3. Ping with Retry to ensure connection is alive
	var lastErr error
	for attempt := 1; attempt <= cfg.RetryAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
		err = dbPool.PingContext(ctx)
		cancel()

		if err == nil {
			logger.Info("Database connected successfully", "pool_max_open", cfg.MaxOpenConns)
			return &DB{DB: dbPool, cfg: cfg}, nil
		}

		lastErr = err
		logger.Error("Failed to ping database", err, "attempt", attempt)
		
		if attempt < cfg.RetryAttempts {
			time.Sleep(cfg.RetryDelay)
		}
	}

	dbPool.Close()
	return nil, fmt.Errorf("failed to connect to database: %w", lastErr)
}

// Close closes the database connection
func (db *DB) Close() error {
	logger.Info("Closing database connection...")
	return db.DB.Close()
}

// ───────────────── Wrappers (Clean Pass-through with Logging) ─────────────────

// Exec executes a query without returning any rows
func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	// Direct call - Context control is now with the Repository/Caller
	res, err := db.DB.ExecContext(ctx, query, args...)
	db.logQuery("exec", query, time.Since(start), err)
	return res, err
}

// Query executes a query that returns rows
func (db *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	db.logQuery("query", query, time.Since(start), err)
	return rows, err
}

// QueryRow executes a query that returns one row
func (db *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	db.logQuery("query_row", query, time.Since(start), nil)
	return row
}

// BeginTx starts a transaction
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.DB.BeginTx(ctx, opts)
}

// ───────────────── Logging Helper ─────────────────

func (db *DB) logQuery(operation, query string, duration time.Duration, err error) {
	if err != nil && err != sql.ErrNoRows {
		logger.Get().Database(operation, duration, err, "query_snippet", getQuerySnippet(query))
	} else if duration > 100*time.Millisecond {
		logger.Get().Database(operation, duration, nil, "query_snippet", getQuerySnippet(query), "slow_query", true)
	}
}

func getQuerySnippet(query string) string {
	if len(query) > 80 {
		return query[:80] + "..."
	}
	return query
}