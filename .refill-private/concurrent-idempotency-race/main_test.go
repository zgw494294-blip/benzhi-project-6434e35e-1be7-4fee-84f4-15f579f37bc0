package concurrent_idempotency_race

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"sync"
	"testing"
	"time"
)

type idempotencyBarrierRepository struct {
	storage.Repository
	mu       sync.Mutex
	armed    bool
	arrivals int
	gate     chan struct{}
}

func (r *idempotencyBarrierRepository) arm() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.armed = true
	r.arrivals = 0
	r.gate = make(chan struct{})
}

func (r *idempotencyBarrierRepository) GetIdempotent(ctx context.Context, batch, key string) ([]byte, bool) {
	v, ok := r.Repository.GetIdempotent(ctx, batch, key)
	r.mu.Lock()
	if !r.armed {
		r.mu.Unlock()
		return v, ok
	}
	r.arrivals++
	gate := r.gate
	if r.arrivals == 2 {
		r.armed = false
		close(gate)
	}
	r.mu.Unlock()
	<-gate
	return v, ok
}

func TestConcurrentIdempotencyKeyPreventsDuplicateCommit(t *testing.T) {
	store, err := storage.Open("file:concurrent-idempotency?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	repo := &idempotencyBarrierRepository{Repository: store}
	service := application.New(repo)
	ctx := context.Background()
	created, err := service.CreateBatch(ctx, application.CreateBatchInput{
		BatchID:   "concurrent-batch",
		Title:     "并发幂等",
		CreatedBy: "技术员",
		Items: []domain.DocumentItem{{
			ItemID: "D1", CallNumber: "C1", Title: "文献", Material: "棉纸", PageCount: 10,
		}},
	}, "")
	if err != nil {
		t.Fatal(err)
	}

	repo.arm()
	type outcome struct {
		result application.Result
		err    error
	}
	outcomes := make(chan outcome, 2)
	for range 2 {
		go func() {
			result, addErr := service.AddSample(ctx, "concurrent-batch", created.Batch.Version, domain.Sample{
				SampleID: "S1", Material: "棉纸", InitialPH: 5, InitialStrength: 100,
				Observer: "检测员", ObservedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			}, "same-request")
			outcomes <- outcome{result: result, err: addErr}
		}()
	}
	first, second := <-outcomes, <-outcomes
	if first.err != nil || second.err != nil || first.result.Sequence != second.result.Sequence || first.result.Batch.Version != second.result.Batch.Version {
		t.Fatalf("并发相同幂等键未复用同一成功响应: first=(%#v,%v) second=(%#v,%v)", first.result, first.err, second.result, second.err)
	}
	latest, err := service.Get(ctx, "concurrent-batch")
	if err != nil {
		t.Fatal(err)
	}
	events, err := service.Events(ctx, "concurrent-batch")
	if err != nil {
		t.Fatal(err)
	}
	if len(latest.Batch.Samples) != 1 || len(events) != 3 || latest.Batch.Version != created.Batch.Version+1 {
		t.Fatalf("并发幂等提交产生了重复状态或审计事件: version=%d samples=%d events=%d", latest.Batch.Version, len(latest.Batch.Samples), len(events))
	}
}
