// Package usecase defines the use cases for managing short links.
package usecase

import (
	"context"

	"github.com/willfreit4s/short_link/internal/entity"
)

type ShortLinkUseCase struct{}

func NewShortLinkUseCase() *ShortLinkUseCase {
	return &ShortLinkUseCase{}
}

func (uc *ShortLinkUseCase) CreateShortLink(ctx context.Context, originalURL string) (*entity.ShortLink, error) {
	shortLink, err := entity.NewShortLink(originalURL)
	if err != nil {
		return nil, err
	}

	return shortLink, nil
}

func (uc *ShortLinkUseCase) GetShortLink(ctx context.Context, hash string) (string, error) {
	// Implementation for retrieving the original URL based on the hash
	return "", nil
}
