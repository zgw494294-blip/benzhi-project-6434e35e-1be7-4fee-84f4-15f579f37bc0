package storage

import (
	"archive-deacidification/internal/domain"
	"context"
	"testing"
)

func TestAuditSequencesAreContinuousPerBatch(t *testing.T) {
	store, err := Open("file:audit?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	for _, id := range []string{"B1", "B2", "B1"} {
		batch := domain.RestorationBatch{BatchID: id, Version: 1}
		if _, err = store.Save(ctx, batch, "test.event", batch); err != nil {
			t.Fatal(err)
		}
	}
	events, err := store.Events(ctx, "B1")
	if err != nil || len(events) != 2 {
		t.Fatalf("events: %#v %v", events, err)
	}
	if events[0].Seq != 1 || events[1].Seq != 2 || events[1].PrevDigest != events[0].Digest {
		t.Fatalf("audit chain is not continuous: %#v", events)
	}
	valid, err := store.VerifyChain(ctx, "B1")
	if err != nil || !valid {
		t.Fatalf("verify chain: %t %v", valid, err)
	}
}
