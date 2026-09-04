package wallet

import "strconv"

// The original app stores several "numeric" columns (earn_logs.user_id/
// amount, user_packages.earn, global_share_wallets_log.user_id/amount) as
// VARCHAR (see schema.md Note 6). These helpers do the same string<->number
// round-tripping PHP's type juggling did implicitly.

func itoa(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func ftoa(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func atof(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
