package config

import (
	"os"
	"strconv"
)

type Config struct {
	Host          string
	Port          int
	GinMode       string
	DBPath        string
	AdminUsername string
	AdminPassword string
	SessionSecret string
	CookieSecure  bool
}

func FromEnv() *Config {
	cfg := &Config{
		Host:          getenv("HOST", "0.0.0.0"),
		Port:          getenvInt("PORT", 8080),
		GinMode:       getenv("GIN_MODE", "release"),
		DBPath:        getenv("DB_PATH", "./data/shorturl.db"),
		AdminUsername: getenv("ADMIN_USERNAME", "admin"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		CookieSecure:  getenvBool("COOKIE_SECURE", false),
	}
	return cfg
}

func (c *Config) Addr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvBool(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

