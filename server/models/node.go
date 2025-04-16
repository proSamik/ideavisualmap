// Package models contains the data models for the application
package models

import (
	"encoding/json"
	"time"
)

// Node represents a node in a mind map
type Node struct {
	ID         string          `json:"id"`
	MindMapID  string          `json:"mind_map_id"`
	ParentID   *string         `json:"parent_id"`
	Content    string          `json:"content"`
	PositionX  float64         `json:"position_x"`
	PositionY  float64         `json:"position_y"`
	NodeType   string          `json:"node_type"`
	StyleData  json.RawMessage `json:"style_data"`
	Metadata   json.RawMessage `json:"metadata"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// NodeCreateRequest represents the data needed to create a new node
type NodeCreateRequest struct {
	MindMapID  string          `json:"mind_map_id" binding:"required"`
	ParentID   *string         `json:"parent_id"`
	Content    string          `json:"content" binding:"required"`
	PositionX  float64         `json:"position_x" binding:"required"`
	PositionY  float64         `json:"position_y" binding:"required"`
	NodeType   string          `json:"node_type"`
	StyleData  json.RawMessage `json:"style_data"`
	Metadata   json.RawMessage `json:"metadata"`
}

// NodeUpdateRequest represents the data that can be updated for a node
type NodeUpdateRequest struct {
	Content    string          `json:"content"`
	PositionX  float64         `json:"position_x"`
	PositionY  float64         `json:"position_y"`
	NodeType   string          `json:"node_type"`
	StyleData  json.RawMessage `json:"style_data"`
	Metadata   json.RawMessage `json:"metadata"`
}

// NodePositionUpdateRequest represents the data needed to update a node's position
type NodePositionUpdateRequest struct {
	ID        string  `json:"id" binding:"required"`
	PositionX float64 `json:"position_x" binding:"required"`
	PositionY float64 `json:"position_y" binding:"required"`
}

// NodeBatchPositionUpdateRequest represents a batch of node position updates
type NodeBatchPositionUpdateRequest struct {
	Positions []NodePositionUpdateRequest `json:"positions" binding:"required"`
}
