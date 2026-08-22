package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	db "github.com/willfreit4s/short_link/internal/db"
	"github.com/willfreit4s/short_link/internal/entity"
	"github.com/willfreit4s/short_link/internal/ports"
	pkgentity "github.com/willfreit4s/short_link/pkg/entity"
)

type shortLinkRepository struct {
	executor db.DBTX
}

func NewShortLinkRepository(executor db.DBTX) ports.ShortLinkRepository {
	return &shortLinkRepository{
		executor: executor,
	}
}

func (r *shortLinkRepository) CreateShortLink(ctx context.Context, shortLink *entity.ShortLink) (*entity.ShortLink, error) {
	createdShortLink, err := r.queries(ctx).CreateShortLink(ctx, db.CreateShortLinkParams{
		ID:          shortLink.ID.String(),
		OriginalUrl: shortLink.OriginalURL,
	})
	if err != nil {
		return nil, err
	}

	return mapLink(createdShortLink), nil
}

func (r *shortLinkRepository) GetShortLink(ctx context.Context, id string) (*entity.ShortLink, error) {
	link, err := r.queries(ctx).GetShortLink(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return mapLink(link), nil
}

func (r *shortLinkRepository) queries(ctx context.Context) *db.Queries {
	return db.New(executorFromContext(ctx, r.executor))
}

func mapLink(link db.Link) *entity.ShortLink {
	return &entity.ShortLink{
		ID:          pkgentity.NanoID(link.ID),
		OriginalURL: link.OriginalUrl,
		CreatedAt:   link.CreatedAt.Time,
	}
}
