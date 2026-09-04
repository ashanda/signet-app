// Package db opens the connection to the EXISTING MySQL database. This
// backend never assumes a fresh schema in production — Connect() just opens
// a pool against whatever database the operator points it at. See
// cmd/migrate for the optional dev-only schema bootstrap.
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
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(5 * time.Minute)
	return conn, nil
}
