// Package models contains the data models for the application
package models

import (
	"encoding/json"
	"time"
)

// Edge represents a connection between two nodes in a mind map
type Edge struct {
	ID        string          `json:"id"`
	MindMapID string          `json:"mind_map_id"`
	SourceID  string          `json:"source_id"`
	TargetID  string          `json:"target_id"`
	EdgeType  string          `json:"edge_type"`
	StyleData json.RawMessage `json:"style_data"`
	CreatedAt time.Time       `json:"created_at"`
}

// EdgeCreateRequest represents the data needed to create a new edge
type EdgeCreateRequest struct {
	MindMapID string          `json:"mind_map_id" binding:"required"`
	SourceID  string          `json:"source_id" binding:"required"`
	TargetID  string          `json:"target_id" binding:"required"`
	EdgeType  string          `json:"edge_type"`
	StyleData json.RawMessage `json:"style_data"`
}

// EdgeBatchCreateRequest represents a batch of edge creation requests
type EdgeBatchCreateRequest struct {
	Edges []EdgeCreateRequest `json:"edges" binding:"required"`
}
