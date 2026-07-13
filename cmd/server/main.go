package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/willfreit4s/short_link/internal/handler"
	"github.com/willfreit4s/short_link/internal/usecase"
	"github.com/willfreit4s/short_link/pkg/logger"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	router := gin.New()
	router.Use(logger.RequestIDMiddleware())
	router.Use(logger.SlogMiddleware(log))
	router.Use(gin.Recovery())

	shortLinkHandler := handler.NewShortLinkHandler(usecase.NewShortLinkUseCase())

	router.GET("/r/:hash", shortLinkHandler.GetShortLink)

	{
		v1 := router.Group("/api/v1")
		v1.POST("/links", shortLinkHandler.CreateShortLink)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server Shutdown", "err", err)
	}
	log.Info("Server exiting")
}
