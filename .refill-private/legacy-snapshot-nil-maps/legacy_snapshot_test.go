package legacysnapshotnilmaps

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLegacySnapshotRestartRehydratesWritableCollections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.json")
	if err := os.WriteFile(path, []byte(`{"events":[]}`), 0o600); err != nil {
		t.Fatalf("写入旧快照失败: %v", err)
	}

	store, err := storage.Open(path)
	if err != nil {
		t.Fatalf("打开旧快照失败: %v", err)
	}
	service := application.New(store)

	createPanic := capturePanic(func() {
		_, _ = service.CreateBatch(context.Background(), application.CreateBatchInput{
			BatchID:       "legacy-batch",
			Title:         "旧快照恢复批次",
			Institution:   "示例档案馆",
			TargetProcess: "水性脱酸",
			CreatedBy:     "修复员甲",
			Items: []domain.DocumentItem{{
				ItemID: "doc-1", Title: "档案一", CallNumber: "A-001", Material: "棉纸", PageCount: 10,
			}},
		}, "")
	})
	if createPanic != nil {
		t.Errorf("重启后创建批次发生 panic: %v", createPanic)
	}

	idempotencyPanic := capturePanic(func() {
		_ = store.PutIdempotent(context.Background(), "legacy-batch", "request-1", []byte(`{"ok":true}`))
	})
	if idempotencyPanic != nil {
		t.Errorf("重启后写入幂等缓存发生 panic: %v", idempotencyPanic)
	}
}

func capturePanic(fn func()) (recovered any) {
	defer func() {
		recovered = recover()
	}()
	fn()
	return nil
}
