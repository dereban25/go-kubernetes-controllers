package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	// Step 7+ API flags
	apiPort       int
	enableAPIOnly bool
)

// Step 7+: JSON API Response structures
type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
	Count  int         `json:"count,omitempty"`
}

type DeploymentSummary struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	UpdatedReplicas   int32             `json:"updated_replicas"`
	Image             string            `json:"image,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreationTime      time.Time         `json:"creation_time"`
	Age               string            `json:"age"`
	Status            string            `json:"status"`
}

// Enhanced EventProcessor with API server
func (e *EventProcessor) StartAPIServer() {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/", e.handleRootAPI)
	mux.HandleFunc("/api/v1/deployments", e.handleDeploymentsAPI)
	mux.HandleFunc("/api/v1/deployments/", e.handleDeploymentByNameAPI)
	mux.HandleFunc("/api/v1/health", e.handleHealthAPI)
	mux.HandleFunc("/api/v1/cache/stats", e.handleCacheStatsAPI)

	// Enable CORS
	handler := enableCORS(mux)

	port := e.config.APIServer.Port
	log.Printf("🌐 Starting API server on port %d", port)
	log.Printf("📋 Available endpoints:")
	log.Printf("  GET / - API information")
	log.Printf("  GET /api/v1/deployments - List all deployments from cache")
	log.Printf("  GET /api/v1/deployments/{namespace}/{name} - Get specific deployment")
	log.Printf("  GET /api/v1/health - Health check")
	log.Printf("  GET /api/v1/cache/stats - Cache statistics")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ API server failed: %v", err)
	}
}

func (e *EventProcessor) handleRootAPI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	apiInfo := map[string]interface{}{
		"service": "k8s-cli API Server",
		"version": "1.0.0",
		"step":    "Step 7+ - Cache Access API",
		"endpoints": map[string]string{
			"GET /api/v1/deployments":                    "List all deployments",
			"GET /api/v1/deployments/{namespace}/{name}": "Get specific deployment",
			"GET /api/v1/health":                         "Health check",
			"GET /api/v1/cache/stats":                    "Cache statistics",
		},
		"features": []string{
			"Informer cache access",
			"Real-time deployment data",
			"Custom event processing",
			"Configuration support",
		},
	}

	writeJSONResponse(w, APIResponse{
		Status: "success",
		Data:   apiInfo,
	})
}

func (e *EventProcessor) handleDeploymentsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	namespaceFilter := r.URL.Query().Get("namespace")
	labelSelector := r.URL.Query().Get("labelSelector")

	var deployments []DeploymentSummary

	// Use informer cache for efficient access
	for _, obj := range e.cacheIndexer.List() {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			// Apply namespace filter
			if namespaceFilter != "" && deployment.Namespace != namespaceFilter {
				continue
			}

			// Apply label selector filter
			if labelSelector != "" && !matchesLabelSelector(deployment.Labels, labelSelector) {
				continue
			}

			summary := e.createDeploymentSummary(deployment)
			deployments = append(deployments, summary)
		}
	}

	writeJSONResponse(w, APIResponse{
		Status: "success",
		Data:   deployments,
		Count:  len(deployments),
	})
}

func (e *EventProcessor) handleDeploymentByNameAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/v1/deployments/{namespace}/{name}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/deployments/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeErrorResponse(w, "Invalid path. Use /api/v1/deployments/{namespace}/{name}", http.StatusBadRequest)
		return
	}

	namespace, name := parts[0], parts[1]
	key := fmt.Sprintf("%s/%s", namespace, name)

	// Check cache first
	if deployment, exists := e.deploymentCache[key]; exists {
		summary := e.createDeploymentSummary(deployment)
		writeJSONResponse(w, APIResponse{
			Status: "success",
			Data:   summary,
		})
		return
	}

	// Fallback to indexer
	obj, exists, err := e.cacheIndexer.GetByKey(key)
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Error accessing cache: %v", err), http.StatusInternalServerError)
		return
	}

	if !exists {
		writeErrorResponse(w, "Deployment not found", http.StatusNotFound)
		return
	}

	if deployment, ok := obj.(*appsv1.Deployment); ok {
		summary := e.createDeploymentSummary(deployment)
		writeJSONResponse(w, APIResponse{
			Status: "success",
			Data:   summary,
		})
	} else {
		writeErrorResponse(w, "Invalid object type in cache", http.StatusInternalServerError)
	}
}

func (e *EventProcessor) handleHealthAPI(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(e.startTime).Round(time.Second)

	writeJSONResponse(w, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"status":       "healthy",
			"service":      "k8s-cli API Server",
			"step":         "Step 7+ - Cache Access",
			"workers":      e.config.Workers,
			"cache_size":   len(e.deploymentCache),
			"indexer_size": len(e.cacheIndexer.List()),
			"uptime":       uptime.String(),
			"start_time":   e.startTime.Format(time.RFC3339),
		},
	})
}

func (e *EventProcessor) handleCacheStatsAPI(w http.ResponseWriter, r *http.Request) {
	// Group by namespace
	namespaceStats := make(map[string]int)
	var totalReplicas int32
	var healthyDeployments, unhealthyDeployments int

	for _, obj := range e.cacheIndexer.List() {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			namespaceStats[deployment.Namespace]++

			if deployment.Spec.Replicas != nil {
				totalReplicas += *deployment.Spec.Replicas
			}

			if deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
				deployment.Status.Replicas > 0 {
				healthyDeployments++
			} else {
				unhealthyDeployments++
			}
		}
	}

	stats := map[string]interface{}{
		"cache_size":            len(e.deploymentCache),
		"indexer_size":          len(e.cacheIndexer.List()),
		"resync_period":         e.config.ResyncPeriod.String(),
		"workers":               e.config.Workers,
		"namespaces":            e.config.Namespaces,
		"by_namespace":          namespaceStats,
		"total_replicas":        totalReplicas,
		"healthy_deployments":   healthyDeployments,
		"unhealthy_deployments": unhealthyDeployments,
		"uptime":                time.Since(e.startTime).Round(time.Second).String(),
		"step_features": map[string]bool{
			"informer_cache":  true,
			"custom_logic":    e.config.CustomLogic.EnableUpdateHandling,
			"delete_handling": e.config.CustomLogic.EnableDeleteHandling,
			"event_logging":   e.config.LogEvents,
		},
	}

	writeJSONResponse(w, APIResponse{
		Status: "success",
		Data:   stats,
	})
}

func (e *EventProcessor) createDeploymentSummary(deployment *appsv1.Deployment) DeploymentSummary {
	replicas := int32(0)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}

	image := ""
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		image = deployment.Spec.Template.Spec.Containers[0].Image
	}

	age := time.Since(deployment.CreationTimestamp.Time).Round(time.Second).String()

	// Determine deployment status
	status := "Unknown"
	if deployment.Status.ReadyReplicas == deployment.Status.Replicas && deployment.Status.Replicas > 0 {
		status = "Healthy"
	} else if deployment.Status.ReadyReplicas == 0 {
		status = "Unhealthy"
	} else {
		status = "Progressing"
	}

	return DeploymentSummary{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Replicas:          replicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		Image:             image,
		Labels:            deployment.Labels,
		CreationTime:      deployment.CreationTimestamp.Time,
		Age:               age,
		Status:            status,
	}
}

// Helper functions
func writeJSONResponse(w http.ResponseWriter, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Error encoding JSON response: %v", err)
	}
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := APIResponse{
		Status: "error",
		Error:  message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Error encoding error response: %v", err)
	}
}

func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func matchesLabelSelector(labels map[string]string, selector string) bool {
	if labels == nil {
		return false
	}

	// Handle simple key=value selectors
	if strings.Contains(selector, "=") {
		parts := strings.SplitN(selector, "=", 2)
		if len(parts) == 2 {
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			return labels[key] == value
		}
	}

	// Handle key existence selectors
	if strings.Contains(selector, "!=") {
		parts := strings.SplitN(selector, "!=", 2)
		if len(parts) == 2 {
			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			return labels[key] != value
		}
	}

	// Simple key existence check
	_, exists := labels[selector]
	return exists
}

// Step 7+: API server command
var apiServerCmd = &cobra.Command{
	Use:   "api-server",
	Short: "Start JSON API server for cache access (Step 7+)",
	Long:  "Start a JSON API server that provides access to deployment data from informer cache",
	Run: func(cmd *cobra.Command, args []string) {
		runAPIServer()
	},
}

func runAPIServer() {
	log.Println("🎯 Starting k8s-cli API server with informer cache...")

	config, err := loadInformerConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Enable API server
	config.APIServer.Enabled = true
	if apiPort > 0 {
		config.APIServer.Port = apiPort
	}

	log.Printf("⚙️ API Configuration - Port: %d, Workers: %d",
		config.APIServer.Port, config.Workers)

	clientset, err := GetKubernetesClient()
	if err != nil {
		log.Fatalf("❌ Failed to create Kubernetes client: %v", err)
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Kubernetes cluster: %v", err)
	}
	log.Printf("✅ Successfully connected to Kubernetes cluster (version: %s)", serverVersion.String())

	processor := NewEventProcessor(clientset, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start informer
	if err := processor.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start event processor: %v", err)
	}

	// Start API server
	go processor.StartAPIServer()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🎉 k8s-cli API server is running. Press Ctrl+C to stop.")
	log.Printf("🌐 JSON API available at: http://localhost:%d/api/v1/", config.APIServer.Port)
	log.Println("📋 Step 7+ Features:")
	log.Println("   - Informer cache access via JSON API ✓")
	log.Println("   - Custom logic for update/delete events ✓")
	log.Println("   - Real-time deployment data ✓")
	log.Printf("📋 Test the API:")
	log.Printf("  curl http://localhost:%d/api/v1/health", config.APIServer.Port)
	log.Printf("  curl http://localhost:%d/api/v1/deployments", config.APIServer.Port)
	log.Printf("  curl http://localhost:%d/api/v1/cache/stats", config.APIServer.Port)

	<-signalChan
	log.Println("\n🛑 Shutdown signal received, stopping...")

	processor.Stop()
	cancel()

	log.Println("👋 k8s-cli API server stopped gracefully")
}

func init() {
	// Add flags for Step 7+ API
	apiServerCmd.Flags().IntVar(&apiPort, "port", 8080, "API server port")
	apiServerCmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")
	apiServerCmd.Flags().DurationVar(&informerResyncPeriod, "resync-period", 0, "Informer resync period")
	apiServerCmd.Flags().IntVar(&informerWorkers, "workers", 0, "Number of worker goroutines")

	// Register command
	RootCmd.AddCommand(apiServerCmd)
}
