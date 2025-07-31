package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example-tool-call/internal/config"
	"example-tool-call/internal/handlers"
	"example-tool-call/internal/services/database"
	"example-tool-call/internal/services/fonnte"
	"example-tool-call/internal/services/openai"
	"example-tool-call/internal/services/tools"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	logLevel, err := logrus.ParseLevel(cfg.WhatsApp.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	logger.Info("Starting WhatsApp AI Bot")

	// Initialize database
	db, err := database.New(cfg.Database.URL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	// Initialize services
	fontteService := fonnte.New(cfg.Fonnte.APIKey, logger)
	openaiService := openai.New(cfg.OpenAI.APIKey, cfg.OpenAI.Model, cfg.OpenAI.MaxTokens, logger)
	
	// Initialize tool manager
	toolManager := tools.NewManager(db, logger)
	
	// Register image generation tool
	imageGenTool := tools.NewImageGenerationTool(cfg.Image.APIKey, logger)
	toolManager.RegisterTool(imageGenTool)

	logger.Info("Services initialized successfully")

	// Initialize handlers
	handler := handlers.NewHandler(db, fontteService, openaiService, toolManager, logger)

	// Setup HTTP server
	if cfg.Server.Host == "0.0.0.0" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoints
	router.GET("/health", handler.Health)
	router.GET("/stats", handler.Stats)

	// Webhook endpoints
	router.POST("/webhook/fonnte", handler.FontteWebhook)

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.WithFields(logrus.Fields{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
		}).Info("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}