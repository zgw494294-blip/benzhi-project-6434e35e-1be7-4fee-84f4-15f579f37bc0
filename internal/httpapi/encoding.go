package httpapi

import (
	"encoding/json"
	"net/http"
)

func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
func writeNoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }
