package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"url-shortener-ob/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	PgErrUniqueViolation = "23505"
)

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) GetOrCreate(ctx context.Context, token, url string) (string, bool, error) {
	q := `
			WITH inserted AS (
				INSERT INTO urls (token, original_url)
				VALUES ($1, $2)
				ON CONFLICT (original_url) DO NOTHING
				RETURNING token
			)
			SELECT token, true AS is_new FROM inserted
			UNION ALL
			SELECT token, false AS is_new FROM urls WHERE original_url = $2
			LIMIT 1;
		  `

	var oldToken string
	var isNew bool

	err := s.db.QueryRowContext(ctx, q, token, url).Scan(&oldToken, &isNew)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == PgErrUniqueViolation {
			if pgErr.ConstraintName == "idx_urls_token" || pgErr.ConstraintName == "urls_token_key" {
				return "", false, repository.ErrTokenExists
			}
		}

		return "", false, fmt.Errorf("create url query: %w", err)
	}

	return oldToken, isNew, nil
}

func (s *Storage) GetURL(ctx context.Context, token string) (string, error) {
	var url string
	q := `SELECT original_url FROM urls WHERE token = $1`
	err := s.db.QueryRowContext(ctx, q, token).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository.ErrNotFound
		}
		return "", fmt.Errorf("get url query: %w", err)
	}

	return url, nil
}
