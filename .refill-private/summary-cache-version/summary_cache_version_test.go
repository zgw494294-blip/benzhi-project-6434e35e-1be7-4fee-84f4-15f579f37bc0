package summarycacheversion

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"testing"
	"time"
)

func TestSummaryCacheRefreshesAfterBatchVersionChanges(t *testing.T) {
	ctx := context.Background()
	store, err := storage.Open("file:summary-cache-version?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	service := application.New(store)

	created, err := service.CreateBatch(ctx, application.CreateBatchInput{
		BatchID:   "summary-cache-batch",
		Title:     "摘要缓存版本测试",
		CreatedBy: "修复技术员",
		Items: []domain.DocumentItem{{
			ItemID: "document-1", CallNumber: "A-001", Title: "测试文献",
			Material: "棉纸", PageCount: 12,
		}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.Summary(ctx, created.Batch.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	if before.SampleCount != 0 {
		t.Fatalf("首次摘要应包含 0 个样本，实际为 %d", before.SampleCount)
	}

	updated, err := service.AddSample(ctx, created.Batch.BatchID, created.Batch.Version, domain.Sample{
		SampleID:        "sample-1",
		Material:        "棉纸",
		InitialPH:       5.2,
		InitialStrength: 100,
		Observer:        "实验检测员",
		ObservedAt:      time.Date(2026, 8, 27, 10, 0, 0, 0, time.UTC),
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Batch.Version != created.Batch.Version+1 {
		t.Fatalf("批次版本未前进：创建版本 %d，写入后版本 %d", created.Batch.Version, updated.Batch.Version)
	}

	after, err := service.Summary(ctx, created.Batch.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	if after.SampleCount != 1 || after.SamplingCoverage.CoveredMaterials != 1 {
		t.Fatalf("版本 %d 的摘要仍为旧缓存：sampleCount=%d coveredMaterials=%d", updated.Batch.Version, after.SampleCount, after.SamplingCoverage.CoveredMaterials)
	}
}
