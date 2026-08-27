package domain

import "strings"

func normalizeWhitespace(v string) string { return strings.Join(strings.Fields(v), " ") }

func NormalizeBatch(b *RestorationBatch) {
	b.Title = normalizeWhitespace(b.Title)
	b.Institution = normalizeWhitespace(b.Institution)
	b.TargetProcess = normalizeWhitespace(b.TargetProcess)
	b.CreatedBy = normalizeWhitespace(b.CreatedBy)
	for i := range b.DocumentItems {
		b.DocumentItems[i].ItemID = normalizeWhitespace(b.DocumentItems[i].ItemID)
		b.DocumentItems[i].Title = normalizeWhitespace(b.DocumentItems[i].Title)
		b.DocumentItems[i].CallNumber = normalizeWhitespace(b.DocumentItems[i].CallNumber)
		b.DocumentItems[i].Material = normalizeWhitespace(b.DocumentItems[i].Material)
	}
}
func NormalizeSample(s *Sample) {
	s.SampleID = normalizeWhitespace(s.SampleID)
	s.BatchID = normalizeWhitespace(s.BatchID)
	s.Material = normalizeWhitespace(s.Material)
	s.Observer = normalizeWhitespace(s.Observer)
}
func NormalizeTrial(t *ProcessTrial) {
	t.TrialID = normalizeWhitespace(t.TrialID)
	t.BatchID = normalizeWhitespace(t.BatchID)
	t.SampleID = normalizeWhitespace(t.SampleID)
	t.Reagent = normalizeWhitespace(t.Reagent)
	t.AnalysisStatus = strings.TrimSpace(t.AnalysisStatus)
}
func NormalizeRetest(r *CorrectionRetest) {
	r.RetestID = normalizeWhitespace(r.RetestID)
	r.BatchID = normalizeWhitespace(r.BatchID)
	r.TrialID = normalizeWhitespace(r.TrialID)
	r.Action = normalizeWhitespace(r.Action)
	r.Result = normalizeWhitespace(r.Result)
	r.SubmittedBy = normalizeWhitespace(r.SubmittedBy)
}
func IsPHInRange(v float64) bool       { return v > 0 && v <= 14 }
func IsStrengthInRange(v float64) bool { return v > 0 && v <= 1000000 }
func ValidateObservationRange(ph, strength float64) error {
	if !IsPHInRange(ph) || !IsStrengthInRange(strength) {
		return ErrInvalidMetrics
	}
	return nil
}
func DeviationCodesUnique(codes []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, c := range codes {
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	return out
}
