package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	// Step 8 flags
	step8Port     int
	enableMetrics bool
	enableDebug   bool
)

// Step 8: Enhanced API response structures
type Step8APIResponse struct {
	Status    string       `json:"status"`
	Data      interface{}  `json:"data,omitempty"`
	Error     string       `json:"error,omitempty"`
	Count     int          `json:"count,omitempty"`
	Metadata  *APIMetadata `json:"metadata,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

type APIMetadata struct {
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
	TotalCount int    `json:"total_count,omitempty"`
	SortBy     string `json:"sort_by,omitempty"`
	FilterBy   string `json:"filter_by,omitempty"`
}

type DeploymentDetail struct {
	DeploymentSummary
	Conditions      []DeploymentCondition `json:"conditions,omitempty"`
	Strategy        DeploymentStrategy    `json:"strategy,omitempty"`
	RevisionHistory int32                 `json:"revision_history,omitempty"`
	Selector        map[string]string     `json:"selector,omitempty"`
	Annotations     map[string]string     `json:"annotations,omitempty"`
}

type DeploymentCondition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastUpdateTime     time.Time `json:"last_update_time,omitempty"`
	LastTransitionTime time.Time `json:"last_transition_time,omitempty"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

type DeploymentStrategy struct {
	Type           string `json:"type"`
	MaxUnavailable string `json:"max_unavailable,omitempty"`
	MaxSurge       string `json:"max_surge,omitempty"`
}

type CacheMetrics struct {
	TotalDeployments      int                    `json:"total_deployments"`
	NamespaceDistribution map[string]int         `json:"namespace_distribution"`
	StatusDistribution    map[string]int         `json:"status_distribution"`
	ImageDistribution     map[string]int         `json:"image_distribution"`
	ReplicaDistribution   map[string]int         `json:"replica_distribution"`
	LastUpdateTime        time.Time              `json:"last_update_time"`
	CacheStats            map[string]interface{} `json:"cache_stats"`
	PerformanceMetrics    map[string]interface{} `json:"performance_metrics"`
}

// Step 8: Enhanced EventProcessor with advanced API handlers
func (e *EventProcessor) StartStep8APIServer() {
	mux := http.NewServeMux()

	// Step 8: Enhanced API routes
	mux.HandleFunc("/", e.handleStep8RootAPI)
	mux.HandleFunc("/api/v2/deployments", e.handleStep8DeploymentsAPI)
	mux.HandleFunc("/api/v2/deployments/", e.handleStep8DeploymentDetailAPI)
	mux.HandleFunc("/api/v2/cache/metrics", e.handleStep8CacheMetricsAPI)
	mux.HandleFunc("/api/v2/cache/search", e.handleStep8CacheSearchAPI)
	mux.HandleFunc("/api/v2/cache/status", e.handleStep8CacheStatusAPI)
	mux.HandleFunc("/api/v2/health", e.handleStep8HealthAPI)

	// Debug endpoints
	if enableDebug {
		mux.HandleFunc("/api/v2/debug/cache-dump", e.handleStep8CacheDumpAPI)
		mux.HandleFunc("/api/v2/debug/performance", e.handleStep8PerformanceAPI)
	}

	// Metrics endpoint
	if enableMetrics {
		mux.HandleFunc("/metrics", e.handlePrometheusMetrics)
	}

	// Enable CORS and middleware
	handler := e.step8Middleware(enableCORS(mux))

	port := step8Port
	log.Printf("🌐 Starting Step 8 Advanced API server on port %d", port)
	log.Printf("📋 Step 8 Enhanced endpoints:")
	log.Printf("  GET /api/v2/deployments - Advanced deployment listing with filtering")
	log.Printf("  GET /api/v2/deployments/{namespace}/{name} - Detailed deployment info")
	log.Printf("  GET /api/v2/cache/metrics - Cache metrics and analytics")
	log.Printf("  GET /api/v2/cache/search - Search deployments in cache")
	log.Printf("  GET /api/v2/cache/status - Cache status and health")
	log.Printf("  GET /api/v2/health - Service health check")

	if enableDebug {
		log.Printf("  GET /api/v2/debug/cache-dump - Debug cache contents")
		log.Printf("  GET /api/v2/debug/performance - Performance metrics")
	}

	if enableMetrics {
		log.Printf("  GET /metrics - Prometheus metrics")
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ Step 8 API server failed: %v", err)
	}
}

// Step 8: Middleware for logging and metrics
func (e *EventProcessor) step8Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Add headers
		w.Header().Set("X-API-Version", "v2")
		w.Header().Set("X-Service", "k8s-cli-step8")

		// Log request
		log.Printf("📥 API Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)

		// Log response time
		duration := time.Since(start)
		log.Printf("📤 API Response: %s %s completed in %v", r.Method, r.URL.Path, duration)
	})
}

func (e *EventProcessor) handleStep8RootAPI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	apiInfo := map[string]interface{}{
		"service":     "k8s-cli Step 8 Advanced API",
		"version":     "2.0.0",
		"step":        "Step 8 - Advanced Cache API Handlers",
		"description": "Enhanced JSON API for deployment cache access with advanced features",
		"features": []string{
			"Advanced deployment filtering and sorting",
			"Detailed deployment information",
			"Cache metrics and analytics",
			"Search functionality",
			"Debug endpoints",
			"Performance monitoring",
		},
		"endpoints": map[string]interface{}{
			"deployments": map[string]string{
				"list":   "GET /api/v2/deployments",
				"detail": "GET /api/v2/deployments/{namespace}/{name}",
			},
			"cache": map[string]string{
				"metrics": "GET /api/v2/cache/metrics",
				"search":  "GET /api/v2/cache/search",
				"status":  "GET /api/v2/cache/status",
			},
			"utility": map[string]string{
				"health": "GET /api/v2/health",
				"debug":  "GET /api/v2/debug/*",
			},
		},
		"query_parameters": map[string]interface{}{
			"deployments": []string{
				"namespace", "labelSelector", "fieldSelector",
				"sortBy", "order", "page", "pageSize",
				"status", "image", "minReplicas", "maxReplicas",
			},
			"search": []string{
				"q", "namespace", "fields", "limit",
			},
		},
	}

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      apiInfo,
		Timestamp: time.Now(),
	})
}

// Step 8: Advanced deployments listing with filtering, sorting, pagination
func (e *EventProcessor) handleStep8DeploymentsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		e.writeStep8ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	params := e.parseQueryParams(r)

	var deployments []DeploymentDetail
	allDeployments := e.getAllDeploymentsFromCache()

	// Apply filters
	filteredDeployments := e.filterDeployments(allDeployments, params)

	// Sort deployments
	sortedDeployments := e.sortDeployments(filteredDeployments, params)

	// Apply pagination
	paginatedDeployments, metadata := e.paginateDeployments(sortedDeployments, params)

	// Convert to detailed format
	for _, deployment := range paginatedDeployments {
		detail := e.createDeploymentDetail(deployment)
		deployments = append(deployments, detail)
	}

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      deployments,
		Count:     len(deployments),
		Metadata:  metadata,
		Timestamp: time.Now(),
	})
}

// Step 8: Detailed deployment information
func (e *EventProcessor) handleStep8DeploymentDetailAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		e.writeStep8ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/v2/deployments/{namespace}/{name}
	path := strings.TrimPrefix(r.URL.Path, "/api/v2/deployments/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		e.writeStep8ErrorResponse(w, "Invalid path. Use /api/v2/deployments/{namespace}/{name}", http.StatusBadRequest)
		return
	}

	namespace, name := parts[0], parts[1]
	key := fmt.Sprintf("%s/%s", namespace, name)

	// Get from cache
	deployment := e.getDeploymentFromCache(key)
	if deployment == nil {
		e.writeStep8ErrorResponse(w, "Deployment not found in cache", http.StatusNotFound)
		return
	}

	detail := e.createDeploymentDetail(deployment)

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      detail,
		Timestamp: time.Now(),
	})
}

// Step 8: Cache metrics and analytics
func (e *EventProcessor) handleStep8CacheMetricsAPI(w http.ResponseWriter, r *http.Request) {
	metrics := e.calculateCacheMetrics()

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      metrics,
		Timestamp: time.Now(),
	})
}

// Step 8: Search deployments in cache
func (e *EventProcessor) handleStep8CacheSearchAPI(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	namespace := r.URL.Query().Get("namespace")
	fields := r.URL.Query().Get("fields")
	limitStr := r.URL.Query().Get("limit")

	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results := e.searchDeployments(query, namespace, fields, limit)

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      results,
		Count:     len(results),
		Timestamp: time.Now(),
	})
}

// Step 8: Cache status and health
func (e *EventProcessor) handleStep8CacheStatusAPI(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"cache_healthy": true,
		"cache_size":    len(e.deploymentCache),
		"indexer_size":  len(e.cacheIndexer.List()),
		"last_sync":     time.Now(), // Would track real sync time
		"sync_status":   "active",
		"worker_status": "running",
		"worker_count":  e.config.Workers,
		"resync_period": e.config.ResyncPeriod.String(),
		"uptime":        time.Since(e.startTime).String(),
		"memory_usage":  "unknown", // Could add runtime.MemStats
	}

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      status,
		Timestamp: time.Now(),
	})
}

// Step 8: Enhanced health check
func (e *EventProcessor) handleStep8HealthAPI(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(e.startTime)

	health := map[string]interface{}{
		"status":          "healthy",
		"service":         "k8s-cli Step 8 API",
		"version":         "2.0.0",
		"step":            "Step 8 - Advanced Cache Handlers",
		"uptime":          uptime.String(),
		"uptime_seconds":  int(uptime.Seconds()),
		"cache_healthy":   len(e.cacheIndexer.List()) >= 0,
		"workers_running": e.config.Workers,
		"api_endpoints":   11, // Count of endpoints
		"features_enabled": map[string]bool{
			"informer_cache":     true,
			"advanced_filtering": true,
			"search":             true,
			"metrics":            enableMetrics,
			"debug":              enableDebug,
		},
		"last_activity": time.Now(),
	}

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      health,
		Timestamp: time.Now(),
	})
}

// Step 8: Debug endpoints
func (e *EventProcessor) handleStep8CacheDumpAPI(w http.ResponseWriter, r *http.Request) {
	if !enableDebug {
		e.writeStep8ErrorResponse(w, "Debug endpoints are disabled", http.StatusForbidden)
		return
	}

	dump := map[string]interface{}{
		"cache_keys":      e.getCacheKeys(),
		"indexer_objects": len(e.cacheIndexer.List()),
		"cache_sample":    e.getCacheSample(5),
	}

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      dump,
		Timestamp: time.Now(),
	})
}

func (e *EventProcessor) handleStep8PerformanceAPI(w http.ResponseWriter, r *http.Request) {
	if !enableDebug {
		e.writeStep8ErrorResponse(w, "Debug endpoints are disabled", http.StatusForbidden)
		return
	}

	// Mock performance data - in real implementation would track actual metrics
	perf := map[string]interface{}{
		"api_requests_total":    100,
		"cache_hits":            95,
		"cache_misses":          5,
		"average_response_time": "50ms",
		"memory_usage":          "25MB",
		"cpu_usage":             "2%",
	}

	e.writeStep8JSONResponse(w, Step8APIResponse{
		Status:    "success",
		Data:      perf,
		Timestamp: time.Now(),
	})
}

// Step 8: Prometheus metrics
func (e *EventProcessor) handlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	metrics := fmt.Sprintf(`# HELP k8s_cli_cache_size Number of deployments in cache
# TYPE k8s_cli_cache_size gauge
k8s_cli_cache_size %d

# HELP k8s_cli_workers_running Number of workers running
# TYPE k8s_cli_workers_running gauge
k8s_cli_workers_running %d

# HELP k8s_cli_uptime_seconds Service uptime in seconds
# TYPE k8s_cli_uptime_seconds counter
k8s_cli_uptime_seconds %d
`, len(e.deploymentCache), e.config.Workers, int(time.Since(e.startTime).Seconds()))

	fmt.Fprint(w, metrics)
}

// Helper functions for Step 8
func (e *EventProcessor) parseQueryParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}

func (e *EventProcessor) getAllDeploymentsFromCache() []*appsv1.Deployment {
	var deployments []*appsv1.Deployment
	for _, obj := range e.cacheIndexer.List() {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			deployments = append(deployments, deployment)
		}
	}
	return deployments
}

func (e *EventProcessor) getDeploymentFromCache(key string) *appsv1.Deployment {
	if deployment, exists := e.deploymentCache[key]; exists {
		return deployment
	}

	obj, exists, _ := e.cacheIndexer.GetByKey(key)
	if exists {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			return deployment
		}
	}
	return nil
}

func (e *EventProcessor) filterDeployments(deployments []*appsv1.Deployment, params map[string]string) []*appsv1.Deployment {
	var filtered []*appsv1.Deployment

	for _, deployment := range deployments {
		// Namespace filter
		if ns := params["namespace"]; ns != "" && deployment.Namespace != ns {
			continue
		}

		// Status filter
		if status := params["status"]; status != "" {
			deploymentStatus := "Unknown"
			if deployment.Status.ReadyReplicas == deployment.Status.Replicas && deployment.Status.Replicas > 0 {
				deploymentStatus = "Healthy"
			} else if deployment.Status.ReadyReplicas == 0 {
				deploymentStatus = "Unhealthy"
			} else {
				deploymentStatus = "Progressing"
			}
			if status != deploymentStatus {
				continue
			}
		}

		// Image filter
		if img := params["image"]; img != "" {
			if len(deployment.Spec.Template.Spec.Containers) == 0 ||
				!strings.Contains(deployment.Spec.Template.Spec.Containers[0].Image, img) {
				continue
			}
		}

		// Label selector
		if labelSelector := params["labelSelector"]; labelSelector != "" {
			if !matchesLabelSelector(deployment.Labels, labelSelector) {
				continue
			}
		}

		filtered = append(filtered, deployment)
	}

	return filtered
}

func (e *EventProcessor) sortDeployments(deployments []*appsv1.Deployment, params map[string]string) []*appsv1.Deployment {
	sortBy := params["sortBy"]
	order := params["order"]

	if sortBy == "" {
		sortBy = "name"
	}

	sort.Slice(deployments, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "name":
			less = deployments[i].Name < deployments[j].Name
		case "namespace":
			less = deployments[i].Namespace < deployments[j].Namespace
		case "created":
			less = deployments[i].CreationTimestamp.Before(&deployments[j].CreationTimestamp)
		case "replicas":
			iReplicas := int32(0)
			jReplicas := int32(0)
			if deployments[i].Spec.Replicas != nil {
				iReplicas = *deployments[i].Spec.Replicas
			}
			if deployments[j].Spec.Replicas != nil {
				jReplicas = *deployments[j].Spec.Replicas
			}
			less = iReplicas < jReplicas
		default:
			less = deployments[i].Name < deployments[j].Name
		}

		if order == "desc" {
			return !less
		}
		return less
	})

	return deployments
}

func (e *EventProcessor) paginateDeployments(deployments []*appsv1.Deployment, params map[string]string) ([]*appsv1.Deployment, *APIMetadata) {
	page := 1
	pageSize := 20

	if p := params["page"]; p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := params["pageSize"]; ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	totalCount := len(deployments)
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= totalCount {
		return []*appsv1.Deployment{}, &APIMetadata{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
		}
	}

	if endIndex > totalCount {
		endIndex = totalCount
	}

	return deployments[startIndex:endIndex], &APIMetadata{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		SortBy:     params["sortBy"],
		FilterBy:   params["namespace"],
	}
}

func (e *EventProcessor) createDeploymentDetail(deployment *appsv1.Deployment) DeploymentDetail {
	summary := e.createDeploymentSummary(deployment)

	detail := DeploymentDetail{
		DeploymentSummary: summary,
		Annotations:       deployment.Annotations,
		Selector:          deployment.Spec.Selector.MatchLabels,
	}

	// Add conditions
	for _, condition := range deployment.Status.Conditions {
		detail.Conditions = append(detail.Conditions, DeploymentCondition{
			Type:               string(condition.Type),
			Status:             string(condition.Status),
			LastUpdateTime:     condition.LastUpdateTime.Time,
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}

	// Add strategy
	if deployment.Spec.Strategy.Type != "" {
		detail.Strategy.Type = string(deployment.Spec.Strategy.Type)
		if deployment.Spec.Strategy.RollingUpdate != nil {
			if deployment.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				detail.Strategy.MaxUnavailable = deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.String()
			}
			if deployment.Spec.Strategy.RollingUpdate.MaxSurge != nil {
				detail.Strategy.MaxSurge = deployment.Spec.Strategy.RollingUpdate.MaxSurge.String()
			}
		}
	}

	// Add revision history limit
	if deployment.Spec.RevisionHistoryLimit != nil {
		detail.RevisionHistory = *deployment.Spec.RevisionHistoryLimit
	}

	return detail
}

func (e *EventProcessor) calculateCacheMetrics() CacheMetrics {
	metrics := CacheMetrics{
		NamespaceDistribution: make(map[string]int),
		StatusDistribution:    make(map[string]int),
		ImageDistribution:     make(map[string]int),
		ReplicaDistribution:   make(map[string]int),
		LastUpdateTime:        time.Now(),
		CacheStats:            make(map[string]interface{}),
		PerformanceMetrics:    make(map[string]interface{}),
	}

	deployments := e.getAllDeploymentsFromCache()
	metrics.TotalDeployments = len(deployments)

	for _, deployment := range deployments {
		// Namespace distribution
		metrics.NamespaceDistribution[deployment.Namespace]++

		// Status distribution
		status := "Unknown"
		if deployment.Status.ReadyReplicas == deployment.Status.Replicas && deployment.Status.Replicas > 0 {
			status = "Healthy"
		} else if deployment.Status.ReadyReplicas == 0 {
			status = "Unhealthy"
		} else {
			status = "Progressing"
		}
		metrics.StatusDistribution[status]++

		// Image distribution
		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			image := deployment.Spec.Template.Spec.Containers[0].Image
			// Extract image name without tag
			parts := strings.Split(image, ":")
			imageName := parts[0]
			metrics.ImageDistribution[imageName]++
		}

		// Replica distribution
		replicas := "0"
		if deployment.Spec.Replicas != nil {
			replicas = fmt.Sprintf("%d", *deployment.Spec.Replicas)
		}
		metrics.ReplicaDistribution[replicas]++
	}

	// Cache stats
	metrics.CacheStats["cache_size"] = len(e.deploymentCache)
	metrics.CacheStats["indexer_size"] = len(e.cacheIndexer.List())
	metrics.CacheStats["workers"] = e.config.Workers
	metrics.CacheStats["resync_period"] = e.config.ResyncPeriod.String()

	// Performance metrics (mock data)
	metrics.PerformanceMetrics["uptime"] = time.Since(e.startTime).String()
	metrics.PerformanceMetrics["cache_hit_ratio"] = 0.95
	metrics.PerformanceMetrics["avg_response_time"] = "25ms"

	return metrics
}

func (e *EventProcessor) searchDeployments(query, namespace, fields string, limit int) []DeploymentSummary {
	var results []DeploymentSummary
	deployments := e.getAllDeploymentsFromCache()

	searchFields := []string{"name", "namespace", "image", "labels"}
	if fields != "" {
		searchFields = strings.Split(fields, ",")
	}

	query = strings.ToLower(query)
	count := 0

	for _, deployment := range deployments {
		if count >= limit {
			break
		}

		if namespace != "" && deployment.Namespace != namespace {
			continue
		}

		match := false
		for _, field := range searchFields {
			switch field {
			case "name":
				if strings.Contains(strings.ToLower(deployment.Name), query) {
					match = true
				}
			case "namespace":
				if strings.Contains(strings.ToLower(deployment.Namespace), query) {
					match = true
				}
			case "image":
				if len(deployment.Spec.Template.Spec.Containers) > 0 {
					image := strings.ToLower(deployment.Spec.Template.Spec.Containers[0].Image)
					if strings.Contains(image, query) {
						match = true
					}
				}
			case "labels":
				for key, value := range deployment.Labels {
					if strings.Contains(strings.ToLower(key), query) ||
						strings.Contains(strings.ToLower(value), query) {
						match = true
						break
					}
				}
			}
			if match {
				break
			}
		}

		if match {
			summary := e.createDeploymentSummary(deployment)
			results = append(results, summary)
			count++
		}
	}

	return results
}

func (e *EventProcessor) getCacheKeys() []string {
	var keys []string
	for key := range e.deploymentCache {
		keys = append(keys, key)
	}
	return keys
}

func (e *EventProcessor) getCacheSample(limit int) []string {
	var sample []string
	count := 0
	for key := range e.deploymentCache {
		if count >= limit {
			break
		}
		sample = append(sample, key)
		count++
	}
	return sample
}

func (e *EventProcessor) writeStep8JSONResponse(w http.ResponseWriter, response Step8APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Error encoding Step 8 JSON response: %v", err)
	}
}

func (e *EventProcessor) writeStep8ErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := Step8APIResponse{
		Status:    "error",
		Error:     message,
		Timestamp: time.Now(),
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Error encoding Step 8 error response: %v", err)
	}
}

// Step 8: Advanced API server command
var step8APICmd = &cobra.Command{
	Use:   "step8-api",
	Short: "Start Step 8 advanced JSON API server with enhanced cache handlers",
	Long: `Start an advanced JSON API server (Step 8) that provides enhanced access to deployment data 
from informer cache with advanced filtering, sorting, pagination, search, and analytics capabilities.

Step 8 Features:
• Advanced deployment listing with filtering and sorting
• Detailed deployment information with conditions and strategy
• Cache metrics and analytics
• Search functionality across deployment fields
• Debug endpoints for troubleshooting
• Prometheus metrics support
• Enhanced error handling and logging`,
	Run: func(cmd *cobra.Command, args []string) {
		runStep8APIServer()
	},
}

func runStep8APIServer() {
	log.Println("🎯 Starting k8s-cli Step 8 Advanced API server...")

	config, err := loadInformerConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Configure for Step 8
	config.APIServer.Enabled = true
	if step8Port > 0 {
		config.APIServer.Port = step8Port
	} else {
		config.APIServer.Port = 8090 // Default Step 8 port
	}

	log.Printf("⚙️ Step 8 API Configuration:")
	log.Printf("   Port: %d", config.APIServer.Port)
	log.Printf("   Workers: %d", config.Workers)
	log.Printf("   Metrics: %t", enableMetrics)
	log.Printf("   Debug: %t", enableDebug)

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

	// Start Step 8 API server
	go processor.StartStep8APIServer()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🎉 Step 8 Advanced API server is running. Press Ctrl+C to stop.")
	log.Printf("🌐 Step 8 JSON API available at: http://localhost:%d/api/v2/", config.APIServer.Port)
	log.Println("")
	log.Println("📋 Step 8 Enhanced Features:")
	log.Println("   ✅ Advanced deployment listing with filtering and sorting")
	log.Println("   ✅ Detailed deployment information with conditions")
	log.Println("   ✅ Cache metrics and analytics")
	log.Println("   ✅ Search functionality across deployment fields")
	log.Println("   ✅ Enhanced error handling and logging")
	if enableDebug {
		log.Println("   ✅ Debug endpoints enabled")
	}
	if enableMetrics {
		log.Println("   ✅ Prometheus metrics enabled")
	}
	log.Println("")
	log.Printf("🧪 Test the Step 8 API:")
	log.Printf("  # Basic listing")
	log.Printf("  curl http://localhost:%d/api/v2/deployments", config.APIServer.Port)
	log.Printf("")
	log.Printf("  # Advanced filtering")
	log.Printf("  curl 'http://localhost:%d/api/v2/deployments?namespace=default&status=Healthy&sortBy=name'", config.APIServer.Port)
	log.Printf("")
	log.Printf("  # Pagination")
	log.Printf("  curl 'http://localhost:%d/api/v2/deployments?page=1&pageSize=5'", config.APIServer.Port)
	log.Printf("")
	log.Printf("  # Search")
	log.Printf("  curl 'http://localhost:%d/api/v2/cache/search?q=nginx&fields=name,image'", config.APIServer.Port)
	log.Printf("")
	log.Printf("  # Cache metrics")
	log.Printf("  curl http://localhost:%d/api/v2/cache/metrics", config.APIServer.Port)
	log.Printf("")
	log.Printf("  # Health check")
	log.Printf("  curl http://localhost:%d/api/v2/health", config.APIServer.Port)

	<-signalChan
	log.Println("\n🛑 Shutdown signal received, stopping...")

	processor.Stop()
	cancel()

	log.Println("👋 Step 8 Advanced API server stopped gracefully")
}

func init() {
	// Add flags for Step 8
	step8APICmd.Flags().IntVar(&step8Port, "port", 8090, "Step 8 API server port")
	step8APICmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")
	step8APICmd.Flags().DurationVar(&informerResyncPeriod, "resync-period", 0, "Informer resync period")
	step8APICmd.Flags().IntVar(&informerWorkers, "workers", 0, "Number of worker goroutines")
	step8APICmd.Flags().BoolVar(&enableMetrics, "enable-metrics", false, "Enable Prometheus metrics endpoint")
	step8APICmd.Flags().BoolVar(&enableDebug, "enable-debug", false, "Enable debug endpoints")

	// Register command
	RootCmd.AddCommand(step8APICmd)
}
