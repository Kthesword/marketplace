// Package config reads and validates application settings from environment
// variables. Everything else in the app receives a Config value instead of
// calling os.Getenv directly.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// minSecretLen: HS256 JWT secrets should be long and random.
const minSecretLen = 32

// Config holds all settings. Field names are Go-style; env names are in Load.
type Config struct {
	Port           string        // e.g. "8080"
	DatabaseURL    string        // postgres://user:pass@host:port/db
	JWTSecret      string        // signing key for tokens (never log this!)
	JWTTTL         time.Duration // how long a token stays valid
	MigrateOnStart bool          // run SQL migrations at boot
}

// Load reads .env (if present) and then the real environment.
// Real environment variables win over .env values.
func Load() (Config, error) {
	// godotenv.Load does NOT override variables that are already set, and a
	// missing file is fine (production has no .env file), so we ignore the error.
	_ = godotenv.Load()

	var cfg Config
	var problems []error // collect every problem so the user fixes them in one go

	cfg.Port = getEnv("PORT", "8080")

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		problems = append(problems, errors.New("DATABASE_URL is required"))
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < minSecretLen || strings.HasPrefix(cfg.JWTSecret, "change-me") {
		problems = append(problems, fmt.Errorf(
			"JWT_SECRET must be at least %d characters and not the placeholder (try: openssl rand -hex 32)", minSecretLen))
	}

	ttl, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil || ttl <= 0 {
		problems = append(problems, errors.New("JWT_TTL must be a positive duration like 30m or 24h"))
	}
	cfg.JWTTTL = ttl

	migrate, err := strconv.ParseBool(getEnv("MIGRATE_ON_START", "true"))
	if err != nil {
		problems = append(problems, errors.New("MIGRATE_ON_START must be true or false"))
	}
	cfg.MigrateOnStart = migrate

	// errors.Join returns nil when the slice is empty.
	if err := errors.Join(problems...); err != nil {
		return Config{}, fmt.Errorf("invalid configuration:\n%w", err)
	}
	return cfg, nil
}

// getEnv returns the variable's value, or fallback when it is unset/empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
