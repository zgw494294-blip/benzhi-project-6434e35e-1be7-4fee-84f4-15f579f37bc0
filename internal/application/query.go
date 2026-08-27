package application

import (
	"archive-deacidification/internal/storage"
	"context"
)

func (s *Service) AuditTrail(ctx context.Context, id string) ([]storage.AuditEvent, error) {
	return s.repo.Events(ctx, id)
}
