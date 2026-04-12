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
	Enabled    bool      `json:"enabled"`
	ClickCount int64     `json:"click_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateLinkInput struct {
	Code      string `json:"code"`
	TargetURL string `json:"target_url"`
}

type UpdateLinkInput struct {
	Code      string `json:"code"`
	TargetURL string `json:"target_url"`
	Enabled   bool   `json:"enabled"`
}

type Repository interface {
	ListLinks(ctx context.Context, limit int) ([]Link, error)
	GetLinkByID(ctx context.Context, id int64) (Link, error)
	GetLinkByCode(ctx context.Context, code string) (Link, error)
	CreateLink(ctx context.Context, code string, targetURL string) (Link, error)
	UpdateLink(ctx context.Context, link Link) (Link, error)
	IncrementClick(ctx context.Context, id int64) error
}
