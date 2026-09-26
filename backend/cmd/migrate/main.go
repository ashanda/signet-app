// Command migrate applies the versioned SQL migrations in backend/migrations
// (embedded into the binary) and records each one in schema_migrations, so
// every migration runs exactly once per database.
//
// Usage (from backend/):
//
//	go run ./cmd/migrate                 apply pending migrations (same as `up`)
//	go run ./cmd/migrate status          list migrations and whether each is applied
//	go run ./cmd/migrate rollback [-yes] roll back the last applied migration (dry run without -yes)
//	go run ./cmd/migrate baseline [N]    record 1..N (default 1) as applied without running them
//	go run ./cmd/migrate create NAME     create migrations/NNN_NAME.sql (-down: .up.sql + .down.sql)
//
// It reads the same DB_* environment variables as the API server (see
// internal/config). A database that already has tables but no migration
// history (the live production database) gets 001_initial.sql recorded as a
// baseline instead of executed — see migrations/README.md.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"signet-backend/internal/config"
	"signet-backend/internal/db"
	"signet-backend/internal/migrate"
	"signet-backend/migrations"
)

const usage = `Usage: go run ./cmd/migrate [command] [flags]

Commands:
  up                 apply pending migrations (default)
  status             list migrations and whether each is applied
  rollback [-yes]    roll back the last applied migration using its .down.sql
                     (prints what it would run unless -yes is given)
  baseline [N]       record migrations 1..N (default 1) as applied WITHOUT running them
  create NAME [-down] create migrations/NNN_NAME.sql (or .up.sql + .down.sql with -down)

Flags:
`

func main() {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	yes := fs.Bool("yes", false, "rollback: actually execute the down migration")
	down := fs.Bool("down", false, "create: also create a .down.sql file")
	lockTimeout := fs.Duration("lock-timeout", time.Minute, "how long to wait for another migrate process to finish")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		fs.PrintDefaults()
	}

	// Flags may appear anywhere (`migrate rollback -yes`, `create NAME -down`):
	// flag.Parse stops at the first positional, so collect it and resume.
	var args []string
	for rest := os.Args[1:]; ; {
		fs.Parse(rest)
		rest = fs.Args()
		if len(rest) == 0 {
			break
		}
		args, rest = append(args, rest[0]), rest[1:]
	}
	cmd := "up"
	if len(args) > 0 {
		cmd, args = args[0], args[1:]
	}

	if cmd == "create" {
		if len(args) != 1 {
			fail("create needs exactly one NAME, e.g. `create add_phone_to_users`")
		}
		if err := create("migrations", args[0], *down); err != nil {
			fail(err.Error())
		}
		return
	}

	// Validate every migration file before touching the database.
	migs, err := migrate.Load(migrations.FS)
	if err != nil {
		fail(err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.Load()
	conn, err := db.Connect(cfg)
	if err != nil {
		fail(fmt.Sprintf("db connect: %v", err))
	}
	defer conn.Close()
	fmt.Printf("Database: %s@%s/%s\n\n", cfg.DBUsername, cfg.DBHost, cfg.DBDatabase)

	m, err := migrate.New(ctx, conn.DB, migs, os.Stdout)
	if err != nil {
		fail(err.Error())
	}
	defer m.Close()

	if cmd == "status" {
		if len(args) != 0 {
			fail("status takes no arguments")
		}
		if err := m.Status(ctx); err != nil {
			fail(err.Error())
		}
		return
	}

	if err := m.Lock(ctx, *lockTimeout); err != nil {
		fail(err.Error())
	}
	defer m.Unlock(context.Background())

	switch cmd {
	case "up":
		err = noArgs(cmd, args)
		if err == nil {
			err = m.Up(ctx)
		}
	case "rollback":
		err = noArgs(cmd, args)
		if err == nil {
			err = m.Rollback(ctx, *yes)
		}
	case "baseline":
		upTo := uint64(migrate.BaselineVersion)
		if len(args) > 1 {
			err = fmt.Errorf("baseline takes at most one version")
		} else if len(args) == 1 {
			upTo, err = strconv.ParseUint(args[0], 10, 64)
		}
		if err == nil {
			err = m.Baseline(ctx, upTo)
		}
	default:
		fs.Usage()
		err = fmt.Errorf("unknown command %q", cmd)
	}
	if err != nil {
		m.Unlock(context.Background())
		fail(err.Error())
	}
}

func noArgs(cmd string, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("%s takes no arguments (got %s)", cmd, strings.Join(args, " "))
	}
	return nil
}

func fail(msg string) {
	fmt.Fprintf(os.Stderr, "\nmigrate: %s\n", msg)
	os.Exit(1)
}

var (
	nameRe     = regexp.MustCompile(`^[a-z0-9][a-z0-9_]*$`)
	existingRe = regexp.MustCompile(`^(\d+)_`)
)

// create writes the next-numbered migration file(s) into dir. It only reads
// file names, so it works even while another new migration is still empty.
func create(dir, name string, withDown bool) error {
	name = strings.ToLower(strings.TrimSpace(name))
	if !nameRe.MatchString(name) {
		return fmt.Errorf("invalid name %q: use lowercase letters, digits and underscores, e.g. add_phone_to_users", name)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s (run this from the backend/ directory): %w", dir, err)
	}
	var next uint64 = 1
	for _, e := range entries {
		if m := existingRe.FindStringSubmatch(e.Name()); m != nil && strings.HasSuffix(e.Name(), ".sql") {
			if v, err := strconv.ParseUint(m[1], 10, 64); err == nil && v >= next {
				next = v + 1
			}
		}
	}
	base := fmt.Sprintf("%03d_%s", next, name)

	files := map[string]string{base + ".sql": upTemplate(base)}
	if withDown {
		files = map[string]string{
			base + ".up.sql":   upTemplate(base),
			base + ".down.sql": fmt.Sprintf("-- %s (down)\n--\n-- Undo exactly what %s.up.sql does. Leave rollback unsupported\n-- (delete this file) rather than write a down that loses data.\n\n", base, base),
		}
	}
	for file, body := range files {
		p := filepath.Join(dir, file)
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		_, err = f.WriteString(body)
		f.Close()
		if err != nil {
			return err
		}
		fmt.Println("Created", p)
	}
	return nil
}

func upTemplate(base string) string {
	return fmt.Sprintf("-- %s\n--\n-- One logical change per migration: MySQL commits DDL immediately, so a\n-- migration that fails part way cannot be undone automatically.\n\n", base)
}
