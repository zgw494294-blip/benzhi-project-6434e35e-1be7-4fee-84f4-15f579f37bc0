package domain

type ThresholdProfile struct {
	Name                 string  `json:"name"`
	MinPHChange          float64 `json:"minPHChange"`
	MinStrengthRetention float64 `json:"minStrengthRetention"`
}

var DefaultThresholds = ThresholdProfile{Name: "档案纸张标准", MinPHChange: MinPHChange, MinStrengthRetention: MinStrengthRetention}

func AnalyzeWithProfile(initialPH, finalPH, before, after float64, p ThresholdProfile) (Metrics, error) {
	if initialPH <= 0 || finalPH <= 0 || before <= 0 || after < 0 {
		return Metrics{}, ErrInvalidMetrics
	}
	m := Metrics{PHChange: finalPH - initialPH, StrengthRetention: after / before, Passed: true}
	if m.PHChange < p.MinPHChange {
		m.Deviations = append(m.Deviations, "PH_CHANGE_LOW")
	}
	if m.StrengthRetention < p.MinStrengthRetention {
		m.Deviations = append(m.Deviations, "STRENGTH_RETENTION_LOW")
	}
	m.Passed = len(m.Deviations) == 0
	return m, nil
}
func DeviationDescription(code string) string {
	switch code {
	case "PH_CHANGE_LOW":
		return "脱酸后酸碱度提升不足"
	case "STRENGTH_RETENTION_LOW":
		return "纸张强度保持率低于阈值"
	default:
		return "未识别的质量偏差"
	}
}
func DeviationSeverity(code string) string {
	if code == "STRENGTH_RETENTION_LOW" {
		return "high"
	}
	return "medium"
}
