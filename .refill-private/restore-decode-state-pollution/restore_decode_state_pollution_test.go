package restore_decode_state_pollution_test

import (
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreDecodeFailurePreservesLiveState(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	storePath := filepath.Join(dir, "live.json")
	st, err := storage.Open(storePath)
	if err != nil {
		t.Fatal(err)
	}

	live := domain.RestorationBatch{BatchID: "live-batch", Title: "在线批次", Version: 1}
	if _, err = st.Save(ctx, live, "batch.created", live); err != nil {
		t.Fatalf("seed live state: %v", err)
	}

	brokenBackup := filepath.Join(dir, "broken-backup.json")
	brokenJSON := []byte(`{"batches":{"partial-batch":{"batchID":"partial-batch","title":"部分备份","version":9}},"events":"invalid-event-list","idempotency":{}}`)
	if err = os.WriteFile(brokenBackup, brokenJSON, 0600); err != nil {
		t.Fatal(err)
	}

	if err = st.Restore(ctx, brokenBackup); err == nil {
		t.Fatal("expected malformed backup restore to fail")
	}
	if _, err = st.Get(ctx, "live-batch"); err != nil {
		t.Fatalf("failed restore replaced the live batch: %v", err)
	}
	if _, err = st.Get(ctx, "partial-batch"); err == nil {
		t.Fatal("failed restore published partially decoded backup data")
	}
}
