package bootstrap

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/willfreit4s/short_link/internal/handler"
	repositorypostgres "github.com/willfreit4s/short_link/internal/repository/postgres"
	"github.com/willfreit4s/short_link/internal/usecase"
	"github.com/willfreit4s/short_link/pkg/logger"
)

func NewRouter(log *slog.Logger, conn *pgxpool.Pool) *gin.Engine {
	shortLinkRepository := repositorypostgres.NewShortLinkRepository(conn)
	transactionManager := repositorypostgres.NewTransactionManager(conn)
	shortLinkUseCase := usecase.NewShortLinkUseCase(shortLinkRepository, transactionManager)
	shortLinkHandler := handler.NewShortLinkHandler(shortLinkUseCase)

	router := gin.New()
	router.Use(logger.RequestIDMiddleware())
	router.Use(logger.SlogMiddleware(log))
	router.Use(gin.Recovery())

	router.GET("/r/:hash", shortLinkHandler.GetShortLink)

	v1 := router.Group("/api/v1")
	v1.POST("/links", shortLinkHandler.CreateShortLink)

	return router
}
