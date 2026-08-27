package storage

import (
	"archive-deacidification/internal/domain"
	"context"
	"encoding/json"
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Batches[p.Batch.BatchID] = p.Batch
	s.data.Events = append(s.data.Events, p.Events...)
	return s.persist()
}
