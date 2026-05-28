package service

import (
	"context"
	"errors"
	"fmt"
	"url-shortener-ob/internal/repository"
)

type Repository interface {
	GetOrCreate(ctx context.Context, token, url string) (string, bool, error)
	GetURL(ctx context.Context, token string) (string, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

type ShortenResult struct {
	Token string
	IsNew bool
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
			if errors.Is(err, repository.ErrStorageFull) {
				return ShortenResult{}, ErrStorageFull
			}
			return ShortenResult{}, fmt.Errorf("service shorten: %w", err)
		}

		return ShortenResult{actualToken, isNew}, nil
	}

	return ShortenResult{}, ErrMaxRetriesReached
}

func (s *Service) GetOriginalURL(ctx context.Context, token string) (string, error) {
	url, err := s.repo.GetURL(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrURLNotFound
		}
		return "", fmt.Errorf("service resolve: %w", err)
	}
	return url, nil
}
