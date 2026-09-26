// Package migrations holds the versioned SQL migrations, embedded into the
// migrate binary so a deployed build always carries its own schema history.
// See README.md in this directory for naming and authoring rules.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
