package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func requireMethod(r *http.Request, w http.ResponseWriter, method string) bool {
	if r.Method != method {
		writeErr(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", fmt.Sprintf("需要 %s", method))
		return false
	}
	return true
}
func parseExpectedVersion(r *http.Request) (int, error) {
	raw := strings.TrimSpace(r.Header.Get("If-Match"))
	raw = strings.Trim(raw, "\"")
	if raw == "" {
		return 0, nil
	}
	v, e := strconv.Atoi(raw)
	if e != nil || v < 1 {
		return 0, fmt.Errorf("If-Match 版本无效")
	}
	return v, nil
}
func idempotencyKey(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("Idempotency-Key"))
}
func requireHeader(r *http.Request, name string) error {
	if strings.TrimSpace(r.Header.Get(name)) == "" {
		return fmt.Errorf("缺少请求头 %s", name)
	}
	return nil
}
func validPathID(id string) bool {
	if id == "" || len(id) > 100 {
		return false
	}
	for _, r := range id {
		if !(r == '-' || r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}
