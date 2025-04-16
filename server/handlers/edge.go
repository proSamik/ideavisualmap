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

// EdgeHandler handles edge-related requests
type EdgeHandler struct {
	DB *database.DB
}

// NewEdgeHandler creates a new EdgeHandler
func NewEdgeHandler(db *database.DB) *EdgeHandler {
	return &EdgeHandler{DB: db}
}

// CreateEdge handles POST /api/edges
func (h *EdgeHandler) CreateEdge(w http.ResponseWriter, r *http.Request) {
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
	var req models.EdgeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.MindMapID == "" {
		http.Error(w, "Mind map ID is required", http.StatusBadRequest)
		return
	}
	if req.SourceID == "" {
		http.Error(w, "Source node ID is required", http.StatusBadRequest)
		return
	}
	if req.TargetID == "" {
		http.Error(w, "Target node ID is required", http.StatusBadRequest)
		return
	}

	// Check if user has access to the mind map
	mindMap, err := h.DB.GetMindMapByID(req.MindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}
	if mindMap.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create edge
	edge, err := h.DB.CreateEdge(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create edge: %v", err), http.StatusInternalServerError)
		return
	}

	// Return created edge
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(edge)
}

// GetEdgesByMindMap handles GET /api/mindmaps/{id}/edges
func (h *EdgeHandler) GetEdgesByMindMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mind map ID from URL
	mindMapID := strings.TrimPrefix(r.URL.Path, "/api/mindmaps/")
	mindMapID = strings.TrimSuffix(mindMapID, "/edges")
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

	// Check if user has access to the mind map
	mindMap, err := h.DB.GetMindMapByID(mindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}
	if mindMap.UserID != userID && !mindMap.IsPublic {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get edges
	edges, err := h.DB.GetEdgesByMindMapID(mindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get edges: %v", err), http.StatusInternalServerError)
		return
	}

	// Return edges
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(edges)
}

// GetEdge handles GET /api/edges/{id}
func (h *EdgeHandler) GetEdge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract edge ID from URL
	edgeID := strings.TrimPrefix(r.URL.Path, "/api/edges/")
	if edgeID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Parse edge ID
	if _, err := uuid.Parse(edgeID); err != nil {
		http.Error(w, "Invalid edge ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get edge
	edge, err := h.DB.GetEdgeByID(edgeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get edge: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the mind map
	mindMap, err := h.DB.GetMindMapByID(edge.MindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}
	if mindMap.UserID != userID && !mindMap.IsPublic {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return edge
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(edge)
}

// DeleteEdge handles DELETE /api/edges/{id}
func (h *EdgeHandler) DeleteEdge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract edge ID from URL
	edgeID := strings.TrimPrefix(r.URL.Path, "/api/edges/")
	if edgeID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Parse edge ID
	if _, err := uuid.Parse(edgeID); err != nil {
		http.Error(w, "Invalid edge ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get edge
	edge, err := h.DB.GetEdgeByID(edgeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get edge: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the mind map
	mindMap, err := h.DB.GetMindMapByID(edge.MindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}
	if mindMap.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete edge
	if err := h.DB.DeleteEdge(edgeID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete edge: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Edge deleted successfully"})
}

// DeleteEdgeByNodes handles DELETE /api/edges/nodes
func (h *EdgeHandler) DeleteEdgeByNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.SourceID == "" {
		http.Error(w, "Source node ID is required", http.StatusBadRequest)
		return
	}
	if req.TargetID == "" {
		http.Error(w, "Target node ID is required", http.StatusBadRequest)
		return
	}

	// Get source node to check mind map ownership
	sourceNode, err := h.DB.GetNodeByID(req.SourceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get source node: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the mind map
	mindMap, err := h.DB.GetMindMapByID(sourceNode.MindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}
	if mindMap.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete edge
	if err := h.DB.DeleteEdgeByNodes(req.SourceID, req.TargetID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete edge: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Edge deleted successfully"})
}
