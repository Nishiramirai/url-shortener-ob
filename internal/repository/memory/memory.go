package memory

import (
	"errors"
	"sync"
)

type Storage struct {
	mu         sync.RWMutex
	tokenToURL map[string]string
	urlToToken map[string]string
}

var ErrNotFound = errors.New("url not found")

func New() *Storage {
	return &Storage{
		tokenToURL: make(map[string]string),
		urlToToken: make(map[string]string),
	}
}

func (s *Storage) GetOrCreate(token, url string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if oldToken, ok := s.urlToToken[url]; ok {
		return oldToken, false, nil
	}
	s.tokenToURL[token] = url
	s.urlToToken[url] = token

	return token, true, nil
}

func (s *Storage) GetURL(token string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.tokenToURL[token]
	if !ok {
		return "", ErrNotFound
	}

	return url, nil
}

func (s *Storage) GetToken(url string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, ok := s.urlToToken[url]
	if !ok {
		return "", ErrNotFound
	}

	return token, nil
}
