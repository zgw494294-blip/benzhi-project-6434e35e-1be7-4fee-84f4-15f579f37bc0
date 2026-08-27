package httpapi

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	app *application.Service
	mux *http.ServeMux
}

func New(app *application.Service) *Server {
	s := &Server{app: app, mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) Handler() http.Handler { return s.mux }
func (s *Server) routes() {
	s.mux.HandleFunc("/v1/batches", s.batches)
	s.mux.HandleFunc("/v1/batches/", s.batchSubresource)
	s.mux.HandleFunc("/healthz", s.health)
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	write(w, 200, map[string]any{"status": "ok"})
}

type createReq struct {
	BatchID, Title, Institution, TargetProcess, CreatedBy string
	DocumentItems                                         []domain.DocumentItem `json:"documentItems"`
}

func (s *Server) batches(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErr(w, 405, "METHOD_NOT_ALLOWED", nil)
		return
	}
	var in createReq
	if !decode(r, &in) {
		writeErr(w, 400, "INVALID_JSON", nil)
		return
	}
	res, e := s.app.CreateBatch(r.Context(), application.CreateBatchInput{BatchID: in.BatchID, Title: in.Title, Institution: in.Institution, TargetProcess: in.TargetProcess, CreatedBy: in.CreatedBy, Items: in.DocumentItems}, r.Header.Get("Idempotency-Key"))
	if e != nil {
		mapErr(w, e)
		return
	}
	write(w, 201, res)
}
func (s *Server) batchSubresource(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/batches/"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeErr(w, 404, "NOT_FOUND", nil)
		return
	}
	id := parts[0]
	if len(parts) == 1 && r.Method == "GET" {
		res, e := s.app.Get(r.Context(), id)
		if e != nil {
			mapErr(w, e)
			return
		}
		write(w, 200, res)
		return
	}
	if len(parts) == 2 && parts[1] == "events" && r.Method == "GET" {
		ev, e := s.app.Events(r.Context(), id)
		if e != nil {
			mapErr(w, e)
			return
		}
		write(w, 200, map[string]any{"events": ev})
		return
	}
	if len(parts) == 2 && parts[1] == "summary" && r.Method == "GET" {
		s.summary(w, r, id)
		return
	}
	if len(parts) == 2 && parts[1] == "reports" && r.Method == "GET" {
		s.reports(w, r, id)
		return
	}
	if len(parts) == 2 && parts[1] == "release" && r.Method == "GET" {
		s.release(w, r, id)
		return
	}
	if len(parts) == 2 && parts[1] == "verify" && r.Method == "GET" {
		ok, e := s.app.Verify(r.Context(), id, r.URL.Query().Get("code"))
		if e != nil {
			mapErr(w, e)
			return
		}
		write(w, 200, map[string]any{"valid": ok})
		return
	}
	if r.Method != "POST" || len(parts) != 2 {
		writeErr(w, 404, "NOT_FOUND", nil)
		return
	}
	v := version(r)
	if v <= 0 {
		writeErr(w, http.StatusPreconditionRequired, "IF_MATCH_REQUIRED", "If-Match 必须包含当前批次版本")
		return
	}
	key := r.Header.Get("Idempotency-Key")
	switch parts[1] {
	case "samples":
		var x domain.Sample
		if !decode(r, &x) {
			writeErr(w, 400, "INVALID_JSON", nil)
			return
		}
		res, e := s.app.AddSample(r.Context(), id, v, x, key)
		respond(w, res, e, 200)
	case "trials":
		var x domain.ProcessTrial
		if !decode(r, &x) {
			writeErr(w, 400, "INVALID_JSON", nil)
			return
		}
		res, e := s.app.AddTrial(r.Context(), id, v, x, key)
		respond(w, res, e, 200)
	case "retests":
		var x domain.CorrectionRetest
		if !decode(r, &x) {
			writeErr(w, 400, "INVALID_JSON", nil)
			return
		}
		res, e := s.app.AddRetest(r.Context(), id, v, x, key)
		respond(w, res, e, 200)
	case "review":
		var x struct {
			Approve          bool `json:"approve"`
			Reviewer, Reason string
		}
		if !decode(r, &x) {
			writeErr(w, 400, "INVALID_JSON", nil)
			return
		}
		res, e := s.app.Review(r.Context(), id, v, x.Approve, x.Reviewer, x.Reason, key)
		respond(w, res, e, 200)
	case "freeze":
		var x struct {
			Reviewer string `json:"reviewer"`
		}
		if !decode(r, &x) {
			writeErr(w, 400, "INVALID_JSON", nil)
			return
		}
		res, e := s.app.Freeze(r.Context(), id, v, x.Reviewer, key)
		respond(w, res, e, 200)
	default:
		writeErr(w, 404, "NOT_FOUND", nil)
	}
}
func version(r *http.Request) int {
	v, _ := strconv.Atoi(strings.Trim(strings.TrimSpace(r.Header.Get("If-Match")), "\""))
	return v
}
func decode(r *http.Request, v any) bool { return json.NewDecoder(r.Body).Decode(v) == nil }
func respond(w http.ResponseWriter, res application.Result, e error, status int) {
	if e != nil {
		mapErr(w, e)
		return
	}
	write(w, status, res)
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, status int, code string, detail any) {
	write(w, status, map[string]any{"error": map[string]any{"code": code, "detail": detail}})
}
func mapErr(w http.ResponseWriter, e error) {
	status, code := 400, "INVALID_REQUEST"
	switch {
	case errors.Is(e, domain.ErrNotFound):
		status, code = 404, "NOT_FOUND"
	case errors.Is(e, domain.ErrVersionConflict):
		status, code = 409, "VERSION_CONFLICT"
	case errors.Is(e, domain.ErrFrozen):
		status, code = 409, "FROZEN"
	case errors.Is(e, domain.ErrUnauthorized):
		status, code = 403, "UNAUTHORIZED"
	case errors.Is(e, domain.ErrInvalidTransition):
		status, code = 409, "INVALID_TRANSITION"
	case errors.Is(e, domain.ErrDeviationOpen):
		status, code = 409, "DEVIATION_OPEN"
	case errors.Is(e, domain.ErrDocumentDuplicate):
		status, code = 409, "DOCUMENT_ITEM_DUPLICATE"
	case errors.Is(e, domain.ErrDocumentScope):
		status, code = 422, "DOCUMENT_SCOPE_INVALID"
	case errors.Is(e, domain.ErrSampleDuplicate):
		status, code = 409, "SAMPLE_ID_CONFLICT"
	case errors.Is(e, domain.ErrTrialDuplicate):
		status, code = 409, "TRIAL_ID_CONFLICT"
	case errors.Is(e, domain.ErrRetestDuplicate):
		status, code = 409, "RETEST_ID_CONFLICT"
	case errors.Is(e, domain.ErrSamplingCoverage):
		status, code = 409, "SAMPLING_COVERAGE_INSUFFICIENT"
	case errors.Is(e, domain.ErrAuditChain):
		status, code = 409, "AUDIT_CHAIN_INVALID"
	case errors.Is(e, domain.ErrInvalidMetrics):
		status, code = 422, "OBSERVATION_OUT_OF_RANGE"
	}
	writeErr(w, status, code, e.Error())
}
