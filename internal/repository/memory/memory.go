package memory

import (
	"context"
	"sync"
	"url-shortener-ob/internal/repository"
)

type Storage struct {
	mu         sync.RWMutex
	tokenToURL map[string]string
	urlToToken map[string]string
	capacity   int
}

func New(capacity int) *Storage {
	return &Storage{
		tokenToURL: make(map[string]string),
		urlToToken: make(map[string]string),
		capacity:   capacity,
	}
}

// Сохраняет ссылку и ссылающийся на нее токен. Если ссылка уже существует, возвращает ее токен,
// если переданный токен уже существует, возвращает ошибку
func (s *Storage) GetOrCreate(ctx context.Context, token, url string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Если переданный url уже есть, возвращается его токен
	if oldToken, ok := s.urlToToken[url]; ok {
		return oldToken, false, nil
	}

	// Проверка на колизию с существующим токеном
	if _, ok := s.tokenToURL[token]; ok {
		return "", false, repository.ErrTokenExists
	}

	// Защита от OOM
	if len(s.tokenToURL) >= s.capacity {
		return "", false, repository.ErrStorageFull
	}

	s.tokenToURL[token] = url
	s.urlToToken[url] = token

	return token, true, nil
}

func (s *Storage) GetURL(ctx context.Context, token string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.tokenToURL[token]
	if !ok {
		return "", repository.ErrNotFound
	}

	return url, nil
}
