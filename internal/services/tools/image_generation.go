package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type ImageGenerationTool struct {
	logger *logrus.Logger
	client *openai.Client
}

type ImageGenerationParams struct {
	Prompt string `json:"prompt"`
	Style  string `json:"style,omitempty"`
	Size   string `json:"size,omitempty"`
}

type ImageGenerationResult struct {
	ImageURL    string `json:"image_url"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

func NewImageGenerationTool(apiKey string, logger *logrus.Logger) *ImageGenerationTool {
	return &ImageGenerationTool{
		logger: logger,
		client: openai.NewClient(apiKey),
	}
}

func (t *ImageGenerationTool) Name() string {
	return "generate_image"
}

func (t *ImageGenerationTool) Description() string {
	return "Generate an image based on text prompt"
}

func (t *ImageGenerationTool) Execute(ctx context.Context, parameters map[string]interface{}) (interface{}, error) {
	// Parse parameters
	var params ImageGenerationParams
	paramBytes, err := json.Marshal(parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	if err := json.Unmarshal(paramBytes, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	if params.Prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Set defaults
	if params.Size == "" {
		params.Size = "1024x1024"
	}

	t.logger.WithFields(logrus.Fields{
		"prompt": params.Prompt,
		"style":  params.Style,
		"size":   params.Size,
	}).Info("Generating image")

	// Create DALL-E request
	req := openai.ImageRequest{
		Prompt:         params.Prompt,
		Size:           params.Size,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	start := time.Now()
	resp, err := t.client.CreateImage(ctx, req)
	duration := time.Since(start)

	if err != nil {
		t.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"duration": duration,
		}).Error("Image generation failed")
		return nil, fmt.Errorf("image generation failed: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no image generated")
	}

	result := ImageGenerationResult{
		ImageURL:      resp.Data[0].URL,
		RevisedPrompt: resp.Data[0].RevisedPrompt,
	}

	t.logger.WithFields(logrus.Fields{
		"duration":       duration,
		"image_url":      result.ImageURL,
		"revised_prompt": result.RevisedPrompt,
	}).Info("Image generated successfully")

	// Validate that the image URL is accessible
	if err := t.validateImageURL(result.ImageURL); err != nil {
		t.logger.WithError(err).Warn("Generated image URL validation failed")
	}

	return result, nil
}

func (t *ImageGenerationTool) validateImageURL(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("image URL returned status %d", resp.StatusCode)
	}

	return nil
}