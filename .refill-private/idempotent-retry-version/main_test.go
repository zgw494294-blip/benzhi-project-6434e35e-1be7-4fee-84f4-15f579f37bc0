package idempotent_retry_version

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/httpapi"
	"archive-deacidification/internal/storage"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func post(t *testing.T, handler http.Handler, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestIdempotentRetryReturnsOriginalResponse(t *testing.T) {
	store, err := storage.Open("file:idempotent-retry?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	handler := httpapi.JSON(httpapi.New(application.New(store)).Handler())
	created := post(t, handler, "/v1/batches", `{
		"batchID":"idempotent-batch","title":"幂等重试","createdBy":"技术员",
		"documentItems":[{"itemID":"D1","callNumber":"C1","title":"文献","material":"棉纸","pageCount":10}]
	}`, nil)
	if created.Code != http.StatusCreated {
		t.Fatalf("创建批次失败: status=%d body=%s", created.Code, created.Body.String())
	}

	body := `{"sampleID":"S1","material":"棉纸","initialPH":5,"initialStrength":100,"observer":"检测员","observedAt":"2026-01-01T00:00:00Z"}`
	headers := map[string]string{"If-Match": "1", "Idempotency-Key": "record-sample-once"}
	first := post(t, handler, "/v1/batches/idempotent-batch/samples", body, headers)
	if first.Code != http.StatusOK {
		t.Fatalf("首次请求失败: status=%d body=%s", first.Code, first.Body.String())
	}
	retry := post(t, handler, "/v1/batches/idempotent-batch/samples", body, headers)
	if retry.Code != first.Code || retry.Body.String() != first.Body.String() {
		t.Fatalf("相同 Idempotency-Key 的精确重试未返回原响应: first=(%d,%s) retry=(%d,%s)", first.Code, first.Body.String(), retry.Code, retry.Body.String())
	}
}
