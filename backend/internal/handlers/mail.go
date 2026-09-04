package handlers

import (
	"fmt"
	"log"
	"net/smtp"

	"signet-backend/internal/app"
	"signet-backend/internal/models"
)

// sendMail is a small best-effort SMTP sender replacing Laravel's queued
// Mailable classes (NewUserPackageMail, PasswordResetMail) — see
// ARCHITECTURE.md "What's intentionally out of scope for a literal 1:1":
// same end effect (an email goes out), simpler mechanism (fire-and-forget
// goroutine instead of a DB-backed queue). Failures are logged, never
// surfaced to the HTTP caller, matching the original's queued/fire-and-forget
// send for the package mail (the password-reset mail was synchronous in the
// original but is still non-critical-path here).
func sendMail(d *app.Deps, to, subject, body string) {
	addr := d.Cfg.MailHost + ":" + d.Cfg.MailPort
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		d.Cfg.MailFrom, to, subject, body)

	var auth smtp.Auth
	if d.Cfg.MailUser != "" {
		auth = smtp.PlainAuth("", d.Cfg.MailUser, d.Cfg.MailPass, d.Cfg.MailHost)
	}
	if err := smtp.SendMail(addr, auth, d.Cfg.MailFrom, []string{to}, []byte(msg)); err != nil {
		log.Printf("mail send failed (to=%s subject=%q): %v", to, subject, err)
	}
}

func sendNewUserPackageMail(d *app.Deps, newUser, parent *models.User, pkg *models.Package) {
	if parent == nil || parent.Email == "" {
		return
	}
	body := fmt.Sprintf(
		"<p>A new user (%s) has purchased the %s package under your referral.</p>",
		newUser.Name, pkg.Name,
	)
	sendMail(d, parent.Email, "New Package Activation", body)
}

func sendPasswordResetMail(d *app.Deps, to, token string) {
	body := fmt.Sprintf(
		`<p>You requested a password reset.</p><p>Reset link: %s/password/reset/%s</p><p>If you did not request this, ignore this email.</p>`,
		d.Cfg.AppURL, token,
	)
	sendMail(d, to, "Password Reset", body)
}
