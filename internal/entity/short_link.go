// Package entity defines the ShortLink entity and its validation logic.
package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/willfreit4s/short_link/pkg/entity"
)

var (
	ErrIDIsRequired          = errors.New("ID is required")
	ErrOriginalURLIsRequired = errors.New("OriginalURL is required")
)

type ShortLink struct {
	ID          entity.NanoID `json:"id"`
	OriginalURL string        `json:"original_url"`
	CreatedAt   time.Time     `json:"created_at"`
}

func NewShortLink(originalURL string) (*ShortLink, error) {
	shortLink := &ShortLink{
		ID:          entity.NewNanoID(),
		OriginalURL: normalizeOriginalURL(originalURL),
		CreatedAt:   time.Now(),
	}

	if err := shortLink.Validate(); err != nil {
		return nil, err
	}

	return shortLink, nil
}

func (s *ShortLink) Validate() error {
	if s.ID == "" {
		return ErrIDIsRequired
	}
	if s.OriginalURL == "" {
		return ErrOriginalURLIsRequired
	}
	return nil
}

func normalizeOriginalURL(originalURL string) string {
	originalURL = strings.TrimSpace(originalURL)

	if originalURL == "" {
		return originalURL
	}

	if strings.HasPrefix(originalURL, "http://") || strings.HasPrefix(originalURL, "https://") {
		return originalURL
	}

	return "https://" + originalURL
}
