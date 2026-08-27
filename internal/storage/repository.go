package storage

import (
	"archive-deacidification/internal/domain"
	"context"
)

type Repository interface {
	Get(context.Context, string) (domain.RestorationBatch, error)
	Save(context.Context, domain.RestorationBatch, string, any) (int64, error)
	SaveEvents(context.Context, domain.RestorationBatch, []EventWrite) ([]int64, error)
	Freeze(context.Context, *domain.RestorationBatch, string) (int64, error)
	Events(context.Context, string) ([]AuditEvent, error)
	VerifyChain(context.Context, string) (bool, error)
	GetIdempotent(context.Context, string, string) ([]byte, bool)
	PutIdempotent(context.Context, string, string, []byte) error
}

type EventWrite struct {
	Type    string
	Payload any
}
