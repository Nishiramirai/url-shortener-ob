package service

import (
	"context"
	"errors"
	"math/rand"
	"url-shortener-ob/internal/repository"
)

var ErrURLNotFound = errors.New("shortened url not found")

type Repository interface {
	GetOrCreate(token, url string) (string, bool, error)
	GetURL(token string) (string, error)
}

type ShortenResult struct {
	Token string
	IsNew bool
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ShortenURL(ctx context.Context, originalURL string) (ShortenResult, error) {
	token := generateToken()

	actualToken, isNew, err := s.repo.GetOrCreate(token, originalURL)
	if err != nil {
		return ShortenResult{}, err
	}

	return ShortenResult{actualToken, isNew}, nil
}

func (s *Service) GetOriginalURL(ctx context.Context, token string) (string, error) {
	url, err := s.repo.GetURL(token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrURLNotFound
		}
		return "", err
	}
	return url, nil
}

func generateToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
