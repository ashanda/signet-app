// Package config loads runtime configuration from environment variables,
// mirroring the keys used by the original Laravel .env file where they map
// to a real equivalent here. See docs/analysis/ARCHITECTURE.md.
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppName string
	AppURL  string
	AppEnv  string
	Port    string

	DBHost     string
	DBPort     string
	DBDatabase string
	DBUsername string
	DBPassword string

	SessionSecret   string
	SessionLifetime time.Duration

	MailHost string
	MailPort string
	MailUser string
	MailPass string
	MailFrom string

	MiningWebhookEnabled bool
	MiningWebhookURL     string
	MiningWebhookSecret  string

	// FrontendOrigin is the Vue dev-server origin allowed for CORS in dev.
	FrontendOrigin string

	// EnableScheduler controls whether this process runs the in-process
	// recurring jobs (mining:update, mining:send-webhook, packages:weekly-sum,
	// share:calculate — see internal/jobs/scheduler.go), replacing the
	// original app's OS-level cron. Defaults to on so a single-instance
	// deployment behaves like the original out of the box; set
	// ENABLE_SCHEDULER=false on every instance but one when running more
	// than one API process against the same database, to avoid duplicate
	// job runs.
	EnableScheduler bool
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func Load() *Config {
	lifetimeMinutes, err := strconv.Atoi(env("SESSION_LIFETIME", "120"))
	if err != nil {
		lifetimeMinutes = 120
	}

	return &Config{
		AppName: env("APP_NAME", "Signetint.net"),
		AppURL:  env("APP_URL", "http://127.0.0.1:8080"),
		AppEnv:  env("APP_ENV", "local"),
		Port:    env("PORT", "8080"),

		DBHost:     env("DB_HOST", "127.0.0.1"),
		DBPort:     env("DB_PORT", "3306"),
		DBDatabase: env("DB_DATABASE", "signet_last"),
		DBUsername: env("DB_USERNAME", "root"),
		DBPassword: env("DB_PASSWORD", ""),

		SessionSecret:   env("SESSION_SECRET", "insecure-dev-secret-change-me"),
		SessionLifetime: time.Duration(lifetimeMinutes) * time.Minute,

		MailHost: env("MAIL_HOST", "127.0.0.1"),
		MailPort: env("MAIL_PORT", "1025"),
		MailUser: env("MAIL_USERNAME", ""),
		MailPass: env("MAIL_PASSWORD", ""),
		MailFrom: env("MAIL_FROM_ADDRESS", "hello@example.com"),

		MiningWebhookEnabled: envBool("MINING_WEBHOOK_ENABLED", false),
		MiningWebhookURL:     env("MINING_WEBHOOK_URL", ""),
		MiningWebhookSecret:  env("MINING_WEBHOOK_SECRET", ""),

		FrontendOrigin: env("FRONTEND_ORIGIN", "http://localhost:5173"),

		EnableScheduler: envBool("ENABLE_SCHEDULER", true),
	}
}

func (c *Config) MySQLDSN() string {
	// parseTime=true so DATETIME/TIMESTAMP columns decode into time.Time;
	// loc=Local matches PHP's date handling (no explicit UTC conversion in
	// the original app).
	return c.DBUsername + ":" + c.DBPassword + "@tcp(" + c.DBHost + ":" + c.DBPort + ")/" + c.DBDatabase +
		"?parseTime=true&loc=Local&charset=utf8mb4"
}
