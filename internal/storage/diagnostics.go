package storage

import (
	"context"
	"fmt"
	"time"
)

type Diagnostic struct {
	Name      string    `json:"name"`
	Healthy   bool      `json:"healthy"`
	Detail    string    `json:"detail"`
	CheckedAt time.Time `json:"checkedAt"`
}

func (s *Store) Diagnostics(ctx context.Context) []Diagnostic {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	return []Diagnostic{{Name: "batch-index", Healthy: s.data.Batches != nil, Detail: fmt.Sprintf("%d 个批次", len(s.data.Batches)), CheckedAt: now}, {Name: "audit-journal", Healthy: s.data.Events != nil, Detail: fmt.Sprintf("%d 条事件", len(s.data.Events)), CheckedAt: now}, {Name: "idempotency", Healthy: s.data.Idempotency != nil, Detail: fmt.Sprintf("%d 个键", len(s.data.Idempotency)), CheckedAt: now}}
}
func (s *Store) Healthy(ctx context.Context) bool {
	for _, d := range s.Diagnostics(ctx) {
		if !d.Healthy {
			return false
		}
	}
	return true
}
func (s *Store) LastEvent(ctx context.Context, id string) (AuditEvent, bool) {
	ev, _ := s.Events(ctx, id)
	if len(ev) == 0 {
		return AuditEvent{}, false
	}
	return ev[len(ev)-1], true
}
func (s *Store) EventCount(ctx context.Context, id string) int {
	ev, _ := s.Events(ctx, id)
	return len(ev)
}
func (s *Store) BatchVersion(ctx context.Context, id string) (int, error) {
	b, e := s.Get(ctx, id)
	return b.Version, e
}
