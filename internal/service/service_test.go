package service_test

import (
	"context"
	"errors"
	"testing"
	"url-shortener-ob/internal/repository"
	"url-shortener-ob/internal/service"
)

type mockRepository struct {
	GetOrCreateFunc func(ctx context.Context, token, url string) (string, bool, error)
	GetURLFunc      func(ctx context.Context, token string) (string, error)
}

func (m *mockRepository) GetOrCreate(ctx context.Context, token, url string) (string, bool, error) {
	return m.GetOrCreateFunc(ctx, token, url)
}

func (m *mockRepository) GetURL(ctx context.Context, token string) (string, error) {
	return m.GetURLFunc(ctx, token)
}

func TestService_ShortenURL(t *testing.T) {
	ctx := context.Background()

	t.Run("success_new_url", func(t *testing.T) {
		repo := &mockRepository{
			GetOrCreateFunc: func(ctx context.Context, token, url string) (string, bool, error) {
				return token, true, nil
			},
		}
		s := service.New(repo)

		res, err := s.ShortenURL(ctx, "https://google.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !res.IsNew {
			t.Error("expected IsNew to be true")
		}
		if len(res.Token) != 10 {
			t.Errorf("expected token length 10, got %d", len(res.Token))
		}
	})

	t.Run("success_existing_url", func(t *testing.T) {
		existingToken := "exist_tkn1"
		repo := &mockRepository{
			GetOrCreateFunc: func(ctx context.Context, token, url string) (string, bool, error) {
				return existingToken, false, nil // Репо вернул старый токен и IsNew = false
			},
		}
		s := service.New(repo)

		res, err := s.ShortenURL(ctx, "https://google.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.IsNew {
			t.Error("expected IsNew to be false")
		}
		if res.Token != existingToken {
			t.Errorf("expected token %s, got %s", existingToken, res.Token)
		}
	})

	t.Run("token_collision_retry_success", func(t *testing.T) {
		calls := 0
		repo := &mockRepository{
			GetOrCreateFunc: func(ctx context.Context, token, url string) (string, bool, error) {
				calls++
				if calls == 1 {
					// Первый раз симулируем коллизию токена в БД
					return "", false, repository.ErrTokenExists
				}
				return token, true, nil // Со второго раза зашло
			},
		}
		s := service.New(repo)

		_, err := s.ShortenURL(ctx, "https://google.com")
		if err != nil {
			t.Fatalf("unexpected error on retry: %v", err)
		}

		if calls != 2 {
			t.Errorf("expected 2 repository calls due to retry, got %d", calls)
		}
	})

	t.Run("storage_full_error", func(t *testing.T) {
		repo := &mockRepository{
			GetOrCreateFunc: func(ctx context.Context, token, url string) (string, bool, error) {
				return "", false, repository.ErrStorageFull
			},
		}
		s := service.New(repo)

		_, err := s.ShortenURL(ctx, "https://google.com")
		if !errors.Is(err, service.ErrStorageFull) {
			t.Errorf("expected service error %v, got %v", service.ErrStorageFull, err)
		}
	})
}

func TestService_GetOriginalURL(t *testing.T) {
	ctx := context.Background()

	t.Run("success_found", func(t *testing.T) {
		expectedURL := "https://yandex.ru"
		repo := &mockRepository{
			GetURLFunc: func(ctx context.Context, token string) (string, error) {
				return expectedURL, nil
			},
		}
		s := service.New(repo)

		url, err := s.GetOriginalURL(ctx, "some_token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if url != expectedURL {
			t.Errorf("expected url %s, got %s", expectedURL, url)
		}
	})

	t.Run("url_not_found", func(t *testing.T) {
		repo := &mockRepository{
			GetURLFunc: func(ctx context.Context, token string) (string, error) {
				return "", repository.ErrNotFound
			},
		}
		s := service.New(repo)

		_, err := s.GetOriginalURL(ctx, "ghostToken")
		if !errors.Is(err, service.ErrURLNotFound) {
			t.Errorf("expected error %v, got %v", service.ErrURLNotFound, err)
		}
	})
}
