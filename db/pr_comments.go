package db

import (
	"database/sql"
	"fmt"
)

func RecordPRComment(db *sql.DB, contextID, repo string, prNumber int, commentURL string) error {
	_, err := db.Exec(`
		INSERT INTO pr_comments (context_id, repo, pr_number, comment_url)
		VALUES ($1, $2, $3, $4)
	`, contextID, repo, prNumber, commentURL)
	if err != nil {
		return fmt.Errorf("RecordPRComment: %w", err)
	}
	return nil
}
