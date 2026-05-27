package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type URLService interface {
	ShortenURL() (string, error)
	GetOriginalURL() (string, error)
}

type Handler struct {
	service URLService
}

func New(service URLService) *Handler {
	return &Handler{service: service}
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body or invalid URL format"})
		return
	}

	url, _ := h.service.ShortenURL()
	c.JSON(http.StatusCreated, url)
}

func (h *Handler) Resolve(c *gin.Context) {
	const shortLen = 10

	shortKey := c.Param("short")
	if len(shortKey) != shortLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key length"})
		return
	}

	url, _ := h.service.ShortenURL()
	c.JSON(http.StatusOK, gin.H{shortKey: url})
	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "url",
	// })
}
