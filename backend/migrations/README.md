# Database migrations

Versioned SQL migrations, applied by `go run ./cmd/migrate` (run from
`backend/`). Each migration runs **once per database**. Applied versions are
recorded in the `schema_migrations` table, and every later run skips them.
The files are embedded into the migrate binary, so rebuild it (or use
`go run`) after adding one.

## Commands

```bash
go run ./cmd/migrate                  # apply pending migrations (same as `up`)
go run ./cmd/migrate status           # applied / pending list (read-only)
go run ./cmd/migrate create NAME      # new NNN_NAME.sql
go run ./cmd/migrate create NAME -down  # new NNN_NAME.up.sql + NNN_NAME.down.sql
go run ./cmd/migrate rollback         # show what rolling back the last migration would run
go run ./cmd/migrate rollback -yes    # ...and actually run it
go run ./cmd/migrate baseline [N]     # record 1..N as applied WITHOUT running them
```

Every command reads the same `DB_*` environment variables as the API.

## Adding a schema change

1. `go run ./cmd/migrate create add_phone_to_users`
2. Write the SQL into the new `NNN_add_phone_to_users.sql`:
   ```sql
   ALTER TABLE users ADD COLUMN phone VARCHAR(50) NULL;
   ```
3. `go run ./cmd/migrate` applies it. Any later run reports it as
   `✓ already applied`.

Rules:

- **Never edit or renumber a migration after it has run anywhere.** Put the
  fix in a new migration. `migrate` warns when an applied file's contents
  have changed, but it never re-runs it.
- Keep **one logical change per migration**. MySQL commits DDL (`CREATE`,
  `ALTER`, `DROP`, `RENAME`, `TRUNCATE`) immediately, so it can't be rolled
  back. If statement 3 of a migration fails, statements 1–2 stay applied
  while the migration is *not* recorded. `migrate` stops and prints exactly
  what happened, and you fix the database or the file by hand before
  re-running. Data changes (`INSERT`/`UPDATE`/`DELETE`) do run inside the
  migration's transaction, so they are undone when it fails.
- File names follow `NNN_name.sql`, where the name uses lowercase letters,
  digits and underscores. A migration that has a rollback uses
  `NNN_name.up.sql` plus `NNN_name.down.sql`. Duplicate versions, invalid
  names, empty files and a down file without its up file all stop `migrate`
  before it connects to the database.
- There is no `DELIMITER` support, so stored procedures and triggers need a
  different route.

## The existing production database (baseline)

`001_initial.sql` is the schema that production already had before
versioned migrations existed. On the first run against a database that
**already has tables but no `schema_migrations` history**, `migrate` records
001 as a *baseline* **without executing it**. No existing table is created,
altered or touched. Only migrations numbered after 001 ever run there.

On an **empty** database (a fresh dev or test setup), 001 runs normally and
creates all 34 tables.

If a baselined database is missing tables that 001 defines, `migrate` lists
them as a warning but never creates them on its own. To add them, write a new
migration.

`go run ./cmd/migrate status` shows beforehand whether a baseline will happen.

## Rollback

`rollback` reverts only the **most recently applied** migration, and only
when that migration has a `.down.sql`. Without `-yes` it prints the SQL it
would run and changes nothing. It refuses to roll back a baselined migration
or one without a down file. On production, prefer a new forward migration
over a rollback. Leave a migration without a down file rather than write a
down that destroys data.

## Concurrency

`migrate` holds a MySQL named lock (`GET_LOCK`) scoped to the database while
it works. A second `migrate` against the same database waits up to
`-lock-timeout` (default 1m) and then either sees the migrations as already
applied or exits with an error. It never runs them a second time.
