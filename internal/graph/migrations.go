package graph

import (
	"database/sql"
	"fmt"
	"strconv"
)

func ensureGraphMigrations(db *sql.DB) error {
	if err := ensureTableColumn(db, "claim_vectors", "provider", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureTableColumn(db, "claim_vectors", "dims", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if _, err := db.Exec(`UPDATE claim_vectors SET dims = length(embedding) / 4 WHERE dims = 0`); err != nil {
		return fmt.Errorf("backfill claim vector dims: %w", err)
	}
	if _, err := db.Exec(
		`UPDATE claim_vectors SET provider = ? WHERE provider = '' AND dims = ?`,
		localEmbeddingProvider, localEmbeddingDims,
	); err != nil {
		return fmt.Errorf("backfill local claim vector provider: %w", err)
	}
	if _, err := db.Exec(
		`UPDATE claim_vectors SET provider = ? WHERE provider = '' AND dims = ?`,
		openAIEmbeddingProvider, openAIEmbeddingDims,
	); err != nil {
		return fmt.Errorf("backfill OpenAI claim vector provider: %w", err)
	}
	if _, err := db.Exec(`UPDATE claim_vectors SET provider = 'unknown' WHERE provider = ''`); err != nil {
		return fmt.Errorf("backfill unknown claim vector provider: %w", err)
	}
	return nil
}

func ensureTableColumn(db *sql.DB, tableName, columnName, columnSQL string) error {
	rows, err := db.Query(`PRAGMA table_info(` + tableName + `)`)
	if err != nil {
		return fmt.Errorf("inspect %s schema: %w", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("scan %s schema: %w", tableName, err)
		}
		if name == columnName {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if _, err := db.Exec(`ALTER TABLE ` + tableName + ` ADD COLUMN ` + columnName + ` ` + columnSQL); err != nil {
		return fmt.Errorf("add %s.%s: %w", tableName, columnName, err)
	}
	return nil
}

func resolveGraphEmbeddingSpec(db *sql.DB) (embeddingSpec, error) {
	desired := desiredEmbeddingSpec()
	if desired.Dims == 0 {
		return embeddingSpec{}, fmt.Errorf("unsupported embedding provider %q", desired.Provider)
	}

	stored, ok, err := readEmbeddingConfig(db)
	if err != nil {
		return embeddingSpec{}, err
	}
	if !ok {
		if inferred, hasInferred, err := inferEmbeddingSpecFromVectors(db); err != nil {
			return embeddingSpec{}, err
		} else if hasInferred {
			stored = inferred
		} else {
			stored = desired
		}
		if err := writeEmbeddingConfig(db, stored); err != nil {
			return embeddingSpec{}, err
		}
	}

	if stored.Provider == "unknown" || stored.Dims == 0 {
		return embeddingSpec{}, fmt.Errorf("graph has legacy vectors with unknown embedding provider; reindex required")
	}
	if desired.Explicit && stored.Provider != desired.Provider {
		return embeddingSpec{}, fmt.Errorf("embedding provider mismatch: graph uses %s, requested %s; reindex required", stored.Provider, desired.Provider)
	}
	return stored, nil
}

func readEmbeddingConfig(db *sql.DB) (embeddingSpec, bool, error) {
	rows, err := db.Query(`SELECT key, value FROM embedding_config WHERE key IN ('provider', 'dims')`)
	if err != nil {
		return embeddingSpec{}, false, fmt.Errorf("read embedding config: %w", err)
	}
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return embeddingSpec{}, false, fmt.Errorf("scan embedding config: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return embeddingSpec{}, false, err
	}
	if values["provider"] == "" || values["dims"] == "" {
		return embeddingSpec{}, false, nil
	}
	dims, err := strconv.Atoi(values["dims"])
	if err != nil {
		return embeddingSpec{}, false, fmt.Errorf("invalid embedding dims %q: %w", values["dims"], err)
	}
	return embeddingSpec{Provider: values["provider"], Dims: dims}, true, nil
}

func writeEmbeddingConfig(db *sql.DB, spec embeddingSpec) error {
	if _, err := db.Exec(`
		INSERT INTO embedding_config (key, value) VALUES ('provider', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, spec.Provider); err != nil {
		return fmt.Errorf("write embedding provider config: %w", err)
	}
	if _, err := db.Exec(`
		INSERT INTO embedding_config (key, value) VALUES ('dims', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, strconv.Itoa(spec.Dims)); err != nil {
		return fmt.Errorf("write embedding dims config: %w", err)
	}
	return nil
}

func inferEmbeddingSpecFromVectors(db *sql.DB) (embeddingSpec, bool, error) {
	var provider string
	var dims int
	err := db.QueryRow(`
		SELECT provider, dims
		FROM claim_vectors
		WHERE provider <> '' AND dims > 0
		ORDER BY id
		LIMIT 1`).Scan(&provider, &dims)
	if err == sql.ErrNoRows {
		return embeddingSpec{}, false, nil
	}
	if err != nil {
		return embeddingSpec{}, false, fmt.Errorf("infer embedding provider: %w", err)
	}
	return embeddingSpec{Provider: provider, Dims: dims}, true, nil
}
