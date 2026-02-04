package api

import (
	"net/http"
)

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(handlers *Handlers) http.Handler {
	mux := http.NewServeMux()

	// OpenAI-compatible endpoints
	mux.HandleFunc("/v1/chat/completions", handlers.HandleChatCompletions)
	mux.HandleFunc("/v1/completions", handlers.HandleCompletions)
	mux.HandleFunc("/v1/models", handlers.HandleModels)

	// Health check
	mux.HandleFunc("/health", handlers.HandleHealth)

	// Apply middleware
	var handler http.Handler = mux
	handler = Logging(handler)
	handler = CORS(handler)

	return handler
}
