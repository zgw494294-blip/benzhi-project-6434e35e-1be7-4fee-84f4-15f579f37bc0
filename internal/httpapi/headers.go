package httpapi

import "net/http"

func setVersion(w http.ResponseWriter, v int) { w.Header().Set("ETag", `"`+itoa(v)+`"`) }
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	b := make([]byte, 0, 12)
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	return string(b)
}
