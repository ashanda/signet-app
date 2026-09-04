// share_calculate.go ports `share:calculate` — CalculateGlobalDirectorShare
// (financial_engine.md §5, monthly Global Director profit-share payout).
package jobs

import (
	"database/sql"
	"log"
	"math"
	"time"

	"github.com/jmoiron/sqlx"

	"signet-backend/internal/wallet"
)

// RunShareCalculate distributes the current month's Global Director profit
// share, gated by wallet.UpdateWallet's 4x-cap check (financial_engine.md
// §1) per recipient. Idempotent per calendar-month period: a period already
// recorded in global_director_share_distributions is skipped outright; if
// any individual credit fails mid-run the period is NOT marked distributed,
// so a safe re-run will re-pay everyone (including already-paid users — no
// per-user idempotency, only whole-period, matching the original verbatim).
func RunShareCalculate(db *sqlx.DB) error {
	period := time.Now().Format("2006-01")

	var alreadyDistributed int
	if err := db.Get(&alreadyDistributed, "SELECT COUNT(*) FROM global_director_share_distributions WHERE period = ?", period); err != nil {
		return err
	}
	if alreadyDistributed > 0 {
		return nil
	}

	now := time.Now()
	var totalPool sql.NullFloat64
	if err := db.Get(&totalPool, `
		SELECT COALESCE(SUM(pool_amount), 0) FROM package_pools
		WHERE YEAR(created_at) = ? AND MONTH(created_at) = ? AND pool_amount > 0`,
		now.Year(), int(now.Month())); err != nil {
		return err
	}
	if totalPool.Float64 <= 0 {
		return nil
	}

	var totalShares sql.NullFloat64
	if err := db.Get(&totalShares, `SELECT COALESCE(SUM(global_director_share), 0) FROM users WHERE global_director_share_status = 1`); err != nil {
		return err
	}
	if totalShares.Float64 <= 0 {
		return nil
	}

	valuePerShare := totalPool.Float64 / totalShares.Float64

	type shareUser struct {
		ID                  uint64  `db:"id"`
		GlobalDirectorShare float64 `db:"global_director_share"`
	}
	var users []shareUser
	if err := db.Select(&users, `SELECT id, global_director_share FROM users WHERE global_director_share_status = 1`); err != nil {
		return err
	}

	succeeded := 0
	anyFailed := false
	for _, u := range users {
		amount := math.Round(u.GlobalDirectorShare*valuePerShare*100) / 100
		if err := wallet.UpdateWallet(db, u.ID, amount, "Global Director Share"); err != nil {
			log.Printf("jobs: share:calculate: credit failed for user %d: %v", u.ID, err)
			anyFailed = true
			continue
		}
		succeeded++
	}

	if anyFailed {
		return nil // period stays undistributed, safe to re-run later
	}

	_, err := db.Exec(`INSERT INTO global_director_share_distributions (period, total_pool, total_shares, value_per_share, user_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())`,
		period, totalPool.Float64, totalShares.Float64, valuePerShare, succeeded)
	return err
}
