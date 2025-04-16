package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"saas-server/models"
	"time"

	"github.com/google/uuid"
)

// CreateMindMap creates a new mind map in the database
func (db *DB) CreateMindMap(userID string, req models.MindMapCreateRequest) (*models.MindMap, error) {
	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO mind_maps (id, user_id, title, description, is_public, created_at, updated_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, title, description, is_public, status, created_at, updated_at`

	var mindMap models.MindMap
	err := db.QueryRow(
		query,
		id,
		userID,
		req.Title,
		req.Description,
		req.IsPublic,
		now,
		now,
		"active",
	).Scan(
		&mindMap.ID,
		&mindMap.UserID,
		&mindMap.Title,
		&mindMap.Description,
		&mindMap.IsPublic,
		&mindMap.Status,
		&mindMap.CreatedAt,
		&mindMap.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &mindMap, nil
}

// GetMindMapsByUserID retrieves all mind maps for a specific user
func (db *DB) GetMindMapsByUserID(userID string) ([]models.MindMap, error) {
	query := `
		SELECT id, user_id, title, description, is_public, status, created_at, updated_at
		FROM mind_maps
		WHERE user_id = $1 AND status != 'deleted'
		ORDER BY updated_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mindMaps []models.MindMap
	for rows.Next() {
		var mindMap models.MindMap
		err := rows.Scan(
			&mindMap.ID,
			&mindMap.UserID,
			&mindMap.Title,
			&mindMap.Description,
			&mindMap.IsPublic,
			&mindMap.Status,
			&mindMap.CreatedAt,
			&mindMap.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		mindMaps = append(mindMaps, mindMap)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return mindMaps, nil
}

// GetMindMapByID retrieves a specific mind map by its ID
func (db *DB) GetMindMapByID(id string) (*models.MindMap, error) {
	query := `
		SELECT id, user_id, title, description, is_public, status, created_at, updated_at
		FROM mind_maps
		WHERE id = $1 AND status != 'deleted'`

	var mindMap models.MindMap
	err := db.QueryRow(query, id).Scan(
		&mindMap.ID,
		&mindMap.UserID,
		&mindMap.Title,
		&mindMap.Description,
		&mindMap.IsPublic,
		&mindMap.Status,
		&mindMap.CreatedAt,
		&mindMap.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &mindMap, nil
}

// GetMindMapWithDetails retrieves a mind map with all its nodes and edges
func (db *DB) GetMindMapWithDetails(id string) (*models.MindMapWithDetails, error) {
	// First get the mind map
	mindMap, err := db.GetMindMapByID(id)
	if err != nil {
		return nil, err
	}

	// Get all nodes for this mind map
	nodesQuery := `
		SELECT id, mind_map_id, parent_id, content, position_x, position_y, 
		       node_type, style_data, metadata, created_at, updated_at
		FROM nodes
		WHERE mind_map_id = $1`

	nodeRows, err := db.Query(nodesQuery, id)
	if err != nil {
		return nil, err
	}
	defer nodeRows.Close()

	var nodes []models.Node
	for nodeRows.Next() {
		var node models.Node
		var parentID sql.NullString
		var styleData, metadata []byte

		err := nodeRows.Scan(
			&node.ID,
			&node.MindMapID,
			&parentID,
			&node.Content,
			&node.PositionX,
			&node.PositionY,
			&node.NodeType,
			&styleData,
			&metadata,
			&node.CreatedAt,
			&node.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			node.ParentID = &parentID.String
		}

		// Convert JSON data
		node.StyleData = json.RawMessage(styleData)
		node.Metadata = json.RawMessage(metadata)

		nodes = append(nodes, node)
	}

	if err = nodeRows.Err(); err != nil {
		return nil, err
	}

	// Get all edges for this mind map
	edgesQuery := `
		SELECT id, mind_map_id, source_id, target_id, edge_type, style_data, created_at
		FROM edges
		WHERE mind_map_id = $1`

	edgeRows, err := db.Query(edgesQuery, id)
	if err != nil {
		return nil, err
	}
	defer edgeRows.Close()

	var edges []models.Edge
	for edgeRows.Next() {
		var edge models.Edge
		var styleData []byte

		err := edgeRows.Scan(
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

		// Convert JSON data
		edge.StyleData = json.RawMessage(styleData)

		edges = append(edges, edge)
	}

	if err = edgeRows.Err(); err != nil {
		return nil, err
	}

	// Combine everything into the result
	result := &models.MindMapWithDetails{
		MindMap: *mindMap,
		Nodes:   nodes,
		Edges:   edges,
	}

	return result, nil
}

// UpdateMindMap updates a mind map's details
func (db *DB) UpdateMindMap(id string, req models.MindMapUpdateRequest) error {
	query := `
		UPDATE mind_maps
		SET title = COALESCE(NULLIF($2, ''), title),
		    description = COALESCE(NULLIF($3, ''), description),
		    is_public = $4,
		    status = COALESCE(NULLIF($5, ''), status),
		    updated_at = $6
		WHERE id = $1 AND status != 'deleted'`

	result, err := db.Exec(
		query,
		id,
		req.Title,
		req.Description,
		req.IsPublic,
		req.Status,
		time.Now(),
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("mind map not found or already deleted")
	}

	return nil
}

// DeleteMindMap soft deletes a mind map by setting its status to 'deleted'
func (db *DB) DeleteMindMap(id string) error {
	query := `
		UPDATE mind_maps
		SET status = 'deleted', updated_at = $2
		WHERE id = $1 AND status != 'deleted'`

	result, err := db.Exec(query, id, time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("mind map not found or already deleted")
	}

	return nil
}
