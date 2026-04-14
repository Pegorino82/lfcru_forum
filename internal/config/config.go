package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL        string
	AppPort            string
	CookieSecure       bool
	SessionLifetime    time.Duration
	BcryptCost         int
	RateLimitWindow    time.Duration
	RateLimitMax       int
	SessionGracePeriod time.Duration
	MaxSessionsPerUser int
	UploadsDir         string
}

func Load() *Config {
	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/lfcru?sslmode=disable"),
		AppPort:            getEnv("APP_PORT", "8080"),
		CookieSecure:       getBool("COOKIE_SECURE", false),
		SessionLifetime:    getDuration("SESSION_LIFETIME", 720*time.Hour),
		BcryptCost:         getInt("BCRYPT_COST", 12),
		RateLimitWindow:    getDuration("RATE_LIMIT_WINDOW", 10*time.Minute),
		RateLimitMax:       getInt("RATE_LIMIT_MAX", 5),
		SessionGracePeriod: getDuration("SESSION_GRACE_PERIOD", 5*time.Minute),
		MaxSessionsPerUser: getInt("MAX_SESSIONS_PER_USER", 10),
		UploadsDir:         getEnv("UPLOADS_DIR", "./uploads"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
