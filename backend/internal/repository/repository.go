// Package repository holds the models and errors returned by the postgres and
// spicedb repositories, so the service doesn't depend on either one directly.
package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRoleNotFound  = errors.New("role not found")
	ErrAlreadyMember = errors.New("user already holds this role")
)

type Role struct {
	ID         uuid.UUID
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}

func (r Role) Archived() bool { return r.ArchivedAt != nil }
