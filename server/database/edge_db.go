package database

import (
	"encoding/json"
	"fmt"
	"saas-server/models"
	"time"

	"github.com/google/uuid"
)

// CreateEdge creates a new edge in the database
func (db *DB) CreateEdge(req models.EdgeCreateRequest) (*models.Edge, error) {
	id := uuid.New().String()
	now := time.Now()

	// Convert JSON data to bytes for storage
	var styleDataBytes []byte
	if req.StyleData != nil {
		styleDataBytes = []byte(req.StyleData)
	} else {
		styleDataBytes = []byte("{}")
	}

	query := `
		INSERT INTO edges (id, mind_map_id, source_id, target_id, edge_type, style_data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, mind_map_id, source_id, target_id, edge_type, style_data, created_at`

	var edge models.Edge
	var styleData []byte

	err := db.QueryRow(
		query,
		id,
		req.MindMapID,
		req.SourceID,
		req.TargetID,
		req.EdgeType,
		styleDataBytes,
		now,
	).Scan(
		&edge.ID,
		&edge.MindMapID,
		&edge.SourceID,
		&edge.TargetID,
		&edge.EdgeType,
		&styleData,
		&edge.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert SQL data to model format
	edge.StyleData = json.RawMessage(styleData)

	return &edge, nil
}

// GetEdgesByMindMapID retrieves all edges for a specific mind map
func (db *DB) GetEdgesByMindMapID(mindMapID string) ([]models.Edge, error) {
	query := `
		SELECT id, mind_map_id, source_id, target_id, edge_type, style_data, created_at
		FROM edges
		WHERE mind_map_id = $1`

	rows, err := db.Query(query, mindMapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []models.Edge
	for rows.Next() {
		var edge models.Edge
		var styleData []byte

		err := rows.Scan(
			&edge.ID,
			&edge.MindMapID,
			&edge.SourceID,
			&edge.TargetID,
			&edge.EdgeType,
			&styleData,
			&edge.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert SQL data to model format
		edge.StyleData = json.RawMessage(styleData)

		edges = append(edges, edge)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return edges, nil
}

// GetEdgeByID retrieves a specific edge by its ID
func (db *DB) GetEdgeByID(id string) (*models.Edge, error) {
	query := `
		SELECT id, mind_map_id, source_id, target_id, edge_type, style_data, created_at
		FROM edges
		WHERE id = $1`

	var edge models.Edge
	var styleData []byte

	err := db.QueryRow(query, id).Scan(
		&edge.ID,
		&edge.MindMapID,
		&edge.SourceID,
		&edge.TargetID,
		&edge.EdgeType,
		&styleData,
		&edge.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Convert SQL data to model format
	edge.StyleData = json.RawMessage(styleData)

	return &edge, nil
}

// DeleteEdge deletes an edge from the database
func (db *DB) DeleteEdge(id string) error {
	query := `DELETE FROM edges WHERE id = $1`

	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("edge not found")
	}

	return nil
}

// DeleteEdgeByNodes deletes an edge between two specific nodes
func (db *DB) DeleteEdgeByNodes(sourceID, targetID string) error {
	query := `DELETE FROM edges WHERE source_id = $1 AND target_id = $2`

	result, err := db.Exec(query, sourceID, targetID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("edge not found between the specified nodes")
	}

	return nil
}
