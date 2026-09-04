package db

import _ "embed"

// SchemaSQL is the dev/test schema bootstrap script (see cmd/migrate) —
// every statement is CREATE TABLE IF NOT EXISTS, so applying it against the
// existing production database is always a safe no-op.
//
//go:embed schema.sql
var SchemaSQL string
