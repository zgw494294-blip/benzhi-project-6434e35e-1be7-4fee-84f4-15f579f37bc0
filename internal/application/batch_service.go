package application

import (
	"archive-deacidification/internal/domain"
	"context"
)

type summaryCacheEntry struct {
	version int
	summary domain.BatchSummary
}

func (s *Service) Summary(ctx context.Context, id string) (domain.BatchSummary, error) {
	s.summaryMu.Lock()
	if cached, ok := s.summaryCache[id]; ok {
		summary := cloneBatchSummary(cached.summary)
		s.summaryMu.Unlock()
		return summary, nil
	}
	s.summaryMu.Unlock()

	b, e := s.Snapshot(ctx, id)
	if e != nil {
		return domain.BatchSummary{}, e
	}
	summary := domain.Summarize(b)
	s.summaryMu.Lock()
	s.summaryCache[id] = summaryCacheEntry{version: b.Version, summary: cloneBatchSummary(summary)}
	s.summaryMu.Unlock()
	return summary, nil
}

func cloneBatchSummary(in domain.BatchSummary) domain.BatchSummary {
	out := in
	out.MaterialCount = make(map[string]int, len(in.MaterialCount))
	for material, count := range in.MaterialCount {
		out.MaterialCount[material] = count
	}
	out.SamplingCoverage.Materials = append([]domain.MaterialCoverage(nil), in.SamplingCoverage.Materials...)
	return out
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
