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
	repository         ports.ShortLinkRepository
	transactionManager ports.TransactionManager
}

func NewShortLinkUseCase(repository ports.ShortLinkRepository, transactionManager ports.TransactionManager) ShortLinkUseCase {
	return &shortLinkUseCase{
		repository:         repository,
		transactionManager: transactionManager,
	}
}

func (uc *shortLinkUseCase) CreateShortLink(ctx context.Context, input usecasedto.CreateShortLinkInput) (usecasedto.CreateShortLinkOutput, error) {
	log := logger.FromContext(ctx)
	log.Info("creating short link")

	shortLink, err := entity.NewShortLink(input.OriginalURL)
	if err != nil {
		return usecasedto.CreateShortLinkOutput{}, err
	}

	createdShortLink, err := uc.repository.CreateShortLink(ctx, shortLink)
	if err != nil {
		return usecasedto.CreateShortLinkOutput{}, err
	}
	if createdShortLink == nil {
		return usecasedto.CreateShortLinkOutput{}, errors.New("short link repository returned nil short link")
	}

	log.Info("short link created", "short_link_id", createdShortLink.ID)

	return usecasedto.CreateShortLinkOutput{
		ID:          createdShortLink.ID.String(),
		Hash:        createdShortLink.ID.String(),
		OriginalURL: createdShortLink.OriginalURL,
	}, nil
}

func (uc *shortLinkUseCase) GetShortLink(ctx context.Context, input usecasedto.GetShortLinkInput) (usecasedto.GetShortLinkOutput, error) {
	log := logger.FromContext(ctx)
	log.Info("resolving short link", "short_link_id", input.Hash)

	shortLink, err := uc.repository.GetShortLink(ctx, input.Hash)
	if err != nil {
		return usecasedto.GetShortLinkOutput{}, err
	}
	if shortLink == nil {
		return usecasedto.GetShortLinkOutput{}, ErrShortLinkNotFound
	}

	log.Info("short link resolved", "short_link_id", shortLink.ID)

	return usecasedto.GetShortLinkOutput{OriginalURL: shortLink.OriginalURL}, nil
}
