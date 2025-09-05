// Package mysql implements a NanoAXM storage backend.
package mysql

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/micromdm/nanoaxm/storage/mysql/sqlc"
)

// MySQLStorage implements a storage.AllStorage using MySQL.
type MySQLStorage struct {
	db *sql.DB
	q  *sqlc.Queries
}

type config struct {
	driver string
	dsn    string
	db     *sql.DB
}

// Option allows configuring a MySQLStorage.
type Option func(*config)

// WithDSN sets the storage MySQL data source name.
func WithDSN(dsn string) Option {
	return func(c *config) {
		c.dsn = dsn
	}
}

// WithDriver sets a custom MySQL driver for the storage.
//
// Default driver is "mysql".
// Value is ignored if WithDB is used.
func WithDriver(driver string) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// WithDB sets a custom MySQL *sql.DB to the storage.
//
// If set, driver passed via WithDriver is ignored.
func WithDB(db *sql.DB) Option {
	return func(c *config) {
		c.db = db
	}
}

// New creates and returns a new MySQLStorage.
func New(opts ...Option) (*MySQLStorage, error) {
	cfg := &config{driver: "mysql"}
	for _, opt := range opts {
		opt(cfg)
	}
	var err error
	if cfg.db == nil {
		cfg.db, err = sql.Open(cfg.driver, cfg.dsn)
		if err != nil {
			return nil, err
		}
	}
	if err = cfg.db.Ping(); err != nil {
		return nil, err
	}
	return &MySQLStorage{db: cfg.db, q: sqlc.New(cfg.db)}, nil
}

// const timestampFormat = "2006-01-02 15:04:05"

// tx wraps g in transactions using db.
// If g returns an err the transaction will be rolled back; otherwise committed.
func tx(ctx context.Context, db *sql.DB, q *sqlc.Queries, g func(ctx context.Context, tx *sql.Tx, qtx *sqlc.Queries) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("tx begin: %w", err)
	}
	if err = g(ctx, tx, q.WithTx(tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx rollback: %w; while trying to handle error: %v", rbErr, err)
		}
		return fmt.Errorf("tx rolled back: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx commit: %w", err)
	}
	return nil
}
