package storage

import (
	"context"
	"fmt"

	"archive-deacidification/internal/domain"
)

func (s *Store) VerifyChain(ctx context.Context, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data.Batches[id]; !ok {
		return false, domain.ErrNotFound
	}
	if e := s.verifyChainLocked(id); e != nil {
		return false, e
	}
	return true, nil
}

func (s *Store) verifyChainLocked(id string) error {
	ev := s.eventsLocked(id)
	prev := ""
	for index, x := range ev {
		if x.Seq != int64(index+1) || x.PrevDigest != prev || x.Digest != eventDigest(x) {
			return fmt.Errorf("%w: 批次 %s 的事件 %d 不连续或摘要不符", domain.ErrAuditChain, id, x.Seq)
		}
		prev = x.Digest
	}
	return nil
}
