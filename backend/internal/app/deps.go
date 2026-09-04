// Package app holds the shared dependency bundle every handler package
// needs, so each `internal/handlers/*.go` file can expose a uniform
// `RegisterXRoutes(r chi.Router, d *app.Deps)` function that cmd/api/main.go
// wires up. Keeping this in its own tiny package (rather than in
// internal/handlers itself) avoids an import cycle between handlers and
// whatever mounts them.
package app

import (
	"github.com/jmoiron/sqlx"

	"signet-backend/internal/auth"
	"signet-backend/internal/config"
)

type Deps struct {
	DB   *sqlx.DB
	Auth *auth.Service
	Cfg  *config.Config
}
