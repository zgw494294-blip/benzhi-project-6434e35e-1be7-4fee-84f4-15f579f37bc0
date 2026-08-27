package domain

import "math"

type Observation struct {
	PH          float64 `json:"ph"`
	Strength    float64 `json:"strength"`
	Temperature float64 `json:"temperature"`
	Duration    int     `json:"duration"`
}

func (o Observation) Valid() bool {
	return o.PH > 0 && o.PH <= 14 && o.Strength > 0 && o.Temperature >= 0 && o.Duration > 0
}
func RoundMetric(v float64) float64 { return math.Round(v*1000) / 1000 }
func CompareMetric(before, after float64) float64 {
	if before == 0 {
		return 0
	}
	return RoundMetric(after / before)
}
func PHDelta(before, after float64) float64 { return RoundMetric(after - before) }
func MergeObservations(base, extra map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}
