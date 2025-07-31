package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"example-tool-call/internal/models"
	"example-tool-call/internal/services/database"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	db     *database.DB
	logger *logrus.Logger
	tools  map[string]Tool
}

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, parameters map[string]interface{}) (interface{}, error)
}

type ExecutionResult struct {
	Success   bool        `json:"success"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	Duration  int64       `json:"duration"`
	ToolName  string      `json:"tool_name"`
}

func NewManager(db *database.DB, logger *logrus.Logger) *Manager {
	manager := &Manager{
		db:     db,
		logger: logger,
		tools:  make(map[string]Tool),
	}

	return manager
}

func (m *Manager) RegisterTool(tool Tool) {
	m.tools[tool.Name()] = tool
	m.logger.WithField("tool", tool.Name()).Info("Tool registered")
}

func (m *Manager) ExecuteTool(ctx context.Context, messageID, toolName string, parameters map[string]interface{}) (*ExecutionResult, error) {
	tool, exists := m.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", toolName)
	}

	m.logger.WithFields(logrus.Fields{
		"tool":       toolName,
		"message_id": messageID,
		"parameters": parameters,
	}).Info("Executing tool")

	start := time.Now()
	result, err := tool.Execute(ctx, parameters)
	duration := time.Since(start)

	execution := &models.ToolExecution{
		MessageID:     messageID,
		ToolName:      toolName,
		ExecutionTime: duration.Milliseconds(),
		Success:       err == nil,
	}

	// Marshal parameters for storage
	if paramBytes, marshalErr := json.Marshal(parameters); marshalErr == nil {
		execution.Parameters = string(paramBytes)
	}

	// Handle result and error
	if err != nil {
		execution.ErrorMsg = err.Error()
		m.logger.WithFields(logrus.Fields{
			"tool":     toolName,
			"error":    err.Error(),
			"duration": duration,
		}).Error("Tool execution failed")
	} else {
		if resultBytes, marshalErr := json.Marshal(result); marshalErr == nil {
			execution.Result = string(resultBytes)
		}
		m.logger.WithFields(logrus.Fields{
			"tool":     toolName,
			"duration": duration,
		}).Info("Tool execution completed")
	}

	// Save execution to database
	if dbErr := m.db.SaveToolExecution(execution); dbErr != nil {
		m.logger.WithError(dbErr).Error("Failed to save tool execution")
	}

	return &ExecutionResult{
		Success:  err == nil,
		Result:   result,
		Error:    func() string { if err != nil { return err.Error() }; return "" }(),
		Duration: duration.Milliseconds(),
		ToolName: toolName,
	}, nil
}

func (m *Manager) GetAvailableTools() []openai.Tool {
	tools := make([]openai.Tool, 0, len(m.tools))
	for _, tool := range m.tools {
		// For now, we'll manually define the OpenAI tool format
		// In a real implementation, you might want tools to provide their own schema
		if tool.Name() == "generate_image" {
			tools = append(tools, openai.Tool{
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
			})
		}
	}
	return tools
}