package memory

import (
	"sync"
	"url-shortener-ob/internal/repository"
)

type Storage struct {
	mu         sync.RWMutex
	tokenToURL map[string]string
	urlToToken map[string]string
}

func New() *Storage {

	// TODO: удалить ёто
	urlMap := make(map[string]string)
	urlMap["aaaaaaaaaa"] = "https://amazon.com"
	urlMap["bbbbbbbbbb"] = "https://bbc.com"
	urlMap["cccccccccc"] = "https://cian.com"
	urlMap["dddddddddd"] = "https://docs.com"
	urlMap["eeeeeeeeee"] = "https://ebay.com"
	//

	return &Storage{
		tokenToURL: urlMap,
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
		return "", repository.ErrNotFound
	}

	return url, nil
}

func (s *Storage) GetToken(url string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, ok := s.urlToToken[url]
	if !ok {
		return "", repository.ErrNotFound
	}

	return token, nil
}
