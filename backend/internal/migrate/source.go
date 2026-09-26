package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Migration is one versioned schema change. Up is required; Down is
// optional and only used by rollback.
type Migration struct {
	Version  uint64
	Name     string // e.g. "add_phone_to_users"
	UpFile   string // e.g. "005_add_phone_to_users.sql" or "005_add_phone_to_users.up.sql"
	DownFile string // "" when the migration has no down file
	Up       string
	Down     string
	Checksum string // sha256 of Up, line endings normalised
}

// fileRe matches NNN_name.sql, NNN_name.up.sql and NNN_name.down.sql.
var fileRe = regexp.MustCompile(`^(\d+)_([a-z0-9][a-z0-9_]*?)(?:\.(up|down))?\.sql$`)

// Load reads every *.sql file in the root of fsys and returns the
// migrations sorted by version. Any file it cannot account for is an
// error, so a typo never turns into a silently skipped migration.
func Load(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	byVersion := map[uint64]*Migration{}
	for _, e := range entries {
		file := e.Name()
		if e.IsDir() || path.Ext(file) != ".sql" {
			continue
		}
		m := fileRe.FindStringSubmatch(file)
		if m == nil {
			return nil, fmt.Errorf("invalid migration file name %q (expected NNN_name.sql, NNN_name.up.sql or NNN_name.down.sql; name in lowercase letters, digits and underscores)", file)
		}
		version, err := strconv.ParseUint(m[1], 10, 64)
		if err != nil || version == 0 {
			return nil, fmt.Errorf("invalid migration version in %q (must be a positive number)", file)
		}
		name, kind := m[2], m[3]

		raw, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		body := strings.ReplaceAll(string(raw), "\r\n", "\n")
		stmts, err := SplitStatements(body)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		if len(stmts) == 0 {
			return nil, fmt.Errorf("%s contains no SQL statements", file)
		}

		mig := byVersion[version]
		if mig == nil {
			mig = &Migration{Version: version, Name: name}
			byVersion[version] = mig
		} else if mig.Name != name {
			return nil, fmt.Errorf("duplicate migration version %d: %q and %q", version, displayFile(mig), file)
		}
		if kind == "down" {
			if mig.DownFile != "" {
				return nil, fmt.Errorf("duplicate down migration for version %d: %q and %q", version, mig.DownFile, file)
			}
			mig.DownFile, mig.Down = file, body
		} else {
			if mig.UpFile != "" {
				return nil, fmt.Errorf("duplicate migration version %d: %q and %q", version, mig.UpFile, file)
			}
			sum := sha256.Sum256([]byte(body))
			mig.UpFile, mig.Up, mig.Checksum = file, body, hex.EncodeToString(sum[:])
		}
	}

	out := make([]Migration, 0, len(byVersion))
	for _, mig := range byVersion {
		if mig.UpFile == "" {
			return nil, fmt.Errorf("%s has no matching up migration (%03d_%s.up.sql)", mig.DownFile, mig.Version, mig.Name)
		}
		if mig.DownFile != "" && !strings.HasSuffix(mig.UpFile, ".up.sql") {
			return nil, fmt.Errorf("%s has a down file, so rename it to %03d_%s.up.sql", mig.UpFile, mig.Version, mig.Name)
		}
		out = append(out, *mig)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

func displayFile(m *Migration) string {
	if m.UpFile != "" {
		return m.UpFile
	}
	return m.DownFile
}
