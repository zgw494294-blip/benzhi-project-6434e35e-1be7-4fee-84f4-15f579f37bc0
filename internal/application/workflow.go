package application

import (
	"archive-deacidification/internal/domain"
	"context"
)

func (s *Service) Snapshot(ctx context.Context, id string) (domain.RestorationBatch, error) {
	r, e := s.Get(ctx, id)
	return r.Batch, e
}
func (s *Service) TrialReport(ctx context.Context, id string) (domain.ComparisonReport, error) {
	b, e := s.Snapshot(ctx, id)
	if e != nil {
		return domain.ComparisonReport{}, e
	}
	return domain.BuildComparisonReport(b)
}
