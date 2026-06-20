package db

import (
	"database/sql"
	"fmt"
)

func GetTeamIDsByUser(db *sql.DB, userID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT team_id FROM team_members WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("GetTeamIDsByUser: %w", err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
