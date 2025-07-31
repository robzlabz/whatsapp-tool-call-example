package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"example-tool-call/internal/models"
	"example-tool-call/internal/services/database"
	"example-tool-call/internal/services/fonnte"
	openaiService "example-tool-call/internal/services/openai"
	"example-tool-call/internal/services/tools"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	db      *database.DB
	fonnte  *fonnte.Service
	openai  *openaiService.Service
	toolMgr *tools.Manager
	logger  *logrus.Logger
}

func NewHandler(db *database.DB, fonnte *fonnte.Service, openai *openaiService.Service, toolMgr *tools.Manager, logger *logrus.Logger) *Handler {
	return &Handler{
		db:      db,
		fonnte:  fonnte,
		openai:  openai,
		toolMgr: toolMgr,
		logger:  logger,
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

func (h *Handler) Stats(c *gin.Context) {
	// Get basic statistics
	var messageCount int64
	var conversationCount int64
	var toolExecutionCount int64

	h.db.Model(&models.Message{}).Count(&messageCount)
	h.db.Model(&models.Conversation{}).Count(&conversationCount)
	h.db.Model(&models.ToolExecution{}).Count(&toolExecutionCount)

	c.JSON(http.StatusOK, gin.H{
		"messages":        messageCount,
		"conversations":   conversationCount,
		"tool_executions": toolExecutionCount,
		"timestamp":       time.Now().UTC(),
	})
}

func (h *Handler) FontteWebhook(c *gin.Context) {
	var webhook fonnte.WebhookMessage
	if err := c.ShouldBindJSON(&webhook); err != nil {
		h.logger.WithError(err).Error("Failed to parse webhook")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"sender":  webhook.Sender,
		"message": webhook.Message,
		"device":  webhook.Device,
	}).Info("Received Fonnte webhook")

	// Process the message
	go h.processMessage(webhook.Sender, webhook.Message)

	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (h *Handler) processMessage(sender, message string) {
	ctx := context.Background()

	// Skip empty messages
	if strings.TrimSpace(message) == "" {
		return
	}

	// Get or create conversation
	conversation, err := h.db.GetOrCreateConversation(sender)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get conversation")
		return
	}

	// Get recent messages for context
	recentMessages, err := h.db.GetMessages(sender, 10)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get recent messages")
		recentMessages = []models.Message{}
	}

	// Build conversation history
	messages := []openaiService.ChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful WhatsApp AI assistant. You can generate images when requested. Be friendly and helpful.",
		},
	}

	// Add recent messages to context (in reverse order to maintain chronology)
	for i := len(recentMessages) - 1; i >= 0; i-- {
		msg := recentMessages[i]
		role := "user"
		if msg.IsFromMe {
			role = "assistant"
		}
		messages = append(messages, openaiService.ChatMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Add current message
	messages = append(messages, openaiService.ChatMessage{
		Role:    "user",
		Content: message,
	})

	// Get available tools
	tools := h.toolMgr.GetAvailableTools()

	// Generate response
	response, err := h.openai.GenerateResponse(ctx, messages, tools)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate response")
		h.sendErrorMessage(sender, "Sorry, I'm having trouble processing your message right now.")
		return
	}

	if len(response.Choices) == 0 {
		h.logger.Error("No response choices received")
		h.sendErrorMessage(sender, "Sorry, I couldn't generate a response.")
		return
	}

	choice := response.Choices[0]
	
	// Handle tool calls
	if len(choice.Message.ToolCalls) > 0 {
		h.handleToolCalls(ctx, sender, choice.Message.ToolCalls, choice.Message.Content)
	} else {
		// Send regular text response
		h.sendTextMessage(sender, choice.Message.Content)
	}

	// Save user message
	userMsg := &models.Message{
		MessageID:   fmt.Sprintf("user_%d", time.Now().UnixNano()),
		FromJID:     sender,
		ToJID:       "bot",
		Content:     message,
		MessageType: "text",
		IsFromMe:    false,
		Timestamp:   time.Now(),
	}
	if err := h.db.SaveMessage(userMsg); err != nil {
		h.logger.WithError(err).Error("Failed to save user message")
	}

	// Save assistant message
	assistantMsg := &models.Message{
		MessageID:   fmt.Sprintf("assistant_%d", time.Now().UnixNano()),
		FromJID:     "bot",
		ToJID:       sender,
		Content:     choice.Message.Content,
		MessageType: "text",
		IsFromMe:    true,
		Timestamp:   time.Now(),
	}
	if err := h.db.SaveMessage(assistantMsg); err != nil {
		h.logger.WithError(err).Error("Failed to save assistant message")
	}

	// Update conversation
	conversation.LastMessage = message
	conversation.MessageCount++
	if err := h.db.UpdateConversation(conversation); err != nil {
		h.logger.WithError(err).Error("Failed to update conversation")
	}
}

func (h *Handler) handleToolCalls(ctx context.Context, sender string, toolCalls []openai.ToolCall, assistantMessage string) {
	for _, toolCall := range toolCalls {
		// Parse tool call parameters
		var parameters map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &parameters); err != nil {
			h.logger.WithError(err).Error("Failed to parse tool call arguments")
			continue
		}

		// Execute tool
		result, err := h.toolMgr.ExecuteTool(ctx, toolCall.ID, toolCall.Function.Name, parameters)
		if err != nil {
			h.logger.WithError(err).Error("Tool execution failed")
			h.sendErrorMessage(sender, fmt.Sprintf("Sorry, I couldn't execute the %s tool.", toolCall.Function.Name))
			continue
		}

		// Handle different tool results
		switch toolCall.Function.Name {
		case "generate_image":
			h.handleImageGenerationResult(sender, result, assistantMessage)
		default:
			h.sendTextMessage(sender, fmt.Sprintf("Tool %s executed successfully", toolCall.Function.Name))
		}
	}
}

func (h *Handler) handleImageGenerationResult(sender string, result *tools.ExecutionResult, assistantMessage string) {
	if !result.Success {
		h.sendErrorMessage(sender, "Sorry, I couldn't generate the image. Please try again with a different prompt.")
		return
	}

	// Parse the result
	var imageResult tools.ImageGenerationResult
	resultBytes, _ := json.Marshal(result.Result)
	if err := json.Unmarshal(resultBytes, &imageResult); err != nil {
		h.logger.WithError(err).Error("Failed to parse image generation result")
		h.sendErrorMessage(sender, "Sorry, there was an error processing the generated image.")
		return
	}

	// Send the image
	caption := assistantMessage
	if caption == "" {
		caption = "Here's your generated image!"
	}

	_, err := h.fonnte.SendImage(sender, imageResult.ImageURL, caption)
	if err != nil {
		h.logger.WithError(err).Error("Failed to send image")
		h.sendErrorMessage(sender, "Sorry, I couldn't send the generated image.")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"sender":    sender,
		"image_url": imageResult.ImageURL,
	}).Info("Image sent successfully")
}

func (h *Handler) sendTextMessage(sender, message string) {
	_, err := h.fonnte.SendMessage(sender, message)
	if err != nil {
		h.logger.WithError(err).Error("Failed to send text message")
	}
}

func (h *Handler) sendErrorMessage(sender, message string) {
	_, err := h.fonnte.SendMessage(sender, message)
	if err != nil {
		h.logger.WithError(err).Error("Failed to send error message")
	}
}