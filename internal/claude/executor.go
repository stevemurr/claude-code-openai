package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Executor handles Claude CLI execution
type Executor struct {
	claudePath string
}

// NewExecutor creates a new Claude executor
func NewExecutor(claudePath string) *Executor {
	return &Executor{claudePath: claudePath}
}

// ExecuteRequest executes a non-streaming request
func (e *Executor) ExecuteRequest(ctx context.Context, prompt string) (*JSONResponse, error) {
	cmd := exec.CommandContext(ctx, e.claudePath, "-p",
		"--output-format", "json",
		"--allowedTools", "WebFetch,WebSearch")

	// Pass prompt via stdin to avoid issues with variadic --allowedTools flag
	cmd.Stdin = strings.NewReader(prompt)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("claude command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute claude: %w", err)
	}

	var resp JSONResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse claude response: %w", err)
	}

	return &resp, nil
}

// StreamCallback is called for each streaming event
type StreamCallback func(event *StreamEvent) error

// ExecuteStreamingRequest executes a streaming request
func (e *Executor) ExecuteStreamingRequest(ctx context.Context, prompt string, callback StreamCallback) error {
	cmd := exec.CommandContext(ctx, e.claudePath, "-p",
		"--output-format", "stream-json",
		"--verbose", "--include-partial-messages",
		"--allowedTools", "WebFetch,WebSearch")

	// Pass prompt via stdin to avoid issues with variadic --allowedTools flag
	cmd.Stdin = strings.NewReader(prompt)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude: %w", err)
	}

	// Read stderr in background for error reporting
	var stderrContent strings.Builder
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			stderrContent.WriteString(scanner.Text())
			stderrContent.WriteString("\n")
		}
	}()

	scanner := bufio.NewScanner(stdout)
	// Increase buffer size for large responses
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event StreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Skip lines that aren't valid JSON
			continue
		}

		if err := callback(&event); err != nil {
			cmd.Process.Kill()
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stdout: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		errMsg := stderrContent.String()
		if errMsg != "" {
			return fmt.Errorf("claude command failed: %s", errMsg)
		}
		return fmt.Errorf("claude command failed: %w", err)
	}

	return nil
}
