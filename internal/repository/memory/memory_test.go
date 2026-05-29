package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"url-shortener-ob/internal/repository"
	"url-shortener-ob/internal/repository/memory"
)

func TestStorage_GetOrCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("success_create_new", func(t *testing.T) {
		storage := memory.New(10)

		token, isNew, err := storage.GetOrCreate(ctx, "abcdefghij", "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isNew {
			t.Error("expected isNew to be true for a new URL")
		}
		if token != "abcdefghij" {
			t.Errorf("expected token 'abcdefghij', got '%s'", token)
		}
	})

	t.Run("success_get_existing_url", func(t *testing.T) {
		storage := memory.New(10)
		// Сохраняем первую пару
		_, _, _ = storage.GetOrCreate(ctx, "abcdefghij", "https://example.com")

		// Пытаемся сохранить тот же URL, но передаем другой 10-значный токен
		token, isNew, err := storage.GetOrCreate(ctx, "xyz1234567", "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isNew {
			t.Error("expected isNew to be false for an existing URL")
		}
		// Должен вернуться первый сохраненный токен
		if token != "abcdefghij" {
			t.Errorf("expected existing token 'abcdefghij', got '%s'", token)
		}
	})

	t.Run("error_token_collision", func(t *testing.T) {
		storage := memory.New(10)
		_, _, _ = storage.GetOrCreate(ctx, "abcdefghij", "https://example.com")

		// Передаем тот же 10-значный токен для другого url
		_, _, err := storage.GetOrCreate(ctx, "abcdefghij", "https://other-url.com")
		if !errors.Is(err, repository.ErrTokenExists) {
			t.Errorf("expected error %v, got %v", repository.ErrTokenExists, err)
		}
	})

	t.Run("error_storage_full", func(t *testing.T) {
		storage := memory.New(1) // Лимит — 1 элемент
		_, _, _ = storage.GetOrCreate(ctx, "abcdefghij", "https://example1.com")

		// Второй токен упирается в капасити
		_, _, err := storage.GetOrCreate(ctx, "xyz1234567", "https://example2.com")
		if !errors.Is(err, repository.ErrStorageFull) {
			t.Errorf("expected error %v, got %v", repository.ErrStorageFull, err)
		}
	})
}

func TestStorage_GetURL(t *testing.T) {
	ctx := context.Background()

	t.Run("success_found", func(t *testing.T) {
		storage := memory.New(10)
		targetURL := "https://example.com"
		_, _, _ = storage.GetOrCreate(ctx, "abcdefghij", targetURL)

		url, err := storage.GetURL(ctx, "abcdefghij")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if url != targetURL {
			t.Errorf("expected url '%s', got '%s'", targetURL, url)
		}
	})

	t.Run("error_not_found", func(t *testing.T) {
		storage := memory.New(10)

		_, err := storage.GetURL(ctx, "notfound12")
		if !errors.Is(err, repository.ErrNotFound) {
			t.Errorf("expected error %v, got %v", repository.ErrNotFound, err)
		}
	})
}

func TestStorage_ConcurrencySafe(t *testing.T) {
	ctx := context.Background()
	storage := memory.New(1000)
	var wg sync.WaitGroup

	// Одновременно пишем из 100 горутин
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			token := fmt.Sprintf("tok%07d", n)
			url := fmt.Sprintf("https://url_%d.com", n)
			_, _, _ = storage.GetOrCreate(ctx, token, url)
		}(i)
	}

	// И одновременно читаем из 100 горутин
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			token := fmt.Sprintf("tok%07d", n)
			_, _ = storage.GetURL(ctx, token)
		}(i)
	}

	wg.Wait()
}
