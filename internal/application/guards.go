package application

import (
	"archive-deacidification/internal/domain"
	"context"
	"fmt"
)

func ensureBatchID(id string) error {
	if id == "" {
		return domain.ErrNotFound
	}
	return nil
}
func ensureVersion(v int) error {
	if v < 0 {
		return fmt.Errorf("版本不能为负数")
	}
	return nil
}
func ensureReviewer(name string) error { return domain.ValidateActor(name) }
func (s *Service) CanModify(ctx context.Context, id string) (bool, error) {
	b, e := s.Snapshot(ctx, id)
	if e != nil {
		return false, e
	}
	return domain.CanEdit(b.Status), nil
}
