package domain

import "fmt"

type ProcessParameters struct {
	Concentration      float64 `json:"concentration"`
	TemperatureCelsius float64 `json:"temperatureCelsius"`
	DurationMinutes    int     `json:"durationMinutes"`
}

type MetricStatistics struct {
	Minimum float64 `json:"minimum"`
	Maximum float64 `json:"maximum"`
	Average float64 `json:"average"`
}

type TrialReport struct {
	Sequence            int                `json:"sequence"`
	TrialID             string             `json:"trialID"`
	SampleID            string             `json:"sampleID"`
	Reagent             string             `json:"reagent"`
	Parameters          ProcessParameters  `json:"parameters"`
	PHChange            float64            `json:"phChange"`
	StrengthRetention   float64            `json:"strengthRetention"`
	Passed              bool               `json:"passed"`
	MetricDeviations    []string           `json:"metricDeviations"`
	ParameterDeviations []string           `json:"parameterDeviations"`
	Deviations          []string           `json:"deviations"`
	Retests             []CorrectionRetest `json:"retests"`
	DeviationClosed     bool               `json:"deviationClosed"`
}

type TrialGroupReport struct {
	SampleID          string            `json:"sampleID"`
	Reagent           string            `json:"reagent"`
	Parameters        ProcessParameters `json:"parameters"`
	TrialIDs          []string          `json:"trialIDs"`
	PHChange          MetricStatistics  `json:"phChange"`
	StrengthRetention MetricStatistics  `json:"strengthRetention"`
	Passed            int               `json:"passed"`
	PassRate          float64           `json:"passRate"`
}

type TrialTrend struct {
	Sequence          int     `json:"sequence"`
	TrialID           string  `json:"trialID"`
	PHChange          float64 `json:"phChange"`
	StrengthRetention float64 `json:"strengthRetention"`
	DeviationCount    int     `json:"deviationCount"`
	DeviationOpen     bool    `json:"deviationOpen"`
}

type ComparisonReport struct {
	BatchID            string             `json:"batchID"`
	TrialCount         int                `json:"trialCount"`
	PassedCount        int                `json:"passedCount"`
	PassRate           float64            `json:"passRate"`
	OpenDeviationCount int                `json:"openDeviationCount"`
	Groups             []TrialGroupReport `json:"groups"`
	Reports            []TrialReport      `json:"reports"`
	Trends             []TrialTrend       `json:"trends"`
}

func BuildComparisonReport(b RestorationBatch) (ComparisonReport, error) {
	report := ComparisonReport{BatchID: b.BatchID, TrialCount: len(b.Trials)}
	groupIndex := map[string]int{}
	groupPH := map[string][]float64{}
	groupStrength := map[string][]float64{}
	for i, trial := range b.Trials {
		metrics, err := Analyze(trial.InitialPH, trial.FinalPH, trial.StrengthBefore, trial.StrengthAfter)
		if err != nil {
			return ComparisonReport{}, err
		}
		parameters := ProcessParameters{Concentration: trial.Concentration, TemperatureCelsius: trial.TemperatureCelsius, DurationMinutes: trial.DurationMinutes}
		metricDeviations := trial.MetricDeviations
		if metricDeviations == nil {
			metricDeviations = metrics.Deviations
		}
		parameterDeviations := trial.ParameterDeviations
		if parameterDeviations == nil {
			parameterDeviations = AnalyzeParameters(trial)
		}
		deviations := DeviationCodesUnique(append(append([]string(nil), metricDeviations...), parameterDeviations...))
		closed := len(deviations) == 0 || HasClosedRetest(&b, trial.TrialID)
		trialReport := TrialReport{
			Sequence: i + 1, TrialID: trial.TrialID, SampleID: trial.SampleID,
			Reagent: trial.Reagent, Parameters: parameters,
			PHChange: RoundMetric(metrics.PHChange), StrengthRetention: RoundMetric(metrics.StrengthRetention),
			Passed: len(deviations) == 0, MetricDeviations: metricDeviations,
			ParameterDeviations: parameterDeviations, Deviations: deviations,
			Retests: retestsForTrial(b, trial.TrialID), DeviationClosed: closed,
		}
		report.Reports = append(report.Reports, trialReport)
		report.Trends = append(report.Trends, TrialTrend{
			Sequence: i + 1, TrialID: trial.TrialID, PHChange: trialReport.PHChange,
			StrengthRetention: trialReport.StrengthRetention,
			DeviationCount:    len(deviations), DeviationOpen: len(deviations) > 0 && !closed,
		})
		if trialReport.Passed {
			report.PassedCount++
		} else if !closed {
			report.OpenDeviationCount += len(deviations)
		}
		key := fmt.Sprintf("%s\x00%s\x00%.9g\x00%.9g\x00%d", trial.SampleID, trial.Reagent, trial.Concentration, trial.TemperatureCelsius, trial.DurationMinutes)
		index, exists := groupIndex[key]
		if !exists {
			index = len(report.Groups)
			groupIndex[key] = index
			report.Groups = append(report.Groups, TrialGroupReport{SampleID: trial.SampleID, Reagent: trial.Reagent, Parameters: parameters})
		}
		group := &report.Groups[index]
		group.TrialIDs = append(group.TrialIDs, trial.TrialID)
		if trialReport.Passed {
			group.Passed++
		}
		groupPH[key] = append(groupPH[key], trialReport.PHChange)
		groupStrength[key] = append(groupStrength[key], trialReport.StrengthRetention)
	}
	if report.TrialCount > 0 {
		report.PassRate = RoundMetric(float64(report.PassedCount) / float64(report.TrialCount))
	}
	for key, index := range groupIndex {
		group := &report.Groups[index]
		group.PHChange = metricStatistics(groupPH[key])
		group.StrengthRetention = metricStatistics(groupStrength[key])
		group.PassRate = RoundMetric(float64(group.Passed) / float64(len(group.TrialIDs)))
	}
	return report, nil
}

func BuildTrialReport(t ProcessTrial) (TrialReport, error) {
	b := RestorationBatch{Trials: []ProcessTrial{t}}
	report, err := BuildComparisonReport(b)
	if err != nil {
		return TrialReport{}, err
	}
	return report.Reports[0], nil
}

func metricStatistics(values []float64) MetricStatistics {
	if len(values) == 0 {
		return MetricStatistics{}
	}
	stats := MetricStatistics{Minimum: values[0], Maximum: values[0]}
	for _, value := range values {
		if value < stats.Minimum {
			stats.Minimum = value
		}
		if value > stats.Maximum {
			stats.Maximum = value
		}
		stats.Average += value
	}
	stats.Average = RoundMetric(stats.Average / float64(len(values)))
	return stats
}

func retestsForTrial(b RestorationBatch, trialID string) []CorrectionRetest {
	out := []CorrectionRetest{}
	for _, retest := range b.Retests {
		if retest.TrialID == trialID {
			out = append(out, retest)
		}
	}
	return out
}
