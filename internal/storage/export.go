package storage

import (
	"archive-deacidification/internal/domain"
	"context"
	"encoding/json"
	"fmt"
)

type EvidencePackage struct {
	Batch  domain.RestorationBatch `json:"batch"`
	Events []AuditEvent            `json:"events"`
	Digest string                  `json:"digest"`
}

func (s *Store) ExportEvidence(ctx context.Context, id string) ([]byte, error) {
	b, e := s.Get(ctx, id)
	if e != nil {
		return nil, e
	}
	ev, e := s.Events(ctx, id)
	if e != nil {
		return nil, e
	}
	d, e := domain.EvidenceDigest(&b)
	if e != nil {
		return nil, e
	}
	return json.Marshal(EvidencePackage{Batch: b, Events: ev, Digest: d})
}
func (s *Store) ImportEvidence(ctx context.Context, raw []byte) error {
	var p EvidencePackage
	if e := json.Unmarshal(raw, &p); e != nil {
		return e
	}
	if p.Batch.BatchID == "" {
		return domain.ErrRequiredEvidence
	}
	if d, e := domain.EvidenceDigest(&p.Batch); e != nil || d != p.Digest {
		return domain.ErrInvalidMetrics
	}
	if e := verifyPackageEvents(p.Batch.BatchID, p.Events); e != nil {
		return fmt.Errorf("%w: %s", domain.ErrAuditChain, e.Error())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data.Batches[p.Batch.BatchID]; exists {
		return domain.ErrVersionConflict
	}
	if hasEventsLocked(s.data.Events, p.Batch.BatchID) {
		return domain.ErrVersionConflict
	}
	batchBefore, batchExisted := s.data.Batches[p.Batch.BatchID]
	prevEventCount := len(s.data.Events)
	s.data.Batches[p.Batch.BatchID] = p.Batch
	s.data.Events = append(s.data.Events, p.Events...)
	if e := s.verifyChainLocked(p.Batch.BatchID); e != nil {
		s.data.Batches[p.Batch.BatchID] = batchBefore
		if !batchExisted {
			delete(s.data.Batches, p.Batch.BatchID)
		}
		s.data.Events = s.data.Events[:prevEventCount]
		return e
	}
	return s.persist()
}

// verifyPackageEvents validates that the events in an evidence package form a
// contiguous, untampered audit chain belonging to the claimed batch. It rejects
// forged digests, misordered sequences, and events sourced from other batches.
func verifyPackageEvents(batchID string, events []AuditEvent) error {
	prev := ""
	for index, event := range events {
		if event.BatchID != batchID {
			return fmt.Errorf("事件 %d 的 batchID %q 与批次 %q 不一致", event.Seq, event.BatchID, batchID)
		}
		if event.Seq != int64(index+1) || event.PrevDigest != prev || event.Digest != eventDigest(event) {
			return fmt.Errorf("事件 %d 不连续或摘要不符", event.Seq)
		}
		prev = event.Digest
	}
	return nil
}

// hasEventsLocked reports whether any event in the global journal already
// belongs to the given batch. It must be called with s.mu held.
func hasEventsLocked(events []AuditEvent, batchID string) bool {
	for _, event := range events {
		if event.BatchID == batchID {
			return true
		}
	}
	return false
}
