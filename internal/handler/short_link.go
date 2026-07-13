// Package handler defines the HTTP handlers for managing short links.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/willfreit4s/short_link/internal/usecase"
)

type ShortLinkHandler struct {
	usecase *usecase.ShortLinkUseCase
}

func NewShortLinkHandler(usecase *usecase.ShortLinkUseCase) *ShortLinkHandler {
	return &ShortLinkHandler{
		usecase: usecase,
	}
}

func (h *ShortLinkHandler) CreateShortLink(c *gin.Context) {
	var req struct {
		URL string `json:"url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	shortLink, err := h.usecase.CreateShortLink(c.Request.Context(), req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        shortLink.ID,
		"short_url": fmt.Sprintf("%s/%s", "http://localhost:8080", shortLink.ID),
	})
}

func (h *ShortLinkHandler) GetShortLink(c *gin.Context) {
	hash := c.Param("hash")

	originalURL, err := h.usecase.GetShortLink(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}
