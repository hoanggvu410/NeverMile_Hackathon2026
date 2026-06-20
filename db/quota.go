package db

import (
	"database/sql"
	"fmt"
	"time"
)

func currentMonth() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func GetMonthlyUsage(db *sql.DB, userID string) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT sync_count FROM sync_quota_usage WHERE user_id = $1 AND month = $2
	`, userID, currentMonth()).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("GetMonthlyUsage: %w", err)
	}
	return count, nil
}

func IncrementQuota(db *sql.DB, userID string, count int) error {
	if count == 0 {
		return nil
	}
	_, err := db.Exec(`
		INSERT INTO sync_quota_usage (user_id, month, sync_count)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, month) DO UPDATE
		SET sync_count = sync_quota_usage.sync_count + EXCLUDED.sync_count
	`, userID, currentMonth(), count)
	if err != nil {
		return fmt.Errorf("IncrementQuota: %w", err)
	}
	return nil
}

func CheckAndIncrementQuota(db *sql.DB, userID string, count int) error {
	if count == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	month := currentMonth()

	var current int
	err = tx.QueryRow(`
		SELECT sync_count FROM sync_quota_usage
		WHERE user_id = $1 AND month = $2
		FOR UPDATE
	`, userID, month).Scan(&current)
	if err == sql.ErrNoRows {
		current = 0
	} else if err != nil {
		return fmt.Errorf("CheckAndIncrementQuota read: %w", err)
	}

	if current+count > 20 {
		return ErrQuotaExceeded
	}

	_, err = tx.Exec(`
		INSERT INTO sync_quota_usage (user_id, month, sync_count)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, month) DO UPDATE
		SET sync_count = sync_quota_usage.sync_count + EXCLUDED.sync_count
	`, userID, month, count)
	if err != nil {
		return fmt.Errorf("CheckAndIncrementQuota write: %w", err)
	}

	return tx.Commit()
}
