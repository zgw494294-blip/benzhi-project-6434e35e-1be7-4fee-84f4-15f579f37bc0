package storage

import (
	"archive-deacidification/internal/domain"
	"context"
)

type Repository interface {
	Get(context.Context, string) (domain.RestorationBatch, error)
	Save(context.Context, domain.RestorationBatch, string, any) (int64, error)
	SaveEvents(context.Context, domain.RestorationBatch, []EventWrite) ([]int64, error)
	SaveEventsIdempotent(context.Context, domain.RestorationBatch, []EventWrite, string, func([]int64) []byte) ([]byte, error)
	MutateIdempotent(ctx context.Context, batchID string, key string, mutate func(domain.RestorationBatch) (domain.RestorationBatch, []EventWrite, error), buildResponse func([]int64, domain.RestorationBatch) []byte) ([]byte, error)
	Freeze(context.Context, *domain.RestorationBatch, string) (int64, error)
	FreezeIdempotent(ctx context.Context, batchID string, key string, mutate func(*domain.RestorationBatch) (string, error), buildResponse func(int64) []byte) ([]byte, error)
	Events(context.Context, string) ([]AuditEvent, error)
	VerifyChain(context.Context, string) (bool, error)
	GetIdempotent(context.Context, string, string) ([]byte, bool)
	PutIdempotent(context.Context, string, string, []byte) error
}

type EventWrite struct {
	Type    string
	Payload any
}
