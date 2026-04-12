package links

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/mine/shorturl/internal/shortcode"
)

const defaultListLimit = 200

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Link, error) {
	return s.repo.ListLinks(ctx, defaultListLimit)
}

func (s *Service) Create(ctx context.Context, input CreateLinkInput) (Link, error) {
	code := strings.TrimSpace(input.Code)
	targetURL := strings.TrimSpace(input.TargetURL)
	remark := strings.TrimSpace(input.Remark)
	tags := normalizeTags(input.Tags)

	if !isValidURL(targetURL) {
		return Link{}, fmt.Errorf("%w: invalid target_url", ErrValidation)
	}
	if code != "" && !shortcode.IsValidCustom(code) {
		return Link{}, fmt.Errorf("%w: invalid code", ErrValidation)
	}

	if code == "" {
		generatedCode, err := generateUniqueCode(ctx, s.repo, 6)
		if err != nil {
			return Link{}, err
		}
		code = generatedCode
	}

	return s.repo.CreateLink(ctx, code, targetURL, remark, tags)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateLinkInput) (Link, error) {
	current, err := s.repo.GetLinkByID(ctx, id)
	if err != nil {
		return Link{}, err
	}

	code := strings.TrimSpace(input.Code)
	targetURL := strings.TrimSpace(input.TargetURL)
	remark := strings.TrimSpace(input.Remark)
	tags := normalizeTags(input.Tags)

	if code == "" {
		return Link{}, fmt.Errorf("%w: code is required", ErrValidation)
	}
	if !shortcode.IsValidCustom(code) {
		return Link{}, fmt.Errorf("%w: invalid code", ErrValidation)
	}
	if !isValidURL(targetURL) {
		return Link{}, fmt.Errorf("%w: invalid target_url", ErrValidation)
	}

	current.Code = code
	current.TargetURL = targetURL
	current.Remark = remark
	current.Tags = tags
	current.Enabled = input.Enabled

	return s.repo.UpdateLink(ctx, current)
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return "", ErrLinkNotFound
	}

	link, err := s.repo.GetLinkByCode(ctx, trimmed)
	if err != nil {
		return "", err
	}
	if !link.Enabled {
		return "", ErrLinkNotFound
	}

	_ = s.repo.IncrementClick(ctx, link.ID)
	return link.TargetURL, nil
}

func isValidURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

func generateUniqueCode(ctx context.Context, repo Repository, length int) (string, error) {
	for range 20 {
		code, err := randomBase62(length)
		if err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}
		if !shortcode.IsValidAuto(code) {
			continue
		}
		if _, err := repo.GetLinkByCode(ctx, code); err == nil {
			continue
		} else if !errors.Is(err, ErrLinkNotFound) {
			return "", err
		}
		return code, nil
	}

	return "", fmt.Errorf("generate code: retries exceeded")
}

func randomBase62(n int) (string, error) {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	for index := range buf {
		buf[index] = alphabet[int(buf[index])%len(alphabet)]
	}

	return string(buf), nil
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.TrimSpace(tag)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}
