package httpapi

import (
	"archive-deacidification/internal/domain"
	"net/http"
	"strings"
)

func (s *Server) summary(w http.ResponseWriter, r *http.Request, id string) {
	x, e := s.app.Summary(r.Context(), id)
	if e != nil {
		mapErr(w, e)
		return
	}
	write(w, 200, x)
}
func (s *Server) reports(w http.ResponseWriter, r *http.Request, id string) {
	x, e := s.app.TrialReport(r.Context(), id)
	if e != nil {
		mapErr(w, e)
		return
	}
	write(w, 200, x)
}
func (s *Server) release(w http.ResponseWriter, r *http.Request, id string) {
	x, e := s.app.ReleaseView(r.Context(), id)
	if e != nil {
		mapErr(w, e)
		return
	}
	write(w, 200, x)
}
func parseSubresource(path string) (string, []string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", parts
	}
	return parts[1], parts
}
func methodAllowed(r *http.Request, allowed ...string) bool {
	for _, m := range allowed {
		if r.Method == m {
			return true
		}
	}
	return false
}
func statusLabel(s domain.BatchStatus) string {
	switch s {
	case domain.StatusDraft:
		return "待建档"
	case domain.StatusSampling:
		return "取样中"
	case domain.StatusTrialing:
		return "试验中"
	case domain.StatusCorrection:
		return "整改中"
	case domain.StatusReview:
		return "复核中"
	case domain.StatusApproved:
		return "已批准"
	case domain.StatusFrozen:
		return "已冻结"
	case domain.StatusRejected:
		return "已退回"
	}
	return "未知"
}
func okStatus(w http.ResponseWriter) { write(w, 200, map[string]any{"status": "ok"}) }
