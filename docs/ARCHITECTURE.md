# Signet — Go + Vue rebuild architecture

Source of truth for behavior: schema.md, api_spec.md, financial_engine.md, ui_spec.md in this
same directory (produced by deep-reading `crypto-app-src`, the original Laravel 12 app).

## Stack

- **Backend:** Go 1.24, `chi` router (stdlib-compatible, lightweight), `sqlx` + `go-sql-driver/mysql`
  for the DB layer (raw SQL, not a heavy ORM — several columns in the live schema are
  intentionally-quirky, e.g. string-typed `earn_logs.amount`; an ORM's type assumptions would
  fight that). `golang-jwt/jwt` for session tokens, `golang.org/x/crypto/bcrypt` for password
  hashing (bcrypt is algorithm-compatible with PHP's `password_hash`/Laravel `Hash::make` —
  existing user password hashes in the live DB will verify correctly without a rehash/migration).
- **Frontend:** Vue 3 + Vite, `vue-router`, `pinia` for state, `axios` for API calls,
  Bootstrap 5 (matching the original's actual styling stack per ui_spec.md — Tailwind was present
  in the original repo but unused, so it's dropped here), SweetAlert2 for the same
  confirm/toast UX, Chart-free (original has no live charts).
- **Database:** the EXISTING MySQL database (`signet_last` per `.env` — configurable via env
  vars), connected directly. No new migrations are run against it by default; the backend
  provides an optional `migrate` subcommand that can create the schema from scratch on a fresh
  DB (matching schema.md exactly) for dev/test setups, but production points at the live DB
  as-is and never runs destructive migrations.

## Auth strategy

Two independent principal types, matching the original's two-guard design:

1. **End-user session** (`web` guard equivalent): login issues a JWT (HS256, `SESSION_SECRET`
   env var) containing `{user_id, role}`, set as an **httpOnly, SameSite=Lax, Secure-in-prod**
   cookie named `signet_session`. The Vue SPA never reads the token directly; it calls
   `GET /api/me` to learn who's logged in. This mirrors Laravel's session-cookie web auth
   closely enough for the SPA to behave the same (login redirects to role dashboard, logout
   clears the cookie) without needing PHP session-table compatibility.
2. **API-client bearer token** (`api`/Sanctum guard equivalent, `ApiUser` model): unchanged
   surface — `POST /api/v1/token` exchanges `username`+`password` (checked against `api_users`)
   for a token, stored in `personal_access_tokens` in the same shape Sanctum uses
   (`tokenable_type='App\Models\ApiUser'`, `tokenable_id`, sha256-hashed token, plaintext
   returned once as `{id}|{plaintext}`), so existing issued tokens and the existing table keep
   working. All `/api/v1/*` integration endpoints (check-user, list users, user details) require
   this bearer token and operate on the `users` table, exactly as in the original.

**Deliberate fix, disclosed, not silent:** the original left ~15 financially-sensitive routes
with no auth middleware at all (relying on a null check that would just crash for a guest) —
see api_spec.md Cross-cutting note 2. The Go rewrite puts these behind the same session-auth
requirement the rest of the authenticated app uses. This is a security fix, not a business-rule
change (the business logic executed by each endpoint is unchanged) — flagged here per the
project's own "preserve business rules" instruction, since it's the one place behavior
intentionally diverges from the literal original.

**Preserved as-is (not "fixed"):** the two divergent wallet-crediting code paths
(`WalletService.updateWallet`, cap-gated; vs. the private `TokenController.updateWallet`,
not cap-gated) are ported as two separate functions in `internal/wallet`, each wired to the
exact same call sites the original used (see financial_engine.md + api_spec.md). Likewise the
rank-threshold hardcoded dates, the `user_id` sentinel values (1, 2/3/4/5), string-typed numeric
columns, and every other quirk documented in schema.md/financial_engine.md are preserved
verbatim — these are business rules/data shape, not bugs to silently fix.

## Backend layout

```
backend/
  cmd/api/main.go            entrypoint: load config, connect DB, build router, serve
  internal/config/           env-var loading
  internal/db/               sqlx.DB setup, migration files (dev-only bootstrap)
  internal/models/           one file per table group — structs matching schema.md exactly
  internal/auth/             JWT session auth, bcrypt, Sanctum-compatible token auth, RoleMiddleware equivalent
  internal/wallet/           WalletService.updateWallet + TokenController-style updateWallet, GlobalShareWallet logic
  internal/tree/             ParentFind/superParentFind genealogy placement, checkWalet, tokenShare, rank()
  internal/handlers/         one file per controller domain (auth, admin, agent, company, user,
                              package, wallet(unused/stub), token, mining, kyc, geneology, roc,
                              salary, directshare, executive, leader, countries, userparentlogs,
                              earnlog, apiv1)
  internal/jobs/             the 5 console commands ported as either HTTP-triggered admin
                              actions and/or a lightweight in-process scheduler (share:calculate,
                              mining:send-webhook, mining:update, packages:weekly-sum,
                              users:generate-secret-keys)
  internal/httpx/            small helpers: JSON responses, pagination envelope matching
                              Laravel's paginator shape, error helpers
  go.mod
```

## API surface

Every original Blade-rendered page becomes a `GET` JSON endpoint returning exactly the data the
page used to receive (documented per-entry in api_spec.md as "Response: view X with ..."), and
every original POST/PUT/DELETE keeps its business logic and validation rules, but always returns
JSON (no more redirect+flash — the Vue SPA renders SweetAlert2 itself from the JSON body).
Routes are grouped under `/api/v1/...` mirroring the original path names where sensible (e.g.
`/api/v1/company/dashboard`, `/api/v1/kyc`, `/api/v1/salaries`) so the api_spec.md document
doubles as the endpoint reference with minimal renaming. Pagination responses use the same
field names as Laravel's paginator (`current_page, data, per_page, total, last_page, ...`) so
the documented "page data contract" in api_spec.md §10 maps directly.

## Frontend layout

```
frontend/
  src/
    api/            axios instance + one module per domain (auth.js, company.js, ...)
    router/          vue-router routes matching the role/route map in ui_spec.md, with a
                      navigation guard replicating RoleMiddleware (redirect to /login or 403 page)
    stores/          pinia: auth store (current user/role), toast/alert helper store
    layouts/         AppLayout (sidebar+topbar, mirrors layouts/app.blade.php +
                      layouts/sidebar.blade.php + layouts/topbar.blade.php), AuthLayout
                      (centered-card chrome for login/register/kyc-create etc.), MarketingLayout
                      (welcome page)
    components/       shared: RankBadgeRow (rank() helper), RocSummary (roc() helper),
                      MiningWidget (the polling mining card), DataTable/Paginator, StatCard,
                      ActivatePackageButton, ToastConfirm (the repeated SweetAlert2 toast pattern)
    views/            one folder per role/domain mirroring resources/views/ 1:1 (admin/, agent/,
                      company/, user/, auth/, kyc/, packages/, tokens/, salaries/, geneology/,
                      countries/, leaders/, executives/, leader_code_logs/, executive_code_logs/, earn/)
  index.html, vite.config.js, package.json
```

## Build/run

- Backend: `go run ./cmd/api` (env-configured: `DB_HOST/DB_PORT/DB_DATABASE/DB_USERNAME/DB_PASSWORD`,
  `SESSION_SECRET`, `APP_URL`, `MINING_WEBHOOK_*`, `MAIL_*` — mirroring the original `.env` keys
  where they map to a real Go equivalent).
- Frontend: `npm install && npm run dev` (Vite dev server proxying `/api` to the Go backend),
  `npm run build` for production static assets, served either by the Go backend as static files
  or any static host.

## What's intentionally out of scope for a literal 1:1 (flagged, not silently dropped)

- Laravel-specific infra with no Go equivalent needed for behavior parity: the queue-driven
  mail send for `NewUserPackageMail` is ported as a synchronous (or simple goroutine-async) SMTP
  send rather than a DB-backed job queue — same end effect (an email goes out), simpler
  mechanism. `Opcodes\LogViewer`, `laravel/nightwatch`, `laravel/pail` are dev-ops tooling with
  no user-facing behavior — not ported.
- Pusher/Reverb websocket broadcasting for `MiningUpdated` is replaced by the same polling
  pattern the frontend already relies on for reconciliation (`GET /mining/user/{id}` every 5s,
  per ui_spec.md's Mining Widget note) — a websocket push can be added later without changing
  the data shape, but the original UI already tolerates poll-only (that's literally how it
  reconciles today), so this isn't a functional gap.
