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

// NodeHandler handles node-related requests
type NodeHandler struct {
	DB *database.DB
}

// NewNodeHandler creates a new NodeHandler
func NewNodeHandler(db *database.DB) *NodeHandler {
	return &NodeHandler{DB: db}
}

// CreateNode handles POST /api/nodes
func (h *NodeHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
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
	var req models.NodeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.MindMapID == "" {
		http.Error(w, "Mind map ID is required", http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
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

	// Create node
	node, err := h.DB.CreateNode(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create node: %v", err), http.StatusInternalServerError)
		return
	}

	// Return created node
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

// GetNodesByMindMap handles GET /api/mindmaps/{id}/nodes
func (h *NodeHandler) GetNodesByMindMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mind map ID from URL
	mindMapID := strings.TrimPrefix(r.URL.Path, "/api/mindmaps/")
	mindMapID = strings.TrimSuffix(mindMapID, "/nodes")
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

	// Get nodes
	nodes, err := h.DB.GetNodesByMindMapID(mindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get nodes: %v", err), http.StatusInternalServerError)
		return
	}

	// Return nodes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

// GetNode handles GET /api/nodes/{id}
func (h *NodeHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract node ID from URL
	nodeID := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
	if nodeID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Parse node ID
	if _, err := uuid.Parse(nodeID); err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get node
	node, err := h.DB.GetNodeByID(nodeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get node: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the mind map
	mindMap, err := h.DB.GetMindMapByID(node.MindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}
	if mindMap.UserID != userID && !mindMap.IsPublic {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return node
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(node)
}

// UpdateNode handles PUT /api/nodes/{id}
func (h *NodeHandler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract node ID from URL
	nodeID := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
	if nodeID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Parse node ID
	if _, err := uuid.Parse(nodeID); err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get node
	node, err := h.DB.GetNodeByID(nodeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get node: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the mind map
	mindMap, err := h.DB.GetMindMapByID(node.MindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}
	if mindMap.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.NodeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update node
	if err := h.DB.UpdateNode(nodeID, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update node: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Node updated successfully"})
}

// DeleteNode handles DELETE /api/nodes/{id}
func (h *NodeHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract node ID from URL
	nodeID := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
	if nodeID == r.URL.Path {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Parse node ID
	if _, err := uuid.Parse(nodeID); err != nil {
		http.Error(w, "Invalid node ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get node
	node, err := h.DB.GetNodeByID(nodeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get node: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if user has access to the mind map
	mindMap, err := h.DB.GetMindMapByID(node.MindMapID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
		return
	}
	if mindMap.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete node
	if err := h.DB.DeleteNode(nodeID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete node: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Node deleted successfully"})
}

// BatchUpdateNodePositions handles POST /api/nodes/positions
func (h *NodeHandler) BatchUpdateNodePositions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
	var req models.NodeBatchPositionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Positions) == 0 {
		http.Error(w, "No positions provided", http.StatusBadRequest)
		return
	}

	// Get the first node to check mind map ownership
	if len(req.Positions) > 0 {
		firstNodeID := req.Positions[0].ID
		node, err := h.DB.GetNodeByID(firstNodeID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get node: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if user has access to the mind map
		mindMap, err := h.DB.GetMindMapByID(node.MindMapID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get mind map: %v", err), http.StatusInternalServerError)
			return
		}
		if mindMap.UserID != userID {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Update node positions
	if err := h.DB.BatchUpdateNodePositions(req.Positions); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update node positions: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Node positions updated successfully"})
}
