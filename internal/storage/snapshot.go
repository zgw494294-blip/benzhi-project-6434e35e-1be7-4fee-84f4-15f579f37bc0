package storage

import (
	"archive-deacidification/internal/domain"
	"context"
)

func (s *Store) Snapshot(ctx context.Context, id string) (domain.RestorationBatch, error) {
	return s.Get(ctx, id)
}
