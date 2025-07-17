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

	"github.com/tienpdinh/gpt-home/internal/api"
	"github.com/tienpdinh/gpt-home/internal/config"
	"github.com/tienpdinh/gpt-home/internal/conversation"
	"github.com/tienpdinh/gpt-home/internal/device"
	"github.com/tienpdinh/gpt-home/internal/llm"
	"github.com/tienpdinh/gpt-home/pkg/homeassistant"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	setupLogging(cfg.LogLevel)

	logrus.Info("Starting GPT-Home...")

	// Initialize components
	haClient := homeassistant.NewClient(cfg.HomeAssistant.URL, cfg.HomeAssistant.Token)
	deviceManager := device.NewManager(haClient)
	llmService := llm.NewServiceWithConfig(cfg.LLM.OllamaURL, cfg.LLM.Model, cfg.LLM)
	conversationManager := conversation.NewManager()

	// Initialize and load LLM
	if err := llmService.LoadModel(); err != nil {
		logrus.Fatalf("Failed to load LLM: %v", err)
	}

	// Setup HTTP server
	router := setupRouter(cfg, deviceManager, llmService, conversationManager)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logrus.Infof("Server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func setupRouter(cfg *config.Config, deviceManager *device.Manager, llmService *llm.Service, conversationManager *conversation.Manager) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Initialize API handlers
	apiHandler := api.NewHandler(deviceManager, llmService, conversationManager)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/chat", apiHandler.HandleChat)
		v1.GET("/devices", apiHandler.GetDevices)
		v1.GET("/devices/:id", apiHandler.GetDevice)
		v1.POST("/devices/:id/action", apiHandler.ControlDevice)
		v1.GET("/conversations/:id", apiHandler.GetConversation)
		v1.DELETE("/conversations/:id", apiHandler.DeleteConversation)
		v1.GET("/health", apiHandler.HealthCheck)
	}

	// Static files for web interface
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "GPT-Home",
		})
	})

	return router
}
