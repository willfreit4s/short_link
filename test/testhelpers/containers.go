//go:build integration

package testhelpers

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupIntegrationDependencies(t *testing.T) (*pgxpool.Pool, *redisclient.Client) {
	t.Helper()

	ctx := context.Background()

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:18.1",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "short_link",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	pgHost, err := postgresContainer.Host(ctx)
	require.NoError(t, err)
	pgPort, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/short_link?sslmode=disable", pgHost, pgPort.Port())
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		pool.Close()
	})

	require.NoError(t, pool.Ping(ctx))

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS links (
			id varchar(10) NOT NULL PRIMARY KEY,
			original_url text NOT NULL,
			created_at timestamp NOT NULL DEFAULT now()
		);
	`)
	require.NoError(t, err)

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:8-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	})

	redisHost, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	redisClient := redisclient.NewClient(&redisclient.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
		DB:   0,
	})
	t.Cleanup(func() {
		if err := redisClient.Close(); err != nil {
			t.Logf("failed to close redis client: %v", err)
		}
	})

	require.NoError(t, redisClient.Ping(ctx).Err())

	return pool, redisClient
}
