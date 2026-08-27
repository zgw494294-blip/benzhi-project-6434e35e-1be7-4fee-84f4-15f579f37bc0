package application

import "archive-deacidification/internal/domain"

type ReviewInput struct {
	Approve          bool
	Reviewer, Reason string
}
type CorrectionInput struct {
	TrialID, Action, Result, SubmittedBy string
	Observations                         map[string]float64
}

func TrialMetrics(t domain.ProcessTrial) (domain.Metrics, error) {
	return domain.Analyze(t.InitialPH, t.FinalPH, t.StrengthBefore, t.StrengthAfter)
}
