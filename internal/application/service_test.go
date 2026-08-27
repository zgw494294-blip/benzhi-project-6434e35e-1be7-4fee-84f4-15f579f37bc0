package application

import (
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkflow(t *testing.T) {
	st, e := storage.Open("file:testapp?mode=memory&cache=shared")
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	s := New(st)
	ctx := context.Background()
	r, e := s.CreateBatch(ctx, CreateBatchInput{BatchID: "b1", Title: "批次", CreatedBy: "技术员", Items: []domain.DocumentItem{{ItemID: "d", CallNumber: "A-1", Title: "文献", Material: "纸", PageCount: 2}}}, "")
	if e != nil {
		t.Fatal(e)
	}
	r, e = s.AddSample(ctx, "b1", r.Batch.Version, domain.Sample{SampleID: "s", Material: "纸", InitialPH: 5, InitialStrength: 100, Observer: "检测员", ObservedAt: time.Now().UTC()}, "")
	if e != nil {
		t.Fatal(e)
	}
	r, e = s.AddTrial(ctx, "b1", r.Batch.Version, domain.ProcessTrial{TrialID: "t", SampleID: "s", Reagent: "药剂", Concentration: 2, TemperatureCelsius: 20, DurationMinutes: 10, InitialPH: 5, FinalPH: 6.5, StrengthBefore: 100, StrengthAfter: 95}, "")
	if e != nil {
		t.Fatal(e)
	}
	r, e = s.Review(ctx, "b1", r.Batch.Version, true, "负责人", "", "")
	if e != nil {
		t.Fatal(e)
	}
	r, e = s.Freeze(ctx, "b1", r.Batch.Version, "负责人", "")
	if e != nil || r.Credential == nil {
		t.Fatalf("freeze: %v", e)
	}
	ok, e := s.Verify(ctx, "b1", r.Credential.VerificationCode)
	if e != nil || !ok {
		t.Fatalf("verify: %v %v", ok, e)
	}
	events, e := s.Events(ctx, "b1")
	if e != nil || len(events) != 6 {
		t.Fatalf("events: %d %v", len(events), e)
	}
	for i, event := range events {
		if event.Seq != int64(i+1) {
			t.Fatalf("audit sequence %d: %d", i, event.Seq)
		}
	}
}

func TestDuplicateScopeDoesNotCreateBatchOrAudit(t *testing.T) {
	st, e := storage.Open("file:duplicate?mode=memory&cache=shared")
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	s := New(st)
	items := []domain.DocumentItem{
		{ItemID: "D1", CallNumber: "C1", Title: "甲", Material: "棉纸", PageCount: 2},
		{ItemID: "D1", CallNumber: "C2", Title: "乙", Material: "竹纸", PageCount: 3},
	}
	_, e = s.CreateBatch(context.Background(), CreateBatchInput{BatchID: "duplicate", Title: "重复清单", CreatedBy: "技术员", Items: items}, "")
	if !errors.Is(e, domain.ErrDocumentDuplicate) {
		t.Fatalf("expected duplicate error, got %v", e)
	}
	if _, e = st.Get(context.Background(), "duplicate"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatalf("duplicate batch was persisted: %v", e)
	}
	events, e := st.Events(context.Background(), "duplicate")
	if e != nil || len(events) != 0 {
		t.Fatalf("duplicate validation emitted audit event: %#v %v", events, e)
	}
}

func TestCoverageDeviationRetestAndFrozenGuard(t *testing.T) {
	st, e := storage.Open("file:extended?mode=memory&cache=shared")
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	s := New(st)
	ctx := context.Background()
	created, e := s.CreateBatch(ctx, CreateBatchInput{BatchID: "extended", Title: "扩展流程", CreatedBy: "技术员", Items: []domain.DocumentItem{
		{ItemID: "D1", CallNumber: "C1", Title: "甲", Material: "棉纸", PageCount: 10},
		{ItemID: "D2", CallNumber: "C2", Title: "乙", Material: "竹纸", PageCount: 20},
	}}, "")
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	one, e := s.AddSample(ctx, "extended", created.Batch.Version, domain.Sample{SampleID: "S1", Material: "棉纸", InitialPH: 5, InitialStrength: 100, Observer: "检测员", ObservedAt: now}, "")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.AddSample(ctx, "extended", one.Batch.Version, domain.Sample{SampleID: "S1", Material: "竹纸"}, ""); !errors.Is(e, domain.ErrSampleDuplicate) {
		t.Fatalf("expected immutable sample conflict, got %v", e)
	}
	trial := domain.ProcessTrial{TrialID: "T1", SampleID: "S1", Reagent: "碳酸镁", Concentration: 2, TemperatureCelsius: 25, DurationMinutes: 30, InitialPH: 5, FinalPH: 6.5, StrengthBefore: 100, StrengthAfter: 95}
	if _, e = s.AddTrial(ctx, "extended", one.Batch.Version, trial, ""); !errors.Is(e, domain.ErrSamplingCoverage) {
		t.Fatalf("expected coverage rejection, got %v", e)
	}
	unchanged, e := s.Get(ctx, "extended")
	if e != nil || unchanged.Batch.Version != one.Batch.Version || len(unchanged.Batch.Trials) != 0 {
		t.Fatalf("coverage rejection changed batch: %#v %v", unchanged.Batch, e)
	}
	two, e := s.AddSample(ctx, "extended", one.Batch.Version, domain.Sample{SampleID: "S2", Material: "竹纸", InitialPH: 5.2, InitialStrength: 90, Observer: "检测员", ObservedAt: now}, "")
	if e != nil {
		t.Fatal(e)
	}
	trial.TemperatureCelsius = 45
	trialResult, e := s.AddTrial(ctx, "extended", two.Batch.Version, trial, "")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Review(ctx, "extended", trialResult.Batch.Version, true, "负责人", "", ""); !errors.Is(e, domain.ErrDeviationOpen) {
		t.Fatalf("expected open deviation, got %v", e)
	}
	retest, e := s.AddRetest(ctx, "extended", trialResult.Batch.Version, domain.CorrectionRetest{TrialID: "T1", Action: "降低温度", Result: "passed", SubmittedBy: "检测员"}, "")
	if e != nil {
		t.Fatal(e)
	}
	reviewed, e := s.Review(ctx, "extended", retest.Batch.Version, true, "负责人", "参数已纠正", "")
	if e != nil {
		t.Fatal(e)
	}
	frozen, e := s.Freeze(ctx, "extended", reviewed.Batch.Version, "负责人", "")
	if e != nil {
		t.Fatal(e)
	}
	eventsBefore, _ := s.Events(ctx, "extended")
	if _, e = s.AddSample(ctx, "extended", reviewed.Batch.Version, domain.Sample{SampleID: "S3"}, ""); !errors.Is(e, domain.ErrFrozen) {
		t.Fatalf("expected frozen guard, got %v", e)
	}
	eventsAfter, _ := s.Events(ctx, "extended")
	if len(eventsBefore) != len(eventsAfter) || frozen.Batch.Version != reviewed.Batch.Version+1 {
		t.Fatalf("frozen rejection changed state: %d -> %d", len(eventsBefore), len(eventsAfter))
	}
	valid, e := s.Verify(ctx, "extended", frozen.Credential.VerificationCode)
	if e != nil || !valid {
		t.Fatalf("verification failed: %t %v", valid, e)
	}
	invalid, e := s.Verify(ctx, "extended", "wrong-code")
	if e != nil || invalid {
		t.Fatalf("wrong credential was accepted: %t %v", invalid, e)
	}
}
