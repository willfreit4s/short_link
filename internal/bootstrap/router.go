package bootstrap

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/willfreit4s/short_link/internal/handler"
	repositorypostgres "github.com/willfreit4s/short_link/internal/repository/postgres"
	repositoryredis "github.com/willfreit4s/short_link/internal/repository/redis"
	"github.com/willfreit4s/short_link/internal/usecase"
	"github.com/willfreit4s/short_link/pkg/logger"
	"github.com/willfreit4s/short_link/pkg/metrics"
)

func NewRouter(log *slog.Logger, conn *pgxpool.Pool, redisConn *redisclient.Client, redisTTL time.Duration) *gin.Engine {
	shortLinkRepository := repositorypostgres.NewShortLinkRepository(conn)
	shortLinkCache := repositoryredis.NewShortLinkCache(redisConn, redisTTL)
	shortLinkUseCase := usecase.NewShortLinkUseCase(shortLinkRepository, shortLinkCache)
	shortLinkHandler := handler.NewShortLinkHandler(shortLinkUseCase)

	router := gin.New()
	router.Use(logger.RequestIDMiddleware())
	router.Use(logger.SlogMiddleware(log))
	router.Use(metrics.Middleware())
	router.Use(gin.Recovery())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/r/:hash", shortLinkHandler.GetShortLink)
	router.GET("/health", shortLinkHandler.GetHealth)

	v1 := router.Group("/api/v1")
	v1.POST("/links", shortLinkHandler.CreateShortLink)

	return router
}
