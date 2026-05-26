package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr            string
	AMQPURL             string
	QueueName           string
	AntispamUserCooldown int64
	AntispamMaxPerIPMin  int
	AntispamMaxQueryMin  int
	MaxTopN              int
}

func Load() *Config {
	return &Config{
		HTTPAddr:             getEnv("HTTP_ADDR", ":8080"),
		AMQPURL:              getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:            getEnv("QUEUE_NAME", "search_logs"),
		AntispamUserCooldown: getEnvInt64("ANTISPAM_USER_COOLDOWN_SEC", 5),
		AntispamMaxPerIPMin:  getEnvInt("ANTISPAM_MAX_PER_IP_PER_MIN", 60),
		AntispamMaxQueryMin:  getEnvInt("ANTISPAM_MAX_QUERY_PER_MIN", 500),
		MaxTopN:              getEnvInt("MAX_TOP_N", 100),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if val, ok := os.LookupEnv(key); ok {
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			return n
		}
	}
	return defaultVal
}
