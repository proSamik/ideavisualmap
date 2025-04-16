// Package models contains the data models for the application
package models

import (
	"time"
)

// MindMap represents a mind map created by a user
type MindMap struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MindMapWithDetails includes the mind map with its nodes and edges
type MindMapWithDetails struct {
	MindMap
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// MindMapCreateRequest represents the data needed to create a new mind map
type MindMapCreateRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// MindMapUpdateRequest represents the data that can be updated for a mind map
type MindMapUpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	Status      string `json:"status"`
}
