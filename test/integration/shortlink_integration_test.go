//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/willfreit4s/short_link/internal/entity"
	repositorypostgres "github.com/willfreit4s/short_link/internal/repository/postgres"
	repositoryredis "github.com/willfreit4s/short_link/internal/repository/redis"
	"github.com/willfreit4s/short_link/internal/usecase"
	usecasedto "github.com/willfreit4s/short_link/internal/usecase/dto"
	"github.com/willfreit4s/short_link/test/testhelpers"
)

func TestShortLinkRepository_Integration_CreateAndGet(t *testing.T) {
	pool, _ := testhelpers.SetupIntegrationDependencies(t)
	ctx := context.Background()
	repo := repositorypostgres.NewShortLinkRepository(pool)

	link, err := entity.NewShortLink("example.com")
	require.NoError(t, err)

	created, err := repo.CreateShortLink(ctx, link)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, link.OriginalURL, created.OriginalURL)

	found, err := repo.GetShortLink(ctx, created.ID.String())
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, created.ID.String(), found.ID.String())
	require.Equal(t, created.OriginalURL, found.OriginalURL)
}

func TestShortLinkUseCase_Integration_CreateAndResolve(t *testing.T) {
	pool, redisClient := testhelpers.SetupIntegrationDependencies(t)
	ctx := context.Background()
	repo := repositorypostgres.NewShortLinkRepository(pool)
	cache := repositoryredis.NewShortLinkCache(redisClient, 5*time.Minute)
	uc := usecase.NewShortLinkUseCase(repo, cache)

	result, err := uc.CreateShortLink(ctx, usecasedto.CreateShortLinkInput{OriginalURL: "example.com"})
	require.NoError(t, err)
	require.NotEmpty(t, result.Hash)
	require.Equal(t, "https://example.com", result.OriginalURL)

	cachedValue, err := redisClient.Get(ctx, "shortlink:"+result.Hash).Result()
	require.NoError(t, err)
	require.Equal(t, "https://example.com", cachedValue)

	resolved, err := uc.GetShortLink(ctx, usecasedto.GetShortLinkInput{Hash: result.Hash})
	require.NoError(t, err)
	require.Equal(t, "https://example.com", resolved.OriginalURL)
}

func TestShortLinkUseCase_Integration_UsesRedisCacheWhenPresent(t *testing.T) {
	pool, redisClient := testhelpers.SetupIntegrationDependencies(t)
	ctx := context.Background()
	repo := repositorypostgres.NewShortLinkRepository(pool)
	cache := repositoryredis.NewShortLinkCache(redisClient, 5*time.Minute)
	uc := usecase.NewShortLinkUseCase(repo, cache)

	err := cache.Set(ctx, "cached-hash", "https://from-cache.example.com")
	require.NoError(t, err)

	result, err := uc.GetShortLink(ctx, usecasedto.GetShortLinkInput{Hash: "cached-hash"})
	require.NoError(t, err)
	require.Equal(t, "https://from-cache.example.com", result.OriginalURL)
}
