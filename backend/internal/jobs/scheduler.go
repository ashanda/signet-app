// scheduler.go launches the recurring jobs in-process on time.Tickers,
// replacing the original app's OS-level cron scheduler (routes/console.php
// / app/Console/Kernel.php). Call StartScheduler once from cmd/api/main.go
// after the DB connection is established.
package jobs

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	"signet-backend/internal/config"
)

// StartScheduler starts one goroutine per recurring job. Interval choices
// (the original's exact cron timing wasn't in scope for this port, see
// financial_engine.md §5):
//
//   - mining:update and mining:send-webhook: every minute — financial_engine.md
//     explicitly calls both "run every minute" / "run frequently, e.g. every
//     minute".
//   - packages:weekly-sum and share:calculate: checked once a day. Both
//     RunWeeklySum and RunShareCalculate carry their own internal
//     week/month-boundary math plus an idempotency guard (a completed
//     period is never re-processed), so a daily check is a safe, simple
//     superset of "run only on the day the period closes" — it acts once
//     per period and no-ops every other day, and self-heals a missed run
//     instead of silently skipping it until the next period.
//
// users:generate-secret-keys (RunGenerateSecretKeys) is a one-off backfill
// and is intentionally NOT scheduled here — see generate_secret_keys.go.
//
// Every job's error is logged via the standard `log` package and never
// crashes the process; a failed run just waits for its next tick.
func StartScheduler(db *sqlx.DB, cfg *config.Config) {
	go runOnTicker("mining:update", time.Minute, func() error { return RunMiningUpdate(db) })
	go runOnTicker("mining:send-webhook", time.Minute, func() error { return RunMiningWebhook(db, cfg) })
	go runOnTicker("packages:weekly-sum", 24*time.Hour, func() error { return RunWeeklySum(db) })
	go runOnTicker("share:calculate", 24*time.Hour, func() error { return RunShareCalculate(db) })
}

// runOnTicker runs `run` immediately, then again every `interval`, for as
// long as the process lives. Errors are logged, never fatal.
func runOnTicker(name string, interval time.Duration, run func() error) {
	if err := run(); err != nil {
		log.Printf("jobs: %s: %v", name, err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if err := run(); err != nil {
			log.Printf("jobs: %s: %v", name, err)
		}
	}
}
