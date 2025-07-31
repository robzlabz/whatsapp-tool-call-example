package fonnte

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Service struct {
	apiKey string
	client *http.Client
	logger *logrus.Logger
}

type SendMessageRequest struct {
	Target  string `json:"target"`
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}

type SendMessageResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

type WebhookMessage struct {
	Device    string `json:"device"`
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Member    string `json:"member"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	File      string `json:"file"`
	Filename  string `json:"filename"`
}

func New(apiKey string, logger *logrus.Logger) *Service {
	return &Service{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (s *Service) SendMessage(target, message string) (*SendMessageResponse, error) {
	req := SendMessageRequest{
		Target:  target,
		Message: message,
		Type:    "text",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.fonnte.com/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	s.logger.WithFields(logrus.Fields{
		"target":  target,
		"message": message,
	}).Debug("Sending message via Fonnte")

	start := time.Now()
	resp, err := s.client.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"duration": duration,
		}).Error("Fonnte request failed")
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response SendMessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"duration": duration,
		"status":   response.Status,
		"id":       response.ID,
	}).Info("Fonnte message sent")

	if !response.Status {
		return nil, fmt.Errorf("fonnte API error: %s", response.Message)
	}

	return &response, nil
}

func (s *Service) SendImage(target, imageURL, caption string) (*SendMessageResponse, error) {
	req := map[string]interface{}{
		"target":  target,
		"file":    imageURL,
		"caption": caption,
		"type":    "image",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.fonnte.com/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	s.logger.WithFields(logrus.Fields{
		"target":    target,
		"image_url": imageURL,
		"caption":   caption,
	}).Debug("Sending image via Fonnte")

	start := time.Now()
	resp, err := s.client.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"duration": duration,
		}).Error("Fonnte image request failed")
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response SendMessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"duration": duration,
		"status":   response.Status,
		"id":       response.ID,
	}).Info("Fonnte image sent")

	if !response.Status {
		return nil, fmt.Errorf("fonnte API error: %s", response.Message)
	}

	return &response, nil
}