package application

import (
	"archive-deacidification/internal/domain"
	"context"
)

type MetricsOverview struct {
	Trials                   int     `json:"trials"`
	Passed                   int     `json:"passed"`
	Failed                   int     `json:"failed"`
	AveragePHChange          float64 `json:"averagePHChange"`
	AverageStrengthRetention float64 `json:"averageStrengthRetention"`
}

func (s *Service) MetricsOverview(ctx context.Context, id string) (MetricsOverview, error) {
	b, e := s.Snapshot(ctx, id)
	if e != nil {
		return MetricsOverview{}, e
	}
	o := MetricsOverview{Trials: len(b.Trials)}
	for _, t := range b.Trials {
		m, e := domain.Analyze(t.InitialPH, t.FinalPH, t.StrengthBefore, t.StrengthAfter)
		if e != nil {
			return o, e
		}
		if m.Passed {
			o.Passed++
		} else {
			o.Failed++
		}
		o.AveragePHChange += m.PHChange
		o.AverageStrengthRetention += m.StrengthRetention
	}
	if o.Trials > 0 {
		o.AveragePHChange /= float64(o.Trials)
		o.AverageStrengthRetention /= float64(o.Trials)
	}
	return o, nil
}
func (s *Service) Deviations(ctx context.Context, id string) (map[string]int, error) {
	b, e := s.Snapshot(ctx, id)
	if e != nil {
		return nil, e
	}
	out := map[string]int{}
	for _, t := range b.Trials {
		for _, c := range t.DeviationCodes {
			out[c]++
		}
	}
	return out, nil
}
