# Database Schema — `signet_last` (MySQL)

Reverse-engineered from `database/migrations/*.php` (56 files, applied in
filename/timestamp order, cumulatively — i.e. this is the schema Laravel's
migrator produces after running every migration in order on a fresh
database) plus `database/seeders/*.php` and
`database/factories/UserFactory.php`.

This is the **final state** of every table. Column *order* within a table
is derived by simulating each migration's `->after('col')` placement (or
append-to-end when no `->after()` is given); order is noted for
completeness but has no functional effect on a Go/SQL rebuild.

Laravel type → MySQL type notes used throughout:
- `id()` → `BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY`
- `foreignId('x')` → `BIGINT UNSIGNED` (only becomes a real FK if chained
  with `->constrained()` / `->references()->on()`; several columns in this
  app use `foreignId()` purely as a naming convention for an unsigned
  bigint with **no actual foreign-key constraint** — flagged explicitly
  below wherever that's the case)
- `string()` → `VARCHAR(255)` unless a length is given
- `boolean()` → `TINYINT(1)`
- `rememberToken()` → `VARCHAR(100) NULL`
- `timestamps()` → `created_at TIMESTAMP NULL`, `updated_at TIMESTAMP NULL`
- `softDeletes()` → `deleted_at TIMESTAMP NULL`
- `morphs('x')` → `x_type VARCHAR(255) NOT NULL`, `x_id BIGINT UNSIGNED NOT NULL`, plus a composite index on `(x_type, x_id)`
- FK with no `->onDelete(...)` specified → MySQL default (RESTRICT / NO ACTION)

---

## users

| Column | Type | Nullable | Default | Unsigned | Extra |
|---|---|---|---|---|---|
| id | BIGINT | NO | — | YES | PK, AUTO_INCREMENT |
| country_id | BIGINT | YES | NULL | YES | FK → countries.id |
| name | VARCHAR(255) | NO | — | — | |
| email | VARCHAR(255) | NO | — | — | UNIQUE |
| roc_status | VARCHAR(255) | YES | NULL | — | |
| global_director_share | DECIMAL(15,2) | NO | 0 | — | |
| global_director_share_status | TINYINT(1) | NO | 0 (false) | — | |
| email_verified_at | TIMESTAMP | YES | NULL | — | |
| password | VARCHAR(255) | NO | — | — | |
| remember_token | VARCHAR(100) | YES | NULL | — | |
| created_at | TIMESTAMP | YES | NULL | — | |
| updated_at | TIMESTAMP | YES | NULL | — | |
| role | VARCHAR(255) | NO | 'user' | — | |
| whatsapp_number | VARCHAR(255) | YES | NULL | — | |
| binance_pay_id | VARCHAR(255) | YES | NULL | — | |
| status | ENUM('pending','active','inactive') | NO | 'pending' | — | |
| referred_by | BIGINT | YES | NULL | YES | FK → users.id (self-referential) |
| leader_code | VARCHAR(255) | YES | NULL | — | |
| leader_status | VARCHAR(255) | NO | 'deactive' | — | |
| executive_code | VARCHAR(255) | YES | NULL | — | |
| google_authenticator_secret | VARCHAR(255) | YES | NULL | — | |
| on_vacation | TINYINT(1) | NO | 0 (false) | — | |

- **PK**: `id`
- **Unique**: `email`
- **FK**: `country_id` → `countries.id`, `ON DELETE SET NULL` (`nullOnDelete()`)
- **FK**: `referred_by` → `users.id`, `ON DELETE SET NULL`
- No soft deletes. No explicit index beyond the FKs' implicit indexes and the `email` unique index.
- `status` enum encodes registration/approval state; `role` is a free-text string (values used in seeders: `'user'` default, `'company'` for the admin — see Notes).

---

## password_reset_tokens

(Laravel's default reset-token table, created alongside `users` in the same migration.)

| Column | Type | Nullable | Default |
|---|---|---|---|
| email | VARCHAR(255) | NO | — |
| token | VARCHAR(255) | NO | — |
| created_at | TIMESTAMP | YES | NULL |

- **PK**: `email`
- No further migrations touch this table.

---

## sessions

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | VARCHAR(255) | NO | — | — |
| user_id | BIGINT | YES | NULL | YES |
| ip_address | VARCHAR(45) | YES | NULL | — |
| user_agent | TEXT | YES | NULL | — |
| payload | LONGTEXT | NO | — | — |
| last_activity | INT | NO | — | — |

- **PK**: `id`
- **Index**: `user_id` (plain index, not a FK constraint)
- **Index**: `last_activity`

---

## cache

| Column | Type | Nullable |
|---|---|---|
| key | VARCHAR(255) | NO |
| value | MEDIUMTEXT | NO |
| expiration | INT | NO |

- **PK**: `key`

## cache_locks

| Column | Type | Nullable |
|---|---|---|
| key | VARCHAR(255) | NO |
| owner | VARCHAR(255) | NO |
| expiration | INT | NO |

- **PK**: `key`

---

## jobs

| Column | Type | Nullable | Unsigned |
|---|---|---|---|
| id | BIGINT | NO | YES (AUTO_INCREMENT) |
| queue | VARCHAR(255) | NO | — |
| payload | LONGTEXT | NO | — |
| attempts | TINYINT | NO | YES |
| reserved_at | INT | YES | YES |
| available_at | INT | NO | YES |
| created_at | INT | NO | YES |

- **PK**: `id`
- **Index**: `queue`

## job_batches

| Column | Type | Nullable |
|---|---|---|
| id | VARCHAR(255) | NO |
| name | VARCHAR(255) | NO |
| total_jobs | INT | NO |
| pending_jobs | INT | NO |
| failed_jobs | INT | NO |
| failed_job_ids | LONGTEXT | NO |
| options | MEDIUMTEXT | YES |
| cancelled_at | INT | YES |
| created_at | INT | NO |
| finished_at | INT | YES |

- **PK**: `id`

## failed_jobs

| Column | Type | Nullable | Default |
|---|---|---|---|
| id | BIGINT UNSIGNED | NO | AUTO_INCREMENT |
| uuid | VARCHAR(255) | NO | — |
| connection | TEXT | NO | — |
| queue | TEXT | NO | — |
| payload | LONGTEXT | NO | — |
| exception | LONGTEXT | NO | — |
| failed_at | TIMESTAMP | NO | CURRENT_TIMESTAMP |

- **PK**: `id`
- **Unique**: `uuid`

---

## password_reset

A **second**, separate password-reset table (distinct from `password_reset_tokens` above), created by its own migration. Both tables coexist in the final schema.

| Column | Type | Nullable |
|---|---|---|
| email | VARCHAR(255) | NO |
| token | VARCHAR(255) | NO |
| created_at | TIMESTAMP | YES |

- **No primary key at all** — the migration never calls `->primary()` or `$table->id()`. This is very likely unintentional in the original app, but it is the actual schema; replicate it as-is (no PK) unless the live DB shows otherwise.
- **Index**: `email` (plain index)

---

## referral_codes

| Column | Type | Nullable | Unsigned |
|---|---|---|---|
| id | BIGINT | NO | YES (AUTO_INCREMENT) |
| user_id | BIGINT | NO | YES |
| code | VARCHAR(255) | NO | — |
| created_at | TIMESTAMP | YES | — |
| updated_at | TIMESTAMP | YES | — |

- **PK**: `id`
- **Unique**: `code`
- **FK**: `user_id` → `users.id`, no `onDelete` specified (MySQL default RESTRICT)

---

## user_packages

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | NO | — | YES |
| package | VARCHAR(255) | NO | — | — |
| ref_id | VARCHAR(255) | YES | NULL | — |
| status | ENUM('pending','active','deactivate') | NO | 'pending' | — |
| company_status | TINYINT | NO | 0 | — |
| earn | VARCHAR(255) | NO | '0' | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |
| activated_at | TIMESTAMP | YES | NULL | — |
| sale | ENUM('first','other') | NO | 'other' | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- `earn` is declared as a **string** column (`$table->string('earn')->default(0)`), not decimal — the default `0` is coerced to the string `'0'`. Preserve as a text/varchar column, not numeric, unless the running app actually stores numeric-looking strings that should be parsed.
- `status` = package lifecycle; `sale` distinguishes first purchase vs. subsequent; `company_status` is a tinyint flag (0/1-style), not an enum — no named values given in the migration.

---

## tokens

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | NO | — | YES |
| token | VARCHAR(255) | NO | — | — |
| status | ENUM('active','deactive') | NO | 'active' | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **Unique**: `token`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`

---

## wallets

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | NO | — | YES |
| balance | DECIMAL(10,2) | NO | 0 | — |
| roc_balance | DECIMAL(12,2) | NO | 0 | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- `roc_balance` migration guards with `Schema::hasColumn('wallets','roc_balance')` — no-op on an already-migrated DB, harmless; final result is the column exists as above.

---

## user_parents

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | NO | — | YES |
| virtual_id | BIGINT | NO | 0 | YES |
| parent_id | BIGINT | NO | 0 | YES |
| node | ENUM('active','deactive','gratitude','correct') | NO | 'deactive' | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- `virtual_id` and `parent_id` are declared with `foreignId()` (so they're `BIGINT UNSIGNED`) but **neither has `->constrained()`** — they are **not real foreign keys**, just unsigned bigints defaulting to `0`. Do not add FK constraints for these two in the Go rebuild unless the live DB shows otherwise.
- `node` enum evolved across three migrations (see Notes): started as `('active','deactive')`, then `('active','deactive','gratitude')`, then finally `('active','deactive','gratitude','correct')` — the last is authoritative.

---

## packages

| Column | Type | Nullable | Default |
|---|---|---|---|
| id | BIGINT UNSIGNED | NO | AUTO_INCREMENT |
| name | VARCHAR(255) | NO | — |
| price | INT | NO | — |
| commission | INT | NO | — |
| status | VARCHAR(255) | NO | 'active' |
| created_at | TIMESTAMP | YES | NULL |
| updated_at | TIMESTAMP | YES | NULL |
| rank | VARCHAR(255) | YES | NULL |
| telegram_link | VARCHAR(255) | YES | NULL |

- **PK**: `id`
- `status` is a plain string column (not an enum) — the seeder/app convention uses `'active'`.
- `price`/`commission` are plain signed `INT` (not decimal, not unsigned) — commission looks like a percentage (seeder values 20–90).
- `rank` is a free-text label (seeder values: "Beginner", "Trader", "Senior Trader", "Pro Trader", "Grand Master Trader", "Tycoon", "Pro Tycoon", "Master Tycoon", "Grand Master Tycoon").

---

## earn_logs

| Column | Type | Nullable | Default |
|---|---|---|---|
| id | BIGINT UNSIGNED | NO | AUTO_INCREMENT |
| user_id | VARCHAR(255) | NO | — |
| amount | VARCHAR(255) | NO | — |
| description | VARCHAR(255) | YES | NULL |
| created_at | TIMESTAMP | YES | NULL |
| updated_at | TIMESTAMP | YES | NULL |

- **PK**: `id`
- **No FK, no index on `user_id`.** Both `user_id` and `amount` are declared as plain **strings**, not `unsignedBigInteger`/`decimal` — inconsistent with the rest of the schema but this is what the migration specifies. Preserve as varchar columns in Go/SQL; do not silently promote to int/decimal without confirming against live data (see Notes).

---

## token_logs

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | NO | — | YES |
| shared_by | BIGINT | NO | — | YES |
| amount | DECIMAL(10,2) | NO | 0.00 | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- **FK**: `shared_by` → `users.id`, `ON DELETE CASCADE`

---

## kycs

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | NO | — | YES |
| full_name | VARCHAR(255) | NO | — | — |
| email | VARCHAR(255) | NO | — | — |
| contact_number1 | VARCHAR(255) | NO | — | — |
| contact_number2 | VARCHAR(255) | YES | NULL | — |
| address | VARCHAR(255) | NO | — | — |
| telegram_username | VARCHAR(255) | NO | — | — |
| document_type | ENUM('nic','passport') | NO | — | — |
| document_number | VARCHAR(255) | NO | — | — |
| nic_front | VARCHAR(255) | YES | NULL | — |
| nic_back | VARCHAR(255) | YES | NULL | — |
| passport_image | VARCHAR(255) | YES | NULL | — |
| is_verified | TINYINT(1) | NO | 0 (false) | — |
| comments | TEXT | YES | NULL | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- **Unique**: `document_number`
- `document_type` has no default — must be supplied on insert.

---

## personal_access_tokens

(Laravel Sanctum's standard table.)

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| tokenable_type | VARCHAR(255) | NO | — | — |
| tokenable_id | BIGINT | NO | — | YES |
| name | TEXT | NO | — | — |
| token | VARCHAR(64) | NO | — | — |
| abilities | TEXT | YES | NULL | — |
| last_used_at | TIMESTAMP | YES | NULL | — |
| expires_at | TIMESTAMP | YES | NULL | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **Unique**: `token`
- **Index**: composite `(tokenable_type, tokenable_id)` (from `morphs()`)

---

## api_users

| Column | Type | Nullable |
|---|---|---|
| id | BIGINT UNSIGNED | NO (AUTO_INCREMENT) |
| username | VARCHAR(255) | NO |
| password | VARCHAR(255) | NO |
| remember_token | VARCHAR(100) | YES |
| created_at | TIMESTAMP | YES |
| updated_at | TIMESTAMP | YES |

- **PK**: `id`
- **Unique**: `username`
- Separate auth table from `users`, used for the API-user login (see `ApiUserSeeder`).

---

## super_parent_logs

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| current_user_id | BIGINT | NO | — | YES |
| package_id | BIGINT | NO | — | YES |
| gratitude_user | BIGINT | NO | 0 | — (signed) |
| user_package | BIGINT | YES | NULL | YES |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **No FK constraints at all** — `current_user_id`, `package_id`, `user_package` are plain unsigned-bigint columns with no `->constrained()`/`->foreign()` calls despite the naming. `gratitude_user` is a signed `BIGINT` (declared with `bigInteger()`, not `unsignedBigInteger()`).

---

## weekly_package_summaries

| Column | Type | Nullable | Default |
|---|---|---|---|
| id | BIGINT UNSIGNED | NO | AUTO_INCREMENT |
| job_id | VARCHAR(255) | NO | — |
| week_start | DATE | NO | — |
| week_end | DATE | NO | — |
| per_week_total | DECIMAL(12,2) | NO | 0 |
| balance_forward | DECIMAL(12,2) | NO | 0 |
| roc_potation | INT UNSIGNED | NO | 0 |
| total_amount | DECIMAL(12,2) | NO | 0 |
| created_at | TIMESTAMP | YES | NULL |
| updated_at | TIMESTAMP | YES | NULL |

- **PK**: `id`
- **Unique**: `job_id`

---

## roc_income_log

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| job_id | VARCHAR(255) | NO | — | — |
| user_id | BIGINT | NO | — | YES |
| amount | DECIMAL(12,2) | NO | 0 | — |
| description | VARCHAR(255) | NO | 'ROC Income' | — |
| status | VARCHAR(255) | NO | 'pending' | — |
| earn_log_id | BIGINT | NO | — | YES |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- `earn_log_id` has **no FK constraint** despite the name (just `unsignedBigInteger`); migration comment implies it should reference `earn_logs.id`, but no constraint is created.
- `status` is a free string column; migration comment lists intended values `pending | paid | active` but this is **not an enforced ENUM** — treat as documentation only, not a DB-level constraint.

---

## user_parent_map_logs

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | YES | NULL | YES |
| parent_id | BIGINT | YES | NULL | YES |
| created_row_ids | JSON | YES | NULL | — |
| note | VARCHAR(255) | YES | NULL | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- No FK constraints on `user_id`/`parent_id`.
- `created_row_ids` is a JSON column storing an array of row ids.

---

## user_secret_keys

| Column | Type | Nullable | Unsigned |
|---|---|---|---|
| id | BIGINT | NO | YES (AUTO_INCREMENT) |
| user_id | BIGINT | NO | YES |
| secret_key | VARCHAR(255) | NO | — |
| created_at | TIMESTAMP | YES | — |
| updated_at | TIMESTAMP | YES | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- **Unique**: `user_id` (one secret key per user)
- `secret_key` comment says "hashed key" — stored pre-hashed by the application, plain varchar in the DB.

---

## user_minings

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | NO | — | YES |
| total_token | BIGINT | NO | 0 | — (signed) |
| mining_token | DECIMAL(20,8) | NO | 0 | — |
| daily_mining | BIGINT | NO | 0 | — (signed) |
| status | ENUM('active','inactive') | NO | 'inactive' | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |
| webhook_sent_at | TIMESTAMP | YES | NULL | — |
| webhook_status | VARCHAR(255) | NO | 'pending' | — |
| webhook_attempts | INT | NO | 0 | YES |
| next_retry_at | TIMESTAMP | YES | NULL | — |
| webhook_response | TEXT | YES | NULL | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- **Table-name bug (see Notes)**: two migration filenames say "…_to_minings_table" but both actually alter `user_minings` (there is no `minings` table anywhere in the schema). One of them (`webhook_attempts`/`next_retry_at`) additionally guards with `Schema::hasColumn('minings', ...)`, checking a **table that doesn't exist** — that check is always `false`, so the guard is always a no-op and the columns are unconditionally added. Net effect on final schema is as listed above; flagged only because the guard code is misleading, not because it changes the outcome.

---

## package_pools

| Column | Type | Nullable | Unsigned |
|---|---|---|---|
| id | BIGINT | NO | YES (AUTO_INCREMENT) |
| user_id | BIGINT | NO | YES |
| package_id | BIGINT | NO | YES |
| pool_amount | DECIMAL(15,2) | NO | — |
| created_at | TIMESTAMP | YES | — |
| updated_at | TIMESTAMP | YES | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- **FK**: `package_id` → `packages.id`, `ON DELETE CASCADE`
- `pool_amount` has no default — must be supplied on insert.

---

## global_share_wallets

| Column | Type | Nullable | Default | Unsigned |
|---|---|---|---|---|
| id | BIGINT | NO | AUTO_INCREMENT | YES |
| user_id | BIGINT | NO | — | YES |
| balance | DECIMAL(15,2) | NO | 0 | — |
| max_out | DECIMAL(15,2) | NO | 0 | — |
| created_at | TIMESTAMP | YES | NULL | — |
| updated_at | TIMESTAMP | YES | NULL | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`

---

## global_share_wallets_log

| Column | Type | Nullable | Default |
|---|---|---|---|
| id | BIGINT UNSIGNED | NO | AUTO_INCREMENT |
| user_id | VARCHAR(255) | NO | — |
| amount | VARCHAR(255) | NO | — |
| description | VARCHAR(255) | YES | NULL |
| created_at | TIMESTAMP | YES | NULL |
| updated_at | TIMESTAMP | YES | NULL |

- **PK**: `id`
- Same pattern as `earn_logs`: `user_id` and `amount` are declared as **strings**, not int/decimal, and there is no FK. Preserve as varchar in the rebuild (see Notes).

---

## salaries

| Column | Type | Nullable | Unsigned |
|---|---|---|---|
| id | BIGINT | NO | YES (AUTO_INCREMENT) |
| user_id | BIGINT | NO | YES |
| amount | DECIMAL(12,2) | NO | — |
| salary_date | DATE | NO | — |
| remarks | VARCHAR(255) | YES | — |
| created_at | TIMESTAMP | YES | — |
| updated_at | TIMESTAMP | YES | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- **Index**: `salary_date`

---

## global_director_share_distributions

| Column | Type | Nullable |
|---|---|---|
| id | BIGINT UNSIGNED | NO (AUTO_INCREMENT) |
| period | VARCHAR(255) | NO |
| total_pool | DECIMAL(18,2) | NO |
| total_shares | DECIMAL(18,2) | NO |
| value_per_share | DECIMAL(18,6) | NO |
| user_count | INT UNSIGNED | NO |
| created_at | TIMESTAMP | YES |
| updated_at | TIMESTAMP | YES |

- **PK**: `id`
- **Unique**: `period` (e.g. `"2026-07"` per migration comment — a month-key string, not a date type)

---

## countries

| Column | Type | Nullable |
|---|---|---|
| id | BIGINT UNSIGNED | NO (AUTO_INCREMENT) |
| code | VARCHAR(10) | NO |
| name | VARCHAR(255) | NO |
| created_at | TIMESTAMP | YES |
| updated_at | TIMESTAMP | YES |
| deleted_at | TIMESTAMP | YES |

- **PK**: `id`
- **Unique**: `code`
- **Soft deletes**: yes (`deleted_at`, added by a follow-up migration)

---

## leader_code_logs

| Column | Type | Nullable | Unsigned |
|---|---|---|---|
| id | BIGINT | NO | YES (AUTO_INCREMENT) |
| user_id | BIGINT | NO | YES |
| old_leader_code | VARCHAR(255) | YES | — |
| new_leader_code | VARCHAR(255) | YES | — |
| changed_by | BIGINT | YES | YES |
| created_at | TIMESTAMP | YES | — |
| updated_at | TIMESTAMP | YES | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- **FK**: `changed_by` → `users.id`, `ON DELETE SET NULL` (`nullOnDelete()`)

---

## executive_code_logs

| Column | Type | Nullable | Unsigned |
|---|---|---|---|
| id | BIGINT | NO | YES (AUTO_INCREMENT) |
| user_id | BIGINT | NO | YES |
| old_executive_code | VARCHAR(255) | YES | — |
| new_executive_code | VARCHAR(255) | YES | — |
| changed_by | BIGINT | YES | YES |
| created_at | TIMESTAMP | YES | — |
| updated_at | TIMESTAMP | YES | — |

- **PK**: `id`
- **FK**: `user_id` → `users.id`, `ON DELETE CASCADE`
- **FK**: `changed_by` → `users.id`, `ON DELETE SET NULL` (`nullOnDelete()`)
- Structurally identical to `leader_code_logs`, one migration behind it chronologically.

---

## Seed / default data conventions

- `database/seeders/DatabaseSeeder.php` **only calls `ApiUserSeeder`** by default (`PackageSeeder` and `UserSeeder` exist as classes but are not wired into `DatabaseSeeder::run()` — they'd need to be called explicitly, e.g. `php artisan db:seed --class=PackageSeeder`).
- `ApiUserSeeder` creates one row in **`api_users`**: `username = 'signet'`, `password = Hash::make('KG8QN9ULa&fSK#aL')`. This is the API-auth credential, separate from the `users` table.
- `UserSeeder` (not auto-run) creates one **admin `users` row**: `name = 'Company Head'`, `email = 'company@example.com'`, `password = Hash::make('password')`, `whatsapp_number = '1234567890'`, `binance_pay_id = 'binance_id_123'`, `status = 'active'`, `role = 'company'`, `referred_by = null` — plus a matching `referral_codes` row with a random 6-char uppercase code. This establishes `'company'` as a valid `role` value alongside the column default `'user'`.
- `PackageSeeder` (not auto-run) inserts 9 rows into **`packages`** — the fixed product/rank ladder: `10/100/1000/5000/10000/100000/1000000/5000000/10000000` USDT tiers, commissions `20,40,60,70,80,90,90,90,90`, ranks `Beginner → Grand Master Tycoon`. Useful as reference data / test fixtures when rebuilding, and confirms `packages.price`/`commission` are always used as plain integers (whole USDT / whole percentage points, no decimals).
- `UserFactory` (used for tests, not seeding directly) generates `name`, unique `email`, `email_verified_at = now()`, `password = Hash::make('password')` (shared/cached across factory calls), `remember_token` — a `->unverified()` state clears `email_verified_at`. No other `users` columns are set by the factory, so factory-built users get all migration-level defaults (`role='user'`, `status='pending'`, `leader_status='deactive'`, etc.).

---

## Notes

1. **Two distinct password-reset tables coexist**: Laravel's default `password_reset_tokens` (PK on `email`, created with `users`) and a separate, app-added `password_reset` (singular, **no primary key**, just an index on `email`). Both must be preserved in the rebuild as separate tables; do not merge them. The missing PK on `password_reset` looks like an oversight in the original migration but is the literal, current schema.

2. **`user_parents.node` enum changed twice via raw `DB::statement(ALTER TABLE ... MODIFY COLUMN)`** rather than through `Schema::table()`:
   - `2025_03_06_140459` (create): `ENUM('active','deactive') DEFAULT 'deactive'`
   - `2025_07_13_151554`: widened to `ENUM('active','deactive','gratitude') DEFAULT 'deactive'`
   - `2025_07_25_101638`: widened again to `ENUM('active','deactive','gratitude','correct') DEFAULT 'deactive'`
   Resolution: the **final, authoritative** value set is `('active','deactive','gratitude','correct')`, default `'deactive'`. Used that in the table above.

3. **`user_parents.virtual_id` / `parent_id`** are declared with `$table->foreignId(...)` (implying a foreign key by Laravel convention) but **never chained with `->constrained()`**, so no FK constraint actually exists on either column in the database — they behave as plain `unsignedBigInteger DEFAULT 0` columns. Same pattern recurs in `super_parent_logs.current_user_id` / `package_id` / `user_package`, `roc_income_log.earn_log_id`, and `user_parent_map_logs.user_id` / `parent_id` — none of these are enforced FKs despite naming that suggests otherwise. Do not add FK constraints for these in the Go rebuild's raw SQL unless you've confirmed the live production DB has them (e.g. added manually outside of migrations) — treat the migrations as ground truth per the task instructions.

4. **Two migrations use misleading/buggy guard checks:**
   - `2025_11_06_072558_add_roc_balance_to_wallets_table.php` guards with `Schema::hasColumn('wallets', 'roc_balance')` — correct table name, so this is a normal idempotent guard; no issue.
   - `2026_05_21_051229_add_webhook_retry_columns_to_minings_table.php` guards with `Schema::hasColumn('minings', 'webhook_attempts')` / `Schema::hasColumn('minings', 'next_retry_at')` — but the enclosing `Schema::table('user_minings', ...)` call, and the actual table being modified, is `user_minings`, not `minings` (`minings` does not exist anywhere in the schema). Because `hasColumn` is checking a nonexistent table, it always returns `false`, so the guard is always a no-op and both columns are unconditionally added to `user_minings`. Net effect on the final schema is unambiguous (both columns exist, as documented under `user_minings` above) — flagged purely so it isn't mistaken for a real conditional migration when read by a human debugging column-order oddities.
   - Related: two migration **filenames** (`..._to_minings_table.php`) reference a "minings" table that never exists — both actually operate on `user_minings` inside their `up()` bodies. This is a naming inconsistency in the source repo, not a second table.

5. **`users.role` and `users.status` are separate concepts**: `role` (plain string, default `'user'`, seeder shows `'company'` also used) drives authorization; `status` (`ENUM('pending','active','inactive')`, default `'pending'`) drives account approval/activation state. Don't conflate them.

6. **String-typed numeric-looking columns**: `user_packages.earn`, `earn_logs.user_id`, `earn_logs.amount`, `global_share_wallets_log.user_id`, `global_share_wallets_log.amount` are all declared as `VARCHAR`/`string()` in their migrations even though the names strongly suggest integer/decimal semantics (a user id, a monetary amount). This is very likely an application-level bug/inconsistency (vs. e.g. `token_logs.amount` or `wallets.balance`, which correctly use `DECIMAL`), but it is what the schema literally specifies. Recommendation for the Go rebuild: model these columns as strings/text at the DB layer (matching the live column type exactly, since raw SQL against the same DB must match), and decide at the application layer whether to parse them as numbers — do not change the column type without a matching DB migration, since the task requires the Go/Vue rebuild to read the **same** MySQL database as-is.

7. **`packages.status` is a plain string** (default `'active'`), not an enum — no fixed value list is enforced at the DB level, unlike `users.status`, `user_packages.status`, `tokens.status`, `user_minings.status` which are true `ENUM` columns.

8. **`roc_income_log.status`** is also a plain string (default `'pending'`); the migration source comment lists intended values `pending | paid | active`, but there's no DB-level `ENUM` constraint — application code enforces this list, not MySQL. Documented as a comment/convention only.

9. **Column ordering via `->after(...)`** was simulated step-by-step for tables that received several ALTERs (`users`, `user_packages`, `user_minings`, `packages`, `earn_logs`, `global_share_wallets_log`, `super_parent_logs`, `tokens`, `wallets`, `countries`). Order has no functional impact on a from-scratch Go/SQL rebuild (SQL column order is cosmetic once a table exists) but is documented above for anyone diffing against `DESCRIBE <table>` on the live DB.

10. **`referral_codes.user_id` FK has no `onDelete` clause** (`->constrained('users')` with nothing chained after), meaning it defaults to MySQL's implicit `RESTRICT`/`NO ACTION` — a user with a referral code cannot be hard-deleted while the code row exists. Contrast with most other FKs in this schema, which explicitly cascade or null out.

11. Migration count: the task prompt estimated "~44" migrations; the actual directory contains **56** files. All 56 were read and applied in order to produce this document; the table count below reflects that full set.
