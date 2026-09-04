// mining_webhook.go ports `mining:send-webhook` — SendMiningWebhook
// (financial_engine.md §5, outbound webhook w/ retry backoff; original run
// frequently, e.g. every minute).
package jobs

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"

	"signet-backend/internal/config"
)

// miningWebhookRunning replaces the original's Cache::lock('mining-webhook-lock', 55s)
// overlap guard: a non-blocking in-process flag, since this job is only
// ever driven by this one process's own ticker — if a previous run is still
// mid-flight (slow webhook endpoint) when the next tick fires, this run is
// skipped rather than queued.
var miningWebhookRunning int32

func RunMiningWebhook(db *sqlx.DB, cfg *config.Config) error {
	if !cfg.MiningWebhookEnabled {
		return nil
	}
	if !atomic.CompareAndSwapInt32(&miningWebhookRunning, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&miningWebhookRunning, 0)

	type record struct {
		ID              uint64       `db:"id"`
		UserID          uint64       `db:"user_id"`
		TotalToken      int64        `db:"total_token"`
		MiningToken     float64      `db:"mining_token"`
		DailyMining     int64        `db:"daily_mining"`
		Status          string       `db:"status"`
		CreatedAt       sql.NullTime `db:"created_at"`
		UpdatedAt       sql.NullTime `db:"updated_at"`
		WebhookAttempts int          `db:"webhook_attempts"`
	}
	var records []record
	if err := db.Select(&records, `
		SELECT id, user_id, total_token, mining_token, daily_mining, status, created_at, updated_at, webhook_attempts
		FROM user_minings
		WHERE webhook_sent_at IS NULL
		  AND (webhook_status = 'pending' OR (webhook_status = 'failed' AND (next_retry_at IS NULL OR next_retry_at <= NOW())))
		  AND webhook_attempts < 5
		ORDER BY id LIMIT 50`); err != nil {
		return err
	}

	client := &http.Client{Timeout: 20 * time.Second}

	for _, rec := range records {
		var secretKey string
		if err := db.Get(&secretKey, "SELECT secret_key FROM user_secret_keys WHERE user_id = ? LIMIT 1", rec.UserID); err != nil {
			log.Printf("jobs: mining:send-webhook: no secret key for user %d, skipping record %d: %v", rec.UserID, rec.ID, err)
			continue
		}

		payload := map[string]interface{}{
			"id":           rec.ID,
			"user_id":      secretKey, // the OPAQUE hashed secret key string, not the numeric user id
			"total_token":  rec.TotalToken,
			"mining_token": rec.MiningToken,
			"daily_mining": rec.DailyMining,
			"status":       rec.Status,
			"created_at":   rec.CreatedAt.Time,
			"updated_at":   rec.UpdatedAt.Time,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("jobs: mining:send-webhook: could not encode payload for record %d: %v", rec.ID, err)
			continue
		}

		req, err := http.NewRequest(http.MethodPost, cfg.MiningWebhookURL, bytes.NewReader(body))
		if err != nil {
			log.Printf("jobs: mining:send-webhook: could not build request for record %d: %v", rec.ID, err)
			continue
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Secret", cfg.MiningWebhookSecret)

		resp, err := client.Do(req)
		if err != nil {
			miningWebhookRecordFailure(db, rec.ID, rec.WebhookAttempts, err.Error())
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if _, err := db.Exec(`UPDATE user_minings SET webhook_sent_at = NOW(), webhook_status = 'sent', webhook_response = ?, next_retry_at = NULL, updated_at = NOW() WHERE id = ?`,
				string(respBody), rec.ID); err != nil {
				log.Printf("jobs: mining:send-webhook: could not mark record %d sent: %v", rec.ID, err)
			}
			continue
		}
		miningWebhookRecordFailure(db, rec.ID, rec.WebhookAttempts, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
	}
	return nil
}

func miningWebhookRecordFailure(db *sqlx.DB, id uint64, priorAttempts int, response string) {
	attempts := priorAttempts + 1
	status := "failed"
	var nextRetryAt interface{}
	if attempts >= 5 {
		status = "permanently_failed"
		nextRetryAt = nil
	} else {
		nextRetryAt = time.Now().Add(miningWebhookBackoff(attempts))
	}
	if _, err := db.Exec(`UPDATE user_minings SET webhook_status = ?, webhook_attempts = ?, next_retry_at = ?, webhook_response = ?, updated_at = NOW() WHERE id = ?`,
		status, attempts, nextRetryAt, response, id); err != nil {
		log.Printf("jobs: mining:send-webhook: could not record failure for record %d: %v", id, err)
	}
}

// miningWebhookBackoff: 1st retry +1m, 2nd +5m, 3rd +15m, 4th +30m, default +1h.
func miningWebhookBackoff(attempts int) time.Duration {
	switch attempts {
	case 1:
		return time.Minute
	case 2:
		return 5 * time.Minute
	case 3:
		return 15 * time.Minute
	case 4:
		return 30 * time.Minute
	default:
		return time.Hour
	}
}
