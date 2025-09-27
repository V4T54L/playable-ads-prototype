package configs

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort      string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSslMode       string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	MaxFileSizeMB   int64
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
}

func LoadConfig() (Config, error) {
	cfg := Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "user"),
		DBPassword:    getEnv("DB_PASSWORD", "password"),
		DBName:        getEnv("DB_NAME", "db"),
		DBSslMode:     getEnv("DB_SSL_MODE", "disable"),
		JWTSecret:     getEnv("JWT_SECRET", "a-very-secret-key-that-is-long-enough"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}

	// Parse durations and numbers
	cfg.AccessTokenTTL = parseDuration("JWT_ACCESS_TOKEN_TTL", "15m")
	cfg.RefreshTokenTTL = parseDuration("JWT_REFRESH_TOKEN_TTL", "720h")
	cfg.MaxFileSizeMB = parseInt64("MAX_FILE_SIZE_MB", 10)
	cfg.RedisDB = parseInt("REDIS_DB", 0)

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseDuration(key, defaultVal string) time.Duration {
	val := getEnv(key, defaultVal)
	d, err := time.ParseDuration(val)
	if err != nil {
		return mustParseDuration(defaultVal)
	}
	return d
}

func mustParseDuration(val string) time.Duration {
	d, err := time.ParseDuration(val)
	if err != nil {
		panic("invalid default duration: " + val)
	}
	return d
}

func parseInt(key string, defaultVal int) int {
	val := getEnv(key, "")
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func parseInt64(key string, defaultVal int64) int64 {
	val := getEnv(key, "")
	if val == "" {
		return defaultVal
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultVal
	}
	return i
}
