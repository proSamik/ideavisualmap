package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"saas-server/database"
	"saas-server/models"
)

// IdeaGenerationHandler handles AI-powered idea generation requests
type IdeaGenerationHandler struct {
	DB *database.DB
}

// NewIdeaGenerationHandler creates a new IdeaGenerationHandler
func NewIdeaGenerationHandler(db *database.DB) *IdeaGenerationHandler {
	return &IdeaGenerationHandler{DB: db}
}

// GenerationRequest represents a request to generate ideas
type GenerationRequest struct {
	Topic      string      `json:"topic"`      // The main topic for idea generation
	Context    string      `json:"context"`    // Additional context or constraints
	NodeID     string      `json:"node_id"`    // ID of the node to expand (optional)
	MindMapID  string      `json:"mind_map_id"` // ID of the mind map
	Count      int         `json:"count"`      // Number of ideas to generate (default: 5)
	Type       string      `json:"type"`       // Type of generation: "new", "expand", "improve", "branch"
	APIKey     string      `json:"api_key"`    // User's OpenAI API key (optional)
	UserID     interface{} `json:"-"`          // User ID (set internally, not from JSON)
}

// GenerationResponse represents the response from the idea generation
type GenerationResponse struct {
	Ideas []Idea `json:"ideas"`
}

// Idea represents a generated idea
type Idea struct {
	Content    string  `json:"content"`
	Confidence float64 `json:"confidence"`
}

// GenerateIdeas handles POST /api/generate
func (h *IdeaGenerationHandler) GenerateIdeas(w http.ResponseWriter, r *http.Request) {
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
	var req GenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.MindMapID == "" {
		http.Error(w, "Mind map ID is required", http.StatusBadRequest)
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

	// Set default count if not provided
	if req.Count <= 0 {
		req.Count = 5
	}

	// Cap the count to a reasonable number
	if req.Count > 10 {
		req.Count = 10
	}

	// Set the user ID in the request
	req.UserID = userID

	// Generate ideas using OpenAI API
	ideas, err := h.generateIdeasWithOpenAI(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate ideas: %v", err), http.StatusInternalServerError)
		return
	}

	// Return generated ideas
	response := GenerationResponse{
		Ideas: ideas,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateIdeasWithOpenAI generates ideas using the OpenAI API
func (h *IdeaGenerationHandler) generateIdeasWithOpenAI(req GenerationRequest) ([]Idea, error) {
	// Determine which API key to use
	apiKey := os.Getenv("OPENAI_API_KEY")
	
	// If the request specifies to use the user's API key
	if req.APIKey != "" {
		// Use the provided API key directly
		apiKey = req.APIKey
	} else {
		// Try to get the user's stored API key for OpenAI
		userID, ok := req.UserID.(string)
		if ok && userID != "" {
			userAPIKey, err := h.DB.GetDecryptedAPIKey(userID, "openai")
			if err == nil && userAPIKey != "" {
				apiKey = userAPIKey
			}
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	// Construct the prompt based on the request type
	var prompt string
	switch req.Type {
	case "expand":
		prompt = fmt.Sprintf("Generate %d detailed sub-ideas that expand on this concept: %s. Context: %s", 
			req.Count, req.Topic, req.Context)
	case "improve":
		prompt = fmt.Sprintf("Improve and refine this idea in %d different ways: %s. Context: %s", 
			req.Count, req.Topic, req.Context)
	case "branch":
		prompt = fmt.Sprintf("Generate %d alternative approaches or directions for this concept: %s. Context: %s", 
			req.Count, req.Topic, req.Context)
	default: // "new"
		prompt = fmt.Sprintf("Generate %d creative ideas about: %s. Context: %s", 
			req.Count, req.Topic, req.Context)
	}

	// Prepare the OpenAI API request
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a creative brainstorming assistant. Generate concise, innovative ideas for the given topic. Each idea should be clear, actionable, and directly relevant to the topic. Format your response as a JSON array of ideas.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  500,
	})
	if err != nil {
		return nil, err
	}

	// Make the API request
	client := &http.Client{}
	apiReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(apiReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	// Parse the response
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no ideas generated")
	}

	// Try to parse the response as JSON
	content := apiResp.Choices[0].Message.Content
	var rawIdeas []map[string]interface{}
	
	// First, try to parse as a JSON array directly
	err = json.Unmarshal([]byte(content), &rawIdeas)
	if err != nil {
		// If that fails, try to extract JSON from the text
		start := 0
		end := len(content)
		
		// Look for JSON array start/end
		startIdx := bytes.Index([]byte(content), []byte("["))
		endIdx := bytes.LastIndex([]byte(content), []byte("]"))
		
		if startIdx >= 0 && endIdx > startIdx {
			start = startIdx
			end = endIdx + 1
			err = json.Unmarshal([]byte(content[start:end]), &rawIdeas)
		}
		
		// If still failing, create a simple structure from the text
		if err != nil {
			// Split by newlines and create ideas
			ideas := make([]Idea, 0, req.Count)
			lines := bytes.Split([]byte(content), []byte("\n"))
			
			for _, line := range lines {
				trimmed := bytes.TrimSpace(line)
				if len(trimmed) > 0 {
					ideas = append(ideas, Idea{
						Content:    string(trimmed),
						Confidence: 0.7,
					})
				}
			}
			
			return ideas, nil
		}
	}
	
	// Convert the raw ideas to our Idea struct
	ideas := make([]Idea, 0, len(rawIdeas))
	for _, raw := range rawIdeas {
		idea := Idea{
			Content:    fmt.Sprintf("%v", raw["idea"]),
			Confidence: 0.7,
		}
		
		// Try to get the content from different possible fields
		if idea.Content == "<nil>" {
			if content, ok := raw["content"].(string); ok {
				idea.Content = content
			} else if text, ok := raw["text"].(string); ok {
				idea.Content = text
			} else if description, ok := raw["description"].(string); ok {
				idea.Content = description
			}
		}
		
		ideas = append(ideas, idea)
	}
	
	return ideas, nil
}

// CreateNodesFromIdeas handles POST /api/generate/nodes
func (h *IdeaGenerationHandler) CreateNodesFromIdeas(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		MindMapID string `json:"mind_map_id"`
		ParentID  string `json:"parent_id"`
		Ideas     []Idea `json:"ideas"`
		StartX    float64 `json:"start_x"`
		StartY    float64 `json:"start_y"`
		Layout    string `json:"layout"` // "radial", "vertical", "horizontal"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.MindMapID == "" {
		http.Error(w, "Mind map ID is required", http.StatusBadRequest)
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

	// Create nodes for each idea
	nodes := make([]models.Node, 0, len(req.Ideas))
	edges := make([]models.Edge, 0, len(req.Ideas))

	// Calculate positions based on layout
	positions := h.calculateNodePositions(req.StartX, req.StartY, len(req.Ideas), req.Layout)

	// Create nodes and edges
	for i, idea := range req.Ideas {
		// Create node
		nodeReq := models.NodeCreateRequest{
			MindMapID: req.MindMapID,
			Content:   idea.Content,
			PositionX: positions[i].X,
			PositionY: positions[i].Y,
			NodeType:  "idea",
		}

		// Set parent ID if provided
		if req.ParentID != "" {
			nodeReq.ParentID = &req.ParentID
		}

		node, err := h.DB.CreateNode(nodeReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create node: %v", err), http.StatusInternalServerError)
			return
		}

		nodes = append(nodes, *node)

		// Create edge if there's a parent
		if req.ParentID != "" {
			edgeReq := models.EdgeCreateRequest{
				MindMapID: req.MindMapID,
				SourceID:  req.ParentID,
				TargetID:  node.ID,
				EdgeType:  "idea",
			}

			edge, err := h.DB.CreateEdge(edgeReq)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create edge: %v", err), http.StatusInternalServerError)
				return
			}

			edges = append(edges, *edge)
		}
	}

	// Return created nodes and edges
	response := struct {
		Nodes []models.Node `json:"nodes"`
		Edges []models.Edge `json:"edges"`
	}{
		Nodes: nodes,
		Edges: edges,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Position represents a 2D position
type Position struct {
	X float64
	Y float64
}

// calculateNodePositions calculates positions for nodes based on the layout
func (h *IdeaGenerationHandler) calculateNodePositions(startX, startY float64, count int, layout string) []Position {
	positions := make([]Position, count)
	
	// Constants for spacing
	const (
		radialRadius = 200.0
		horizontalSpacing = 250.0
		verticalSpacing = 150.0
	)
	
	switch layout {
	case "radial":
		// Arrange nodes in a circle around the start position
		angleStep := 2 * math.Pi / float64(count)
		for i := 0; i < count; i++ {
			angle := float64(i) * angleStep
			positions[i] = Position{
				X: startX + radialRadius * math.Cos(angle),
				Y: startY + radialRadius * math.Sin(angle),
			}
		}
	case "horizontal":
		// Arrange nodes horizontally
		for i := 0; i < count; i++ {
			positions[i] = Position{
				X: startX + float64(i-count/2) * horizontalSpacing,
				Y: startY,
			}
		}
	case "vertical":
		// Arrange nodes vertically
		for i := 0; i < count; i++ {
			positions[i] = Position{
				X: startX,
				Y: startY + float64(i-count/2) * verticalSpacing,
			}
		}
	default:
		// Default to grid layout
		cols := int(math.Ceil(math.Sqrt(float64(count))))
		for i := 0; i < count; i++ {
			row := i / cols
			col := i % cols
			positions[i] = Position{
				X: startX + float64(col-cols/2) * horizontalSpacing,
				Y: startY + float64(row-count/(2*cols)) * verticalSpacing,
			}
		}
	}
	
	return positions
}
