package cache

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const (
	cacheTTL     = 24 * time.Hour
	maxCacheSize = 1000
)

// Cache is a SQLite-backed semantic query cache.
// It deduplicates search queries by cosine similarity on embeddings.
type Cache struct {
	db *sql.DB
}

// NewCache opens (or creates) the semantic.db at dbPath.
func NewCache(dbPath string) (*Cache, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open cache db: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS query_cache (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			namespace TEXT    NOT NULL DEFAULT 'search',
			embedding BLOB    NOT NULL,
			results   TEXT    NOT NULL,
			cached_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_query_cache_namespace_cached
			ON query_cache(namespace, cached_at);
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create cache table: %w", err)
	}
	if err := ensureNamespaceColumn(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Cache{db: db}, nil
}

// Close releases the database connection.
func (c *Cache) Close() error { return c.db.Close() }

// Get scans the cache for a stored embedding with cosine similarity >= threshold.
// Returns the cached JSON, true on hit; nil, false on miss.
func (c *Cache) Get(embedding []float32, threshold float64) (json.RawMessage, bool, error) {
	return c.GetInNamespace("search", embedding, threshold)
}

// GetInNamespace is Get scoped to a caller-owned cache namespace. Search and
// tripwire results have different JSON shapes, so they must not share entries.
func (c *Cache) GetInNamespace(namespace string, embedding []float32, threshold float64) (json.RawMessage, bool, error) {
	if namespace == "" {
		namespace = "search"
	}
	cutoff := time.Now().Add(-cacheTTL).UTC().Format(time.RFC3339)
	rows, err := c.db.Query(
		`SELECT embedding, results FROM query_cache
		 WHERE namespace = ? AND cached_at > ?
		 ORDER BY cached_at DESC
		 LIMIT 1000`,
		namespace, cutoff,
	)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var blob []byte
		var results string
		if err := rows.Scan(&blob, &results); err != nil {
			continue
		}
		stored, err := DecodeEmbedding(blob)
		if err != nil {
			continue
		}
		if CosineSim(embedding, stored) >= threshold {
			return json.RawMessage(results), true, nil
		}
	}
	return nil, false, nil
}

// Set stores a query embedding and its serialized results.
func (c *Cache) Set(embedding []float32, results []byte) error {
	return c.SetInNamespace("search", embedding, results)
}

// SetInNamespace stores a query embedding and serialized results in namespace.
func (c *Cache) SetInNamespace(namespace string, embedding []float32, results []byte) error {
	if namespace == "" {
		namespace = "search"
	}
	blob := EncodeEmbedding(embedding)
	_, err := c.db.Exec(
		`INSERT INTO query_cache (namespace, embedding, results, cached_at) VALUES (?, ?, ?, ?)`,
		namespace, blob, string(results), time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	return c.prune()
}

// ClearNamespacePrefixes deletes cache entries whose namespace starts with any
// of the supplied prefixes.
func (c *Cache) ClearNamespacePrefixes(prefixes ...string) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	for _, prefix := range prefixes {
		if prefix == "" {
			continue
		}
		if _, err := tx.Exec(
			`DELETE FROM query_cache WHERE namespace = ? OR namespace LIKE ?`,
			prefix, prefix+"%",
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// prune removes expired entries and caps the table at maxCacheSize rows.
// Both deletes run in a single transaction so a crash mid-prune can't leave
// the table temporarily oversized.
func (c *Cache) prune() error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	cutoff := time.Now().Add(-cacheTTL).UTC().Format(time.RFC3339)
	if _, err := tx.Exec(`DELETE FROM query_cache WHERE cached_at <= ?`, cutoff); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM query_cache WHERE id NOT IN (
			SELECT id FROM query_cache ORDER BY cached_at DESC LIMIT ?
		)`, maxCacheSize); err != nil {
		return err
	}
	return tx.Commit()
}

func ensureNamespaceColumn(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(query_cache)`)
	if err != nil {
		return fmt.Errorf("inspect cache schema: %w", err)
	}
	defer rows.Close()

	hasNamespace := false
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("scan cache schema: %w", err)
		}
		if name == "namespace" {
			hasNamespace = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if hasNamespace {
		return nil
	}
	if _, err := db.Exec(`ALTER TABLE query_cache ADD COLUMN namespace TEXT NOT NULL DEFAULT 'search'`); err != nil {
		return fmt.Errorf("migrate cache namespace: %w", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_query_cache_namespace_cached ON query_cache(namespace, cached_at)`)
	return err
}

// CosineSim computes the cosine similarity between two float32 vectors.
func CosineSim(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		fa, fb := float64(a[i]), float64(b[i])
		dot += fa * fb
		normA += fa * fa
		normB += fb * fb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// EncodeEmbedding serialises a float32 slice to little-endian bytes.
func EncodeEmbedding(v []float32) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, v)
	return buf.Bytes()
}

// DecodeEmbedding deserialises little-endian bytes to a float32 slice.
func DecodeEmbedding(b []byte) ([]float32, error) {
	if len(b)%4 != 0 {
		return nil, fmt.Errorf("embedding blob length %d not divisible by 4", len(b))
	}
	v := make([]float32, len(b)/4)
	if err := binary.Read(bytes.NewReader(b), binary.LittleEndian, v); err != nil {
		return nil, err
	}
	return v, nil
}
