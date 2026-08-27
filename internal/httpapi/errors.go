package httpapi

type ErrorBody struct {
	Code   string `json:"code"`
	Detail any    `json:"detail,omitempty"`
}

func errorBody(code string, detail any) map[string]any {
	return map[string]any{"error": ErrorBody{Code: code, Detail: detail}}
}
