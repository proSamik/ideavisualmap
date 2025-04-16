package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"saas-server/models"
)

// CreateAPIKey creates a new API key for a user
func (db *DB) CreateAPIKey(userID string, req models.APIKeyCreateRequest) (*models.APIKeyResponse, error) {
	// Encrypt the API key
	encryptedKey, err := encryptAPIKey(req.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API key: %v", err)
	}

	// Check if the user already has an API key for this service
	var existingID string
	err = db.QueryRow(
		"SELECT id FROM api_keys WHERE user_id = $1 AND service = $2",
		userID, req.Service,
	).Scan(&existingID)

	if err == nil {
		// Update the existing API key
		_, err = db.Exec(
			"UPDATE api_keys SET encrypted_key = $1, is_active = true, updated_at = NOW() WHERE id = $2",
			encryptedKey, existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update API key: %v", err)
		}

		// Get the updated API key
		return db.GetAPIKeyByID(existingID)
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing API key: %v", err)
	}

	// Insert a new API key
	var id string
	err = db.QueryRow(
		`INSERT INTO api_keys (user_id, service, encrypted_key, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		RETURNING id`,
		userID, req.Service, encryptedKey,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %v", err)
	}

	// Get the created API key
	return db.GetAPIKeyByID(id)
}

// GetAPIKeyByID gets an API key by ID
func (db *DB) GetAPIKeyByID(id string) (*models.APIKeyResponse, error) {
	var apiKey models.APIKeyResponse
	err := db.QueryRow(
		`SELECT id, user_id, service, is_active, created_at, updated_at
		FROM api_keys
		WHERE id = $1`,
		id,
	).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Service,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %v", err)
	}

	return &apiKey, nil
}

// GetAPIKeyByUserAndService gets an API key by user ID and service
func (db *DB) GetAPIKeyByUserAndService(userID, service string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := db.QueryRow(
		`SELECT id, user_id, service, encrypted_key, is_active, created_at, updated_at
		FROM api_keys
		WHERE user_id = $1 AND service = $2`,
		userID, service,
	).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Service,
		&apiKey.EncryptedKey,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %v", err)
	}

	return &apiKey, nil
}

// GetAPIKeysByUserID gets all API keys for a user
func (db *DB) GetAPIKeysByUserID(userID string) ([]models.APIKeyResponse, error) {
	rows, err := db.Query(
		`SELECT id, user_id, service, is_active, created_at, updated_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %v", err)
	}
	defer rows.Close()

	var apiKeys []models.APIKeyResponse
	for rows.Next() {
		var apiKey models.APIKeyResponse
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Service,
			&apiKey.IsActive,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %v", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API keys: %v", err)
	}

	return apiKeys, nil
}

// UpdateAPIKey updates an API key
func (db *DB) UpdateAPIKey(id string, req models.APIKeyUpdateRequest) (*models.APIKeyResponse, error) {
	// Check if the API key exists
	_, err := db.GetAPIKeyByID(id)
	if err != nil {
		return nil, err
	}

	// Update the API key
	if req.Key != "" {
		// Encrypt the new API key
		encryptedKey, err := encryptAPIKey(req.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %v", err)
		}

		_, err = db.Exec(
			"UPDATE api_keys SET encrypted_key = $1, updated_at = NOW() WHERE id = $2",
			encryptedKey, id,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update API key: %v", err)
		}
	}

	// Update the is_active status
	_, err = db.Exec(
		"UPDATE api_keys SET is_active = $1, updated_at = NOW() WHERE id = $2",
		req.IsActive, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update API key status: %v", err)
	}

	// Get the updated API key
	return db.GetAPIKeyByID(id)
}

// DeleteAPIKey deletes an API key
func (db *DB) DeleteAPIKey(id string) error {
	_, err := db.Exec("DELETE FROM api_keys WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %v", err)
	}
	return nil
}

// GetDecryptedAPIKey gets a decrypted API key by user ID and service
func (db *DB) GetDecryptedAPIKey(userID, service string) (string, error) {
	apiKey, err := db.GetAPIKeyByUserAndService(userID, service)
	if err != nil {
		return "", err
	}

	if !apiKey.IsActive {
		return "", fmt.Errorf("API key is not active")
	}

	// Decrypt the API key
	decryptedKey, err := decryptAPIKey(apiKey.EncryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt API key: %v", err)
	}

	return decryptedKey, nil
}

// encryptAPIKey encrypts an API key using AES-256-GCM
func encryptAPIKey(plaintext string) (string, error) {
	// Get the encryption key from environment variable
	key := []byte(os.Getenv("API_KEY_ENCRYPTION_KEY"))
	if len(key) < 32 {
		// Pad the key to 32 bytes if it's too short
		paddedKey := make([]byte, 32)
		copy(paddedKey, key)
		key = paddedKey
	} else if len(key) > 32 {
		// Truncate the key to 32 bytes if it's too long
		key = key[:32]
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode the ciphertext as base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptAPIKey decrypts an API key using AES-256-GCM
func decryptAPIKey(ciphertext string) (string, error) {
	// Get the encryption key from environment variable
	key := []byte(os.Getenv("API_KEY_ENCRYPTION_KEY"))
	if len(key) < 32 {
		// Pad the key to 32 bytes if it's too short
		paddedKey := make([]byte, 32)
		copy(paddedKey, key)
		key = paddedKey
	} else if len(key) > 32 {
		// Truncate the key to 32 bytes if it's too long
		key = key[:32]
	}

	// Decode the ciphertext from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Check if the ciphertext is valid
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	// Extract the nonce and ciphertext
	nonce, ciphertextBytes := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	// Decrypt the ciphertext
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
