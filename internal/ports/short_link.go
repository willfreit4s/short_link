package ports

import (
	"context"

	db "github.com/willfreit4s/short_link/internal/db"
)

type ShortLinkRepository interface {
	CreateShortLink(ctx context.Context, arg db.CreateShortLinkParams) (db.Link, error)
	GetShortLink(ctx context.Context, id string) (db.Link, error)
}