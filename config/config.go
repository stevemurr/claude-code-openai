package config

import (
	"os"
)

type Config struct {
	Port       string
	ClaudePath string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	claudePath := os.Getenv("CLAUDE_PATH")
	if claudePath == "" {
		claudePath = "claude"
	}

	return &Config{
		Port:       port,
		ClaudePath: claudePath,
	}
}
