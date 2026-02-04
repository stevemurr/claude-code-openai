# claude-code-openai

A Go server that wraps the Claude CLI (`claude -p`) and exposes it as an OpenAI-compatible HTTP API with full streaming support.

## Requirements

- Go 1.21+
- [Claude CLI](https://github.com/anthropics/claude-code) installed and authenticated

## Installation

```bash
go install github.com/stevemurr/claude-code-openai@latest
```

Or build from source:

```bash
git clone https://github.com/stevemurr/claude-code-openai.git
cd claude-code-openai
go build -o claude-code-openai
```

## Usage

```bash
# Start server (default port 8080)
./claude-code-openai

# Or with custom config
PORT=3000 CLAUDE_PATH=/path/to/claude ./claude-code-openai
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PORT` | `8080` | Server port |
| `CLAUDE_PATH` | `claude` | Path to Claude CLI binary |

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/chat/completions` | POST | Chat completions (streaming + non-streaming) |
| `/v1/completions` | POST | Legacy completions API |
| `/v1/models` | GET | List available models |
| `/health` | GET | Health check |

## Examples

### Non-streaming request

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-cli",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Streaming request

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-cli",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

### Using with OpenAI Python client

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="not-needed"  # Claude CLI handles auth
)

response = client.chat.completions.create(
    model="claude-cli",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)
```

## Limitations

The following OpenAI parameters are accepted but ignored:
- `temperature`, `top_p`, `presence_penalty`, `frequency_penalty`
- `n` (multiple completions)
- `logprobs`
- Function calling / tools

## License

MIT
