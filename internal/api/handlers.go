package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"claude-cli-as-openai-api/internal/claude"
	"claude-cli-as-openai-api/internal/converter"
	"claude-cli-as-openai-api/internal/openai"
	"claude-cli-as-openai-api/pkg/sse"
)

// Handlers contains HTTP handlers
type Handlers struct {
	executor *claude.Executor
}

// NewHandlers creates new handlers
func NewHandlers(executor *claude.Executor) *Handlers {
	return &Handlers{executor: executor}
}

// HandleChatCompletions handles /v1/chat/completions
func (h *Handlers) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}

	var req openai.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "invalid_request_error")
		return
	}

	if len(req.Messages) == 0 {
		h.writeError(w, http.StatusBadRequest, "messages array is required", "invalid_request_error")
		return
	}

	prompt := converter.MessagesToPrompt(req.Messages)
	requestID := fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	model := "claude-cli"

	if req.Stream {
		h.handleStreamingChat(w, r, prompt, requestID, model)
	} else {
		h.handleNonStreamingChat(w, r, prompt, requestID, model)
	}
}

func (h *Handlers) handleNonStreamingChat(w http.ResponseWriter, r *http.Request, prompt, requestID, model string) {
	resp, err := h.executor.ExecuteRequest(r.Context(), prompt)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error(), "api_error")
		return
	}

	response := converter.ConvertFinalResponse(resp, requestID, model)
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handlers) handleStreamingChat(w http.ResponseWriter, r *http.Request, prompt, requestID, model string) {
	sseWriter, err := sse.NewWriter(w)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error(), "api_error")
		return
	}

	streamConverter := converter.NewStreamConverter(requestID, model)

	err = h.executor.ExecuteStreamingRequest(r.Context(), prompt, func(event *claude.StreamEvent) error {
		response := streamConverter.ConvertEvent(event)
		if response != nil {
			return sseWriter.WriteEvent(response)
		}
		return nil
	})

	if err != nil {
		// For streaming, we can't send error as JSON after headers are sent
		// Just log and close the connection
		return
	}

	sseWriter.WriteDone()
}

// HandleCompletions handles /v1/completions (legacy)
func (h *Handlers) HandleCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}

	var req openai.CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error(), "invalid_request_error")
		return
	}

	prompt := converter.PromptStringToPrompt(req.Prompt)
	requestID := fmt.Sprintf("cmpl-%d", time.Now().UnixNano())
	model := "claude-cli"

	if req.Stream {
		h.handleStreamingCompletion(w, r, prompt, requestID, model)
	} else {
		h.handleNonStreamingCompletion(w, r, prompt, requestID, model)
	}
}

func (h *Handlers) handleNonStreamingCompletion(w http.ResponseWriter, r *http.Request, prompt, requestID, model string) {
	resp, err := h.executor.ExecuteRequest(r.Context(), prompt)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error(), "api_error")
		return
	}

	response := converter.ConvertToCompletionResponse(resp, requestID, model)
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handlers) handleStreamingCompletion(w http.ResponseWriter, r *http.Request, prompt, requestID, model string) {
	sseWriter, err := sse.NewWriter(w)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error(), "api_error")
		return
	}

	streamConverter := converter.NewStreamConverter(requestID, model)

	err = h.executor.ExecuteStreamingRequest(r.Context(), prompt, func(event *claude.StreamEvent) error {
		response := streamConverter.ConvertEvent(event)
		if response != nil {
			// Convert chat completion chunk to completion chunk for legacy API
			if len(response.Choices) > 0 && response.Choices[0].Delta != nil {
				legacyResp := &openai.CompletionResponse{
					ID:      response.ID,
					Object:  "text_completion",
					Created: response.Created,
					Model:   response.Model,
					Choices: []openai.CompletionChoice{
						{
							Index:        0,
							Text:         response.Choices[0].Delta.Content,
							FinishReason: response.Choices[0].FinishReason,
						},
					},
				}
				return sseWriter.WriteEvent(legacyResp)
			}
		}
		return nil
	})

	if err != nil {
		return
	}

	sseWriter.WriteDone()
}

// HandleModels handles /v1/models
func (h *Handlers) HandleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}

	response := openai.ModelList{
		Object: "list",
		Data: []openai.Model{
			{
				ID:      "claude-cli",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "anthropic",
			},
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// HandleHealth handles /health
func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handlers) writeError(w http.ResponseWriter, status int, message, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(openai.ErrorResponse{
		Error: openai.ErrorDetail{
			Message: message,
			Type:    errType,
		},
	})
}
