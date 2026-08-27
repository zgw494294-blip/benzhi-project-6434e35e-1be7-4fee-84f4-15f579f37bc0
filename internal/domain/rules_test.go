package domain

import (
	"errors"
	"testing"
	"time"
)

func TestAnalyze(t *testing.T) {
	m, e := Analyze(5, 6.5, 100, 95)
	if e != nil || !m.Passed {
		t.Fatalf("expected pass: %#v %v", m, e)
	}
	m, e = Analyze(5, 5.2, 100, 70)
	if e != nil || m.Passed || len(m.Deviations) != 2 {
		t.Fatalf("expected deviations: %#v", m)
	}
}

func TestDocumentScopeAndSamplingCoverage(t *testing.T) {
	b := RestorationBatch{DocumentItems: []DocumentItem{
		{ItemID: " A 1 ", CallNumber: " C 1 ", Title: " 甲  文献 ", Material: " 棉 纸 ", PageCount: 10},
		{ItemID: "A2", CallNumber: "C2", Title: "乙", Material: "竹纸", PageCount: 20},
	}}
	NormalizeBatch(&b)
	if e := ValidateDocumentItems(b.DocumentItems); e != nil {
		t.Fatal(e)
	}
	if b.DocumentItems[0].Title != "甲 文献" || b.DocumentItems[0].Material != "棉 纸" {
		t.Fatalf("whitespace was not normalized: %#v", b.DocumentItems[0])
	}
	b.Status = StatusDraft
	if e := b.AddSample(Sample{SampleID: "S1", Material: "棉 纸", Observer: "检测员", ObservedAt: time.Now(), InitialPH: 5, InitialStrength: 100}); e != nil {
		t.Fatal(e)
	}
	coverage := CalculateSamplingCoverage(b)
	if coverage.Satisfied || coverage.Rate != 0.5 {
		t.Fatalf("unexpected coverage: %#v", coverage)
	}
	trial := ProcessTrial{TrialID: "T1", SampleID: "S1", Reagent: "碳酸镁", Concentration: 2, TemperatureCelsius: 20, DurationMinutes: 30, InitialPH: 5, FinalPH: 6.5, StrengthBefore: 100, StrengthAfter: 95}
	if e := b.AddTrial(trial); !errors.Is(e, ErrSamplingCoverage) {
		t.Fatalf("expected sampling coverage error, got %v", e)
	}
	if len(b.Trials) != 0 {
		t.Fatal("coverage rejection mutated trials")
	}
}

func TestComparisonReportSeparatesParameterDeviation(t *testing.T) {
	b := RestorationBatch{
		Status:        StatusSampling,
		DocumentItems: []DocumentItem{{ItemID: "D1", CallNumber: "C1", Title: "文献", Material: "纸", PageCount: 1}},
		Samples:       []Sample{{SampleID: "S1", Material: "纸"}},
	}
	base := ProcessTrial{SampleID: "S1", Reagent: "碳酸镁", TemperatureCelsius: 25, DurationMinutes: 30, InitialPH: 5, FinalPH: 6.5, StrengthBefore: 100, StrengthAfter: 95}
	first := base
	first.TrialID, first.Concentration = "T1", 2
	if e := b.AddTrial(first); e != nil {
		t.Fatal(e)
	}
	second := base
	second.TrialID, second.Concentration, second.TemperatureCelsius = "T2", 3, 45
	if e := b.AddTrial(second); e != nil {
		t.Fatal(e)
	}
	report, e := BuildComparisonReport(b)
	if e != nil {
		t.Fatal(e)
	}
	if len(report.Groups) != 2 || report.PassedCount != 1 || report.OpenDeviationCount != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if len(report.Reports[1].ParameterDeviations) != 1 || len(report.Reports[1].MetricDeviations) != 0 {
		t.Fatalf("deviation types were not separated: %#v", report.Reports[1])
	}
}
func TestTransition(t *testing.T) {
	b := RestorationBatch{Status: StatusDraft, DocumentItems: []DocumentItem{{ItemID: "1", Title: "x", PageCount: 1}}}
	if e := ValidateBatchForSampling(&b); e != nil {
		t.Fatal(e)
	}
	if e := b.Transition(StatusSampling); e != nil {
		t.Fatal(e)
	}
	if e := b.Transition(StatusFrozen); e == nil {
		t.Fatal("invalid transition accepted")
	}
}
