package persistfailurestateleak_test

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistFailureDoesNotLeakUncommittedState(t *testing.T) {
	root := t.TempDir()
	activeDir := filepath.Join(root, "active")
	if err := os.Mkdir(activeDir, 0700); err != nil {
		t.Fatal(err)
	}

	store, err := storage.Open(filepath.Join(activeDir, "store.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := application.New(store)
	ctx := context.Background()

	created, err := service.CreateBatch(ctx, application.CreateBatchInput{
		BatchID:   "persist-failure",
		Title:     "持久化失败隔离测试",
		CreatedBy: "技术员",
		Items: []domain.DocumentItem{{
			ItemID: "document-1", CallNumber: "CALL-1", Title: "文献一", Material: "棉纸", PageCount: 8,
		}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	baselineEvents, err := service.Events(ctx, created.Batch.BatchID)
	if err != nil {
		t.Fatal(err)
	}

	if err = os.Rename(activeDir, filepath.Join(root, "detached")); err != nil {
		t.Fatal(err)
	}
	_, err = service.AddSample(ctx, created.Batch.BatchID, created.Batch.Version, domain.Sample{
		SampleID:        "sample-uncommitted",
		Material:        "棉纸",
		InitialPH:       5.2,
		InitialStrength: 100,
		Observer:        "检测员",
		ObservedAt:      time.Date(2026, 8, 27, 10, 0, 0, 0, time.UTC),
	}, "")
	if err == nil {
		t.Fatal("持久化目标失效后写入意外成功")
	}

	after, err := service.Get(ctx, created.Batch.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	afterEvents, err := service.Events(ctx, created.Batch.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Batch.Version != created.Batch.Version || len(after.Batch.Samples) != 0 || len(afterEvents) != len(baselineEvents) {
		t.Fatalf("失败写入污染内存状态: version=%d samples=%d events=%d, want version=%d samples=0 events=%d",
			after.Batch.Version, len(after.Batch.Samples), len(afterEvents), created.Batch.Version, len(baselineEvents))
	}
}
