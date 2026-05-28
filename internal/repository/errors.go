package repository

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrTokenExists = errors.New("token already exists")
	ErrStorageFull = errors.New("storage capacity reached")
)
