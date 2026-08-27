package domain

import (
	"fmt"
	"strings"
)

const MinPHChange = 1.0
const MinStrengthRetention = 0.85

const (
	MinSamplesPerMaterial = 1
	MinConcentration      = 0.5
	MaxConcentration      = 5.0
	MinTemperatureCelsius = 15.0
	MaxTemperatureCelsius = 35.0
	MinDurationMinutes    = 5
	MaxDurationMinutes    = 120
)

func Analyze(initialPH, finalPH, before, after float64) (Metrics, error) {
	if initialPH <= 0 || finalPH <= 0 || before <= 0 || after < 0 {
		return Metrics{}, ErrInvalidMetrics
	}
	m := Metrics{PHChange: finalPH - initialPH, StrengthRetention: after / before, Passed: true}
	if m.PHChange < MinPHChange {
		m.Deviations = append(m.Deviations, "PH_CHANGE_LOW")
	}
	if m.StrengthRetention < MinStrengthRetention {
		m.Deviations = append(m.Deviations, "STRENGTH_RETENTION_LOW")
	}
	m.Passed = len(m.Deviations) == 0
	return m, nil
}

func ValidateBatchForSampling(b *RestorationBatch) error {
	if b.Status != StatusDraft {
		return ErrInvalidTransition
	}
	if len(b.DocumentItems) == 0 {
		return ErrRequiredEvidence
	}
	return nil
}
func ValidateBatchForTrial(b *RestorationBatch) error {
	if b.Status != StatusSampling && b.Status != StatusTrialing {
		return ErrInvalidTransition
	}
	coverage := CalculateSamplingCoverage(*b)
	if !coverage.Satisfied {
		return fmt.Errorf("%w: 覆盖率 %.3f", ErrSamplingCoverage, coverage.Rate)
	}
	return nil
}
func ValidateBatchForReview(b *RestorationBatch) error {
	if b.Status != StatusTrialing && b.Status != StatusCorrection && b.Status != StatusReview {
		return ErrInvalidTransition
	}
	if len(b.Trials) == 0 {
		return ErrRequiredEvidence
	}
	for _, t := range b.Trials {
		if len(t.DeviationCodes) > 0 && !HasClosedRetest(b, t.TrialID) {
			return fmt.Errorf("%w: %s", ErrDeviationOpen, t.TrialID)
		}
	}
	return nil
}
func HasClosedRetest(b *RestorationBatch, id string) bool {
	for _, r := range b.Retests {
		if r.TrialID == id && retestPassed(r.Result) {
			return true
		}
	}
	return false
}

func retestPassed(result string) bool {
	switch strings.ToLower(strings.TrimSpace(result)) {
	case "passed", "pass", "closed", "合格", "通过", "已闭环":
		return true
	default:
		return false
	}
}
func (b *RestorationBatch) Transition(next BatchStatus) error {
	if b.Status == StatusFrozen {
		return ErrFrozen
	}
	allowed := map[BatchStatus][]BatchStatus{StatusDraft: {StatusSampling}, StatusSampling: {StatusTrialing}, StatusTrialing: {StatusCorrection, StatusReview}, StatusCorrection: {StatusTrialing, StatusReview}, StatusReview: {StatusApproved, StatusRejected}, StatusApproved: {StatusFrozen}, StatusRejected: {StatusCorrection}}
	for _, s := range allowed[b.Status] {
		if s == next {
			b.Status = next
			b.Version++
			return nil
		}
	}
	return ErrInvalidTransition
}
func (b *RestorationBatch) AddSample(s Sample) error {
	if b.Status == StatusFrozen {
		return ErrFrozen
	}
	for _, existing := range b.Samples {
		if existing.SampleID == s.SampleID {
			return fmt.Errorf("%w: %s", ErrSampleDuplicate, s.SampleID)
		}
	}
	if !materialInScope(*b, s.Material) {
		return fmt.Errorf("%w: 样本材质 %q 不在批次清单中", ErrDocumentScope, s.Material)
	}
	b.Samples = append(b.Samples, s)
	if b.Status == StatusDraft {
		b.Status = StatusSampling
	}
	b.Version++
	return nil
}
func (b *RestorationBatch) AddTrial(t ProcessTrial) error {
	if b.Status == StatusFrozen {
		return ErrFrozen
	}
	if t.SampleID == "" || t.Reagent == "" {
		return ErrRequiredEvidence
	}
	if e := ValidateBatchForTrial(b); e != nil {
		return e
	}
	if _, ok := SampleByID(*b, t.SampleID); !ok {
		return fmt.Errorf("%w: sampleID %q 不在批次取样范围中", ErrDocumentScope, t.SampleID)
	}
	for _, existing := range b.Trials {
		if existing.TrialID == t.TrialID {
			return fmt.Errorf("%w: %s", ErrTrialDuplicate, t.TrialID)
		}
	}
	m, e := Analyze(t.InitialPH, t.FinalPH, t.StrengthBefore, t.StrengthAfter)
	if e != nil {
		return e
	}
	t.MetricDeviations = append([]string(nil), m.Deviations...)
	t.ParameterDeviations = AnalyzeParameters(t)
	t.DeviationCodes = DeviationCodesUnique(append(append([]string(nil), t.MetricDeviations...), t.ParameterDeviations...))
	t.AnalysisStatus = map[bool]string{true: "passed", false: "deviation"}[len(t.DeviationCodes) == 0]
	b.Trials = append(b.Trials, t)
	b.Status = StatusTrialing
	b.Version++
	return nil
}
func (b *RestorationBatch) AddRetest(r CorrectionRetest) error {
	if b.Status == StatusFrozen {
		return ErrFrozen
	}
	if r.RetestID == "" || r.TrialID == "" || r.Action == "" || r.Result == "" || r.SubmittedBy == "" {
		return ErrRequiredEvidence
	}
	t, ok := TrialByID(*b, r.TrialID)
	if !ok {
		return fmt.Errorf("%w: trialID %q 不在批次试验范围中", ErrDocumentScope, r.TrialID)
	}
	if len(t.DeviationCodes) == 0 {
		return fmt.Errorf("%w: trialID %q 没有待整改偏差", ErrDocumentScope, r.TrialID)
	}
	for _, existing := range b.Retests {
		if existing.RetestID == r.RetestID {
			return fmt.Errorf("%w: %s", ErrRetestDuplicate, r.RetestID)
		}
	}
	b.Retests = append(b.Retests, r)
	b.Status = StatusCorrection
	b.Version++
	return nil
}

func AnalyzeParameters(t ProcessTrial) []string {
	var out []string
	if t.Concentration < MinConcentration || t.Concentration > MaxConcentration {
		out = append(out, "PARAM_CONCENTRATION_OUT_OF_RANGE")
	}
	if t.TemperatureCelsius < MinTemperatureCelsius || t.TemperatureCelsius > MaxTemperatureCelsius {
		out = append(out, "PARAM_TEMPERATURE_OUT_OF_RANGE")
	}
	if t.DurationMinutes < MinDurationMinutes || t.DurationMinutes > MaxDurationMinutes {
		out = append(out, "PARAM_DURATION_OUT_OF_RANGE")
	}
	return out
}

func materialInScope(b RestorationBatch, material string) bool {
	for _, item := range b.DocumentItems {
		if item.Material == material {
			return true
		}
	}
	return false
}
