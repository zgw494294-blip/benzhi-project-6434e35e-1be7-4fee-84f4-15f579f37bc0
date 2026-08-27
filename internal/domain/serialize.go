package domain

import "encoding/json"

func MarshalBatch(b RestorationBatch) ([]byte, error) { return json.Marshal(b) }
func UnmarshalBatch(raw []byte) (RestorationBatch, error) {
	var b RestorationBatch
	e := json.Unmarshal(raw, &b)
	return b, e
}
