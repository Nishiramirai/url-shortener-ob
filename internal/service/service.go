package service

import (
	"errors"
	"url-shortener-ob/internal/storage"
)

var (
	ErrNotFound      = errors.New("url not found")
	ErrAlreadyExists = errors.New("url already exists")
)

type Service struct {
	repo storage.Repository
}

func New(repo storage.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ShortenURL() (string, error) {
	// TODO:
	// 1. Проверить через s.repo.GetByOriginal, нет ли уже такого URL.
	// 2. Если есть — вернуть старую короткую ссылку (или ошибку ErrAlreadyExists, смотря как решишь).
	// 3. Если нет — сгенерировать 10 символов.
	// 4. Сохранить через s.repo.Save.
	return "abcdefg", nil
}

func (s *Service) GetOriginalURL() (string, error) {
	// TODO: Вызвать s.repo.GetByShort. Если в базе пусто, вернуть ErrNotFound.
	return "https://google.com", nil
}
