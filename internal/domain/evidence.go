package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func EvidenceDigest(b *RestorationBatch) (string, error) {
	clone := struct {
		BatchID       string
		Version       int
		Status        BatchStatus
		DocumentItems []DocumentItem
		Samples       []Sample
		Trials        []ProcessTrial
		Retests       []CorrectionRetest
	}{b.BatchID, b.Version, b.Status, b.DocumentItems, b.Samples, b.Trials, b.Retests}
	raw, e := json.Marshal(clone)
	if e != nil {
		return "", e
	}
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:]), nil
}
func VerificationCode(digest, credentialID string) string {
	h := sha256.Sum256([]byte(digest + ":" + credentialID))
	return hex.EncodeToString(h[:])[:20]
}
