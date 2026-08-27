package domain

import "errors"

var (
	ErrInvalidTransition = errors.New("非法状态转换")
	ErrRequiredEvidence  = errors.New("缺少必需证据")
	ErrDeviationOpen     = errors.New("仍有未闭环偏差")
	ErrFrozen            = errors.New("批次已冻结")
	ErrInvalidMetrics    = errors.New("观测指标无效")
	ErrNotFound          = errors.New("资源不存在")
	ErrVersionConflict   = errors.New("版本冲突")
	ErrUnauthorized      = errors.New("缺少负责人身份")
	ErrDocumentDuplicate = errors.New("文献清单存在重复项")
	ErrDocumentScope     = errors.New("文献处理范围无效")
	ErrSampleDuplicate   = errors.New("样本编号已存在")
	ErrTrialDuplicate    = errors.New("试验编号已存在")
	ErrRetestDuplicate   = errors.New("复验编号已存在")
	ErrSamplingCoverage  = errors.New("代表性取样覆盖不足")
	ErrAuditChain        = errors.New("审计链验真失败")
)
