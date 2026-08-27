package application

import (
	"archive-deacidification/internal/domain"
	"context"
)

func (s *Service) Summary(ctx context.Context, id string) (domain.BatchSummary, error) {
	b, e := s.Snapshot(ctx, id)
	if e != nil {
		return domain.BatchSummary{}, e
	}
	return domain.Summarize(b), nil
}
func (s *Service) EnsureExpected(ctx context.Context, id string, expected int) error {
	_, e := s.loadVersion(ctx, id, expected)
	return e
}
func (s *Service) RejectReason(ctx context.Context, id string) (string, error) {
	b, e := s.Snapshot(ctx, id)
	if e != nil {
		return "", e
	}
	if b.Status != domain.StatusRejected {
		return "", nil
	}
	ev, e := s.Events(ctx, id)
	if e != nil {
		return "", e
	}
	for i := len(ev) - 1; i >= 0; i-- {
		if ev[i].Type == "review.completed" {
			var p struct {
				Reason string `json:"reason"`
			}
			_ = unmarshal(ev[i].Payload, &p)
			return p.Reason, nil
		}
	}
	return "", nil
}
