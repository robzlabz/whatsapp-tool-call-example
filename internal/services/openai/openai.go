package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Service struct {
	client   *openai.Client
	model    string
	maxTokens int
	logger   *logrus.Logger
}

type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func New(apiKey, baseURL, model string, maxTokens int, logger *logrus.Logger) *Service {
	config := openai.DefaultConfig(apiKey)
	
	// Set custom base URL if provided
	if baseURL != "" {
		config.BaseURL = baseURL
		logger.WithField("base_url", baseURL).Info("Using custom OpenAI-compatible API endpoint")
	}
	
	client := openai.NewClientWithConfig(config)
	return &Service{
		client:    client,
		model:     model,
		maxTokens: maxTokens,
		logger:    logger,
	}
}

func (s *Service) GenerateResponse(ctx context.Context, messages []ChatMessage, tools []openai.Tool) (*openai.ChatCompletionResponse, error) {
	// Convert our messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:     s.model,
		Messages:  openaiMessages,
		MaxTokens: s.maxTokens,
		Tools:     tools,
	}

	s.logger.WithFields(logrus.Fields{
		"model":      s.model,
		"messages":   len(messages),
		"tools":      len(tools),
		"max_tokens": s.maxTokens,
	}).Debug("Sending request to OpenAI")

	start := time.Now()
	resp, err := s.client.CreateChatCompletion(ctx, req)
	duration := time.Since(start)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"duration": duration,
		}).Error("OpenAI request failed")
		return nil, fmt.Errorf("OpenAI request failed: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"duration":     duration,
		"usage_tokens": resp.Usage.TotalTokens,
		"choices":      len(resp.Choices),
	}).Info("OpenAI request completed")

	return &resp, nil
}

func (s *Service) GetAvailableTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "generate_image",
				Description: "Generate an image based on text prompt",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"prompt": map[string]interface{}{
							"type":        "string",
							"description": "Text description of the image to generate",
						},
						"style": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"realistic", "cartoon", "artistic"},
							"description": "Style of the generated image",
						},
						"size": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"256x256", "512x512", "1024x1024"},
							"description": "Size of the generated image",
						},
					},
					"required": []string{"prompt"},
				},
			},
		},
	}
}

func (s *Service) ParseToolCall(toolCall openai.ToolCall) (*ToolCall, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		return nil, fmt.Errorf("failed to parse tool call parameters: %w", err)
	}

	return &ToolCall{
		Name:       toolCall.Function.Name,
		Parameters: params,
	}, nil
}