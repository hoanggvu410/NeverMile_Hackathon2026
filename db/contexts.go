package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func GetExistingLocalIDs(db *sql.DB, userID string, localIDs []string) ([]string, error) {
	if len(localIDs) == 0 {
		return []string{}, nil
	}
	rows, err := db.Query(`
		SELECT local_id FROM contexts WHERE user_id = $1 AND local_id = ANY($2)
	`, userID, pq.Array(localIDs))
	if err != nil {
		return nil, fmt.Errorf("GetExistingLocalIDs: %w", err)
	}
	defer rows.Close()

	existing := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		existing = append(existing, id)
	}
	return existing, rows.Err()
}

func SyncContexts(db *sql.DB, userID string, ctxs []ContextRow) ([]string, error) {
	var synced []string
	for _, c := range ctxs {
		filesJSON, err := json.Marshal(c.Files)
		if err != nil {
			continue
		}
		commitsJSON, err := json.Marshal(c.Commits)
		if err != nil {
			continue
		}

		contextTS := c.ContextTS
		if contextTS.IsZero() {
			contextTS = time.Now().UTC()
		}

		_, err = db.Exec(`
			INSERT INTO contexts (
				local_id, user_id, team_id, prompt, reasoning, decisions,
				rejected_alternatives, trade_offs, files, commits,
				domain, topic, agent, model, repo_name, context_ts
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			ON CONFLICT (local_id) DO UPDATE SET
				reasoning             = EXCLUDED.reasoning,
				decisions             = EXCLUDED.decisions,
				rejected_alternatives = EXCLUDED.rejected_alternatives,
				trade_offs            = EXCLUDED.trade_offs,
				files                 = EXCLUDED.files,
				commits               = EXCLUDED.commits,
				domain                = EXCLUDED.domain,
				topic                 = EXCLUDED.topic,
				updated_at            = NOW()
			WHERE contexts.user_id = EXCLUDED.user_id
		`, c.LocalID, userID, c.TeamID, c.Prompt, c.Reasoning, c.Decisions,
			c.RejectedAlternatives, c.TradeOffs,
			string(filesJSON), string(commitsJSON),
			c.Domain, c.Topic, c.Agent, c.Model, c.RepoName, contextTS)
		if err != nil {
			continue
		}
		synced = append(synced, c.LocalID)
	}
	if synced == nil {
		synced = []string{}
	}
	return synced, nil
}

func GetContextByLocalID(db *sql.DB, localID string) (*ContextRow, error) {
	c := &ContextRow{}
	var filesJSON, commitsJSON []byte
	var teamID sql.NullString

	err := db.QueryRow(`
		SELECT id, local_id, user_id, team_id, prompt,
		       COALESCE(reasoning,''), COALESCE(decisions,''),
		       COALESCE(rejected_alternatives,''), COALESCE(trade_offs,''),
		       files, commits,
		       COALESCE(domain,''), COALESCE(topic,''),
		       COALESCE(agent,''), COALESCE(model,''),
		       is_published, COALESCE(repo_name,''),
		       context_ts, created_at, updated_at
		FROM contexts WHERE local_id = $1
	`, localID).Scan(
		&c.ID, &c.LocalID, &c.UserID, &teamID, &c.Prompt,
		&c.Reasoning, &c.Decisions, &c.RejectedAlternatives, &c.TradeOffs,
		&filesJSON, &commitsJSON,
		&c.Domain, &c.Topic, &c.Agent, &c.Model,
		&c.IsPublished, &c.RepoName,
		&c.ContextTS, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetContextByLocalID: %w", err)
	}
	if teamID.Valid {
		c.TeamID = &teamID.String
	}
	json.Unmarshal(filesJSON, &c.Files)
	json.Unmarshal(commitsJSON, &c.Commits)
	if c.Files == nil {
		c.Files = []string{}
	}
	if c.Commits == nil {
		c.Commits = []string{}
	}
	return c, nil
}

func SearchContexts(db *sql.DB, userID string, teamIDs []string, query string, limit int) ([]ContextRow, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	q := "%" + query + "%"

	var (
		rows *sql.Rows
		err  error
	)
	if len(teamIDs) == 0 {
		rows, err = db.Query(`
			SELECT id, local_id, user_id, team_id, prompt,
			       COALESCE(reasoning,''), COALESCE(decisions,''),
			       COALESCE(rejected_alternatives,''), COALESCE(trade_offs,''),
			       files, commits,
			       COALESCE(domain,''), COALESCE(topic,''),
			       COALESCE(agent,''), COALESCE(model,''),
			       is_published, COALESCE(repo_name,''),
			       context_ts, created_at, updated_at
			FROM contexts
			WHERE user_id = $1
			  AND (prompt ILIKE $2 OR reasoning ILIKE $2 OR decisions ILIKE $2 OR topic ILIKE $2)
			ORDER BY context_ts DESC
			LIMIT $3
		`, userID, q, limit)
	} else {
		rows, err = db.Query(`
			SELECT id, local_id, user_id, team_id, prompt,
			       COALESCE(reasoning,''), COALESCE(decisions,''),
			       COALESCE(rejected_alternatives,''), COALESCE(trade_offs,''),
			       files, commits,
			       COALESCE(domain,''), COALESCE(topic,''),
			       COALESCE(agent,''), COALESCE(model,''),
			       is_published, COALESCE(repo_name,''),
			       context_ts, created_at, updated_at
			FROM contexts
			WHERE (user_id = $1 OR (is_published = TRUE AND team_id = ANY($2::uuid[])))
			  AND (prompt ILIKE $3 OR reasoning ILIKE $3 OR decisions ILIKE $3 OR topic ILIKE $3)
			ORDER BY context_ts DESC
			LIMIT $4
		`, userID, pq.Array(teamIDs), q, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("SearchContexts: %w", err)
	}
	defer rows.Close()
	return scanContextRows(rows)
}

func PublishContexts(db *sql.DB, userID string, localIDs []string) error {
	if len(localIDs) == 0 {
		return nil
	}
	_, err := db.Exec(`
		UPDATE contexts SET is_published = TRUE, updated_at = NOW()
		WHERE user_id = $1 AND local_id = ANY($2)
	`, userID, pq.Array(localIDs))
	if err != nil {
		return fmt.Errorf("PublishContexts: %w", err)
	}
	return nil
}

func GetPublishedContextsByTeam(db *sql.DB, teamID string) ([]ContextRow, error) {
	rows, err := db.Query(`
		SELECT id, local_id, user_id, team_id, prompt,
		       COALESCE(reasoning,''), COALESCE(decisions,''),
		       COALESCE(rejected_alternatives,''), COALESCE(trade_offs,''),
		       files, commits,
		       COALESCE(domain,''), COALESCE(topic,''),
		       COALESCE(agent,''), COALESCE(model,''),
		       is_published, COALESCE(repo_name,''),
		       context_ts, created_at, updated_at
		FROM contexts
		WHERE team_id = $1 AND is_published = TRUE
		ORDER BY context_ts DESC
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("GetPublishedContextsByTeam: %w", err)
	}
	defer rows.Close()
	return scanContextRows(rows)
}

func scanContextRows(rows *sql.Rows) ([]ContextRow, error) {
	var out []ContextRow
	for rows.Next() {
		var c ContextRow
		var filesJSON, commitsJSON []byte
		var teamID sql.NullString

		if err := rows.Scan(
			&c.ID, &c.LocalID, &c.UserID, &teamID, &c.Prompt,
			&c.Reasoning, &c.Decisions, &c.RejectedAlternatives, &c.TradeOffs,
			&filesJSON, &commitsJSON,
			&c.Domain, &c.Topic, &c.Agent, &c.Model,
			&c.IsPublished, &c.RepoName,
			&c.ContextTS, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if teamID.Valid {
			c.TeamID = &teamID.String
		}
		json.Unmarshal(filesJSON, &c.Files)
		json.Unmarshal(commitsJSON, &c.Commits)
		if c.Files == nil {
			c.Files = []string{}
		}
		if c.Commits == nil {
			c.Commits = []string{}
		}
		out = append(out, c)
	}
	if out == nil {
		out = []ContextRow{}
	}
	return out, rows.Err()
}
