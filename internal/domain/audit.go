package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type AuditRecord struct {
	Sequence       int64  `json:"sequence"`
	EventType      string `json:"eventType"`
	BatchID        string `json:"batchID"`
	Payload        any    `json:"payload"`
	PreviousDigest string `json:"previousDigest"`
	Digest         string `json:"digest"`
}

func NewAuditRecord(seq int64, batch, event, prev string, payload any) AuditRecord {
	r := AuditRecord{Sequence: seq, BatchID: batch, EventType: event, PreviousDigest: prev, Payload: payload}
	raw, _ := json.Marshal(struct {
		P string
		V any
	}{prev, payload})
	h := sha256.Sum256(raw)
	r.Digest = hex.EncodeToString(h[:])
	return r
}
func AuditDigest(records []AuditRecord) string {
	prev := ""
	for _, r := range records {
		prev = r.Digest
	}
	return prev
}
func AuditContiguous(records []AuditRecord) bool {
	for i, r := range records {
		if r.Sequence != int64(i+1) {
			return false
		}
		if i > 0 && r.PreviousDigest != records[i-1].Digest {
			return false
		}
	}
	return true
}
