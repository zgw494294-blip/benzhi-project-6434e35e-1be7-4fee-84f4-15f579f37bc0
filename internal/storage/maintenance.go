package storage

import (
	"context"
	"time"
)

type MaintenanceStats struct {
	Batches, Events, Idempotency int
	CheckedAt                    time.Time
}

func (s *Store) Maintenance(ctx context.Context) MaintenanceStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return MaintenanceStats{Batches: len(s.data.Batches), Events: len(s.data.Events), Idempotency: len(s.data.Idempotency), CheckedAt: time.Now().UTC()}
}
func (s *Store) Compact(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.persist()
}
func (s *Store) RemoveIdempotency(ctx context.Context, batch, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data.Idempotency, batch+"|"+key)
	return s.persist()
}
func (s *Store) HasBatch(ctx context.Context, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data.Batches[id]
	return ok
}
func (s *Store) HasEvent(ctx context.Context, seq int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return seq > 0 && seq <= int64(len(s.data.Events))
}
