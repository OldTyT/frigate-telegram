package config

import (
	"os"
	"strconv"
	"strings"
)

const Version string = "v0.0.1"

type Config struct {
	Debug              bool
	TelegramBotToken   string
	FrigateURL         string
	FrigateEventLimit  int
	TelegramChatID     int64
	SleepTime          int
	FrigateExternalURL string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	RedisProtocol      int
	RedisTTL           int
	TimeWaitSave       int
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		FrigateURL:         getEnv("FRIGATE_URL", "http://localhost:5000"),
		FrigateEventLimit:  getEnvAsInt("FRIGATE_EVENT_LIMIT", 20),
		Debug:              getEnvAsBool("DEBUG", false),
		TelegramChatID:     getEnvAsInt64("TELEGRAM_CHAT_ID", 0),
		SleepTime:          getEnvAsInt("SLEEP_TIME", 5),
		FrigateExternalURL: getEnv("FRIGATE_EXTERNAL_URL", "http://localhost:5000"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvAsInt("REDIS_DB", 0),
		RedisProtocol:      getEnvAsInt("REDIS_PROTOCOL", 3),
		RedisTTL:           getEnvAsInt("REDIS_TTL", 1209600), // 7 days
		TimeWaitSave:       getEnvAsInt("TIME_WAIT_SAVE", 60), // 5 seconds
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt64(name string, defaultVal int64) int64 {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return int64(value)
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Helper to read an environment variable into a string slice or return default value
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
