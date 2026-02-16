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

// Connect initializes the database connection with retry logic
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	var db *sql.DB
	var err error

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode, int(cfg.ConnectTimeout.Seconds()),
	)

	for attempt := 1; attempt <= cfg.RetryAttempts; attempt++ {
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

// ───────────────── Context Helper ─────────────────

// withQueryContext creates a context with timeout for DB operations
// It ALWAYS uses background context to avoid HTTP request cancellation issues
func (db *DB) withQueryContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), db.cfg.QueryTimeout)
}

// ───────────────── Instrumented Wrappers ─────────────────

// Exec executes a query without returning any rows
func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	
	// Use background context with timeout for independence
	queryCtx, cancel := db.withQueryContext()
	defer cancel()
	
	res, err := db.DB.ExecContext(queryCtx, query, args...)
	
	db.logQuery("exec", query, time.Since(start), err)
	return res, err
}

// Query executes a query that returns rows
func (db *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	// Use background context with timeout for independence
	queryCtx, _ := db.withQueryContext()
	rows, err := db.DB.QueryContext(queryCtx, query, args...)
	db.logQuery("query", query, time.Since(start), err)
	return rows, err
}

// QueryRow executes a query that is expected to return at most one row
func (db *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	queryCtx, _ := db.withQueryContext()
	// No defer - context will timeout automatically
	row := db.DB.QueryRowContext(queryCtx, query, args...)
	db.logQuery("query_row", query, time.Since(start), nil)
	return row
}

// ───────────────── Transaction Support ─────────────────

// BeginTx starts a transaction
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// For transactions, use background context but with longer timeout
	txCtx, cancel := context.WithTimeout(context.Background(), db.cfg.QueryTimeout*5)
	defer cancel()
	
	return db.DB.BeginTx(txCtx, opts)
}

// ───────────────── Logging Helper ─────────────────

func (db *DB) logQuery(operation, query string, duration time.Duration, err error) {
	if err != nil && err != sql.ErrNoRows {
		logger.Get().Database(operation, duration, err, 
			"query_snippet", getQuerySnippet(query),
		)
	} else {
		// Only log slow queries in production (> 100ms)
		if duration > 100*time.Millisecond {
			logger.Get().Database(operation, duration, nil,
				"query_snippet", getQuerySnippet(query),
				"slow_query", true,
			)
		}
	}
}

// Helper to keep logs clean
func getQuerySnippet(query string) string {
	if len(query) > 80 {
		return query[:80] + "..."
	}
	return query
}

// ───────────────── Health Check ─────────────────

// Ping checks if database is reachable
func (db *DB) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.DB.PingContext(pingCtx)
}

// Stats returns connection pool statistics
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}