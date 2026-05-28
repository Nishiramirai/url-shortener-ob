package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener-ob/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

type ShortenRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func (h *Handler) Shorten(c *gin.Context) {
	var req ShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url format, must be a valid absolute URI"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed JSON body or missing required fields"})
		return
	}

	result, err := h.service.ShortenURL(c.Request.Context(), req.URL)
	if err != nil {
		h.logger.Error("failed to shorten url",
			slog.Any("err", err),
			slog.String("requested_url", req.URL),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if result.IsNew {
		c.JSON(http.StatusCreated, ShortenResponse{Result: result.Token})
	} else {
		c.JSON(http.StatusOK, ShortenResponse{Result: result.Token})
	}
}

func (h *Handler) Resolve(c *gin.Context) {
	shortKey := c.Param("short")

	if len(shortKey) != 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token format, length must be exactly 10 characters"})
		return
	}

	originalURL, err := h.service.GetOriginalURL(c.Request.Context(), shortKey)
	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
			return
		}

		h.logger.Error("failed to resolve url",
			slog.Any("err", err),
			slog.String("requested_url", shortKey),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"original_url": originalURL})
}
