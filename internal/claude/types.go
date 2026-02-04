package claude

// StreamEvent represents any event from Claude's stream-json output
type StreamEvent struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`

	// For init events
	SessionID string `json:"session_id,omitempty"`
	Tools     []any  `json:"tools,omitempty"`
	Model     string `json:"model,omitempty"`

	// For assistant message events
	Message *AssistantMessage `json:"message,omitempty"`

	// For stream_event wrapper (when using --include-partial-messages)
	Event *InnerStreamEvent `json:"event,omitempty"`

	// For result events
	ResultText    string  `json:"result,omitempty"`
	CostUSD       float64 `json:"cost_usd,omitempty"`
	TotalCostUSD  float64 `json:"total_cost_usd,omitempty"`
	IsError       bool    `json:"is_error,omitempty"`
	DurationMS    int     `json:"duration_ms,omitempty"`
	DurationAPIMS int     `json:"duration_api_ms,omitempty"`
	NumTurns      int     `json:"num_turns,omitempty"`
}

// InnerStreamEvent represents the inner event from stream_event wrapper
type InnerStreamEvent struct {
	Type    string `json:"type"`
	Index   int    `json:"index,omitempty"`
	Message *AssistantMessage `json:"message,omitempty"`

	// For content_block_start
	ContentBlock *ContentBlock `json:"content_block,omitempty"`

	// For content_block_delta
	Delta *ContentDelta `json:"delta,omitempty"`
}

// AssistantMessage represents an assistant message in streaming
type AssistantMessage struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Role    string `json:"role,omitempty"`
	Content []any  `json:"content,omitempty"`
	Model   string `json:"model,omitempty"`
}

// ContentBlock represents a content block
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ContentDelta represents a delta in content
type ContentDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Usage represents token usage in Claude response
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// JSONResponse represents a non-streaming JSON response from Claude
type JSONResponse struct {
	Type       string  `json:"type"`
	Subtype    string  `json:"subtype,omitempty"`
	CostUSD    float64 `json:"cost_usd,omitempty"`
	IsError    bool    `json:"is_error,omitempty"`
	DurationMS int     `json:"duration_ms,omitempty"`
	NumTurns   int     `json:"num_turns,omitempty"`
	Result     string  `json:"result,omitempty"`
	SessionID  string  `json:"session_id,omitempty"`
}
