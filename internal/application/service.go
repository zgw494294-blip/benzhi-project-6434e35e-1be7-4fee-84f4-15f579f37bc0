package application

import (
	"archive-deacidification/internal/domain"
	"archive-deacidification/internal/storage"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Service struct{ repo storage.Repository }

func New(repo storage.Repository) *Service { return &Service{repo: repo} }

type Result struct {
	Batch      domain.RestorationBatch   `json:"batch"`
	Sequence   int64                     `json:"auditSequence"`
	Credential *domain.ReleaseCredential `json:"credential,omitempty"`
}
type CreateBatchInput struct {
	BatchID, Title, Institution, TargetProcess, CreatedBy string
	Items                                                 []domain.DocumentItem
}

func (s *Service) CreateBatch(ctx context.Context, in CreateBatchInput, key string) (Result, error) {
	b := domain.RestorationBatch{BatchID: in.BatchID, Title: in.Title, Institution: in.Institution, TargetProcess: in.TargetProcess, CreatedBy: in.CreatedBy, DocumentItems: in.Items, Status: domain.StatusDraft, Version: 1}
	if b.BatchID == "" {
		b.BatchID = newID()
	}
	domain.NormalizeBatch(&b)
	if b.Title == "" || b.CreatedBy == "" {
		return Result{}, domain.ErrRequiredEvidence
	}
	if e := domain.ValidateDocumentItems(b.DocumentItems); e != nil {
		return Result{}, e
	}
	if _, e := s.repo.Get(ctx, b.BatchID); e == nil {
		return Result{}, domain.ErrVersionConflict
	} else if !errors.Is(e, domain.ErrNotFound) {
		return Result{}, e
	}
	b.DocumentStats = domain.DocumentStatisticsFor(b.DocumentItems)
	b.CreatedAt = ptr(time.Now().UTC())
	return s.saveCreated(ctx, b, key)
}
func (s *Service) AddSample(ctx context.Context, id string, expected int, sample domain.Sample, key string) (Result, error) {
	sample.BatchID = id
	domain.NormalizeSample(&sample)
	mutate := func(b domain.RestorationBatch) (domain.RestorationBatch, []storage.EventWrite, error) {
		if e := checkVersion(b, expected); e != nil {
			return b, nil, e
		}
		if _, exists := domain.SampleByID(b, sample.SampleID); exists {
			return b, nil, fmt.Errorf("%w: %s", domain.ErrSampleDuplicate, sample.SampleID)
		}
		if e := domain.ValidateSample(sample); e != nil {
			return b, nil, e
		}
		if e := b.AddSample(sample); e != nil {
			return b, nil, e
		}
		return b, []storage.EventWrite{{Type: "sample.recorded", Payload: b}}, nil
	}
	return s.mutate(ctx, id, key, mutate)
}
func (s *Service) AddTrial(ctx context.Context, id string, expected int, t domain.ProcessTrial, key string) (Result, error) {
	t.BatchID = id
	domain.NormalizeTrial(&t)
	mutate := func(b domain.RestorationBatch) (domain.RestorationBatch, []storage.EventWrite, error) {
		if e := checkVersion(b, expected); e != nil {
			return b, nil, e
		}
		if _, exists := domain.TrialByID(b, t.TrialID); exists {
			return b, nil, fmt.Errorf("%w: %s", domain.ErrTrialDuplicate, t.TrialID)
		}
		if e := domain.ValidateTrial(t); e != nil {
			return b, nil, e
		}
		if e := b.AddTrial(t); e != nil {
			return b, nil, e
		}
		return b, []storage.EventWrite{{Type: "trial.analyzed", Payload: b}}, nil
	}
	return s.mutate(ctx, id, key, mutate)
}
func (s *Service) AddRetest(ctx context.Context, id string, expected int, r domain.CorrectionRetest, key string) (Result, error) {
	r.BatchID = id
	if strings.TrimSpace(r.RetestID) == "" {
		r.RetestID = newID()
	}
	domain.NormalizeRetest(&r)
	if r.SubmittedAt.IsZero() {
		r.SubmittedAt = time.Now().UTC()
	}
	mutate := func(b domain.RestorationBatch) (domain.RestorationBatch, []storage.EventWrite, error) {
		if e := checkVersion(b, expected); e != nil {
			return b, nil, e
		}
		for _, existing := range b.Retests {
			if existing.RetestID == r.RetestID {
				return b, nil, fmt.Errorf("%w: %s", domain.ErrRetestDuplicate, r.RetestID)
			}
		}
		if e := b.AddRetest(r); e != nil {
			return b, nil, e
		}
		return b, []storage.EventWrite{{Type: "retest.submitted", Payload: b}}, nil
	}
	return s.mutate(ctx, id, key, mutate)
}
func (s *Service) Review(ctx context.Context, id string, expected int, approve bool, reviewer, reason, key string) (Result, error) {
	if e := domain.ValidateActor(reviewer); e != nil {
		return Result{}, e
	}
	reviewer = strings.Join(strings.Fields(reviewer), " ")
	reason = strings.TrimSpace(reason)
	mutate := func(b domain.RestorationBatch) (domain.RestorationBatch, []storage.EventWrite, error) {
		if e := checkVersion(b, expected); e != nil {
			return b, nil, e
		}
		if e := domain.ValidateBatchForReview(&b); e != nil {
			return b, nil, e
		}
		if b.Status != domain.StatusReview {
			if e := b.Transition(domain.StatusReview); e != nil {
				return b, nil, e
			}
		}
		next := domain.StatusRejected
		if approve {
			next = domain.StatusApproved
		}
		if e := b.Transition(next); e != nil {
			return b, nil, e
		}
		payload := map[string]any{"reviewer": reviewer, "approve": approve, "reason": reason}
		return b, []storage.EventWrite{{Type: "review.completed", Payload: payload}}, nil
	}
	idempotencyKey := keyWithPayload(key, map[string]any{"reviewer": reviewer, "approve": approve, "reason": reason})
	return s.mutate(ctx, id, idempotencyKey, mutate)
}
func (s *Service) Freeze(ctx context.Context, id string, expected int, reviewer, key string) (Result, error) {
	if e := domain.ValidateActor(reviewer); e != nil {
		return Result{}, e
	}
	reviewer = strings.Join(strings.Fields(reviewer), " ")
	var finalBatch *domain.RestorationBatch
	mutate := func(b *domain.RestorationBatch) (string, error) {
		if e := checkVersion(*b, expected); e != nil {
			return "", e
		}
		if b.Status != domain.StatusApproved {
			return "", domain.ErrInvalidTransition
		}
		if e := b.Transition(domain.StatusFrozen); e != nil {
			return "", e
		}
		digest, e := domain.EvidenceDigest(b)
		if e != nil {
			return "", e
		}
		cred := &domain.ReleaseCredential{CredentialID: newID(), BatchID: id, EvidenceDigest: digest, Decision: "approved", Reviewer: reviewer, IssuedAt: time.Now().UTC()}
		cred.VerificationCode = domain.VerificationCode(digest, cred.CredentialID)
		b.ReleasedAt = &cred.IssuedAt
		b.Credential = cred
		finalBatch = b
		return "release.frozen", nil
	}
	raw, e := s.repo.FreezeIdempotent(ctx, id, key, mutate, func(seq int64) []byte {
		r := Result{Sequence: seq}
		if finalBatch != nil {
			r.Batch = *finalBatch
			r.Credential = finalBatch.Credential
		}
		out, _ := json.Marshal(r)
		return out
	})
	if e != nil {
		return Result{}, e
	}
	var r Result
	if json.Unmarshal(raw, &r) != nil {
		return Result{}, fmt.Errorf("internal: failed to decode idempotent response")
	}
	return r, nil
}
func (s *Service) Get(ctx context.Context, id string) (Result, error) {
	b, e := s.repo.Get(ctx, id)
	return Result{Batch: b, Credential: b.Credential}, e
}
func (s *Service) Events(ctx context.Context, id string) ([]storage.AuditEvent, error) {
	if _, e := s.repo.Get(ctx, id); e != nil {
		return nil, e
	}
	return s.repo.Events(ctx, id)
}
func (s *Service) Verify(ctx context.Context, id, code string) (bool, error) {
	b, e := s.repo.Get(ctx, id)
	if e != nil {
		return false, e
	}
	if b.Credential == nil {
		return false, nil
	}
	digest, e := domain.EvidenceDigest(&b)
	if e != nil {
		return false, e
	}
	chainValid, e := s.repo.VerifyChain(ctx, id)
	if e != nil {
		return false, e
	}
	return chainValid && b.Credential.EvidenceDigest == digest && b.Credential.VerificationCode == code, nil
}
func (s *Service) loadVersion(ctx context.Context, id string, v int) (domain.RestorationBatch, error) {
	b, e := s.repo.Get(ctx, id)
	if e != nil {
		return b, e
	}
	if b.Status == domain.StatusFrozen {
		return b, domain.ErrFrozen
	}
	if v > 0 && b.Version != v {
		return b, domain.ErrVersionConflict
	}
	return b, nil
}
func checkVersion(b domain.RestorationBatch, v int) error {
	if v > 0 && b.Version != v {
		return domain.ErrVersionConflict
	}
	return nil
}
func (s *Service) mutate(ctx context.Context, id, key string, mutate func(domain.RestorationBatch) (domain.RestorationBatch, []storage.EventWrite, error)) (Result, error) {
	buildResponse := func(sequences []int64, b domain.RestorationBatch) []byte {
		seq := int64(0)
		if len(sequences) > 0 {
			seq = sequences[len(sequences)-1]
		}
		r := Result{Batch: b, Sequence: seq, Credential: b.Credential}
		raw, _ := json.Marshal(r)
		return raw
	}
	raw, e := s.repo.MutateIdempotent(ctx, id, key, mutate, buildResponse)
	if e != nil {
		return Result{}, e
	}
	var r Result
	if json.Unmarshal(raw, &r) != nil {
		return Result{}, fmt.Errorf("internal: failed to decode idempotent response")
	}
	return r, nil
}

func (s *Service) saveCreated(ctx context.Context, b domain.RestorationBatch, key string) (Result, error) {
	writes := []storage.EventWrite{
		{Type: "batch.created", Payload: b},
		{Type: "batch.scope.validated", Payload: map[string]any{
			"version": b.Version, "valid": true, "statistics": b.DocumentStats,
		}},
	}
	buildResponse := func(sequences []int64) []byte {
		r := Result{Batch: b}
		if len(sequences) > 0 {
			r.Sequence = sequences[len(sequences)-1]
		}
		raw, _ := json.Marshal(r)
		return raw
	}
	raw, e := s.repo.SaveEventsIdempotent(ctx, b, writes, key, buildResponse)
	if e != nil {
		return Result{}, e
	}
	var r Result
	if json.Unmarshal(raw, &r) != nil {
		return Result{}, fmt.Errorf("internal: failed to decode idempotent response")
	}
	return r, nil
}
func keyWithPayload(k string, v any) string {
	if k == "" {
		return ""
	}
	return fmt.Sprintf("%s:%v", k, v)
}
func ptr(t time.Time) *time.Time { return &t }
func newID() string              { b := make([]byte, 8); _, _ = rand.Read(b); return hex.EncodeToString(b) }
