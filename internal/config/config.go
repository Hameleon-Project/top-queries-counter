package config

import "os"

type Config struct {
	HTTPAddr string
	AMQPURL  string
}

func Load() *Config {
	return &Config{
		HTTPAddr: getEnv("HTTP_ADDR", ":8080"),
		AMQPURL:  getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
