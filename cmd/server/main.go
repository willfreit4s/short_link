package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/willfreit4s/short_link/internal/handler"
	"github.com/willfreit4s/short_link/internal/usecase"
)

func main() {
	router := gin.Default()

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
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
