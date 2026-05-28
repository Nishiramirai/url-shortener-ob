package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"url-shortener-ob/internal/repository"
)

var ErrURLNotFound = errors.New("shortened url not found")

type Repository interface {
	GetOrCreate(ctx context.Context, token, url string) (string, bool, error)
	GetURL(ctx context.Context, token string) (string, error)
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
	const maxRetries = 5

	for i := 0; i < maxRetries; i++ {
		token, err := generateToken()
		if err != nil {
			return ShortenResult{}, fmt.Errorf("failed to generate token: %w", err)
		}

		actualToken, isNew, err := s.repo.GetOrCreate(ctx, token, originalURL)
		if err != nil {
			if errors.Is(err, repository.ErrTokenExists) {
				continue
			}
			return ShortenResult{}, fmt.Errorf("repository: %v", err)
		}

		return ShortenResult{actualToken, isNew}, nil
	}

	return ShortenResult{}, errors.New("failed to generate unique token after retries")
}

func (s *Service) GetOriginalURL(ctx context.Context, token string) (string, error) {
	url, err := s.repo.GetURL(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrURLNotFound
		}
		return "", err
	}
	return url, nil
}

func generateToken() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	const tokenLength = 10

	bytes := make([]byte, tokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	lenAlphabet := len(alphabet)
	for i, b := range bytes {
		bytes[i] = alphabet[b%byte(lenAlphabet)]
	}

	return string(bytes), nil
}
