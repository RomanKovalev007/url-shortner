package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/RomanKovalev007/url-shortner/internal/domain"
	"github.com/google/uuid"
)

type entry struct {
	url       domain.URL
	expiresAt time.Time
}

func (e entry) isExpired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}

type Store struct {
	mu         sync.RWMutex
	byAlias    map[string]entry
	byOriginal map[string]string
	ttl        time.Duration
}

func New(ttl time.Duration) *Store {
	return &Store{
		byAlias:    make(map[string]entry),
		byOriginal: make(map[string]string),
		ttl:        ttl,
	}
}

func (s *Store) StartCleanup(ctx context.Context, interval time.Duration) {
	if s.ttl == 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.deleteExpired()
			}
		}
	}()
}

func (s *Store) deleteExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for alias, e := range s.byAlias {
		if e.isExpired() {
			delete(s.byOriginal, e.url.Original)
			delete(s.byAlias, alias)
		}
	}
}

func (s *Store) SaveAlias(_ context.Context, alias, original string) (domain.URL, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byAlias[alias]; ok {
		return domain.URL{}, false, domain.ErrAliasAlreadyExists
	}

	if existingAlias, ok := s.byOriginal[original]; ok {
		e := s.byAlias[existingAlias]
		if s.ttl > 0 {
			e.expiresAt = time.Now().Add(s.ttl)
			s.byAlias[existingAlias] = e
		}
		return e.url, false, nil
	}

	url := domain.URL{
		ID:        uuid.New(),
		Alias:     alias,
		Original:  original,
		CreatedAt: time.Now(),
	}

	var exp time.Time
	if s.ttl > 0 {
		exp = time.Now().Add(s.ttl)
	}

	s.byAlias[alias] = entry{url: url, expiresAt: exp}
	s.byOriginal[original] = alias

	return url, true, nil
}

func (s *Store) GetByAlias(_ context.Context, alias string) (domain.URL, error) {
	if s.ttl == 0 {
		s.mu.RLock()
		defer s.mu.RUnlock()

		e, ok := s.byAlias[alias]
		if !ok {
			return domain.URL{}, domain.ErrNotFound
		}
		return e.url, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.byAlias[alias]
	if !ok {
		return domain.URL{}, domain.ErrNotFound
	}

	e.expiresAt = time.Now().Add(s.ttl)
	s.byAlias[alias] = e

	return e.url, nil
}
