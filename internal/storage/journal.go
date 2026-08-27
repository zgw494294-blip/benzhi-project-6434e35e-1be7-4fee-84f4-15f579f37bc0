package storage

import (
	"archive-deacidification/internal/domain"
	"context"
	"sort"
)

func (s *Store) LatestSequence(ctx context.Context) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return int64(len(s.data.Events))
}
func (s *Store) EventsSince(ctx context.Context, seq int64) []AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []AuditEvent{}
	for _, e := range s.data.Events {
		if e.Seq > seq {
			out = append(out, e)
		}
	}
	return out
}
func (s *Store) BatchIDs(ctx context.Context) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.data.Batches))
	for id := range s.data.Batches {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
func (s *Store) Reset(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = fileData{Batches: map[string]domain.RestorationBatch{}, Idempotency: map[string][]byte{}}
	return s.persist()
}
