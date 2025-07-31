# OpenAI-Compatible API Support

The WhatsApp AI Bot now supports OpenAI-compatible APIs, allowing you to use alternative AI providers while maintaining the same functionality.

## Configuration

To use a custom OpenAI-compatible API endpoint, set the `OPENAI_BASE_URL` environment variable:

```bash
# For ai.sumopod.com
OPENAI_BASE_URL=https://ai.sumopod.com/v1

# For other providers
OPENAI_BASE_URL=https://your-provider.com/v1
```

## Supported Providers

Any provider that implements the OpenAI Chat Completions API should work, including:

- **ai.sumopod.com** - Alternative OpenAI provider
- **OpenRouter** - Multiple model access
- **Together AI** - Open source models
- **Anyscale** - Ray-based inference
- **Local deployments** (Ollama, vLLM, etc.)

## Example Configuration

### Using ai.sumopod.com

```bash
# .env file
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_BASE_URL=https://ai.sumopod.com/v1
OPENAI_MODEL=gpt-4o-mini
OPENAI_MAX_TOKENS=1000
```

### Using OpenAI (default)

```bash
# .env file
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_BASE_URL=
OPENAI_MODEL=gpt-4-turbo-preview
OPENAI_MAX_TOKENS=1000
```

## Testing the Configuration

You can test your API configuration using curl:

```bash
curl https://ai.sumopod.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {
        "role": "user",
        "content": "Say hello in a creative way"
      }
    ],
    "max_tokens": 150,
    "temperature": 0.7
  }'
```

## Model Selection

Different providers support different models. Make sure to:

1. Check which models are available from your provider
2. Update the `OPENAI_MODEL` environment variable accordingly
3. Verify the model supports the features you need (tool calling, etc.)

## Common Models by Provider

### ai.sumopod.com
- `gpt-4o-mini`
- `gpt-4o`
- `gpt-3.5-turbo`

### OpenRouter
- `openai/gpt-4-turbo-preview`
- `anthropic/claude-3-opus`
- `meta-llama/llama-2-70b-chat`

### Together AI
- `meta-llama/Llama-2-70b-chat-hf`
- `mistralai/Mixtral-8x7B-Instruct-v0.1`
- `NousResearch/Nous-Hermes-2-Mixtral-8x7B-DPO`

## Troubleshooting

### Authentication Issues
- Verify your API key is correct for the provider
- Check if the provider requires a different authentication format

### Model Not Found
- Ensure the model name matches exactly what the provider expects
- Some providers use different naming conventions

### Tool Calling Support
- Not all models/providers support OpenAI's tool calling feature
- Image generation may not work with all providers
- Check provider documentation for feature support

### Rate Limits
- Different providers have different rate limits
- Monitor your usage and adjust accordingly

## Environment Variables Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `OPENAI_API_KEY` | API key for your provider | `sk-...` |
| `OPENAI_BASE_URL` | Custom API endpoint (optional) | `https://ai.sumopod.com/v1` |
| `OPENAI_MODEL` | Model to use | `gpt-4o-mini` |
| `OPENAI_MAX_TOKENS` | Maximum tokens per response | `1000` |

## Implementation Details

The bot uses the `github.com/sashabaranov/go-openai` library with custom configuration:

```go
config := openai.DefaultConfig(apiKey)
if baseURL != "" {
    config.BaseURL = baseURL
}
client := openai.NewClientWithConfig(config)
```

This ensures compatibility with any OpenAI-compatible API endpoint while maintaining all existing functionality including tool calling and conversation management.