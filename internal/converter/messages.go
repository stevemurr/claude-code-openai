package converter

import (
	"fmt"
	"strings"

	"claude-cli-as-openai-api/internal/openai"
)

// MessagesToPrompt converts OpenAI messages to a Claude prompt string
func MessagesToPrompt(messages []openai.Message) string {
	var parts []string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			parts = append(parts, fmt.Sprintf("[System: %s]", msg.Content))
		case "user":
			if msg.Name != "" {
				parts = append(parts, fmt.Sprintf("%s: %s", msg.Name, msg.Content))
			} else {
				parts = append(parts, msg.Content)
			}
		case "assistant":
			parts = append(parts, fmt.Sprintf("[Previous assistant response: %s]", msg.Content))
		}
	}

	return strings.Join(parts, "\n\n")
}

// PromptStringToPrompt handles the legacy completions API prompt field
func PromptStringToPrompt(prompt any) string {
	switch p := prompt.(type) {
	case string:
		return p
	case []any:
		var parts []string
		for _, item := range p {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "\n")
	default:
		return fmt.Sprintf("%v", prompt)
	}
}
