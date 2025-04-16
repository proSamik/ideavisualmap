package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"saas-server/database"
	"saas-server/models"
	"strings"

	"github.com/google/uuid"
)

// MindMapHandler handles mind map-related requests
type MindMapHandler struct {
	DB *database.DB
}

// NewMindMapHandler creates a new MindMapHandler
func NewMindMapHandler(db *database.DB) *MindMapHandler {
	return &MindMapHandler{DB: db}
}

// CreateMindMap handles POST /api/mindmaps
func (h *MindMapHandler) CreateMindMap(w http.ResponseWriter, r *http.Request) {
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
	var req models.MindMapCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Create mind map
	mindMap, err := h.DB.CreateMindMap(userID, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create mind map: %v", err), http.StatusInternalServerError)
		return
	}

	// Return created mind map
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mindMap)
}

// GetMindMaps handles GET /api/mindmaps
func (h *MindMapHandler) GetMindMaps(w http.ResponseWriter, r *http.Request) {
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

	// Get mind maps
	mindMaps, err := h.DB.GetMindMapsByUserID(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind maps: %v", err), http.StatusInternalServerError)
		return
	}

	// Return mind maps
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mindMaps)
}

// GetMindMap handles GET /api/mindmaps/{id}
func (h *MindMapHandler) GetMindMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mind map ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/mindmaps/")
	if path == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Check if we need to get details
	isDetails := false
	if strings.HasSuffix(path, "/details") {
		isDetails = true
		path = strings.TrimSuffix(path, "/details")
	}

	// Parse mind map ID
	mindMapID := path
	if _, err := uuid.Parse(mindMapID); err != nil {
		http.Error(w, "Invalid mind map ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if isDetails {
		// Get mind map with details
		mindMapWithDetails, err := h.DB.GetMindMapWithDetails(mindMapID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if user has access
		if mindMapWithDetails.UserID != userID && !mindMapWithDetails.IsPublic {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Return mind map with details
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mindMapWithDetails)
		return
	}

	// Get mind map
	mindMap, err := h.DB.GetMindMapByID(mindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if mindMap.UserID != userID && !mindMap.IsPublic {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return mind map
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mindMap)
}

// UpdateMindMap handles PUT /api/mindmaps/{id}
func (h *MindMapHandler) UpdateMindMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mind map ID from URL
	mindMapID := strings.TrimPrefix(r.URL.Path, "/api/mindmaps/")
	if mindMapID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Parse mind map ID
	if _, err := uuid.Parse(mindMapID); err != nil {
		http.Error(w, "Invalid mind map ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get mind map to check ownership
	mindMap, err := h.DB.GetMindMapByID(mindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if mindMap.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.MindMapUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update mind map
	if err := h.DB.UpdateMindMap(mindMapID, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update mind map: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Mind map updated successfully"})
}

// DeleteMindMap handles DELETE /api/mindmaps/{id}
func (h *MindMapHandler) DeleteMindMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mind map ID from URL
	mindMapID := strings.TrimPrefix(r.URL.Path, "/api/mindmaps/")
	if mindMapID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Parse mind map ID
	if _, err := uuid.Parse(mindMapID); err != nil {
		http.Error(w, "Invalid mind map ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get mind map to check ownership
	mindMap, err := h.DB.GetMindMapByID(mindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access
	if mindMap.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete mind map
	if err := h.DB.DeleteMindMap(mindMapID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete mind map: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Mind map deleted successfully"})
}
