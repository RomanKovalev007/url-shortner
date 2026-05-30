package postgresrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/RomanKovalev007/url-shortner/internal/domain"
)

type Store struct {
	db *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{db: pool}
}

func (s *Store) SaveAlias(ctx context.Context, alias, original string) (domain.URL, bool, error) {
	q := `
		INSERT INTO urls (alias, original)
		VALUES ($1, $2)
		ON CONFLICT (original) DO NOTHING
		RETURNING id, alias, original, created_at
	`

	var url domain.URL
	if err := s.db.QueryRow(ctx, q, alias, original).Scan(
		&url.ID,
		&url.Alias,
		&url.Original,
		&url.CreatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.URL{}, false, domain.ErrAliasAlreadyExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			existing, err := s.getByOriginal(ctx, original)
			return existing, false, err
		}

		return domain.URL{}, false, err
	}

	return url, true, nil
}

func (s *Store) GetByAlias(ctx context.Context, alias string) (domain.URL, error) {
	q := `
		SELECT id, alias, original, created_at
		FROM urls
		WHERE alias = $1
	`

	var url domain.URL
	err := s.db.QueryRow(ctx, q, alias).Scan(
		&url.ID,
		&url.Alias,
		&url.Original,
		&url.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.URL{}, domain.ErrNotFound
		}
		return domain.URL{}, err
	}

	return url, nil
}

func (s *Store) getByOriginal(ctx context.Context, original string) (domain.URL, error) {
	q := `
		SELECT id, alias, original, created_at
		FROM urls
		WHERE original = $1
	`

	var url domain.URL
	err := s.db.QueryRow(ctx, q, original).Scan(
		&url.ID,
		&url.Alias,
		&url.Original,
		&url.CreatedAt,
	)
	if err != nil {
		return domain.URL{}, err
	}

	return url, nil
}
