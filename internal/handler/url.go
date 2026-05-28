package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener-ob/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

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
		if errors.Is(err, service.ErrStorageFull) {
			h.logger.Error("storage limit reached",
				slog.Any("err", err),
			)
			c.JSON(http.StatusInsufficientStorage, gin.H{"error": "storage limit reached"})
			return
		}
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

	if !isValidToken(shortKey) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token format, must be 10 characters alphanumeric or underscore"})
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
			slog.String("token", shortKey),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// В задании написано, что GET должен возвращать оригинальный URL, не совсем понятно что именно имеется ввиду,
	// так как сокращатели ссылок обычно делают redirect. Решил сделать все же redirect
	c.Redirect(http.StatusTemporaryRedirect, originalURL)

	// Но в случае чего всегда можно убрать строку выше и расскоменить строку ниже, будет дословое выполнение задания
	// c.JSON(http.StatusOK, gin.H{"url": originalURL})
}

func isValidToken(token string) bool {
	if len(token) != 10 {
		return false
	}

	for i := 0; i < len(token); i++ {
		b := token[i]
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || (b == '_')) {
			return false
		}
	}
	return true
}
