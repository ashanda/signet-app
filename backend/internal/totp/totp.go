// Package totp implements standard RFC 6238 TOTP generation/verification,
// compatible with Google Authenticator and with the original app's
// sonata-project/google-authenticator PHP library (same algorithm: HMAC-
// SHA1, 30s step, 6 digits, base32 secret).
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"strings"
	"time"
)

// GenerateSecret returns a 32-character base32 secret (160 bits), the same
// size sonata-project/google-authenticator's generateSecret() produces.
func GenerateSecret() string {
	raw := make([]byte, 20)
	_, _ = rand.Read(raw)
	return strings.TrimRight(base32.StdEncoding.EncodeToString(raw), "=")
}

// OtpauthURL builds the otpauth://totp/... URI the QR code encodes, matching
// the original's `signetint:{email}?secret={secret}&issuer=signetint` shape.
func OtpauthURL(issuer, account, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, account, secret, issuer)
}

func code(secret string, counter int64) (string, error) {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", err
	}
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = byte(counter & 0xff)
		counter >>= 8
	}
	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := (int(sum[offset])&0x7f)<<24 | (int(sum[offset+1])&0xff)<<16 | (int(sum[offset+2])&0xff)<<8 | (int(sum[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", value%1000000), nil
}

// CheckCode verifies a 6-digit code against the secret, allowing ±1 step
// (30s) of clock drift — the standard tolerance most TOTP apps/libraries
// (including sonata's) use by default.
func CheckCode(secret, submitted string) bool {
	submitted = strings.TrimSpace(submitted)
	if submitted == "" {
		return false
	}
	now := time.Now().Unix() / 30
	for _, step := range []int64{now - 1, now, now + 1} {
		expected, err := code(secret, step)
		if err == nil && expected == submitted {
			return true
		}
	}
	return false
}
