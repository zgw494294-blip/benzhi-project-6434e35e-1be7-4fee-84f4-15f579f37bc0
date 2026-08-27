package httpapi

type RouteDoc struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

func RouteDocs() []RouteDoc {
	return []RouteDoc{{"POST", "/v1/batches", "建立修复批次并登记文献件"}, {"GET", "/v1/batches/{id}", "查询批次及冻结凭据"}, {"POST", "/v1/batches/{id}/samples", "登记代表性取样"}, {"POST", "/v1/batches/{id}/trials", "提交脱酸工艺试验"}, {"POST", "/v1/batches/{id}/retests", "提交偏差整改复验"}, {"POST", "/v1/batches/{id}/review", "批准或退回工艺结论"}, {"POST", "/v1/batches/{id}/freeze", "冻结证据并签发凭据"}, {"GET", "/v1/batches/{id}/events", "读取审计时间线"}, {"GET", "/v1/batches/{id}/verify", "验真放行凭据"}, {"GET", "/v1/batches/{id}/summary", "查看批次摘要"}, {"GET", "/v1/batches/{id}/reports", "查看试验指标报告"}, {"GET", "/v1/batches/{id}/release", "查看放行视图"}, {"GET", "/healthz", "服务健康检查"}}
}
func HeaderDocs() map[string]string {
	return map[string]string{"If-Match": "写入时的批次版本号", "Idempotency-Key": "避免重复提交的幂等键", "Content-Type": "请求体必须为 application/json"}
}
func ErrorDocs() map[string]string {
	return map[string]string{"VERSION_CONFLICT": "版本已变化，请重新读取批次", "DEVIATION_OPEN": "存在未完成的偏差复验", "FROZEN": "证据已冻结，禁止继续写入", "INVALID_TRANSITION": "当前状态不允许该操作", "NOT_FOUND": "批次或凭据不存在"}
}
