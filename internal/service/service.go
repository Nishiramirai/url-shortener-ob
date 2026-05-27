package service

import (
	"context"
	"errors"
	"url-shortener-ob/internal/repository"
)

type ShortenResult struct {
	Token string
	IsNew bool
}

type Service struct {
	repo repository.Repository
}

var (
	ErrNotFound = errors.New("not found")
)

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ShortenURL(ctx context.Context, originalURL string) (ShortenResult, error) {
	// TODO:
	// 1. Проверить через s.repo.GetByOriginal, нет ли уже такого URL.
	// 2. Если есть — вернуть старую короткую ссылку (или ошибку ErrAlreadyExists, смотря как решишь).
	// 3. Если нет — сгенерировать 10 символов.
	// 4. Сохранить через s.repo.Save.
	return ShortenResult{"abcdefg", true}, nil
}

func (s *Service) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	// TODO: Вызвать s.repo.GetByShort. Если в базе пусто, вернуть ErrNotFound.
	return "https://google.com", nil
}
