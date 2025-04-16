package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"saas-server/models"
	"time"

	"github.com/google/uuid"
)

// CreateNode creates a new node in the database
func (db *DB) CreateNode(req models.NodeCreateRequest) (*models.Node, error) {
	id := uuid.New().String()
	now := time.Now()

	// Convert JSON data to bytes for storage
	var styleDataBytes, metadataBytes []byte
	var err error

	if req.StyleData != nil {
		styleDataBytes = []byte(req.StyleData)
	} else {
		styleDataBytes = []byte("{}")
	}

	if req.Metadata != nil {
		metadataBytes = []byte(req.Metadata)
	} else {
		metadataBytes = []byte("{}")
	}

	query := `
		INSERT INTO nodes (id, mind_map_id, parent_id, content, position_x, position_y, 
		                  node_type, style_data, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, mind_map_id, parent_id, content, position_x, position_y, 
		         node_type, style_data, metadata, created_at, updated_at`

	var node models.Node
	var parentID sql.NullString
	var styleData, metadata []byte

	if req.ParentID != nil {
		parentID.String = *req.ParentID
		parentID.Valid = true
	}

	err = db.QueryRow(
		query,
		id,
		req.MindMapID,
		parentID,
		req.Content,
		req.PositionX,
		req.PositionY,
		req.NodeType,
		styleDataBytes,
		metadataBytes,
		now,
		now,
	).Scan(
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

	// Convert SQL data to model format
	if parentID.Valid {
		node.ParentID = &parentID.String
	}
	node.StyleData = json.RawMessage(styleData)
	node.Metadata = json.RawMessage(metadata)

	return &node, nil
}

// GetNodesByMindMapID retrieves all nodes for a specific mind map
func (db *DB) GetNodesByMindMapID(mindMapID string) ([]models.Node, error) {
	query := `
		SELECT id, mind_map_id, parent_id, content, position_x, position_y, 
		       node_type, style_data, metadata, created_at, updated_at
		FROM nodes
		WHERE mind_map_id = $1`

	rows, err := db.Query(query, mindMapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []models.Node
	for rows.Next() {
		var node models.Node
		var parentID sql.NullString
		var styleData, metadata []byte

		err := rows.Scan(
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

		// Convert SQL data to model format
		if parentID.Valid {
			node.ParentID = &parentID.String
		}
		node.StyleData = json.RawMessage(styleData)
		node.Metadata = json.RawMessage(metadata)

		nodes = append(nodes, node)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

// GetNodeByID retrieves a specific node by its ID
func (db *DB) GetNodeByID(id string) (*models.Node, error) {
	query := `
		SELECT id, mind_map_id, parent_id, content, position_x, position_y, 
		       node_type, style_data, metadata, created_at, updated_at
		FROM nodes
		WHERE id = $1`

	var node models.Node
	var parentID sql.NullString
	var styleData, metadata []byte

	err := db.QueryRow(query, id).Scan(
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

	// Convert SQL data to model format
	if parentID.Valid {
		node.ParentID = &parentID.String
	}
	node.StyleData = json.RawMessage(styleData)
	node.Metadata = json.RawMessage(metadata)

	return &node, nil
}

// UpdateNode updates a node's details
func (db *DB) UpdateNode(id string, req models.NodeUpdateRequest) error {
	// Convert JSON data to bytes for storage
	var styleDataBytes, metadataBytes []byte
	var err error

	if req.StyleData != nil {
		styleDataBytes = []byte(req.StyleData)
	}

	if req.Metadata != nil {
		metadataBytes = []byte(req.Metadata)
	}

	query := `
		UPDATE nodes
		SET content = COALESCE(NULLIF($2, ''), content),
		    position_x = COALESCE($3, position_x),
		    position_y = COALESCE($4, position_y),
		    node_type = COALESCE(NULLIF($5, ''), node_type),
		    style_data = COALESCE($6, style_data),
		    metadata = COALESCE($7, metadata),
		    updated_at = $8
		WHERE id = $1`

	// Use zero values for float64 to indicate no update
	var posX, posY *float64
	if req.PositionX != 0 {
		posX = &req.PositionX
	}
	if req.PositionY != 0 {
		posY = &req.PositionY
	}

	result, err := db.Exec(
		query,
		id,
		req.Content,
		posX,
		posY,
		req.NodeType,
		styleDataBytes,
		metadataBytes,
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
		return fmt.Errorf("node not found")
	}

	return nil
}

// DeleteNode deletes a node from the database
func (db *DB) DeleteNode(id string) error {
	query := `DELETE FROM nodes WHERE id = $1`

	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("node not found")
	}

	return nil
}

// BatchUpdateNodePositions updates the positions of multiple nodes in a single transaction
func (db *DB) BatchUpdateNodePositions(positions []models.NodePositionUpdateRequest) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	query := `
		UPDATE nodes
		SET position_x = $2,
		    position_y = $3,
		    updated_at = $4
		WHERE id = $1`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, pos := range positions {
		_, err = stmt.Exec(pos.ID, pos.PositionX, pos.PositionY, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
