package audit_event_payload_alias_test

import (
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"testing"
)

func TestAuditEventsCannotMutateStoredPayload(t *testing.T) {
	store, err := storage.Open("file:audit-event-payload-alias?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()
	batch := domain.RestorationBatch{BatchID: "payload-alias", Version: 1}
	if _, err := store.Save(ctx, batch, "sample.recorded", map[string]string{"sampleID": "S1"}); err != nil {
		t.Fatal(err)
	}
	events, err := store.Events(ctx, batch.BatchID)
	if err != nil || len(events) != 1 {
		t.Fatalf("events: %#v %v", events, err)
	}
	valid, err := store.VerifyChain(ctx, batch.BatchID)
	if err != nil || !valid {
		t.Fatalf("baseline verify: %t %v", valid, err)
	}
	if len(events[0].Payload) == 0 {
		t.Fatal("event payload is empty")
	}
	events[0].Payload[0] = '['

	valid, err = store.VerifyChain(ctx, batch.BatchID)
	if err != nil {
		t.Fatalf("verify returned error: %v", err)
	}
	if !valid {
		t.Fatalf("stored audit chain was changed through Events result")
	}
}
