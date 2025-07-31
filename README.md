# WhatsApp AI Bot

A WhatsApp bot application built with Go and whatsmeow that integrates with Fonnte.com API for messaging capabilities, OpenAI for intelligent responses, and includes tool calling functionality for image generation.

## Features

- 🤖 **AI-Powered Responses**: Uses OpenAI GPT models for intelligent conversations
- 🔄 **OpenAI-Compatible APIs**: Support for alternative AI providers (ai.sumopod.com, OpenRouter, etc.)
- 🎨 **Image Generation**: Generate images using OpenAI DALL-E through tool calling
- 📱 **WhatsApp Integration**: Direct WhatsApp connection using whatsmeow library
- 🔗 **Fonnte.com Support**: Alternative messaging through Fonnte API
- 💾 **Persistent Storage**: SQLite/PostgreSQL support for conversation history
- 🔧 **Tool System**: Extensible tool calling architecture
- 📊 **Monitoring**: Health checks and statistics endpoints
- 🐳 **Docker Ready**: Containerized deployment support

## Quick Start

### Prerequisites

- Go 1.21 or higher
- OpenAI API key
- Fonnte.com API key (optional)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd wa-test
```

2. Copy environment configuration:
```bash
cp .env.example .env
```

3. Edit `.env` file with your API keys:
```env
OPENAI_API_KEY=your_openai_api_key_here
FONNTE_API_KEY=your_fonnte_api_key_here
IMAGE_API_KEY=your_openai_api_key_here
```

4. Install dependencies:
```bash
go mod tidy
```

5. Run the application:
```bash
go run cmd/bot/main.go
```

### Docker Deployment

1. Build the Docker image:
```bash
docker build -t whatsapp-ai-bot .
```

2. Run the container:
```bash
docker run -d \
  --name whatsapp-bot \
  -p 8080:8080 \
  -e OPENAI_API_KEY=your_key_here \
  -e FONNTE_API_KEY=your_key_here \
  -v $(pwd)/sessions:/root/sessions \
  whatsapp-ai-bot
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server host address | `0.0.0.0` |
| `SERVER_PORT` | Server port | `8080` |
| `OPENAI_API_KEY` | OpenAI API key | Required |
| `OPENAI_BASE_URL` | Custom OpenAI-compatible API endpoint | Optional |
| `OPENAI_MODEL` | OpenAI model to use | `gpt-4-turbo-preview` |
| `OPENAI_MAX_TOKENS` | Maximum tokens per response | `1000` |
| `FONNTE_API_KEY` | Fonnte.com API key | Required |
| `IMAGE_API_PROVIDER` | Image generation provider | `openai` |
| `IMAGE_API_KEY` | Image generation API key | Required |
| `DATABASE_URL` | Database connection URL | `sqlite://./bot.db` |
| `WHATSAPP_SESSION_PATH` | WhatsApp session storage path | `./sessions` |
| `WHATSAPP_LOG_LEVEL` | Logging level | `INFO` |

### OpenAI-Compatible APIs

The bot supports OpenAI-compatible API endpoints, allowing you to use alternative AI providers:

```bash
# Using ai.sumopod.com
OPENAI_BASE_URL=https://ai.sumopod.com/v1
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_MODEL=gpt-4o-mini

# Using OpenRouter
OPENAI_BASE_URL=https://openrouter.ai/api/v1
OPENAI_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_MODEL=openai/gpt-4-turbo-preview

# Using default OpenAI (leave OPENAI_BASE_URL empty)
OPENAI_BASE_URL=
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_MODEL=gpt-4-turbo-preview
```

For detailed configuration and supported providers, see [OpenAI-Compatible APIs Documentation](docs/openai-compatible-apis.md).

## API Endpoints

### Health Check
```
GET /health
```
Returns server health status.

### Statistics
```
GET /stats
```
Returns usage statistics.

### Fonnte Webhook
```
POST /webhook/fonnte
```
Receives incoming messages from Fonnte.com.

## Usage

### Text Conversations
Simply send any text message to the bot, and it will respond using OpenAI's language model.

### Image Generation
Ask the bot to generate images using natural language:
- "Generate an image of a sunset over mountains"
- "Create a cartoon style image of a cat wearing a hat"
- "Make a realistic image of a futuristic city"

The bot supports different styles:
- `realistic` - Photorealistic images
- `cartoon` - Cartoon-style images  
- `artistic` - Artistic/stylized images

And different sizes:
- `256x256` - Small images
- `512x512` - Medium images
- `1024x1024` - Large images (default)

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   WhatsApp      │    │   WhatsApp AI   │    │   External APIs │
│   Users         │◄──►│   Bot           │◄──►│                 │
│                 │    │                 │    │ • OpenAI        │
└─────────────────┘    │ ┌─────────────┐ │    │ • Fonnte.com    │
                       │ │ whatsmeow   │ │    │ • DALL-E        │
┌─────────────────┐    │ │ library     │ │    │                 │
│   Fonnte.com    │◄──►│ └─────────────┘ │    └─────────────────┘
│   Webhook       │    │                 │
└─────────────────┘    │ ┌─────────────┐ │
                       │ │ Tool Call   │ │
                       │ │ Manager     │ │
                       │ └─────────────┘ │
                       └─────────────────┘
```

## Development

### Project Structure
```
example-tool-call/
├── cmd/
│   └── bot/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handlers/
│   │   └── handlers.go          # HTTP handlers
│   ├── models/
│   │   └── models.go            # Database models
│   └── services/
│       ├── database/
│       │   └── database.go      # Database service
│       ├── fonnte/
│       │   └── fonnte.go        # Fonnte API client
│       ├── openai/
│       │   └── openai.go        # OpenAI service
│       ├── tools/
│       │   ├── manager.go       # Tool manager
│       │   └── image_generation.go # Image generation tool
│       └── whatsapp/
│           └── whatsapp.go      # WhatsApp service
├── .env.example                 # Environment template
├── Dockerfile                   # Docker configuration
├── go.mod                       # Go module
├── go.sum                       # Go dependencies
└── README.md                    # This file
```

### Adding New Tools

1. Implement the `Tool` interface:
```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, parameters map[string]interface{}) (interface{}, error)
}
```

2. Register the tool in the tool manager:
```go
toolManager.RegisterTool(yourNewTool)
```

3. Add the tool definition to OpenAI function calling schema.

## Monitoring

The application provides several monitoring endpoints:

- `/health` - Health check endpoint
- `/stats` - Usage statistics

Logs are structured in JSON format and include:
- Request/response logging
- Tool execution logs
- Error tracking
- Performance metrics

## Security

- API keys are managed through environment variables
- Input sanitization for all user messages
- Rate limiting support
- Webhook signature validation (when configured)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For issues and questions:
1. Check the existing issues
2. Create a new issue with detailed information
3. Include logs and configuration (without sensitive data)

## Changelog

### v1.0.0
- Initial release
- OpenAI integration
- Fonnte.com support
- Image generation tool
- Docker support
- Basic monitoring