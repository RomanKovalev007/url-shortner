package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/RomanKovalev007/url-shortner/internal/domain"
	"github.com/google/uuid"
)

type Store struct {
	mu         sync.RWMutex
	byAlias    map[string]domain.URL
	byOriginal map[string]string
}

func New() *Store {
	return &Store{
		byAlias:    make(map[string]domain.URL),
		byOriginal: make(map[string]string),
	}
}

func (s *Store) SaveAlias(_ context.Context, alias, original string) (domain.URL, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byAlias[alias]; ok {
		return domain.URL{}, false, domain.ErrAliasAlreadyExists
	}

	if existingAlias, ok := s.byOriginal[original]; ok {
		return s.byAlias[existingAlias], false, nil
	}

	url := domain.URL{
		ID:        uuid.New(),
		Alias:     alias,
		Original:  original,
		CreatedAt: time.Now(),
	}
	s.byAlias[alias] = url
	s.byOriginal[original] = alias

	return url, true, nil
}

func (s *Store) GetByAlias(_ context.Context, alias string) (domain.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.byAlias[alias]
	if !ok {
		return domain.URL{}, domain.ErrNotFound
	}

	return url, nil
}
