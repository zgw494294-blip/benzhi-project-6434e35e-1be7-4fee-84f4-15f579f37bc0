package get_alias_state_pollution

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"testing"
	"time"
)

func TestReadResultCannotMutateFrozenEvidence(t *testing.T) {
	store, err := storage.Open("file:get-alias?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := application.New(store)
	ctx := context.Background()
	created, err := service.CreateBatch(ctx, application.CreateBatchInput{
		BatchID: "alias-batch", Title: "冻结证据别名", CreatedBy: "技术员",
		Items: []domain.DocumentItem{{ItemID: "D1", CallNumber: "C1", Title: "原始文献", Material: "棉纸", PageCount: 10}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	sampled, err := service.AddSample(ctx, "alias-batch", created.Batch.Version, domain.Sample{
		SampleID: "S1", Material: "棉纸", InitialPH: 5, InitialStrength: 100,
		Observer: "原始检测员", ObservedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	trialed, err := service.AddTrial(ctx, "alias-batch", sampled.Batch.Version, domain.ProcessTrial{
		TrialID: "T1", SampleID: "S1", Reagent: "碳酸镁", Concentration: 2,
		TemperatureCelsius: 25, DurationMinutes: 30, InitialPH: 5, FinalPH: 5.2,
		StrengthBefore: 100, StrengthAfter: 95,
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	retested, err := service.AddRetest(ctx, "alias-batch", trialed.Batch.Version, domain.CorrectionRetest{
		RetestID: "R1", TrialID: "T1", Action: "补充脱酸", Result: "passed", SubmittedBy: "检测员",
		Observations: map[string]float64{"finalPH": 6.5},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	reviewed, err := service.Review(ctx, "alias-batch", retested.Batch.Version, true, "质量负责人", "", "")
	if err != nil {
		t.Fatal(err)
	}
	frozen, err := service.Freeze(ctx, "alias-batch", reviewed.Batch.Version, "质量负责人", "")
	if err != nil {
		t.Fatal(err)
	}

	read, err := service.Get(ctx, "alias-batch")
	if err != nil {
		t.Fatal(err)
	}
	read.Batch.DocumentItems[0].Title = "外部篡改"
	read.Batch.DocumentStats.MaterialCount["棉纸"] = 999
	read.Batch.Samples[0].Observer = "外部篡改"
	read.Batch.Trials[0].Reagent = "外部篡改"
	read.Batch.Retests[0].Observations["finalPH"] = 1
	read.Batch.Credential.Reviewer = "外部篡改"

	after, err := service.Get(ctx, "alias-batch")
	if err != nil {
		t.Fatal(err)
	}
	verified, verifyErr := service.Verify(ctx, "alias-batch", frozen.Credential.VerificationCode)
	if verifyErr != nil {
		t.Fatal(verifyErr)
	}
	if after.Batch.DocumentItems[0].Title != "原始文献" ||
		after.Batch.DocumentStats.MaterialCount["棉纸"] != 1 ||
		after.Batch.Samples[0].Observer != "原始检测员" ||
		after.Batch.Trials[0].Reagent != "碳酸镁" ||
		after.Batch.Retests[0].Observations["finalPH"] != 6.5 ||
		after.Batch.Credential.Reviewer != "质量负责人" || !verified {
		t.Fatalf("读取结果的共享引用污染了冻结证据: batch=%#v verified=%t", after.Batch, verified)
	}
}
