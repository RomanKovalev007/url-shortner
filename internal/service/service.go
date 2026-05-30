package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/RomanKovalev007/url-shortner/internal/domain"
)

type UrlRepo interface {
	SaveAlias(ctx context.Context, alias, original string) (domain.URL, bool, error)
	GetByAlias(ctx context.Context, alias string) (domain.URL, error)
}

type Service struct {
	urlRepo UrlRepo
}

func NewService(urlRepo UrlRepo) *Service {
	return &Service{urlRepo: urlRepo}
}

const maxRetries = 5

func (s *Service) Shorten(ctx context.Context, original string) (domain.URL, bool, error) {
	for range maxRetries {
		alias := generateAlias()
		res, created, err := s.urlRepo.SaveAlias(ctx, alias, original)
		if err == nil {
			slog.InfoContext(ctx, "url shortened", "original", original, "alias", res.Alias, "created", created)
			return res, created, nil
		}
		if !errors.Is(err, domain.ErrAliasAlreadyExists) {
			return domain.URL{}, false, err
		}
	}
	return domain.URL{}, false, errors.New("failed to generate unique alias")
}

func (s *Service) GetOriginal(ctx context.Context, alias string) (domain.URL, error) {
	res, err := s.urlRepo.GetByAlias(ctx, alias)
	if err != nil {
		return domain.URL{}, err
	}
	slog.InfoContext(ctx, "alias resolved", "alias", alias, "original", res.Original)
	return res, nil
}


