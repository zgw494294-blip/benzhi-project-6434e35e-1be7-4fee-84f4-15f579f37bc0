package domain

import "sort"

type MaterialCoverage struct {
	Material string `json:"material"`
	Required int    `json:"required"`
	Observed int    `json:"observed"`
	Covered  bool   `json:"covered"`
}

type SamplingCoverage struct {
	RequiredMaterials  int                `json:"requiredMaterials"`
	CoveredMaterials   int                `json:"coveredMaterials"`
	Rate               float64            `json:"rate"`
	MinimumPerMaterial int                `json:"minimumPerMaterial"`
	Satisfied          bool               `json:"satisfied"`
	Materials          []MaterialCoverage `json:"materials"`
}

type BatchSummary struct {
	BatchID          string           `json:"batchID"`
	Status           BatchStatus      `json:"status"`
	DocumentCount    int              `json:"documentCount"`
	TotalPages       int              `json:"totalPages"`
	MaterialCount    map[string]int   `json:"materialCount"`
	SamplingCoverage SamplingCoverage `json:"samplingCoverage"`
	SampleCount      int              `json:"sampleCount"`
	TrialCount       int              `json:"trialCount"`
	RetestCount      int              `json:"retestCount"`
	PassedTrials     int              `json:"passedTrials"`
	OpenDeviations   int              `json:"openDeviations"`
	Ready            bool             `json:"ready"`
}

func Summarize(b RestorationBatch) BatchSummary {
	stats := DocumentStatisticsFor(b.DocumentItems)
	x := BatchSummary{
		BatchID: b.BatchID, Status: b.Status, DocumentCount: stats.ItemCount,
		TotalPages: stats.TotalPages, MaterialCount: stats.MaterialCount,
		SamplingCoverage: CalculateSamplingCoverage(b), SampleCount: len(b.Samples),
		TrialCount: len(b.Trials), RetestCount: len(b.Retests),
	}
	for _, t := range b.Trials {
		if len(t.DeviationCodes) == 0 {
			x.PassedTrials++
		} else if !HasClosedRetest(&b, t.TrialID) {
			x.OpenDeviations += len(t.DeviationCodes)
		}
	}
	x.Ready = ReadyForFreeze(b)
	return x
}

func CalculateSamplingCoverage(b RestorationBatch) SamplingCoverage {
	required := map[string]struct{}{}
	observed := map[string]int{}
	for _, item := range b.DocumentItems {
		required[item.Material] = struct{}{}
	}
	for _, sample := range b.Samples {
		observed[sample.Material]++
	}
	materials := make([]string, 0, len(required))
	for material := range required {
		materials = append(materials, material)
	}
	sort.Strings(materials)
	coverage := SamplingCoverage{RequiredMaterials: len(materials), MinimumPerMaterial: MinSamplesPerMaterial}
	for _, material := range materials {
		count := observed[material]
		covered := count >= MinSamplesPerMaterial
		if covered {
			coverage.CoveredMaterials++
		}
		coverage.Materials = append(coverage.Materials, MaterialCoverage{Material: material, Required: MinSamplesPerMaterial, Observed: count, Covered: covered})
	}
	if coverage.RequiredMaterials > 0 {
		coverage.Rate = RoundMetric(float64(coverage.CoveredMaterials) / float64(coverage.RequiredMaterials))
	}
	coverage.Satisfied = coverage.RequiredMaterials > 0 && coverage.CoveredMaterials == coverage.RequiredMaterials
	return coverage
}

func SortedTrialIDs(b RestorationBatch) []string {
	out := make([]string, 0, len(b.Trials))
	for _, t := range b.Trials {
		out = append(out, t.TrialID)
	}
	sort.Strings(out)
	return out
}

func TrialByID(b RestorationBatch, id string) (ProcessTrial, bool) {
	for _, t := range b.Trials {
		if t.TrialID == id {
			return t, true
		}
	}
	return ProcessTrial{}, false
}

func SampleByID(b RestorationBatch, id string) (Sample, bool) {
	for _, s := range b.Samples {
		if s.SampleID == id {
			return s, true
		}
	}
	return Sample{}, false
}

func DocumentByID(b RestorationBatch, id string) (DocumentItem, bool) {
	for _, d := range b.DocumentItems {
		if d.ItemID == id {
			return d, true
		}
	}
	return DocumentItem{}, false
}
