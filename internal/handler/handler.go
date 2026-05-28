package handler

import (
	"context"
	"log/slog"
	"url-shortener-ob/internal/service"

	"github.com/gin-gonic/gin"
)

type URLService interface {
	ShortenURL(ctx context.Context, originalURL string) (service.ShortenResult, error)
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
}

type Handler struct {
	service URLService
	logger  *slog.Logger
}

func New(service URLService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/", h.Shorten)
	r.GET("/:short", h.Resolve)
}
