// generate_secret_keys.go ports `users:generate-secret-keys` —
// GenerateOldUserSecretKeys (financial_engine.md §5, one-off backfill).
//
// NOTE (preserved from the original, flagged to the product owner per
// financial_engine.md): the generated plaintext key is never stored or
// returned anywhere by this command — only its bcrypt hash is persisted.
// This command is a backfill for pre-existing users only; it has no way to
// hand the plaintext back to the user. Not wired into StartScheduler (it's
// a one-off, not a recurring job) — call RunGenerateSecretKeys directly
// (e.g. from an admin/ops entrypoint) when a backfill is actually needed.
package jobs

import (
	"log"

	"github.com/jmoiron/sqlx"

	"signet-backend/internal/auth"
)

func RunGenerateSecretKeys(db *sqlx.DB) error {
	const chunkSize = 100
	var lastID uint64
	for {
		var ids []uint64
		if err := db.Select(&ids, "SELECT id FROM users WHERE id > ? ORDER BY id ASC LIMIT ?", lastID, chunkSize); err != nil {
			return err
		}
		if len(ids) == 0 {
			break
		}
		for _, id := range ids {
			lastID = id

			var exists int
			if err := db.Get(&exists, "SELECT COUNT(*) FROM user_secret_keys WHERE user_id = ?", id); err != nil {
				log.Printf("jobs: users:generate-secret-keys: could not check existing key for user %d: %v", id, err)
				continue
			}
			if exists > 0 {
				continue
			}

			plainKey := "USER-" + jobsItoa(id) + "-" + jobsRandomString(40)
			hashed, err := auth.HashPassword(plainKey)
			if err != nil {
				log.Printf("jobs: users:generate-secret-keys: could not hash key for user %d: %v", id, err)
				continue
			}
			if _, err := db.Exec("INSERT INTO user_secret_keys (user_id, secret_key, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", id, hashed); err != nil {
				log.Printf("jobs: users:generate-secret-keys: could not insert key for user %d: %v", id, err)
			}
		}
		if len(ids) < chunkSize {
			break
		}
	}
	return nil
}
