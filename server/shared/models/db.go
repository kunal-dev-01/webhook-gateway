package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"shared/config"

	_ "github.com/lib/pq" 
)

// DB wraps sql.DB with additional functionality
type DB struct {
	*sql.DB
	Config *config.DatabaseConfig
}

var instance *DB

// Connect establishes database connection with retry logic
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	var db *sql.DB
	var err error

	// Retry logic
	for attempt := 1; attempt <= cfg.RetryAttempts; attempt++ {
		log.Printf("🔄 Connecting to database (attempt %d/%d)...", attempt, cfg.RetryAttempts)

		db, err = sql.Open("postgres", cfg.GetDSN())
		if err != nil {
			log.Printf("❌ Failed to open database: %v", err)
			if attempt < cfg.RetryAttempts {
				log.Printf("⏳ Retrying in %v...", cfg.RetryDelay)
				time.Sleep(cfg.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to open database after %d attempts: %w", cfg.RetryAttempts, err)
		}

		// Test connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			log.Printf("❌ Failed to ping database: %v", err)
			db.Close()
			if attempt < cfg.RetryAttempts {
				log.Printf("⏳ Retrying in %v...", cfg.RetryDelay)
				time.Sleep(cfg.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to ping database after %d attempts: %w", cfg.RetryAttempts, err)
		}

		// Connection successful
		break
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	instance = &DB{
		DB:     db,
		Config: cfg,
	}

	log.Println("✅ Database connected successfully!")
	logConnectionStats(instance)

	return instance, nil
}

// GetDB returns the singleton database instance
func GetDB() *DB {
	if instance == nil {
		log.Fatal("❌ Database not initialized. Call Connect() first.")
	}
	return instance
}

// Close gracefully closes database connection
func (db *DB) Close() error {
	if db.DB != nil {
		log.Println("🔌 Closing database connection...")
		return db.DB.Close()
	}
	return nil
}

// Ping checks database connectivity
func (db *DB) Ping(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Stats returns database connection statistics
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// ExecuteWithTimeout executes query with timeout
func (db *DB) ExecuteWithTimeout(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, db.Config.QueryTimeout)
	defer cancel()

	return db.ExecContext(ctx, query, args...)
}

// QueryWithTimeout executes query with timeout
func (db *DB) QueryWithTimeout(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, db.Config.QueryTimeout)
	defer cancel()

	return db.QueryContext(ctx, query, args...)
}

// QueryRowWithTimeout executes single row query with timeout
func (db *DB) QueryRowWithTimeout(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, db.Config.QueryTimeout)
	defer cancel()

	return db.QueryRowContext(ctx, query, args...)
}

// Transaction executes function within a transaction
func (db *DB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BulkInsert performs optimized bulk insert
func (db *DB) BulkInsert(ctx context.Context, table string, columns []string, values [][]interface{}) error {
	if len(values) == 0 {
		return nil
	}

	return db.Transaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, buildBulkInsertQuery(table, columns, len(values[0])))
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, row := range values {
			if _, err := stmt.ExecContext(ctx, row...); err != nil {
				return fmt.Errorf("failed to insert row: %w", err)
			}
		}

		return nil
	})
}

// Helper: Build bulk insert query
func buildBulkInsertQuery(table string, columns []string, valueCount int) string {
	query := fmt.Sprintf("INSERT INTO %s (", table)
	for i, col := range columns {
		if i > 0 {
			query += ", "
		}
		query += col
	}
	query += ") VALUES ("
	for i := 0; i < valueCount; i++ {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+1)
	}
	query += ")"
	return query
}

// LogConnectionStats logs database statistics
func logConnectionStats(db *DB) {
	stats := db.Stats()
	log.Printf("📊 Connection Pool Stats:")
	log.Printf("   Max Open Connections: %d", db.Config.MaxOpenConns)
	log.Printf("   Max Idle Connections: %d", db.Config.MaxIdleConns)
	log.Printf("   Open Connections: %d", stats.OpenConnections)
	log.Printf("   In Use: %d", stats.InUse)
	log.Printf("   Idle: %d", stats.Idle)
}