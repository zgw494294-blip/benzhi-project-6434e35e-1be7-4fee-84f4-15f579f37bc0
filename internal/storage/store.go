package storage

import (
	"archive-deacidification/internal/domain"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
)

type Store struct {
	path string
	mu   sync.Mutex
	data fileData
}
type fileData struct {
	Batches     map[string]domain.RestorationBatch `json:"batches"`
	Events      []AuditEvent                       `json:"events"`
	Idempotency map[string][]byte                  `json:"idempotency"`
}
type AuditEvent struct {
	Seq        int64           `json:"seq"`
	BatchID    string          `json:"batchID"`
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	PrevDigest string          `json:"prevDigest"`
	Digest     string          `json:"digest"`
	CreatedAt  time.Time       `json:"createdAt"`
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, data: fileData{Batches: map[string]domain.RestorationBatch{}, Idempotency: map[string][]byte{}}}
	if path != "" && !strings.HasPrefix(path, "file:") {
		if raw, e := os.ReadFile(path); e == nil {
			_ = json.Unmarshal(raw, &s.data)
		}
	}
	return s, nil
}
func (s *Store) persist() error {
	if s.path == "" || strings.HasPrefix(s.path, "file:") {
		return nil
	}
	raw, e := json.Marshal(s.data)
	if e != nil {
		return e
	}
	tmp := s.path + ".wal"
	if e = os.WriteFile(tmp, raw, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, s.path)
}
func (s *Store) Close() error { s.mu.Lock(); defer s.mu.Unlock(); return s.persist() }

// cloneBatch 返回 b 的深拷贝，使调用方对切片、映射和凭据指针的修改
// 不会影响存储内部的冻结状态或后续读取与凭据验真。
func cloneBatch(b domain.RestorationBatch) domain.RestorationBatch {
	out := b
	if b.DocumentItems != nil {
		out.DocumentItems = make([]domain.DocumentItem, len(b.DocumentItems))
		copy(out.DocumentItems, b.DocumentItems)
	}
	if b.Samples != nil {
		out.Samples = make([]domain.Sample, len(b.Samples))
		copy(out.Samples, b.Samples)
	}
	if b.Trials != nil {
		out.Trials = make([]domain.ProcessTrial, len(b.Trials))
		for i := range b.Trials {
			t := b.Trials[i]
			if t.MetricDeviations != nil {
				t.MetricDeviations = append([]string(nil), t.MetricDeviations...)
			}
			if t.ParameterDeviations != nil {
				t.ParameterDeviations = append([]string(nil), t.ParameterDeviations...)
			}
			if t.DeviationCodes != nil {
				t.DeviationCodes = append([]string(nil), t.DeviationCodes...)
			}
			out.Trials[i] = t
		}
	}
	if b.Retests != nil {
		out.Retests = make([]domain.CorrectionRetest, len(b.Retests))
		for i := range b.Retests {
			r := b.Retests[i]
			if r.Observations != nil {
				r.Observations = cloneFloatMap(r.Observations)
			}
			out.Retests[i] = r
		}
	}
	if b.DocumentStats.MaterialCount != nil {
		out.DocumentStats.MaterialCount = cloneIntMap(b.DocumentStats.MaterialCount)
	}
	if b.CreatedAt != nil {
		c := *b.CreatedAt
		out.CreatedAt = &c
	}
	if b.ReleasedAt != nil {
		c := *b.ReleasedAt
		out.ReleasedAt = &c
	}
	if b.Credential != nil {
		c := *b.Credential
		out.Credential = &c
	}
	return out
}

func cloneIntMap(m map[string]int) map[string]int {
	out := make(map[string]int, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneFloatMap(m map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
func (s *Store) Get(ctx context.Context, id string) (domain.RestorationBatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.data.Batches[id]
	if !ok {
		return domain.RestorationBatch{}, domain.ErrNotFound
	}
	return cloneBatch(b), nil
}
func (s *Store) Save(ctx context.Context, b domain.RestorationBatch, eventType string, payload any) (int64, error) {
	sequences, err := s.SaveEvents(ctx, b, []EventWrite{{Type: eventType, Payload: payload}})
	if err != nil {
		return 0, err
	}
	return sequences[0], nil
}

func (s *Store) SaveEvents(ctx context.Context, b domain.RestorationBatch, writes []EventWrite) ([]int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(writes) == 0 {
		return nil, domain.ErrRequiredEvidence
	}
	previous, existed := s.data.Batches[b.BatchID]
	previousEventCount := len(s.data.Events)
	s.data.Batches[b.BatchID] = b
	sequences := make([]int64, 0, len(writes))
	for _, write := range writes {
		ev, err := s.appendEventLocked(b.BatchID, write.Type, write.Payload)
		if err != nil {
			s.rollbackBatchEventsLocked(b.BatchID, previous, existed, previousEventCount)
			return nil, err
		}
		sequences = append(sequences, ev.Seq)
	}
	if err := s.persist(); err != nil {
		s.rollbackBatchEventsLocked(b.BatchID, previous, existed, previousEventCount)
		return nil, err
	}
	return sequences, nil
}

func (s *Store) rollbackBatchEventsLocked(batchID string, previous domain.RestorationBatch, existed bool, eventCount int) {
	if existed {
		s.data.Batches[batchID] = previous
	} else {
		delete(s.data.Batches, batchID)
	}
	s.data.Events = s.data.Events[:eventCount]
}

func (s *Store) Freeze(ctx context.Context, b *domain.RestorationBatch, eventType string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b == nil || b.Credential == nil {
		return 0, domain.ErrRequiredEvidence
	}
	current, ok := s.data.Batches[b.BatchID]
	if !ok {
		return 0, domain.ErrNotFound
	}
	if current.Version+1 != b.Version {
		return 0, domain.ErrVersionConflict
	}
	if err := s.verifyChainLocked(b.BatchID); err != nil {
		return 0, err
	}
	events := s.eventsLocked(b.BatchID)
	b.Credential.AuditSequence = int64(len(events) + 1)
	previousEventCount := len(s.data.Events)
	s.data.Batches[b.BatchID] = *b
	ev, err := s.appendEventLocked(b.BatchID, eventType, *b)
	if err != nil {
		s.data.Batches[b.BatchID] = current
		return 0, err
	}
	if err = s.persist(); err != nil {
		s.data.Batches[b.BatchID] = current
		s.data.Events = s.data.Events[:previousEventCount]
		return 0, err
	}
	return ev.Seq, nil
}

func (s *Store) appendEventLocked(batchID, eventType string, payload any) (AuditEvent, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return AuditEvent{}, err
	}
	events := s.eventsLocked(batchID)
	prev := ""
	if len(events) > 0 {
		prev = events[len(events)-1].Digest
	}
	ev := AuditEvent{Seq: int64(len(events) + 1), BatchID: batchID, Type: eventType, Payload: p, PrevDigest: prev, CreatedAt: time.Now().UTC()}
	ev.Digest = eventDigest(ev)
	s.data.Events = append(s.data.Events, ev)
	return ev, nil
}

func (s *Store) eventsLocked(id string) []AuditEvent {
	out := []AuditEvent{}
	for _, event := range s.data.Events {
		if event.BatchID == id {
			out = append(out, event)
		}
	}
	return out
}

func eventDigest(event AuditEvent) string {
	raw, _ := json.Marshal(struct {
		Sequence       int64           `json:"sequence"`
		BatchID        string          `json:"batchID"`
		Type           string          `json:"type"`
		Payload        json.RawMessage `json:"payload"`
		PreviousDigest string          `json:"previousDigest"`
		CreatedAt      time.Time       `json:"createdAt"`
	}{event.Seq, event.BatchID, event.Type, event.Payload, event.PrevDigest, event.CreatedAt})
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}
func (s *Store) Events(ctx context.Context, id string) ([]AuditEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.eventsLocked(id), nil
}
func (s *Store) GetIdempotent(ctx context.Context, batch, key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data.Idempotency[batch+"|"+key]
	return v, ok
}
func (s *Store) PutIdempotent(ctx context.Context, batch, key string, response []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Idempotency[batch+"|"+key] = response
	return s.persist()
}
func (s *Store) Health(ctx context.Context) error { return nil }
func (s *Store) Count(ctx context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data.Batches), nil
}
