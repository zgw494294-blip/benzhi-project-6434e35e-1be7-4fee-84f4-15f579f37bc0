package import_evidence_chain_bypass

import (
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestImportEvidenceRejectsUnverifiedEvents(t *testing.T) {
	ctx := context.Background()
	source, err := storage.Open("file:import-source?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	batch := domain.RestorationBatch{BatchID: "import-batch", Title: "证据导入", CreatedBy: "技术员", Status: domain.StatusDraft, Version: 1}
	if _, err = source.Save(ctx, batch, "batch.created", batch); err != nil {
		t.Fatal(err)
	}
	raw, err := source.ExportEvidence(ctx, batch.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	var pkg storage.EvidencePackage
	if err = json.Unmarshal(raw, &pkg); err != nil {
		t.Fatal(err)
	}
	if len(pkg.Events) != 1 {
		t.Fatalf("source evidence events=%d", len(pkg.Events))
	}
	pkg.Events[0].Digest = "forged-digest"
	tampered, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}

	target, err := storage.Open("file:import-target?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	importErr := target.ImportEvidence(ctx, tampered)
	if !errors.Is(importErr, domain.ErrAuditChain) {
		t.Fatalf("带伪造事件摘要的证据包未被拒绝: err=%v", importErr)
	}
	if target.HasBatch(ctx, batch.BatchID) || target.EventCount(ctx, batch.BatchID) != 0 {
		t.Fatalf("拒绝失败后仍污染了目标存储: batch=%t events=%d", target.HasBatch(ctx, batch.BatchID), target.EventCount(ctx, batch.BatchID))
	}
}
