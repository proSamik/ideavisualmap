// Package models contains the data models for the application
package models

import (
	"time"
)

// APIKey represents a user's API key for a specific service
type APIKey struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Service      string    `json:"service"`
	EncryptedKey string    `json:"-"` // Not exposed in JSON
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// APIKeyCreateRequest represents the data needed to create a new API key
type APIKeyCreateRequest struct {
	Service string `json:"service" binding:"required"`
	Key     string `json:"key" binding:"required"`
}

// APIKeyUpdateRequest represents the data that can be updated for an API key
type APIKeyUpdateRequest struct {
	Key      string `json:"key"`
	IsActive bool   `json:"is_active"`
}

// APIKeyResponse represents the data returned to the client
type APIKeyResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Service   string    `json:"service"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
