package handler_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"url-shortener-ob/internal/handler"
	"url-shortener-ob/internal/service"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockService struct {
	ShortenURLFunc     func(ctx context.Context, originalURL string) (service.ShortenResult, error)
	GetOriginalURLFunc func(ctx context.Context, token string) (string, error)
}

func (m *mockService) ShortenURL(ctx context.Context, originalURL string) (service.ShortenResult, error) {
	return m.ShortenURLFunc(ctx, originalURL)
}

func (m *mockService) GetOriginalURL(ctx context.Context, token string) (string, error) {
	return m.GetOriginalURLFunc(ctx, token)
}

// Тест хендлера создания ссылки (POST)
func TestHandler_Shorten(t *testing.T) {
	// Создаем глухой логгер, чтобы не засорять вывод тестов
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success_created_new_url", func(t *testing.T) {
		mockSvc := &mockService{
			ShortenURLFunc: func(ctx context.Context, originalURL string) (service.ShortenResult, error) {
				return service.ShortenResult{Token: "abcdefghij", IsNew: true}, nil
			},
		}

		r := gin.New()
		h := handler.New(mockSvc, discardLogger)
		r.POST("/shorten", h.Shorten)

		body := `{"url": "https://google.com"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
		expectedBody := `{"result":"abcdefghij"}`
		if strings.TrimSpace(w.Body.String()) != expectedBody {
			t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
		}
	})

	t.Run("success_existing_url", func(t *testing.T) {
		mockSvc := &mockService{
			ShortenURLFunc: func(ctx context.Context, originalURL string) (service.ShortenResult, error) {
				return service.ShortenResult{Token: "abcdefghij", IsNew: false}, nil
			},
		}

		r := gin.New()
		h := handler.New(mockSvc, discardLogger)
		r.POST("/shorten", h.Shorten)

		body := `{"url": "https://google.com"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("error_invalid_url_format", func(t *testing.T) {
		r := gin.New()
		h := handler.New(&mockService{}, discardLogger)
		r.POST("/shorten", h.Shorten)

		// Передаем невалидный URL, триггерим validator.ValidationErrors
		body := `{"url": "not-a-valid-url"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "invalid url format") {
			t.Errorf("unexpected error message: %s", w.Body.String())
		}
	})

	t.Run("error_storage_full", func(t *testing.T) {
		mockSvc := &mockService{
			ShortenURLFunc: func(ctx context.Context, originalURL string) (service.ShortenResult, error) {
				return service.ShortenResult{}, service.ErrStorageFull
			},
		}

		r := gin.New()
		h := handler.New(mockSvc, discardLogger)
		r.POST("/shorten", h.Shorten)

		body := `{"url": "https://google.com"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusInsufficientStorage {
			t.Errorf("expected status 507, got %d", w.Code)
		}
	})
}

// Тест хендлера получения оригинальной ссылки (GET)
func TestHandler_Resolve(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success_redirect", func(t *testing.T) {
		targetURL := "https://yandex.ru"
		mockSvc := &mockService{
			GetOriginalURLFunc: func(ctx context.Context, token string) (string, error) {
				return targetURL, nil
			},
		}

		r := gin.New()
		h := handler.New(mockSvc, discardLogger)
		r.GET("/:short", h.Resolve)

		req := httptest.NewRequest(http.MethodGet, "/abcdefghij", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Errorf("expected status 307, got %d", w.Code)
		}

		// Проверяем, что заголовок Location ведет куда надо
		if w.Header().Get("Location") != targetURL {
			t.Errorf("expected redirect location %s, got %s", targetURL, w.Header().Get("Location"))
		}
	})

	t.Run("error_invalid_token_format", func(t *testing.T) {
		r := gin.New()
		h := handler.New(&mockService{}, discardLogger)
		r.GET("/:short", h.Resolve)

		// Невалидный токен
		req := httptest.NewRequest(http.MethodGet, "/abc-defghi", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "invalid token format") {
			t.Errorf("unexpected error message: %s", w.Body.String())
		}
	})

	t.Run("error_url_not_found", func(t *testing.T) {
		mockSvc := &mockService{
			GetOriginalURLFunc: func(ctx context.Context, token string) (string, error) {
				return "", service.ErrURLNotFound
			},
		}

		r := gin.New()
		h := handler.New(mockSvc, discardLogger)
		r.GET("/:short", h.Resolve)

		req := httptest.NewRequest(http.MethodGet, "/notfound12", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "url not found") {
			t.Errorf("unexpected error message: %s", w.Body.String())
		}
	})
}
