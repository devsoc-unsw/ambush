// Package repository holds the models and errors returned by the postgres and
// spicedb repositories, so the service doesn't depend on either one directly.
package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrgNotFound   = errors.New("org not found")
	ErrTeamNotFound  = errors.New("team not found")
	ErrRoleNotFound  = errors.New("role not found")
	ErrAlreadyMember = errors.New("user already holds this role")
)

type Org struct {
	ID         uuid.UUID
	Name       string
	Slug       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}

func (o Org) Archived() bool { return o.ArchivedAt != nil }

type Team struct {
	ID         uuid.UUID
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}

func (t Team) Archived() bool { return t.ArchivedAt != nil }

type Role struct {
	ID         uuid.UUID
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}

func (r Role) Archived() bool { return r.ArchivedAt != nil }
