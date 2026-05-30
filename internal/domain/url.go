package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("url not found")
	ErrAliasAlreadyExists = errors.New("alias already exists")
)

type URL struct {
	ID        uuid.UUID
	Alias     string
	Original  string
	CreatedAt time.Time
}
