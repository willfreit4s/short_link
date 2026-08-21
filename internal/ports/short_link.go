package ports

import (
	"context"

	"github.com/willfreit4s/short_link/internal/entity"
)

type ShortLinkRepository interface {
	CreateShortLink(ctx context.Context, shortLink *entity.ShortLink) (*entity.ShortLink, error)
	GetShortLink(ctx context.Context, id string) (*entity.ShortLink, error)
}

type ShortLinkCache interface {
	Get(ctx context.Context, hash string) (string, bool, error)
	Set(ctx context.Context, hash string, originalURL string) error
	Delete(ctx context.Context, hash string) error
}
