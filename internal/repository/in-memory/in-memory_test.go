package inmemory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/RomanKovalev007/url-shortner/internal/domain"
	inmemory "github.com/RomanKovalev007/url-shortner/internal/repository/in-memory"
)

func TestSaveAlias(t *testing.T) {
	ctx := context.Background()

	t.Run("saves new url and returns it", func(t *testing.T) {
		s := inmemory.New()

		got, _, err := s.SaveAlias(ctx, "abc123_XYZ", "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Alias != "abc123_XYZ" || got.Original != "https://example.com" {
			t.Errorf("unexpected result: %+v", got)
		}
	})

	t.Run("returns existing record on duplicate original", func(t *testing.T) {
		s := inmemory.New()
		first, _, err := s.SaveAlias(ctx, "firstAlias_", "https://example.com")
		if err != nil {
			t.Fatalf("first save: %v", err)
		}

		got, _, err := s.SaveAlias(ctx, "otherAlias_", "https://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Alias != first.Alias {
			t.Errorf("got alias %q, want existing %q", got.Alias, first.Alias)
		}
	})

	t.Run("returns ErrAliasAlreadyExists on alias collision", func(t *testing.T) {
		s := inmemory.New()
		if _, _, err := s.SaveAlias(ctx, "collidAlias", "https://first.com"); err != nil {
			t.Fatalf("first save: %v", err)
		}

		_, _, err := s.SaveAlias(ctx, "collidAlias", "https://second.com")
		if !errors.Is(err, domain.ErrAliasAlreadyExists) {
			t.Errorf("got %v, want ErrAliasAlreadyExists", err)
		}
	})
}

func TestGetByAlias(t *testing.T) {
	ctx := context.Background()

	t.Run("returns url for known alias", func(t *testing.T) {
		s := inmemory.New()
		if _, _, err := s.SaveAlias(ctx, "knownAlias_", "https://example.com"); err != nil {
			t.Fatalf("save: %v", err)
		}

		got, err := s.GetByAlias(ctx, "knownAlias_")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Original != "https://example.com" {
			t.Errorf("got %q, want %q", got.Original, "https://example.com")
		}
	})

	t.Run("returns ErrNotFound for unknown alias", func(t *testing.T) {
		s := inmemory.New()
		_, err := s.GetByAlias(ctx, "doesNotExist")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestConcurrentSaveAlias(t *testing.T) {
	const n = 100
	s := inmemory.New()
	ctx := context.Background()

	type entry struct {
		alias    string
		original string
	}
	entries := make([]entry, n)
	for i := range n {
		entries[i] = entry{
			alias:    fmt.Sprintf("alias%04d", i),
			original: fmt.Sprintf("https://example.com/%d", i),
		}
	}

	var wg sync.WaitGroup
	for _, e := range entries {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, _ = s.SaveAlias(ctx, e.alias, e.original)
		}()
	}
	wg.Wait()

	for _, e := range entries {
		got, err := s.GetByAlias(ctx, e.alias)
		if err != nil {
			t.Errorf("alias %q not found after concurrent save: %v", e.alias, err)
			continue
		}
		if got.Original != e.original {
			t.Errorf("alias %q: got original %q, want %q", e.alias, got.Original, e.original)
		}
	}
}
