package main

import (
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

	router.Run()
}
