package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RateLimitIP       int
	RateLimitToken    int
	BlockDurationSecs int
	RedisAddr         string
	RedisPassword     string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		RateLimitIP:       getEnvAsInt("RATE_LIMIT_IP", 5),
		RateLimitToken:    getEnvAsInt("RATE_LIMIT_TOKEN", 10),
		BlockDurationSecs: getEnvAsInt("BLOCK_DURATION_SECONDS", 300),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
	}
}

func getEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvAsInt(name string, defaultVal int) int {
	valStr := getEnv(name, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}
