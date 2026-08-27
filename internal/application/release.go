package application

import (
	"archive-deacidification/internal/domain"
	"context"
)

type ReleaseView struct {
	Credential domain.ReleaseCredential `json:"credential"`
	Summary    domain.BatchSummary      `json:"summary"`
	Verified   bool                     `json:"verified"`
}

func (s *Service) ReleaseView(ctx context.Context, id string) (ReleaseView, error) {
	r, e := s.Get(ctx, id)
	if e != nil {
		return ReleaseView{}, e
	}
	if r.Credential == nil {
		return ReleaseView{}, domain.ErrRequiredEvidence
	}
	digest, e := domain.EvidenceDigest(&r.Batch)
	if e != nil {
		return ReleaseView{}, e
	}
	chainValid, e := s.repo.VerifyChain(ctx, id)
	if e != nil {
		return ReleaseView{}, e
	}
	verified := chainValid && digest == r.Credential.EvidenceDigest && r.Batch.Status == domain.StatusFrozen
	return ReleaseView{Credential: *r.Credential, Summary: domain.Summarize(r.Batch), Verified: verified}, nil
}
func (s *Service) IsReleased(ctx context.Context, id string) (bool, error) {
	b, e := s.Snapshot(ctx, id)
	if e != nil {
		return false, e
	}
	return b.Status == domain.StatusFrozen && b.Credential != nil, nil
}
