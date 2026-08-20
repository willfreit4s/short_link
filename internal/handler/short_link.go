// Package handler defines the HTTP handlers for managing short links.
package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/willfreit4s/short_link/configs"
	"github.com/willfreit4s/short_link/internal/usecase"
	usecasedto "github.com/willfreit4s/short_link/internal/usecase/dto"
)

type ShortLinkResponse struct {
	ShortURL string `json:"short_url"`
}

type ShortLinkRequest struct {
	URL string `json:"url"`
}

type ShortLinkHandler struct {
	usecase usecase.ShortLinkUseCase
}

func NewShortLinkHandler(usecase usecase.ShortLinkUseCase) *ShortLinkHandler {
	return &ShortLinkHandler{
		usecase: usecase,
	}
}

func (h *ShortLinkHandler) CreateShortLink(c *gin.Context) {
	var req ShortLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"message": err.Error(),
		})
		return
	}

	input := usecasedto.CreateShortLinkInput{
		OriginalURL: req.URL,
	}

	shortLink, err := h.usecase.CreateShortLink(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create short link",
			"message": err.Error(),
		})
		return
	}

	var httpOrHttps string = "http"
	env := configs.Instance.GetEnv()

	if env != "local" {
		httpOrHttps = "https"
	}

	shortLinkResponse := ShortLinkResponse{
		ShortURL: fmt.Sprintf("%s://%s/r/%s", httpOrHttps, c.Request.Host, shortLink.Hash),
	}

	c.JSON(http.StatusCreated, shortLinkResponse)
}

func (h *ShortLinkHandler) GetShortLink(c *gin.Context) {
	hash := c.Param("hash")

	shortLink, err := h.usecase.GetShortLink(c.Request.Context(), usecasedto.GetShortLinkInput{Hash: hash})
	if err != nil {
		if errors.Is(err, usecase.ErrShortLinkNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "short link not found",
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, shortLink.OriginalURL)
}
