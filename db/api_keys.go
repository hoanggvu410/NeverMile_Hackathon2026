package db

import (
	"database/sql"
	"fmt"
)

func CreateAPIKey(db *sql.DB, userID, name, keyHash string) (*APIKey, error) {
	k := &APIKey{}
	err := db.QueryRow(`
		INSERT INTO api_keys (user_id, name, key_hash)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, key_hash, COALESCE(name,''), last_used_at, revoked_at, created_at
	`, userID, name, keyHash).Scan(
		&k.ID, &k.UserID, &k.KeyHash, &k.Name,
		&k.LastUsedAt, &k.RevokedAt, &k.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateAPIKey: %w", err)
	}
	return k, nil
}

func GetAPIKeyByHash(db *sql.DB, keyHash string) (*APIKey, error) {
	k := &APIKey{}
	err := db.QueryRow(`
		SELECT id, user_id, key_hash, COALESCE(name,''), last_used_at, revoked_at, created_at
		FROM api_keys WHERE key_hash = $1
	`, keyHash).Scan(
		&k.ID, &k.UserID, &k.KeyHash, &k.Name,
		&k.LastUsedAt, &k.RevokedAt, &k.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetAPIKeyByHash: %w", err)
	}
	return k, nil
}

func ListAPIKeysByUser(db *sql.DB, userID string) ([]APIKey, error) {
	rows, err := db.Query(`
		SELECT id, user_id, key_hash, COALESCE(name,''), last_used_at, revoked_at, created_at
		FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("ListAPIKeysByUser: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.KeyHash, &k.Name,
			&k.LastUsedAt, &k.RevokedAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func RevokeAPIKey(db *sql.DB, keyID, userID string) error {
	res, err := db.Exec(`
		UPDATE api_keys SET revoked_at = NOW()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, keyID, userID)
	if err != nil {
		return fmt.Errorf("RevokeAPIKey: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("key not found or already revoked")
	}
	return nil
}

func UpdateAPIKeyLastUsed(db *sql.DB, keyID string) error {
	_, err := db.Exec(`UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, keyID)
	if err != nil {
		return fmt.Errorf("UpdateAPIKeyLastUsed: %w", err)
	}
	return nil
}

func CountActiveAPIKeys(db *sql.DB, userID string) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM api_keys WHERE user_id = $1 AND revoked_at IS NULL
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountActiveAPIKeys: %w", err)
	}
	return count, nil
}
