package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/willfreit4s/short_link/configs"
	"github.com/willfreit4s/short_link/internal/usecase"
	usecasedto "github.com/willfreit4s/short_link/internal/usecase/dto"
)

type stubShortLinkUseCase struct {
	createFn func(ctx context.Context, input usecasedto.CreateShortLinkInput) (usecasedto.CreateShortLinkOutput, error)
	getFn    func(ctx context.Context, input usecasedto.GetShortLinkInput) (usecasedto.GetShortLinkOutput, error)
}

func (s *stubShortLinkUseCase) CreateShortLink(ctx context.Context, input usecasedto.CreateShortLinkInput) (usecasedto.CreateShortLinkOutput, error) {
	if s.createFn != nil {
		return s.createFn(ctx, input)
	}
	return usecasedto.CreateShortLinkOutput{}, nil
}

func (s *stubShortLinkUseCase) GetShortLink(ctx context.Context, input usecasedto.GetShortLinkInput) (usecasedto.GetShortLinkOutput, error) {
	if s.getFn != nil {
		return s.getFn(ctx, input)
	}
	return usecasedto.GetShortLinkOutput{}, nil
}

func TestShortLinkHandler_GetHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	h := NewShortLinkHandler(&stubShortLinkUseCase{})

	h.GetHealth(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestShortLinkHandler_CreateShortLink(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldInstance := configs.Instance
	configs.Instance = &configs.Config{Environment: "local"}
	t.Cleanup(func() {
		configs.Instance = oldInstance
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/links", strings.NewReader(`{"url":"example.com"}`))
		req.Host = "example.com"
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h := NewShortLinkHandler(&stubShortLinkUseCase{
			createFn: func(ctx context.Context, input usecasedto.CreateShortLinkInput) (usecasedto.CreateShortLinkOutput, error) {
				require.Equal(t, "example.com", input.OriginalURL)
				return usecasedto.CreateShortLinkOutput{Hash: "abc123"}, nil
			},
		})

		h.CreateShortLink(c)

		require.Equal(t, http.StatusCreated, w.Code)
		require.JSONEq(t, `{"short_url":"http://example.com/r/abc123"}`, w.Body.String())
	})

	t.Run("invalid payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/links", strings.NewReader(`{"url":`))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h := NewShortLinkHandler(&stubShortLinkUseCase{})
		h.CreateShortLink(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "Invalid request payload")
	})

	t.Run("usecase error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/links", strings.NewReader(`{"url":"example.com"}`))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		h := NewShortLinkHandler(&stubShortLinkUseCase{
			createFn: func(ctx context.Context, input usecasedto.CreateShortLinkInput) (usecasedto.CreateShortLinkOutput, error) {
				return usecasedto.CreateShortLinkOutput{}, errors.New("repository failed")
			},
		})

		h.CreateShortLink(c)

		require.Equal(t, http.StatusBadRequest, w.Code)
		require.Contains(t, w.Body.String(), "Failed to create short link")
		require.Contains(t, w.Body.String(), "repository failed")
	})
}

func TestShortLinkHandler_GetShortLink(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success redirect", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/abc123", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "hash", Value: "abc123"}}

		h := NewShortLinkHandler(&stubShortLinkUseCase{
			getFn: func(ctx context.Context, input usecasedto.GetShortLinkInput) (usecasedto.GetShortLinkOutput, error) {
				require.Equal(t, "abc123", input.Hash)
				return usecasedto.GetShortLinkOutput{OriginalURL: "https://example.com"}, nil
			},
		})

		h.GetShortLink(c)

		require.Equal(t, http.StatusFound, w.Code)
		require.Equal(t, "https://example.com", w.Header().Get("Location"))
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/missing", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "hash", Value: "missing"}}

		h := NewShortLinkHandler(&stubShortLinkUseCase{
			getFn: func(ctx context.Context, input usecasedto.GetShortLinkInput) (usecasedto.GetShortLinkOutput, error) {
				return usecasedto.GetShortLinkOutput{}, usecase.ErrShortLinkNotFound
			},
		})

		h.GetShortLink(c)

		require.Equal(t, http.StatusNotFound, w.Code)
		require.JSONEq(t, `{"error":"short link not found"}`, w.Body.String())
	})

	t.Run("generic error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/r/abc123", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "hash", Value: "abc123"}}

		h := NewShortLinkHandler(&stubShortLinkUseCase{
			getFn: func(ctx context.Context, input usecasedto.GetShortLinkInput) (usecasedto.GetShortLinkOutput, error) {
				return usecasedto.GetShortLinkOutput{}, errors.New("repository failure")
			},
		})

		h.GetShortLink(c)

		require.Equal(t, http.StatusNotFound, w.Code)
		require.Contains(t, w.Body.String(), "repository failure")
	})
}
