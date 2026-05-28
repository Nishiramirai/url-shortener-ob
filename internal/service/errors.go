package service

import "errors"

var (
	ErrURLNotFound       = errors.New("url not found")
	ErrMaxRetriesReached = errors.New("failed to generate unique token after max retries")
	ErrStorageFull       = errors.New("storage capacity limit reached")
)
