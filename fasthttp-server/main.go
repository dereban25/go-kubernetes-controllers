package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

const (
	requestIDKey = "requestID"
)

var (
	serverPort int
	logLevel   string
	logFile    *os.File
	startTime  time.Time
)

// Logger for structured logging
type Logger struct {
	level string
}

// NewLogger creates a new logger instance
func NewLogger(level string) *Logger {
	return &Logger{level: level}
}

// Setup logging to both console and file
func setupLogging() error {
	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Create log file with timestamp
	startTime = time.Now()
	timestamp := startTime.Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("server_%s.log", timestamp)
	logFilePath := filepath.Join(logsDir, logFileName)

	var err error
	logFile, err = os.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}

	// Create multi-writer to write to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// Set log format with timestamp
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Printf("[SYSTEM] Logging started - Console and File: %s", logFilePath)
	return nil
}

// Close logging and write stop message
func closeLogging() {
	if logFile != nil {
		stopTime := time.Now()
		duration := stopTime.Sub(startTime)

		log.Printf("[SYSTEM] Server stopped at %s", stopTime.Format("2006-01-02 15:04:05"))
		log.Printf("[SYSTEM] Total uptime: %v", duration)
		log.Printf("[SYSTEM] Logging ended")

		logFile.Close()
	}
}

// LogRequest logs HTTP request details with request ID
func (l *Logger) LogRequest(ctx *fasthttp.RequestCtx, requestID string, startTime time.Time) {
	duration := time.Since(startTime)

	log.Printf("[REQUEST] ID=%s | %s %s | Status=%d | Duration=%v | IP=%s | UserAgent=%s | Size=%d bytes",
		requestID,
		string(ctx.Method()),
		string(ctx.RequestURI()),
		ctx.Response.StatusCode(),
		duration,
		ctx.RemoteIP().String(),
		string(ctx.UserAgent()),
		len(ctx.Response.Body()),
	)
}

// LogError logs error details with request ID
func (l *Logger) LogError(requestID, message string, err error) {
	log.Printf("[ERROR] ID=%s | %s | Error: %v", requestID, message, err)
}

// LogInfo logs informational messages with request ID
func (l *Logger) LogInfo(requestID, message string) {
	log.Printf("[INFO] ID=%s | %s", requestID, message)
}

// Middleware for request logging and tracing
func loggingMiddleware(next fasthttp.RequestHandler, logger *Logger) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		startTime := time.Now()

		// Generate a unique request ID for tracing
		requestID := uuid.New().String()
		ctx.SetUserValue(requestIDKey, requestID)

		// Log incoming request
		logger.LogInfo(requestID, fmt.Sprintf("Incoming request: %s %s from %s",
			string(ctx.Method()), string(ctx.RequestURI()), ctx.RemoteIP().String()))

		// Set request ID in response header for client-side tracing
		ctx.Response.Header.Set("X-Request-ID", requestID)

		// Call next handler
		next(ctx)

		// Log request completion
		logger.LogRequest(ctx, requestID, startTime)
	}
}

// Recovery middleware to handle panics
func recoveryMiddleware(next fasthttp.RequestHandler, logger *Logger) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				requestID := getRequestID(ctx)
				logger.LogError(requestID, "Panic recovered", fmt.Errorf("panic: %v", r))

				ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
			}
		}()

		next(ctx)
	}
}

// Helper function to get request ID from context
func getRequestID(ctx *fasthttp.RequestCtx) string {
	if requestID := ctx.UserValue(requestIDKey); requestID != nil {
		return requestID.(string)
	}
	return "unknown"
}

// Main request handler
func mainHandler(ctx *fasthttp.RequestCtx) {
	requestID := getRequestID(ctx)
	logger := NewLogger(logLevel)

	switch string(ctx.Path()) {
	case "/":
		handleRoot(ctx, requestID, logger)
	case "/health":
		handleHealth(ctx, requestID, logger)
	case "/api/v1/status":
		handleStatus(ctx, requestID, logger)
	default:
		handleNotFound(ctx, requestID, logger)
	}
}

// Root endpoint handler
func handleRoot(ctx *fasthttp.RequestCtx, requestID string, logger *Logger) {
	logger.LogInfo(requestID, "Handling root endpoint")

	response := fmt.Sprintf(`{
		"message": "Welcome to FastHTTP Server",
		"request_id": "%s",
		"timestamp": "%s",
		"version": "1.0.0"
	}`, requestID, time.Now().Format(time.RFC3339))

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.WriteString(response)
}

// Health check endpoint handler
func handleHealth(ctx *fasthttp.RequestCtx, requestID string, logger *Logger) {
	logger.LogInfo(requestID, "Health check requested")

	response := fmt.Sprintf(`{
		"status": "healthy",
		"request_id": "%s",
		"timestamp": "%s"
	}`, requestID, time.Now().Format(time.RFC3339))

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.WriteString(response)
}

// Status endpoint handler
func handleStatus(ctx *fasthttp.RequestCtx, requestID string, logger *Logger) {
	logger.LogInfo(requestID, "Status endpoint requested")

	response := fmt.Sprintf(`{
		"server": "fasthttp",
		"uptime": "%s",
		"request_id": "%s",
		"timestamp": "%s",
		"go_version": "go1.21+",
		"memory_usage": "calculated_in_production"
	}`, time.Since(startTime).String(), requestID, time.Now().Format(time.RFC3339))

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.WriteString(response)
}

// 404 handler
func handleNotFound(ctx *fasthttp.RequestCtx, requestID string, logger *Logger) {
	logger.LogInfo(requestID, fmt.Sprintf("404 Not Found: %s", string(ctx.RequestURI())))

	response := fmt.Sprintf(`{
		"error": "Not Found",
		"message": "The requested resource was not found",
		"request_id": "%s",
		"timestamp": "%s"
	}`, requestID, time.Now().Format(time.RFC3339))

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.WriteString(response)
}

// Server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the FastHTTP server",
	Long:  "Start the FastHTTP server with comprehensive request logging and tracing capabilities",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

// Start the HTTP server
func startServer() {
	// Setup logging to both console and file
	if err := setupLogging(); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	// Ensure logging is closed on exit
	defer closeLogging()

	logger := NewLogger(logLevel)

	// Create request handler with middleware chain
	handler := loggingMiddleware(
		recoveryMiddleware(mainHandler, logger),
		logger,
	)

	// Configure server
	server := &fasthttp.Server{
		Handler:            handler,
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
		IdleTimeout:        60 * time.Second,
		MaxConnsPerIP:      100,
		MaxRequestsPerConn: 1000,
		TCPKeepalive:       true,
		MaxRequestBodySize: 10 * 1024 * 1024, // 10MB
		ReduceMemoryUsage:  true,
		LogAllErrors:       true,
		ErrorHandler: func(ctx *fasthttp.RequestCtx, err error) {
			requestID := getRequestID(ctx)
			logger.LogError(requestID, "Server error", err)
			ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		},
	}

	// Server address
	addr := fmt.Sprintf(":%d", serverPort)

	// Log server startup information
	log.Printf("[SERVER] Starting FastHTTP server at %s", startTime.Format("2006-01-02 15:04:05"))
	log.Printf("[SERVER] Server port: %d", serverPort)
	log.Printf("[SERVER] Logging level: %s", logLevel)
	log.Printf("[SERVER] Process ID: %d", os.Getpid())
	log.Printf("[SERVER] Available endpoints:")
	log.Printf("[SERVER]   GET  /           - Root endpoint")
	log.Printf("[SERVER]   GET  /health     - Health check")
	log.Printf("[SERVER]   GET  /api/v1/status - Server status")

	// Start server in goroutine
	go func() {
		log.Printf("[SERVER] Server listening on %s", addr)
		if err := server.ListenAndServe(addr); err != nil {
			log.Fatalf("[SERVER] Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[SERVER] Received shutdown signal at %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("[SERVER] Shutting down server...")

	// Graceful shutdown with timeout
	if err := server.Shutdown(); err != nil {
		log.Printf("[SERVER] Error during shutdown: %v", err)
	} else {
		log.Println("[SERVER] Server shutdown completed successfully")
	}
}

// Root command
var rootCmd = &cobra.Command{
	Use:   "fasthttp-server",
	Short: "A FastHTTP server with comprehensive logging",
	Long:  "A high-performance HTTP server built with FastHTTP and Cobra, featuring request tracing and comprehensive logging",
}

// Initialize commands and flags
func init() {
	// Add server command to root
	rootCmd.AddCommand(serverCmd)

	// Server command flags
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Server port")
	serverCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
}

// Main function
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
