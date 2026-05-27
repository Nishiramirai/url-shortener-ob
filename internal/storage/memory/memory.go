package memory

import "sync"

type Storage struct {
	mu sync.RWMutex
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Save() error {
	return nil
}

func (s *Storage) GetByShort() (string, error) {
	return "https://google.com", nil
}

func (s *Storage) GetByOriginal() (string, error) {
	return "short_stub", nil
}
