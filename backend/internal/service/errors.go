package service

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrEmptyName    = errors.New("name is empty")
	ErrOrgArchived  = errors.New("org is archived")
	ErrTeamArchived = errors.New("team is archived")
	ErrRoleArchived = errors.New("role is archived")
)

// rollback runs undo after a SpiceDB write fails, so postgres doesn't keep a
// row that SpiceDB never linked. It ignores ctx cancellation, since that may be
// why the write failed.
func rollback(ctx context.Context, err error, undo func(context.Context) error) error {
	if undoErr := undo(context.WithoutCancel(ctx)); undoErr != nil {
		return errors.Join(err, fmt.Errorf("rollback: %w", undoErr))
	}
	return err
}
