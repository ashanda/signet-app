// Package db opens the connection to the EXISTING MySQL database. This
// backend never assumes a fresh schema in production — Connect() just opens
// a pool against whatever database the operator points it at. Schema
// changes are versioned migrations applied by cmd/migrate (see
// backend/migrations).
package db

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"signet-backend/internal/config"
)

func Connect(cfg *config.Config) (*sqlx.DB, error) {
	dsn := cfg.MySQLDSN()
	conn, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to mysql: %w", err)
	}
	// Unsafe: a column with no matching struct field is ignored instead of
	// failing the query ("missing destination name ..."). The live database
	// predates this code and may carry columns the Go models don't, and
	// many queries use SELECT *; Eloquent never failed on those either.
	// Transactions started from this pool inherit the setting.
	conn = conn.Unsafe()
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(5 * time.Minute)
	return conn, nil
}
