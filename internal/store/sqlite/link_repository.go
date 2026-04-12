package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

func (r *LinkRepository) RecordVisit(ctx context.Context, linkID int64, meta links.VisitMeta) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO link_visits(link_id, ip, referer, referer_host, user_agent, client_name, client_type, device_type, os, visited_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		linkID,
		meta.IP,
		meta.Referer,
		meta.RefererHost,
		meta.UserAgent,
		meta.ClientName,
		meta.ClientType,
		meta.DeviceType,
		meta.OS,
		meta.VisitedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (r *LinkRepository) GetLinkAnalytics(ctx context.Context, id int64, since time.Time, limit int) (links.LinkAnalytics, error) {
	link, err := r.GetLinkByID(ctx, id)
	if err != nil {
		return links.LinkAnalytics{}, err
	}

	analytics := links.LinkAnalytics{
		Link:         link,
		TimeSeries:   []links.VisitPoint{},
		TopReferrers: []links.VisitBreakdown{},
		TopClients:   []links.VisitBreakdown{},
		RecentVisits: []links.VisitRecord{},
	}

	var lastVisitedAt sql.NullString
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*), COUNT(DISTINCT CASE WHEN ip <> '' THEN ip END), COALESCE(MAX(strftime('%Y-%m-%dT%H:%M:%fZ', visited_at)), '')
		 FROM link_visits
		 WHERE link_id = ? AND visited_at >= ?`,
		id,
		since.UTC(),
	).Scan(&analytics.RecentClicks, &analytics.UniqueIPs, &lastVisitedAt); err != nil {
		return links.LinkAnalytics{}, err
	}
	if lastVisitedAt.Valid && lastVisitedAt.String != "" {
		value, err := parseSQLiteTime(lastVisitedAt.String)
		if err != nil {
			return links.LinkAnalytics{}, err
		}
		analytics.LastVisitedAt = &value
	}

	timeSeriesRows, err := r.db.QueryContext(
		ctx,
		`SELECT DATE(visited_at) AS bucket, COUNT(*)
		 FROM link_visits
		 WHERE link_id = ? AND visited_at >= ?
		 GROUP BY DATE(visited_at)
		 ORDER BY bucket ASC`,
		id,
		since.UTC(),
	)
	if err != nil {
		return links.LinkAnalytics{}, err
	}
	defer timeSeriesRows.Close()

	seriesByDate := make(map[string]int64)
	for timeSeriesRows.Next() {
		var point links.VisitPoint
		if err := timeSeriesRows.Scan(&point.Bucket, &point.Clicks); err != nil {
			return links.LinkAnalytics{}, err
		}
		seriesByDate[point.Bucket] = point.Clicks
	}
	if err := timeSeriesRows.Err(); err != nil {
		return links.LinkAnalytics{}, err
	}

	totalDays := max(int(time.Now().UTC().Sub(since.UTC()).Hours()/24)+1, 1)
	for day := 0; day < totalDays; day++ {
		bucket := since.UTC().AddDate(0, 0, day).Format("2006-01-02")
		analytics.TimeSeries = append(analytics.TimeSeries, links.VisitPoint{
			Bucket: bucket,
			Clicks: seriesByDate[bucket],
		})
	}

	referrerRows, err := r.db.QueryContext(
		ctx,
		`SELECT CASE WHEN referer_host = '' THEN '直接访问' ELSE referer_host END AS referrer_name, COUNT(*) AS total
		 FROM link_visits
		 WHERE link_id = ? AND visited_at >= ?
		 GROUP BY referrer_name
		 ORDER BY total DESC, referrer_name ASC
		 LIMIT 8`,
		id,
		since.UTC(),
	)
	if err != nil {
		return links.LinkAnalytics{}, err
	}
	defer referrerRows.Close()

	for referrerRows.Next() {
		var item links.VisitBreakdown
		if err := referrerRows.Scan(&item.Name, &item.Count); err != nil {
			return links.LinkAnalytics{}, err
		}
		analytics.TopReferrers = append(analytics.TopReferrers, item)
	}
	if err := referrerRows.Err(); err != nil {
		return links.LinkAnalytics{}, err
	}

	clientRows, err := r.db.QueryContext(
		ctx,
		`SELECT CASE WHEN client_name = '' THEN '未知客户端' ELSE client_name END AS client_name, COUNT(*) AS total
		 FROM link_visits
		 WHERE link_id = ? AND visited_at >= ?
		 GROUP BY client_name
		 ORDER BY total DESC, client_name ASC
		 LIMIT 8`,
		id,
		since.UTC(),
	)
	if err != nil {
		return links.LinkAnalytics{}, err
	}
	defer clientRows.Close()

	for clientRows.Next() {
		var item links.VisitBreakdown
		if err := clientRows.Scan(&item.Name, &item.Count); err != nil {
			return links.LinkAnalytics{}, err
		}
		analytics.TopClients = append(analytics.TopClients, item)
	}
	if err := clientRows.Err(); err != nil {
		return links.LinkAnalytics{}, err
	}

	visitRows, err := r.db.QueryContext(
		ctx,
		`SELECT strftime('%Y-%m-%dT%H:%M:%fZ', visited_at), ip, referer, referer_host, user_agent, client_name, client_type, device_type, os
		 FROM link_visits
		 WHERE link_id = ?
		 ORDER BY visited_at DESC
		 LIMIT ?`,
		id,
		limit,
	)
	if err != nil {
		return links.LinkAnalytics{}, err
	}
	defer visitRows.Close()

	for visitRows.Next() {
		var (
			record  links.VisitRecord
			rawTime string
			rawIP   string
		)
		if err := visitRows.Scan(
			&rawTime,
			&rawIP,
			&record.Referer,
			&record.RefererHost,
			&record.UserAgent,
			&record.ClientName,
			&record.ClientType,
			&record.DeviceType,
			&record.OS,
		); err != nil {
			return links.LinkAnalytics{}, err
		}
		recordTime, err := parseSQLiteTime(rawTime)
		if err != nil {
			return links.LinkAnalytics{}, err
		}
		record.VisitedAt = recordTime
		record.IPMasked = links.MaskVisitIP(rawIP)
		analytics.RecentVisits = append(analytics.RecentVisits, record)
	}
	if err := visitRows.Err(); err != nil {
		return links.LinkAnalytics{}, err
	}

	return analytics, nil
}

func parseSQLiteTime(raw string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("parse sqlite time: %s", raw)
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
