package application

import "encoding/json"

func marshal(v any) ([]byte, error)     { return json.Marshal(v) }
func unmarshal(raw []byte, v any) error { return json.Unmarshal(raw, v) }
func cloneResult(r Result) (Result, error) {
	raw, e := marshal(r)
	if e != nil {
		return Result{}, e
	}
	var out Result
	e = unmarshal(raw, &out)
	return out, e
}
