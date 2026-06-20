package db

import (
	"database/sql"
	"fmt"
)

func GetUserByID(db *sql.DB, id string) (*User, error) {
	u := &User{}
	err := db.QueryRow(`
		SELECT id, email, COALESCE(github_id,''), COALESCE(github_login,''),
		       plan, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.GitHubID, &u.GitHubLogin,
		&u.Plan, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetUserByID: %w", err)
	}
	return u, nil
}

func GetUserByGitHubID(db *sql.DB, githubID string) (*User, error) {
	u := &User{}
	err := db.QueryRow(`
		SELECT id, email, COALESCE(github_id,''), COALESCE(github_login,''),
		       plan, is_active, created_at, updated_at
		FROM users WHERE github_id = $1
	`, githubID).Scan(&u.ID, &u.Email, &u.GitHubID, &u.GitHubLogin,
		&u.Plan, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetUserByGitHubID: %w", err)
	}
	return u, nil
}

func UpsertUserFromGitHub(db *sql.DB, githubID, login, email string) (*User, error) {
	u := &User{}
	err := db.QueryRow(`
		INSERT INTO users (email, github_id, github_login)
		VALUES ($1, $2, $3)
		ON CONFLICT (github_id) DO UPDATE SET
			github_login = EXCLUDED.github_login,
			email        = CASE WHEN EXCLUDED.email <> '' THEN EXCLUDED.email ELSE users.email END,
			updated_at   = NOW()
		RETURNING id, email, COALESCE(github_id,''), COALESCE(github_login,''),
		          plan, is_active, created_at, updated_at
	`, email, githubID, login).Scan(&u.ID, &u.Email, &u.GitHubID, &u.GitHubLogin,
		&u.Plan, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("UpsertUserFromGitHub: %w", err)
	}
	return u, nil
}

func UpdateUserPlan(db *sql.DB, userID, plan string) error {
	_, err := db.Exec(`UPDATE users SET plan = $1, updated_at = NOW() WHERE id = $2`, plan, userID)
	if err != nil {
		return fmt.Errorf("UpdateUserPlan: %w", err)
	}
	return nil
}
