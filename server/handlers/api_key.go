package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"saas-server/database"
	"saas-server/models"
	"strings"
)

// APIKeyHandler handles API key-related requests
type APIKeyHandler struct {
	DB *database.DB
}

// NewAPIKeyHandler creates a new APIKeyHandler
func NewAPIKeyHandler(db *database.DB) *APIKeyHandler {
	return &APIKeyHandler{DB: db}
}

// CreateAPIKey handles POST /api/apikeys
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.APIKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Service == "" {
		http.Error(w, "Service is required", http.StatusBadRequest)
		return
	}
	if req.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	// Create API key
	apiKey, err := h.DB.CreateAPIKey(userID, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Return created API key
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiKey)
}

// GetAPIKeys handles GET /api/apikeys
func (h *APIKeyHandler) GetAPIKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get API keys
	apiKeys, err := h.DB.GetAPIKeysByUserID(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get API keys: %v", err), http.StatusInternalServerError)
		return
	}

	// Return API keys
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiKeys)
}

// GetAPIKey handles GET /api/apikeys/{id}
func (h *APIKeyHandler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract API key ID from URL
	apiKeyID := strings.TrimPrefix(r.URL.Path, "/api/apikeys/")
	if apiKeyID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get API key
	apiKey, err := h.DB.GetAPIKeyByID(apiKeyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the API key
	if apiKey.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return API key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiKey)
}

// UpdateAPIKey handles PUT /api/apikeys/{id}
func (h *APIKeyHandler) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract API key ID from URL
	apiKeyID := strings.TrimPrefix(r.URL.Path, "/api/apikeys/")
	if apiKeyID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get API key to check ownership
	apiKey, err := h.DB.GetAPIKeyByID(apiKeyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the API key
	if apiKey.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.APIKeyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update API key
	updatedAPIKey, err := h.DB.UpdateAPIKey(apiKeyID, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Return updated API key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedAPIKey)
}

// DeleteAPIKey handles DELETE /api/apikeys/{id}
func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract API key ID from URL
	apiKeyID := strings.TrimPrefix(r.URL.Path, "/api/apikeys/")
	if apiKeyID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get API key to check ownership
	apiKey, err := h.DB.GetAPIKeyByID(apiKeyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the API key
	if apiKey.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete API key
	if err := h.DB.DeleteAPIKey(apiKeyID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "API key deleted successfully"})
}

// GetAPIKeyByService handles GET /api/apikeys/service/{service}
func (h *APIKeyHandler) GetAPIKeyByService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract service from URL
	service := strings.TrimPrefix(r.URL.Path, "/api/apikeys/service/")
	if service == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get API key
	apiKey, err := h.DB.GetAPIKeyByUserAndService(userID, service)
	if err != nil {
		// If the API key doesn't exist, return an empty response
		if strings.Contains(err.Error(), "not found") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{}"))
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Return API key (without the encrypted key)
	response := models.APIKeyResponse{
		ID:        apiKey.ID,
		UserID:    apiKey.UserID,
		Service:   apiKey.Service,
		IsActive:  apiKey.IsActive,
		CreatedAt: apiKey.CreatedAt,
		UpdatedAt: apiKey.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
