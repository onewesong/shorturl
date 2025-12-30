package db

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"time"

	"github.com/mine/shorturl/internal/shortcode"
)

type Link struct {
	ID         int64
	Code       string
	TargetURL  string
	Enabled    bool
	ClickCount int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func ListLinks(db *sql.DB, limit int) ([]Link, error) {
	rows, err := db.Query(
		`SELECT id, code, target_url, enabled, click_count, created_at, updated_at
		 FROM links
		 ORDER BY id DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Link
	for rows.Next() {
		var l Link
		var enabledInt int
		if err := rows.Scan(&l.ID, &l.Code, &l.TargetURL, &enabledInt, &l.ClickCount, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		l.Enabled = enabledInt != 0
		out = append(out, l)
	}
	return out, rows.Err()
}

func GetLinkByID(db *sql.DB, id int64) (*Link, error) {
	var l Link
	var enabledInt int
	err := db.QueryRow(
		`SELECT id, code, target_url, enabled, click_count, created_at, updated_at
		 FROM links
		 WHERE id = ?`,
		id,
	).Scan(&l.ID, &l.Code, &l.TargetURL, &enabledInt, &l.ClickCount, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	l.Enabled = enabledInt != 0
	return &l, nil
}

func GetLinkByCode(db *sql.DB, code string) (*Link, error) {
	var l Link
	var enabledInt int
	err := db.QueryRow(
		`SELECT id, code, target_url, enabled, click_count, created_at, updated_at
		 FROM links
		 WHERE code = ?`,
		code,
	).Scan(&l.ID, &l.Code, &l.TargetURL, &enabledInt, &l.ClickCount, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	l.Enabled = enabledInt != 0
	return &l, nil
}

func CreateLink(db *sql.DB, code, targetURL string) error {
	_, err := db.Exec(
		`INSERT INTO links(code, target_url, enabled) VALUES(?, ?, 1)`,
		code, targetURL,
	)
	return err
}

func UpdateLink(db *sql.DB, id int64, targetURL string, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := db.Exec(
		`UPDATE links
		 SET target_url = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		targetURL, enabledInt, id,
	)
	return err
}

func IncrementClick(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE links SET click_count = click_count + 1 WHERE id = ?`, id)
	return err
}

func GenerateUniqueCode(db *sql.DB, length int) (string, error) {
	for range 20 {
		code, err := randomBase62(length)
		if err != nil {
			return "", err
		}
		if !shortcode.IsValidAuto(code) {
			continue
		}
		var exists int
		if err := db.QueryRow(`SELECT COUNT(1) FROM links WHERE code = ?`, code).Scan(&exists); err != nil {
			return "", err
		}
		if exists == 0 {
			return code, nil
		}
	}
	return "", errors.New("generate code retries exceeded")
}

func randomBase62(n int) (string, error) {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}

