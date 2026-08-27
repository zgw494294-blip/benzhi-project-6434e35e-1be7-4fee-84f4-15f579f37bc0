package storage

import (
	"archive-deacidification/internal/domain"
	"context"
	"encoding/json"
	"os"
	"time"
)

type BackupManifest struct {
	CreatedAt  time.Time `json:"createdAt"`
	BatchCount int       `json:"batchCount"`
	EventCount int       `json:"eventCount"`
}

func (s *Store) Backup(ctx context.Context, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, e := json.Marshal(s.data)
	if e != nil {
		return e
	}
	return os.WriteFile(path, raw, 0600)
}
func (s *Store) BackupManifest(ctx context.Context) BackupManifest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return BackupManifest{CreatedAt: time.Now().UTC(), BatchCount: len(s.data.Batches), EventCount: len(s.data.Events)}
}
func (s *Store) Restore(ctx context.Context, path string) error {
	raw, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	var d fileData
	decodeErr := json.Unmarshal(raw, &d)
	if d.Batches == nil {
		d.Batches = map[string]domain.RestorationBatch{}
	}
	if d.Idempotency == nil {
		d.Idempotency = map[string][]byte{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = d
	if decodeErr != nil {
		return decodeErr
	}
	return s.persist()
}
