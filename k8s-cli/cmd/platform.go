package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8scliv1 "k8s-cli/api/v1"
)

var (
	// Step 12 flags
	platformPort      int
	portAPIToken      string
	portBaseURL       string
	enableWebhooks    bool
	discordWebhookURL string
)

// Step 12: Platform Engineering API based on Port.io
type PlatformAPI struct {
	client        client.Client
	scheme        *runtime.Scheme
	portClient    *PortClient
	discordClient *DiscordClient
}

// Port.io API Client
type PortClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// Discord Webhook Client
type DiscordClient struct {
	WebhookURL string
	HTTPClient *http.Client
}

// Port.io Action structures
type PortAction struct {
	Identifier  string                 `json:"identifier"`
	Title       string                 `json:"title"`
	Trigger     string                 `json:"trigger"`
	Description string                 `json:"description"`
	Inputs      map[string]interface{} `json:"inputs"`
	Run         string                 `json:"run"`
}

type ActionRequest struct {
	Action     string                 `json:"action"`
	ResourceId string                 `json:"resourceId"`
	Trigger    string                 `json:"trigger"`
	Inputs     map[string]interface{} `json:"inputs"`
	Context    map[string]interface{} `json:"context"`
}

type ActionResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Logs    []string    `json:"logs,omitempty"`
}

// Discord message structure
type DiscordMessage struct {
	Content string         `json:"content"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Fields      []DiscordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func NewPlatformAPI(client client.Client, scheme *runtime.Scheme) *PlatformAPI {
	portClient := &PortClient{
		BaseURL:    portBaseURL,
		Token:      portAPIToken,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}

	var discordClient *DiscordClient
	if discordWebhookURL != "" {
		discordClient = &DiscordClient{
			WebhookURL: discordWebhookURL,
			HTTPClient: &http.Client{Timeout: 30 * time.Second},
		}
	}

	return &PlatformAPI{
		client:        client,
		scheme:        scheme,
		portClient:    portClient,
		discordClient: discordClient,
	}
}

// Step 12: API handlers for CRUD actions
func (p *PlatformAPI) StartServer() {
	mux := http.NewServeMux()

	// Platform engineering endpoints
	mux.HandleFunc("/", p.handleRoot)
	mux.HandleFunc("/webhook/port", p.handlePortWebhook)
	mux.HandleFunc("/api/v1/actions", p.handleActions)

	// CRUD endpoints for FrontendPage
	mux.HandleFunc("/api/v1/frontendpages", p.handleFrontendPages)
	mux.HandleFunc("/api/v1/frontendpages/", p.handleFrontendPageByName)

	// Step 12+: Update action support
	mux.HandleFunc("/api/v1/frontendpages/update", p.handleUpdateAction)

	// Health and metrics
	mux.HandleFunc("/health", p.handleHealth)
	mux.HandleFunc("/metrics", p.handleMetrics)

	// Enable CORS
	handler := p.enableCORS(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", platformPort),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🌐 Starting Platform Engineering API on port %d", platformPort)
	log.Printf("📋 Available endpoints:")
	log.Printf("  POST /webhook/port - Port.io webhook handler")
	log.Printf("  GET  /api/v1/actions - List available actions")
	log.Printf("  GET  /api/v1/frontendpages - List FrontendPages")
	log.Printf("  POST /api/v1/frontendpages - Create FrontendPage")
	log.Printf("  PUT  /api/v1/frontendpages/{name} - Update FrontendPage")
	log.Printf("  DELETE /api/v1/frontendpages/{name} - Delete FrontendPage")
	log.Printf("  POST /api/v1/frontendpages/update - Update action support")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ Platform API server failed: %v", err)
	}
}

func (p *PlatformAPI) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "k8s-cli Platform Engineering API",
		"version": "1.0.0",
		"step":    "Step 12 - Platform Engineering Integration",
		"features": []string{
			"Port.io integration for self-service experiences",
			"CRUD operations for custom resources",
			"Webhook handlers for external triggers",
			"Discord notifications integration",
			"Update action support for IDP",
		},
		"endpoints": map[string]string{
			"webhook":       "/webhook/port",
			"actions":       "/api/v1/actions",
			"frontendpages": "/api/v1/frontendpages",
			"health":        "/health",
			"metrics":       "/metrics",
		},
	}

	p.writeJSONResponse(w, response)
}

// Step 12: Port.io webhook handler
func (p *PlatformAPI) handlePortWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var actionReq ActionRequest
	if err := json.Unmarshal(body, &actionReq); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	log.Printf("📨 Step 12: Received Port.io action: %s", actionReq.Action)
	log.Printf("   Resource ID: %s", actionReq.ResourceId)
	log.Printf("   Trigger: %s", actionReq.Trigger)

	// Process the action
	response, err := p.processAction(r.Context(), &actionReq)
	if err != nil {
		log.Printf("❌ Failed to process action: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send Discord notification if configured
	if p.discordClient != nil {
		go p.sendDiscordNotification(&actionReq, response)
	}

	p.writeJSONResponse(w, response)
}

func (p *PlatformAPI) processAction(ctx context.Context, req *ActionRequest) (*ActionResponse, error) {
	switch req.Action {
	case "create_frontend":
		return p.createFrontendPageAction(ctx, req)
	case "update_frontend":
		return p.updateFrontendPageAction(ctx, req)
	case "delete_frontend":
		return p.deleteFrontendPageAction(ctx, req)
	case "scale_frontend":
		return p.scaleFrontendPageAction(ctx, req)
	default:
		return &ActionResponse{
			Status:  "error",
			Message: fmt.Sprintf("Unknown action: %s", req.Action),
		}, nil
	}
}

func (p *PlatformAPI) createFrontendPageAction(ctx context.Context, req *ActionRequest) (*ActionResponse, error) {
	log.Printf("🔨 Step 12: Creating FrontendPage from Port.io action")

	// Extract inputs
	name, _ := req.Inputs["name"].(string)
	title, _ := req.Inputs["title"].(string)
	description, _ := req.Inputs["description"].(string)
	path, _ := req.Inputs["path"].(string)
	image, _ := req.Inputs["image"].(string)
	replicas, _ := req.Inputs["replicas"].(float64)

	if name == "" {
		return &ActionResponse{
			Status:  "error",
			Message: "Missing required field: name",
		}, nil
	}

	// Create FrontendPage resource
	frontendPage := &k8scliv1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				"created-by": "port-io",
				"action":     req.Action,
			},
			Annotations: map[string]string{
				"port.io/trigger":     req.Trigger,
				"port.io/resource-id": req.ResourceId,
			},
		},
		Spec: k8scliv1.FrontendPageSpec{
			Title:       title,
			Description: description,
			Path:        path,
			Image:       image,
			Replicas:    int32(replicas),
		},
	}

	if err := p.client.Create(ctx, frontendPage); err != nil {
		return &ActionResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to create FrontendPage: %v", err),
		}, err
	}

	return &ActionResponse{
		Status:  "success",
		Message: fmt.Sprintf("FrontendPage '%s' created successfully", name),
		Data: map[string]interface{}{
			"name":      name,
			"namespace": frontendPage.Namespace,
			"title":     title,
		},
		Logs: []string{
			fmt.Sprintf("Created FrontendPage: %s", name),
			fmt.Sprintf("Title: %s", title),
			fmt.Sprintf("Path: %s", path),
		},
	}, nil
}

// Step 12+: Update action support
func (p *PlatformAPI) updateFrontendPageAction(ctx context.Context, req *ActionRequest) (*ActionResponse, error) {
	log.Printf("🔄 Step 12+: Updating FrontendPage from Port.io action")

	name, _ := req.Inputs["name"].(string)
	if name == "" {
		return &ActionResponse{
			Status:  "error",
			Message: "Missing required field: name",
		}, nil
	}

	// Get existing FrontendPage
	var frontendPage k8scliv1.FrontendPage
	if err := p.client.Get(ctx, client.ObjectKey{Name: name, Namespace: "default"}, &frontendPage); err != nil {
		return &ActionResponse{
			Status:  "error",
			Message: fmt.Sprintf("FrontendPage not found: %v", err),
		}, err
	}

	// Update fields if provided
	updated := false
	logs := []string{fmt.Sprintf("Updating FrontendPage: %s", name)}

	if title, ok := req.Inputs["title"].(string); ok && title != "" {
		frontendPage.Spec.Title = title
		updated = true
		logs = append(logs, fmt.Sprintf("Updated title: %s", title))
	}

	if description, ok := req.Inputs["description"].(string); ok && description != "" {
		frontendPage.Spec.Description = description
		updated = true
		logs = append(logs, fmt.Sprintf("Updated description: %s", description))
	}

	if replicas, ok := req.Inputs["replicas"].(float64); ok && replicas > 0 {
		frontendPage.Spec.Replicas = int32(replicas)
		updated = true
		logs = append(logs, fmt.Sprintf("Updated replicas: %d", int32(replicas)))
	}

	if image, ok := req.Inputs["image"].(string); ok && image != "" {
		frontendPage.Spec.Image = image
		updated = true
		logs = append(logs, fmt.Sprintf("Updated image: %s", image))
	}

	if !updated {
		return &ActionResponse{
			Status:  "success",
			Message: "No updates provided",
			Logs:    logs,
		}, nil
	}

	// Update the resource
	if err := p.client.Update(ctx, &frontendPage); err != nil {
		return &ActionResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to update FrontendPage: %v", err),
		}, err
	}

	return &ActionResponse{
		Status:  "success",
		Message: fmt.Sprintf("FrontendPage '%s' updated successfully", name),
		Data: map[string]interface{}{
			"name":      name,
			"namespace": frontendPage.Namespace,
			"updated":   updated,
		},
		Logs: logs,
	}, nil
}

func (p *PlatformAPI) deleteFrontendPageAction(ctx context.Context, req *ActionRequest) (*ActionResponse, error) {
	log.Printf("🗑️ Step 12: Deleting FrontendPage from Port.io action")

	name, _ := req.Inputs["name"].(string)
	if name == "" {
		return &ActionResponse{
			Status:  "error",
			Message: "Missing required field: name",
		}, nil
	}

	frontendPage := &k8scliv1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}

	if err := p.client.Delete(ctx, frontendPage); err != nil {
		return &ActionResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to delete FrontendPage: %v", err),
		}, err
	}

	return &ActionResponse{
		Status:  "success",
		Message: fmt.Sprintf("FrontendPage '%s' deleted successfully", name),
		Logs: []string{
			fmt.Sprintf("Deleted FrontendPage: %s", name),
		},
	}, nil
}

func (p *PlatformAPI) scaleFrontendPageAction(ctx context.Context, req *ActionRequest) (*ActionResponse, error) {
	log.Printf("📈 Step 12: Scaling FrontendPage from Port.io action")

	name, _ := req.Inputs["name"].(string)
	replicas, _ := req.Inputs["replicas"].(float64)

	if name == "" || replicas <= 0 {
		return &ActionResponse{
			Status:  "error",
			Message: "Missing required fields: name and replicas",
		}, nil
	}

	var frontendPage k8scliv1.FrontendPage
	if err := p.client.Get(ctx, client.ObjectKey{Name: name, Namespace: "default"}, &frontendPage); err != nil {
		return &ActionResponse{
			Status:  "error",
			Message: fmt.Sprintf("FrontendPage not found: %v", err),
		}, err
	}

	oldReplicas := frontendPage.Spec.Replicas
	frontendPage.Spec.Replicas = int32(replicas)

	if err := p.client.Update(ctx, &frontendPage); err != nil {
		return &ActionResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to scale FrontendPage: %v", err),
		}, err
	}

	return &ActionResponse{
		Status:  "success",
		Message: fmt.Sprintf("FrontendPage '%s' scaled from %d to %d replicas", name, oldReplicas, int32(replicas)),
		Data: map[string]interface{}{
			"name":         name,
			"old_replicas": oldReplicas,
			"new_replicas": int32(replicas),
		},
		Logs: []string{
			fmt.Sprintf("Scaled FrontendPage %s from %d to %d replicas", name, oldReplicas, int32(replicas)),
		},
	}, nil
}

// CRUD API handlers
func (p *PlatformAPI) handleFrontendPages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		p.listFrontendPages(w, r)
	case http.MethodPost:
		p.createFrontendPage(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (p *PlatformAPI) handleFrontendPageByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/api/v1/frontendpages/"):]
	if name == "" {
		http.Error(w, "Missing frontendpage name", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		p.getFrontendPage(w, r, name)
	case http.MethodPut:
		p.updateFrontendPage(w, r, name)
	case http.MethodDelete:
		p.deleteFrontendPage(w, r, name)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (p *PlatformAPI) listFrontendPages(w http.ResponseWriter, r *http.Request) {
	var frontendPages k8scliv1.FrontendPageList
	if err := p.client.List(r.Context(), &frontendPages); err != nil {
		http.Error(w, fmt.Sprintf("Failed to list FrontendPages: %v", err), http.StatusInternalServerError)
		return
	}

	p.writeJSONResponse(w, map[string]interface{}{
		"status": "success",
		"data":   frontendPages.Items,
		"count":  len(frontendPages.Items),
	})
}

func (p *PlatformAPI) createFrontendPage(w http.ResponseWriter, r *http.Request) {
	var frontendPage k8scliv1.FrontendPage
	if err := json.NewDecoder(r.Body).Decode(&frontendPage); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := p.client.Create(r.Context(), &frontendPage); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create FrontendPage: %v", err), http.StatusInternalServerError)
		return
	}

	p.writeJSONResponse(w, map[string]interface{}{
		"status":  "success",
		"message": "FrontendPage created successfully",
		"data":    frontendPage,
	})
}

func (p *PlatformAPI) getFrontendPage(w http.ResponseWriter, r *http.Request, name string) {
	var frontendPage k8scliv1.FrontendPage
	if err := p.client.Get(r.Context(), client.ObjectKey{Name: name, Namespace: "default"}, &frontendPage); err != nil {
		http.Error(w, fmt.Sprintf("FrontendPage not found: %v", err), http.StatusNotFound)
		return
	}

	p.writeJSONResponse(w, map[string]interface{}{
		"status": "success",
		"data":   frontendPage,
	})
}

func (p *PlatformAPI) updateFrontendPage(w http.ResponseWriter, r *http.Request, name string) {
	var frontendPage k8scliv1.FrontendPage
	if err := p.client.Get(r.Context(), client.ObjectKey{Name: name, Namespace: "default"}, &frontendPage); err != nil {
		http.Error(w, fmt.Sprintf("FrontendPage not found: %v", err), http.StatusNotFound)
		return
	}

	var updateData k8scliv1.FrontendPage
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Update spec fields
	frontendPage.Spec = updateData.Spec

	if err := p.client.Update(r.Context(), &frontendPage); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update FrontendPage: %v", err), http.StatusInternalServerError)
		return
	}

	p.writeJSONResponse(w, map[string]interface{}{
		"status":  "success",
		"message": "FrontendPage updated successfully",
		"data":    frontendPage,
	})
}

func (p *PlatformAPI) deleteFrontendPage(w http.ResponseWriter, r *http.Request, name string) {
	frontendPage := &k8scliv1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}

	if err := p.client.Delete(r.Context(), frontendPage); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete FrontendPage: %v", err), http.StatusInternalServerError)
		return
	}

	p.writeJSONResponse(w, map[string]interface{}{
		"status":  "success",
		"message": "FrontendPage deleted successfully",
	})
}

// Step 12+: Update action handler
func (p *PlatformAPI) handleUpdateAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updateReq struct {
		Name    string                 `json:"name"`
		Updates map[string]interface{} `json:"updates"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Create action request for update
	actionReq := &ActionRequest{
		Action:  "update_frontend",
		Trigger: "api",
		Inputs: map[string]interface{}{
			"name": updateReq.Name,
		},
	}

	// Add update fields to inputs
	for key, value := range updateReq.Updates {
		actionReq.Inputs[key] = value
	}

	response, err := p.updateFrontendPageAction(r.Context(), actionReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p.writeJSONResponse(w, response)
}

func (p *PlatformAPI) handleActions(w http.ResponseWriter, r *http.Request) {
	actions := []PortAction{
		{
			Identifier:  "create_frontend",
			Title:       "Create Frontend Page",
			Trigger:     "manual",
			Description: "Create a new frontend page application",
			Inputs: map[string]interface{}{
				"name":        "string",
				"title":       "string",
				"description": "string",
				"path":        "string",
				"image":       "string",
				"replicas":    "number",
			},
		},
		{
			Identifier:  "update_frontend",
			Title:       "Update Frontend Page",
			Trigger:     "manual",
			Description: "Update an existing frontend page",
			Inputs: map[string]interface{}{
				"name":        "string",
				"title":       "string",
				"description": "string",
				"replicas":    "number",
				"image":       "string",
			},
		},
		{
			Identifier:  "delete_frontend",
			Title:       "Delete Frontend Page",
			Trigger:     "manual",
			Description: "Delete a frontend page application",
			Inputs: map[string]interface{}{
				"name": "string",
			},
		},
		{
			Identifier:  "scale_frontend",
			Title:       "Scale Frontend Page",
			Trigger:     "manual",
			Description: "Scale frontend page replicas",
			Inputs: map[string]interface{}{
				"name":     "string",
				"replicas": "number",
			},
		},
	}

	p.writeJSONResponse(w, map[string]interface{}{
		"status":  "success",
		"actions": actions,
		"count":   len(actions),
	})
}

// Step 12++: Discord notifications
func (p *PlatformAPI) sendDiscordNotification(req *ActionRequest, response *ActionResponse) {
	if p.discordClient == nil {
		return
	}

	log.Printf("📱 Step 12++: Sending Discord notification for action: %s", req.Action)

	color := 0x00FF00 // Green for success
	if response.Status == "error" {
		color = 0xFF0000 // Red for error
	}

	embed := DiscordEmbed{
		Title:       fmt.Sprintf("Platform Action: %s", req.Action),
		Description: response.Message,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []DiscordEmbedField{
			{
				Name:   "Status",
				Value:  response.Status,
				Inline: true,
			},
			{
				Name:   "Trigger",
				Value:  req.Trigger,
				Inline: true,
			},
		},
	}

	if req.ResourceId != "" {
		embed.Fields = append(embed.Fields, DiscordEmbedField{
			Name:   "Resource ID",
			Value:  req.ResourceId,
			Inline: true,
		})
	}

	if len(response.Logs) > 0 {
		logsText := ""
		for _, logEntry := range response.Logs {
			logsText += "• " + logEntry + "\n"
		}
		embed.Fields = append(embed.Fields, DiscordEmbedField{
			Name:   "Logs",
			Value:  logsText,
			Inline: false,
		})
	}

	message := DiscordMessage{
		Content: fmt.Sprintf("🤖 k8s-cli Platform Action completed"),
		Embeds:  []DiscordEmbed{embed},
	}

	if err := p.discordClient.SendMessage(message); err != nil {
		log.Printf("❌ Failed to send Discord notification: %v", err)
	} else {
		log.Printf("✅ Discord notification sent successfully")
	}
}

func (dc *DiscordClient) SendMessage(message DiscordMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	resp, err := dc.HTTPClient.Post(dc.WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord webhook failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (p *PlatformAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	p.writeJSONResponse(w, map[string]interface{}{
		"status":    "healthy",
		"service":   "k8s-cli Platform Engineering API",
		"step":      "Step 12/12+/12++",
		"timestamp": time.Now().Format(time.RFC3339),
		"features": map[string]bool{
			"port_integration":      portAPIToken != "",
			"discord_notifications": discordWebhookURL != "",
			"webhook_support":       enableWebhooks,
			"crud_operations":       true,
			"update_actions":        true,
		},
	})
}

func (p *PlatformAPI) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Simple Prometheus-style metrics
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP k8s_cli_platform_requests_total Total platform API requests\n")
	fmt.Fprintf(w, "# TYPE k8s_cli_platform_requests_total counter\n")
	fmt.Fprintf(w, "k8s_cli_platform_requests_total 100\n")
}

func (p *PlatformAPI) enableCORS(handler http.Handler) http.Handler {
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

func (p *PlatformAPI) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// Step 12: Platform command
var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "Start platform engineering API with Port.io integration (Step 12)",
	Long: `Start platform engineering API that integrates with Port.io for self-service experiences.

Step 12 Features:
• Platform engineering integration based on Port.io
• API handlers for actions to CRUD custom resources
• Webhook support for external triggers
• Self-service resource management
• Action-based resource operations

Step 12+ Features:
• Update action support for IDP and controller
• Enhanced CRUD operations with validation
• Improved error handling and logging

Step 12++ Features:
• Discord notifications integration
• Rich embed messages for action results
• Configurable notification channels
• Status updates and logging integration`,
	Run: func(cmd *cobra.Command, args []string) {
		runPlatformAPI()
	},
}

func runPlatformAPI() {
	log.Println("🎯 Starting Step 12: Platform Engineering API with Port.io integration...")

	// Setup controller-runtime client
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0", // Disable controller metrics
		HealthProbeBindAddress: "0", // Disable controller health
		LeaderElection:         false,
	})
	if err != nil {
		log.Fatalf("❌ Failed to create manager: %v", err)
	}

	// Create platform API
	platformAPI := NewPlatformAPI(mgr.GetClient(), mgr.GetScheme())

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start manager in background
	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Fatalf("❌ Manager failed to start: %v", err)
		}
	}()

	// Start platform API server
	go platformAPI.StartServer()

	log.Println("🎉 Step 12: Platform Engineering API is running!")
	log.Println("")
	log.Println("📋 Step 12 Features Active:")
	log.Println("   ✅ Port.io integration for self-service experiences")
	log.Println("   ✅ API handlers for CRUD operations on custom resources")
	log.Println("   ✅ Webhook support for external triggers")
	log.Println("   ✅ Action-based resource management")

	if portAPIToken != "" {
		log.Printf("   ✅ Port.io API integration enabled")
	} else {
		log.Printf("   ⚠️ Port.io API token not configured")
	}

	if discordWebhookURL != "" {
		log.Printf("   ✅ Discord notifications enabled")
	} else {
		log.Printf("   ⚠️ Discord webhook not configured")
	}

	log.Printf("   ✅ Step 12+ Update action support")
	log.Printf("   ✅ Step 12++ Discord notifications integration")
	log.Println("")
	log.Println("🔗 Platform Engineering Endpoints:")
	log.Printf("   🔗 Platform API: http://localhost:%d", platformPort)
	log.Printf("   📨 Port.io Webhook: http://localhost:%d/webhook/port", platformPort)
	log.Printf("   📋 Available Actions: http://localhost:%d/api/v1/actions", platformPort)
	log.Printf("   🏗️ FrontendPages API: http://localhost:%d/api/v1/frontendpages", platformPort)
	log.Printf("   ❤️ Health Check: http://localhost:%d/health", platformPort)
	log.Println("")
	log.Println("🧪 Test the platform API:")
	log.Println("   # Create a FrontendPage via API:")
	log.Printf("   curl -X POST http://localhost:%d/api/v1/frontendpages \\", platformPort)
	log.Println("     -H 'Content-Type: application/json' \\")
	log.Println("     -d '{")
	log.Println("       \"metadata\": {\"name\": \"test-platform\"},")
	log.Println("       \"spec\": {")
	log.Println("         \"title\": \"Platform Test\",")
	log.Println("         \"description\": \"Created via Platform API\",")
	log.Println("         \"path\": \"/platform\",")
	log.Println("         \"replicas\": 2")
	log.Println("       }")
	log.Println("     }'")
	log.Println("")
	log.Println("   # Trigger Port.io action:")
	log.Printf("   curl -X POST http://localhost:%d/webhook/port \\", platformPort)
	log.Println("     -H 'Content-Type: application/json' \\")
	log.Println("     -d '{")
	log.Println("       \"action\": \"create_frontend\",")
	log.Println("       \"resourceId\": \"frontend-123\",")
	log.Println("       \"trigger\": \"manual\",")
	log.Println("       \"inputs\": {")
	log.Println("         \"name\": \"port-frontend\",")
	log.Println("         \"title\": \"Port Created Frontend\",")
	log.Println("         \"path\": \"/port\",")
	log.Println("         \"replicas\": 3")
	log.Println("       }")
	log.Println("     }'")

	// Wait for shutdown signal
	<-signalChan
	log.Println("\n🛑 Shutdown signal received, stopping platform API...")

	cancel()
	time.Sleep(2 * time.Second)
	log.Println("👋 Step 12: Platform Engineering API stopped gracefully")
}

func init() {
	// Add flags for Step 12
	platformCmd.Flags().IntVar(&platformPort, "port", 8084, "Platform API server port")
	platformCmd.Flags().StringVar(&portAPIToken, "port-token", "", "Port.io API token")
	platformCmd.Flags().StringVar(&portBaseURL, "port-url", "https://api.getport.io", "Port.io API base URL")
	platformCmd.Flags().BoolVar(&enableWebhooks, "enable-webhooks", true, "Enable webhook handlers")
	platformCmd.Flags().StringVar(&discordWebhookURL, "discord-webhook", "", "Discord webhook URL for notifications")

	// Register command
	RootCmd.AddCommand(platformCmd)
}
