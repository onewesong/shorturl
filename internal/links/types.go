package links

import (
	"context"
	"errors"
	"time"
)

var (
	ErrLinkNotFound = errors.New("link not found")
	ErrLinkExists   = errors.New("link already exists")
	ErrValidation   = errors.New("validation failed")
)

type Link struct {
	ID         int64     `json:"id"`
	Code       string    `json:"code"`
	TargetURL  string    `json:"target_url"`
	Remark     string    `json:"remark"`
	Tags       []string  `json:"tags"`
	Enabled    bool      `json:"enabled"`
	ClickCount int64     `json:"click_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateLinkInput struct {
	Code      string   `json:"code"`
	TargetURL string   `json:"target_url"`
	Remark    string   `json:"remark"`
	Tags      []string `json:"tags"`
}

type UpdateLinkInput struct {
	Code      string   `json:"code"`
	TargetURL string   `json:"target_url"`
	Remark    string   `json:"remark"`
	Tags      []string `json:"tags"`
	Enabled   bool     `json:"enabled"`
}

type VisitMeta struct {
	VisitedAt   time.Time
	IP          string
	Referer     string
	RefererHost string
	UserAgent   string
	ClientName  string
	ClientType  string
	DeviceType  string
	OS          string
}

type VisitPoint struct {
	Bucket string `json:"bucket"`
	Clicks int64  `json:"clicks"`
}

type VisitBreakdown struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type VisitRecord struct {
	VisitedAt   time.Time `json:"visited_at"`
	IPMasked    string    `json:"ip_masked"`
	Referer     string    `json:"referer"`
	RefererHost string    `json:"referer_host"`
	UserAgent   string    `json:"user_agent"`
	ClientName  string    `json:"client_name"`
	ClientType  string    `json:"client_type"`
	DeviceType  string    `json:"device_type"`
	OS          string    `json:"os"`
}

type LinkAnalytics struct {
	Link          Link             `json:"link"`
	RangeDays     int              `json:"range_days"`
	RecentClicks  int64            `json:"recent_clicks"`
	UniqueIPs     int64            `json:"unique_ips"`
	LastVisitedAt *time.Time       `json:"last_visited_at,omitempty"`
	TimeSeries    []VisitPoint     `json:"time_series"`
	TopReferrers  []VisitBreakdown `json:"top_referrers"`
	TopClients    []VisitBreakdown `json:"top_clients"`
	RecentVisits  []VisitRecord    `json:"recent_visits"`
}

type Repository interface {
	ListLinks(ctx context.Context, limit int) ([]Link, error)
	GetLinkByID(ctx context.Context, id int64) (Link, error)
	GetLinkByCode(ctx context.Context, code string) (Link, error)
	CreateLink(ctx context.Context, code string, targetURL string, remark string, tags []string) (Link, error)
	UpdateLink(ctx context.Context, link Link) (Link, error)
	DeleteLink(ctx context.Context, id int64) error
	IncrementClick(ctx context.Context, id int64) error
	RecordVisit(ctx context.Context, linkID int64, meta VisitMeta) error
	GetLinkAnalytics(ctx context.Context, id int64, since time.Time, limit int) (LinkAnalytics, error)
}
