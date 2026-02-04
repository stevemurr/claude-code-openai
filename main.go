package main

import (
	"log"
	"net/http"

	"claude-cli-as-openai-api/config"
	"claude-cli-as-openai-api/internal/api"
	"claude-cli-as-openai-api/internal/claude"
)

func main() {
	cfg := config.Load()

	executor := claude.NewExecutor(cfg.ClaudePath)
	handlers := api.NewHandlers(executor)
	router := api.NewRouter(handlers)

	addr := ":" + cfg.Port
	log.Printf("Starting server on %s", addr)
	log.Printf("Claude CLI path: %s", cfg.ClaudePath)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
