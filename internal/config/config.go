package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr   string
	TokenFile    string
	ImgproxyURL  string
	TempDir      string
	MaxFileSize  int64
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	LogFile      string
}

func Load() *Config {
	return &Config{
		ListenAddr:   env("LISTEN_ADDR", "127.0.0.1:9000"),
		TokenFile:    env("TOKEN_FILE", "tokens.json"),
		ImgproxyURL:  env("IMGPROXY_URL", "http://127.0.0.1:48080"),
		TempDir:      env("TEMP_DIR", "/var/www/imgproxy/uploads"),
		MaxFileSize:  envInt64("MAX_FILE_SIZE", 20<<20),
		ReadTimeout:  envDuration("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: envDuration("WRITE_TIMEOUT", 60*time.Second),
		LogFile:      env("LOG_FILE", ""),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
