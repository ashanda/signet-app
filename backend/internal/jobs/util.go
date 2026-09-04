// Package jobs ports the five scheduled console commands documented in
// financial_engine.md §5 (app/Console/Commands/*.php) to standalone Go
// functions plus a StartScheduler that runs them on tickers in-process —
// see scheduler.go for interval choices. Kept separate from
// internal/handlers (no shared package, no import-cycle risk) since these
// run on timers, not HTTP requests.
package jobs

import (
	"crypto/rand"
	"math/big"
	"strconv"
)

const jobsUpperAlnum = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const jobsAlnum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// jobsRandomUpper ports Str::random() drawn from uppercase-alnum, used for
// the weekly job-id suffix (WeeklyActivePackageSum: "JOB-{year}W{week}-{5
// random uppercase chars}").
func jobsRandomUpper(n int) string {
	return jobsRandomFrom(jobsUpperAlnum, n)
}

// jobsRandomString is mixed-case, used for GenerateOldUserSecretKeys'
// plaintext key suffix ("USER-{id}-" + random(40)).
func jobsRandomString(n int) string {
	return jobsRandomFrom(jobsAlnum, n)
}

func jobsRandomFrom(alphabet string, n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[idx.Int64()]
	}
	return string(b)
}

func jobsItoa(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func jobsFtoa(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
