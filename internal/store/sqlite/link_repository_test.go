package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/mine/shorturl/internal/links"
)

func TestGetLinkAnalytics(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "analytics-test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	if err := Init(ctx, database); err != nil {
		t.Fatalf("init db: %v", err)
	}

	repo := NewLinkRepository(database)
	link, err := repo.CreateLink(ctx, "demo", "https://example.com/demo", "", nil)
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	if err := repo.RecordVisit(ctx, link.ID, links.VisitMeta{
		VisitedAt:   time.Now().UTC(),
		IP:          "10.2.3.4",
		Referer:     "https://mp.weixin.qq.com/s/demo",
		RefererHost: "mp.weixin.qq.com",
		UserAgent:   "Mozilla/5.0 MicroMessenger/8.0.49",
		ClientName:  "微信",
		ClientType:  "app",
		DeviceType:  "mobile",
		OS:          "iOS",
	}); err != nil {
		t.Fatalf("record visit: %v", err)
	}

	analytics, err := repo.GetLinkAnalytics(ctx, link.ID, time.Now().UTC().Add(-24*time.Hour), 20)
	if err != nil {
		t.Fatalf("get analytics: %v", err)
	}

	if analytics.RecentClicks != 1 {
		t.Fatalf("expected 1 recent click, got %d", analytics.RecentClicks)
	}
	if len(analytics.RecentVisits) != 1 {
		t.Fatalf("expected 1 recent visit, got %d", len(analytics.RecentVisits))
	}
}

func TestDeleteLink(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "delete-test.db")
	database, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	if err := Init(ctx, database); err != nil {
		t.Fatalf("init db: %v", err)
	}

	repo := NewLinkRepository(database)
	link, err := repo.CreateLink(ctx, "delete-me", "https://example.com/delete", "", nil)
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	if err := repo.DeleteLink(ctx, link.ID); err != nil {
		t.Fatalf("delete link: %v", err)
	}

	if _, err := repo.GetLinkByID(ctx, link.ID); err != links.ErrLinkNotFound {
		t.Fatalf("expected ErrLinkNotFound, got %v", err)
	}

	if err := repo.DeleteLink(ctx, link.ID); err != links.ErrLinkNotFound {
		t.Fatalf("expected ErrLinkNotFound for missing delete, got %v", err)
	}
}
