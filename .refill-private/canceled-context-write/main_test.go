package canceled_context_write

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"errors"
	"testing"
	"time"
)

func TestCanceledContextDoesNotPersistSample(t *testing.T) {
	store, err := storage.Open("file:canceled-context?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := application.New(store)
	created, err := service.CreateBatch(context.Background(), application.CreateBatchInput{
		BatchID: "context-batch", Title: "取消写入", CreatedBy: "技术员",
		Items: []domain.DocumentItem{{ItemID: "D1", CallNumber: "C1", Title: "文献", Material: "棉纸", PageCount: 10}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	_, writeErr := service.AddSample(canceled, "context-batch", created.Batch.Version, domain.Sample{
		SampleID: "S1", Material: "棉纸", InitialPH: 5, InitialStrength: 100,
		Observer: "检测员", ObservedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}, "")
	after, getErr := service.Get(context.Background(), "context-batch")
	if getErr != nil {
		t.Fatal(getErr)
	}
	if !errors.Is(writeErr, context.Canceled) || after.Batch.Version != created.Batch.Version || len(after.Batch.Samples) != 0 {
		t.Fatalf("已取消的 context 仍改变了批次: err=%v version=%d samples=%d", writeErr, after.Batch.Version, len(after.Batch.Samples))
	}
}
