// Package usecase defines the use cases for managing short links.
package usecase

import (
	"context"
	"errors"

	"github.com/willfreit4s/short_link/internal/entity"
	"github.com/willfreit4s/short_link/internal/ports"
	usecasedto "github.com/willfreit4s/short_link/internal/usecase/dto"
	"github.com/willfreit4s/short_link/pkg/logger"
)

var ErrShortLinkNotFound = errors.New("short link not found")

type ShortLinkUseCase interface {
	CreateShortLink(ctx context.Context, input usecasedto.CreateShortLinkInput) (usecasedto.CreateShortLinkOutput, error)
	GetShortLink(ctx context.Context, input usecasedto.GetShortLinkInput) (usecasedto.GetShortLinkOutput, error)
}

type shortLinkUseCase struct {
	repository ports.ShortLinkRepository
	cache      ports.ShortLinkCache
}

func NewShortLinkUseCase(repository ports.ShortLinkRepository, cache ports.ShortLinkCache) ShortLinkUseCase {
	return &shortLinkUseCase{
		repository: repository,
		cache:      cache,
	}
}

func (uc *shortLinkUseCase) CreateShortLink(ctx context.Context, input usecasedto.CreateShortLinkInput) (usecasedto.CreateShortLinkOutput, error) {
	log := logger.FromContext(ctx)

	shortLink, err := entity.NewShortLink(input.OriginalURL)
	if err != nil {
		log.Error("failed to create short link entity", "error", err)
		return usecasedto.CreateShortLinkOutput{}, err
	}

	createdShortLink, err := uc.repository.CreateShortLink(ctx, shortLink)
	if err != nil {
		log.Error("failed to create short link in repository", "error", err)
		return usecasedto.CreateShortLinkOutput{}, err
	}
	if createdShortLink == nil {
		log.Error("short link repository returned nil short link")
		return usecasedto.CreateShortLinkOutput{}, errors.New("short link repository returned nil short link")
	}

	if err := uc.cache.Set(ctx, createdShortLink.ID.String(), createdShortLink.OriginalURL); err != nil {
		log.Error("failed to cache created short link", "error", err, "hash", createdShortLink.ID.String())
	}

	return usecasedto.CreateShortLinkOutput{
		ID:          createdShortLink.ID.String(),
		Hash:        createdShortLink.ID.String(),
		OriginalURL: createdShortLink.OriginalURL,
	}, nil
}

func (uc *shortLinkUseCase) GetShortLink(ctx context.Context, input usecasedto.GetShortLinkInput) (usecasedto.GetShortLinkOutput, error) {
	log := logger.FromContext(ctx)

	if cachedURL, found, err := uc.cache.Get(ctx, input.Hash); err != nil {
		log.Error("failed to get short link from cache", "error", err, "hash", input.Hash)
	} else if found {
		return usecasedto.GetShortLinkOutput{OriginalURL: cachedURL}, nil
	}

	shortLink, err := uc.repository.GetShortLink(ctx, input.Hash)
	if err != nil {
		log.Error("failed to get short link from repository", "error", err)
		return usecasedto.GetShortLinkOutput{}, err
	}
	if shortLink == nil {
		log.Error("short link not found", "hash", input.Hash)
		return usecasedto.GetShortLinkOutput{}, ErrShortLinkNotFound
	}

	if err := uc.cache.Set(ctx, input.Hash, shortLink.OriginalURL); err != nil {
		log.Error("failed to update short link cache after repository hit", "error", err, "hash", input.Hash)
	}

	return usecasedto.GetShortLinkOutput{OriginalURL: shortLink.OriginalURL}, nil
}
