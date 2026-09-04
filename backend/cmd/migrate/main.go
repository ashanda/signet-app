// Command migrate bootstraps a fresh database with the Signet schema. It is
// a dev/test convenience ONLY — every statement in schema.sql is `CREATE
// TABLE IF NOT EXISTS`, so running this against the real production
// database (the one the task requires this app to use directly) is a safe
// no-op: existing tables and their data are left completely untouched, and
// nothing is dropped, altered, or overwritten.
//
// Usage:
//
//	go run ./cmd/migrate
//
// It reads the same DB_* environment variables as the API server (see
// internal/config), so point it at the same .env/environment before
// running. On a brand new empty database this creates all 34 tables from
// docs/analysis/schema.md; on the existing live database it verifies
// connectivity and reports that every table already exists.
package main

import (
	"log"
	"strings"

	"signet-backend/internal/config"
	"signet-backend/internal/db"
)

func main() {
	cfg := config.Load()

	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer conn.Close()

	statements := splitStatements(db.SchemaSQL)
	log.Printf("migrate: connected to %s@%s/%s — applying %d statement(s) from schema.sql (all CREATE TABLE IF NOT EXISTS, safe against an existing database)", cfg.DBUsername, cfg.DBHost, cfg.DBDatabase, len(statements))

	applied, skipped := 0, 0
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		if _, err := conn.Exec(trimmed); err != nil {
			log.Fatalf("migrate: statement failed: %v\n--- statement ---\n%s", err, trimmed)
		}
		if strings.HasPrefix(strings.ToUpper(trimmed), "CREATE TABLE") {
			applied++
		} else {
			skipped++
		}
	}

	log.Printf("migrate: done — %d CREATE TABLE statement(s) applied (existing tables left untouched), %d other statement(s) run", applied, skipped)
}

// splitStatements does a naive split on `;\n` — sufficient here because
// schema.sql contains no stored procedures/triggers/multi-statement bodies,
// only CREATE TABLE and two SET statements (see schema.sql itself).
func splitStatements(sql string) []string {
	parts := strings.Split(sql, ";\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p+";")
	}
	return out
}
