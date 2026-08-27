package storage

import "context"

func (s *Store) WithLock(ctx context.Context, fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn()
}
func (s *Store) Transactional(ctx context.Context, fn func() error) error { return s.WithLock(ctx, fn) }
