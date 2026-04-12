package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mine/shorturl/internal/links"
)

type LinkRepository struct {
	db *sql.DB
}

func NewLinkRepository(db *sql.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) ListLinks(ctx context.Context, limit int) ([]links.Link, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, code, target_url, remark, tags_json, enabled, click_count, created_at, updated_at
		 FROM links
		 ORDER BY id DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]links.Link, 0, limit)
	for rows.Next() {
		link, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, link)
	}

	return result, rows.Err()
}

func (r *LinkRepository) GetLinkByID(ctx context.Context, id int64) (links.Link, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, code, target_url, remark, tags_json, enabled, click_count, created_at, updated_at
		 FROM links
		 WHERE id = ?`,
		id,
	)

	link, err := scanLink(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return links.Link{}, links.ErrLinkNotFound
		}
		return links.Link{}, err
	}

	return link, nil
}

func (r *LinkRepository) GetLinkByCode(ctx context.Context, code string) (links.Link, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, code, target_url, remark, tags_json, enabled, click_count, created_at, updated_at
		 FROM links
		 WHERE code = ?`,
		code,
	)

	link, err := scanLink(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return links.Link{}, links.ErrLinkNotFound
		}
		return links.Link{}, err
	}

	return link, nil
}

func (r *LinkRepository) CreateLink(ctx context.Context, code string, targetURL string, remark string, tags []string) (links.Link, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return links.Link{}, fmt.Errorf("marshal link tags: %w", err)
	}

	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO links(code, target_url, remark, tags_json, enabled) VALUES(?, ?, ?, ?, 1)`,
		code,
		targetURL,
		remark,
		string(tagsJSON),
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return links.Link{}, links.ErrLinkExists
		}
		return links.Link{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return links.Link{}, err
	}

	return r.GetLinkByID(ctx, id)
}

func (r *LinkRepository) UpdateLink(ctx context.Context, link links.Link) (links.Link, error) {
	enabled := 0
	if link.Enabled {
		enabled = 1
	}

	tagsJSON, err := json.Marshal(link.Tags)
	if err != nil {
		return links.Link{}, fmt.Errorf("marshal link tags: %w", err)
	}

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE links
		 SET code = ?, target_url = ?, remark = ?, tags_json = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		link.Code,
		link.TargetURL,
		link.Remark,
		string(tagsJSON),
		enabled,
		link.ID,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return links.Link{}, links.ErrLinkExists
		}
		return links.Link{}, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return links.Link{}, err
	}
	if rowsAffected == 0 {
		return links.Link{}, links.ErrLinkNotFound
	}

	return r.GetLinkByID(ctx, link.ID)
}

func (r *LinkRepository) IncrementClick(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE links SET click_count = click_count + 1 WHERE id = ?`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanLink(scanTarget scanner) (links.Link, error) {
	var link links.Link
	var tagsJSON string
	var enabled int

	err := scanTarget.Scan(
		&link.ID,
		&link.Code,
		&link.TargetURL,
		&link.Remark,
		&tagsJSON,
		&enabled,
		&link.ClickCount,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err != nil {
		return links.Link{}, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &link.Tags); err != nil {
		return links.Link{}, fmt.Errorf("decode link tags: %w", err)
	}
	if len(link.Tags) == 0 {
		link.Tags = []string{}
	}

	link.Enabled = enabled != 0
	return link, nil
}
