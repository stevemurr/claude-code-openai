# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this repository.

## Project Overview

This is a Go server that wraps the Claude CLI and exposes it as an OpenAI-compatible HTTP API. It enables any application that supports the OpenAI API format to use Claude CLI as a backend.

## Architecture

```
main.go                     → Entry point, wires up dependencies
config/config.go            → Environment configuration (PORT, CLAUDE_PATH)
internal/
  api/
    handlers.go             → HTTP handlers for all endpoints
    router.go               → Route definitions
    middleware.go           → CORS and logging middleware
  openai/types.go           → OpenAI API request/response structs
  claude/
    executor.go             → Spawns and manages Claude CLI processes
    types.go                → Claude CLI output format structs
  converter/
    messages.go             → Converts OpenAI messages to Claude prompt
    stream.go               → Converts Claude stream events to OpenAI SSE
pkg/sse/writer.go           → Server-Sent Events response writer
```

## Key Implementation Details

### Claude CLI Invocation

- Non-streaming: `echo "<prompt>" | claude -p --output-format json --allowedTools WebFetch,WebSearch`
- Streaming: `echo "<prompt>" | claude -p --output-format stream-json --verbose --include-partial-messages --allowedTools WebFetch,WebSearch`

The prompt is passed via stdin because `--allowedTools` is a variadic flag that would consume positional arguments.
The `--include-partial-messages` flag is required for true token-by-token streaming.
The `--allowedTools` flag enables Claude to use web tools (WebFetch, WebSearch) for retrieving online content.

### Stream Event Flow

Claude CLI outputs newline-delimited JSON with these event types:
1. `system` (init) → ignored
2. `stream_event` with inner `message_start` → OpenAI role chunk
3. `stream_event` with inner `content_block_delta` → OpenAI content chunk
4. `result` → OpenAI finish_reason chunk + `[DONE]`

## Common Commands

```bash
# Run server
go run main.go

# Build binary
go build -o claude-code-openai

# Run with custom port
PORT=3000 go run main.go
```

## Testing

Test endpoints manually:

```bash
# Health check
curl http://localhost:8080/health

# Models
curl http://localhost:8080/v1/models

# Chat completion
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-cli","messages":[{"role":"user","content":"Hi"}]}'

# Streaming
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-cli","messages":[{"role":"user","content":"Hi"}],"stream":true}'
```
