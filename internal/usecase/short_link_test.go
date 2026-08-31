package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/willfreit4s/short_link/internal/entity"
	usecasedto "github.com/willfreit4s/short_link/internal/usecase/dto"
)

type stubRepository struct {
	createFn func(ctx context.Context, shortLink *entity.ShortLink) (*entity.ShortLink, error)
	getFn    func(ctx context.Context, id string) (*entity.ShortLink, error)
}

func (s *stubRepository) CreateShortLink(ctx context.Context, shortLink *entity.ShortLink) (*entity.ShortLink, error) {
	if s.createFn != nil {
		return s.createFn(ctx, shortLink)
	}
	return nil, nil
}

func (s *stubRepository) GetShortLink(ctx context.Context, id string) (*entity.ShortLink, error) {
	if s.getFn != nil {
		return s.getFn(ctx, id)
	}
	return nil, nil
}

type stubCache struct {
	getFn func(ctx context.Context, hash string) (string, bool, error)
	setFn func(ctx context.Context, hash string, originalURL string) error
	deleteFn func(ctx context.Context, hash string) error
}

func (s *stubCache) Get(ctx context.Context, hash string) (string, bool, error) {
	if s.getFn != nil {
		return s.getFn(ctx, hash)
	}
	return "", false, nil
}

func (s *stubCache) Set(ctx context.Context, hash string, originalURL string) error {
	if s.setFn != nil {
		return s.setFn(ctx, hash, originalURL)
	}
	return nil
}

func (s *stubCache) Delete(ctx context.Context, hash string) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, hash)
	}
	return nil
}

func TestShortLinkUseCase_CreateShortLink(t *testing.T) {
	t.Parallel()

	t.Run("success creates and caches short link", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		var repoCalled bool
		var cacheCalled bool
		createdShortLink := &entity.ShortLink{ID: "abc123", OriginalURL: "https://example.com"}
		repo := &stubRepository{
			createFn: func(ctx context.Context, shortLink *entity.ShortLink) (*entity.ShortLink, error) {
				repoCalled = true
				require.NotNil(t, shortLink)
				require.Equal(t, "https://example.com", shortLink.OriginalURL)
				return createdShortLink, nil
			},
		}
		cache := &stubCache{
			setFn: func(ctx context.Context, hash string, originalURL string) error {
				cacheCalled = true
				require.Equal(t, createdShortLink.ID.String(), hash)
				require.Equal(t, "https://example.com", originalURL)
				return nil
			},
		}

		uc := NewShortLinkUseCase(repo, cache)
		result, err := uc.CreateShortLink(ctx, usecasedto.CreateShortLinkInput{OriginalURL: "example.com"})

		require.NoError(t, err)
		require.True(t, repoCalled)
		require.True(t, cacheCalled)
		require.NotEmpty(t, result.ID)
		require.Equal(t, result.Hash, result.ID)
		require.Equal(t, "https://example.com", result.OriginalURL)
	})

	t.Run("returns validation error for empty url", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &stubRepository{}
		cache := &stubCache{}
		uc := NewShortLinkUseCase(repo, cache)

		_, err := uc.CreateShortLink(ctx, usecasedto.CreateShortLinkInput{OriginalURL: "  "})

		require.Error(t, err)
		require.EqualError(t, err, entity.ErrOriginalURLIsRequired.Error())
	})

	t.Run("propagates repository error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repoErr := errors.New("database unavailable")
		repo := &stubRepository{
			createFn: func(ctx context.Context, shortLink *entity.ShortLink) (*entity.ShortLink, error) {
				return nil, repoErr
			},
		}
		cache := &stubCache{}
		uc := NewShortLinkUseCase(repo, cache)

		_, err := uc.CreateShortLink(ctx, usecasedto.CreateShortLinkInput{OriginalURL: "example.com"})

		require.ErrorIs(t, err, repoErr)
	})

	t.Run("returns error when repository returns nil link", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &stubRepository{
			createFn: func(ctx context.Context, shortLink *entity.ShortLink) (*entity.ShortLink, error) {
				return nil, nil
			},
		}
		cache := &stubCache{}
		uc := NewShortLinkUseCase(repo, cache)

		_, err := uc.CreateShortLink(ctx, usecasedto.CreateShortLinkInput{OriginalURL: "example.com"})

		require.Error(t, err)
		require.Contains(t, err.Error(), "repository returned nil short link")
	})

	t.Run("ignores cache write errors and still returns success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		repo := &stubRepository{
			createFn: func(ctx context.Context, shortLink *entity.ShortLink) (*entity.ShortLink, error) {
				return shortLink, nil
			},
		}
		cache := &stubCache{
			setFn: func(ctx context.Context, hash string, originalURL string) error {
				return errors.New("redis unavailable")
			},
		}
		uc := NewShortLinkUseCase(repo, cache)

		result, err := uc.CreateShortLink(ctx, usecasedto.CreateShortLinkInput{OriginalURL: "example.com"})

		require.NoError(t, err)
		require.Equal(t, "https://example.com", result.OriginalURL)
		require.NotEmpty(t, result.Hash)
	})
}

func TestShortLinkUseCase_GetShortLink(t *testing.T) {
	t.Parallel()

	t.Run("returns cached url when cache hit", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cache := &stubCache{
			getFn: func(ctx context.Context, hash string) (string, bool, error) {
				require.Equal(t, "abc123", hash)
				return "https://cached.example.com", true, nil
			},
		}
		repo := &stubRepository{}
		uc := NewShortLinkUseCase(repo, cache)

		result, err := uc.GetShortLink(ctx, usecasedto.GetShortLinkInput{Hash: "abc123"})

		require.NoError(t, err)
		require.Equal(t, "https://cached.example.com", result.OriginalURL)
	})

	t.Run("reads from repository when cache misses", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cache := &stubCache{
			getFn: func(ctx context.Context, hash string) (string, bool, error) {
				return "", false, nil
			},
			setFn: func(ctx context.Context, hash string, originalURL string) error {
				require.Equal(t, "abc123", hash)
				require.Equal(t, "https://example.com", originalURL)
				return nil
			},
		}
		repo := &stubRepository{
			getFn: func(ctx context.Context, id string) (*entity.ShortLink, error) {
				require.Equal(t, "abc123", id)
				return &entity.ShortLink{ID: "abc123", OriginalURL: "https://example.com"}, nil
			},
		}
		uc := NewShortLinkUseCase(repo, cache)

		result, err := uc.GetShortLink(ctx, usecasedto.GetShortLinkInput{Hash: "abc123"})

		require.NoError(t, err)
		require.Equal(t, "https://example.com", result.OriginalURL)
	})

	t.Run("returns not found when repository has no record", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cache := &stubCache{
			getFn: func(ctx context.Context, hash string) (string, bool, error) {
				return "", false, nil
			},
		}
		repo := &stubRepository{
			getFn: func(ctx context.Context, id string) (*entity.ShortLink, error) {
				return nil, nil
			},
		}
		uc := NewShortLinkUseCase(repo, cache)

		_, err := uc.GetShortLink(ctx, usecasedto.GetShortLinkInput{Hash: "missing"})

		require.ErrorIs(t, err, ErrShortLinkNotFound)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cache := &stubCache{
			getFn: func(ctx context.Context, hash string) (string, bool, error) {
				return "", false, nil
			},
		}
		repoErr := errors.New("database read failed")
		repo := &stubRepository{
			getFn: func(ctx context.Context, id string) (*entity.ShortLink, error) {
				return nil, repoErr
			},
		}
		uc := NewShortLinkUseCase(repo, cache)

		_, err := uc.GetShortLink(ctx, usecasedto.GetShortLinkInput{Hash: "abc123"})

		require.ErrorIs(t, err, repoErr)
	})

	t.Run("ignores cache write errors after repository hit", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		cache := &stubCache{
			getFn: func(ctx context.Context, hash string) (string, bool, error) {
				return "", false, nil
			},
			setFn: func(ctx context.Context, hash string, originalURL string) error {
				return errors.New("redis unavailable")
			},
		}
		repo := &stubRepository{
			getFn: func(ctx context.Context, id string) (*entity.ShortLink, error) {
				return &entity.ShortLink{ID: "abc123", OriginalURL: "https://example.com"}, nil
			},
		}
		uc := NewShortLinkUseCase(repo, cache)

		result, err := uc.GetShortLink(ctx, usecasedto.GetShortLinkInput{Hash: "abc123"})

		require.NoError(t, err)
		require.Equal(t, "https://example.com", result.OriginalURL)
	})
}

