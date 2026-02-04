package converter

import (
	"time"

	"claude-cli-as-openai-api/internal/claude"
	"claude-cli-as-openai-api/internal/openai"
)

// StreamConverter converts Claude stream events to OpenAI format
type StreamConverter struct {
	requestID   string
	model       string
	created     int64
	sentRole    bool
}

// NewStreamConverter creates a new stream converter
func NewStreamConverter(requestID, model string) *StreamConverter {
	return &StreamConverter{
		requestID: requestID,
		model:     model,
		created:   time.Now().Unix(),
		sentRole:  false,
	}
}

// ConvertEvent converts a Claude stream event to an OpenAI stream response
// Returns nil if the event should not produce output
func (c *StreamConverter) ConvertEvent(event *claude.StreamEvent) *openai.ChatCompletionStreamResponse {
	switch event.Type {
	case "stream_event":
		// Handle nested stream events from --include-partial-messages
		if event.Event == nil {
			return nil
		}
		return c.convertInnerEvent(event.Event)

	case "result":
		// Final event with finish reason
		finishReason := "stop"
		return &openai.ChatCompletionStreamResponse{
			ID:      c.requestID,
			Object:  "chat.completion.chunk",
			Created: c.created,
			Model:   c.model,
			Choices: []openai.Choice{
				{
					Index:        0,
					Delta:        &openai.Delta{},
					FinishReason: &finishReason,
				},
			},
		}
	}

	return nil
}

func (c *StreamConverter) convertInnerEvent(event *claude.InnerStreamEvent) *openai.ChatCompletionStreamResponse {
	switch event.Type {
	case "message_start":
		// Send role on first message
		if !c.sentRole {
			c.sentRole = true
			return &openai.ChatCompletionStreamResponse{
				ID:      c.requestID,
				Object:  "chat.completion.chunk",
				Created: c.created,
				Model:   c.model,
				Choices: []openai.Choice{
					{
						Index: 0,
						Delta: &openai.Delta{
							Role: "assistant",
						},
						FinishReason: nil,
					},
				},
			}
		}

	case "content_block_delta":
		if event.Delta != nil && event.Delta.Type == "text_delta" {
			return &openai.ChatCompletionStreamResponse{
				ID:      c.requestID,
				Object:  "chat.completion.chunk",
				Created: c.created,
				Model:   c.model,
				Choices: []openai.Choice{
					{
						Index: 0,
						Delta: &openai.Delta{
							Content: event.Delta.Text,
						},
						FinishReason: nil,
					},
				},
			}
		}
	}

	return nil
}

// ConvertFinalResponse converts a Claude JSON response to an OpenAI response
func ConvertFinalResponse(resp *claude.JSONResponse, requestID, model string) *openai.ChatCompletionResponse {
	finishReason := "stop"
	return &openai.ChatCompletionResponse{
		ID:      requestID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []openai.Choice{
			{
				Index: 0,
				Message: &openai.Message{
					Role:    "assistant",
					Content: resp.Result,
				},
				FinishReason: &finishReason,
			},
		},
	}
}

// ConvertToCompletionResponse converts a Claude response to a legacy completion response
func ConvertToCompletionResponse(resp *claude.JSONResponse, requestID, model string) *openai.CompletionResponse {
	finishReason := "stop"
	return &openai.CompletionResponse{
		ID:      requestID,
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []openai.CompletionChoice{
			{
				Index:        0,
				Text:         resp.Result,
				FinishReason: &finishReason,
			},
		},
	}
}
