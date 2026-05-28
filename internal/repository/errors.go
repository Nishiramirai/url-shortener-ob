package repository

import "errors"

var (
	ErrNotFound    = errors.New("url not found")
	ErrTokenExists = errors.New("token already exists")
)
