package domain

import (
	"fmt"
	"strings"
)

func ValidateDocumentItems(items []DocumentItem) error {
	if len(items) == 0 {
		return ErrRequiredEvidence
	}
	itemIDs := make(map[string]struct{}, len(items))
	callNumbers := make(map[string]struct{}, len(items))
	for index, i := range items {
		if strings.TrimSpace(i.ItemID) == "" || strings.TrimSpace(i.CallNumber) == "" || strings.TrimSpace(i.Title) == "" || strings.TrimSpace(i.Material) == "" || i.PageCount <= 0 {
			return fmt.Errorf("%w: 第 %d 条缺少 itemID、callNumber、title、material 或有效 pageCount", ErrDocumentScope, index+1)
		}
		if _, ok := itemIDs[i.ItemID]; ok {
			return fmt.Errorf("%w: itemID %q", ErrDocumentDuplicate, i.ItemID)
		}
		if _, ok := callNumbers[i.CallNumber]; ok {
			return fmt.Errorf("%w: callNumber %q", ErrDocumentDuplicate, i.CallNumber)
		}
		itemIDs[i.ItemID] = struct{}{}
		callNumbers[i.CallNumber] = struct{}{}
	}
	return nil
}
func ValidateSample(s Sample) error {
	if s.SampleID == "" || s.BatchID == "" || s.Material == "" || s.Observer == "" || s.ObservedAt.IsZero() {
		return ErrRequiredEvidence
	}
	return ValidateObservationRange(s.InitialPH, s.InitialStrength)
}
func ValidateTrial(t ProcessTrial) error {
	if t.TrialID == "" || t.BatchID == "" || t.SampleID == "" || t.Reagent == "" || t.Concentration <= 0 || t.TemperatureCelsius < 0 || t.DurationMinutes <= 0 {
		return ErrRequiredEvidence
	}
	if e := ValidateObservationRange(t.InitialPH, t.StrengthBefore); e != nil {
		return e
	}
	return ValidateObservationRange(t.FinalPH, t.StrengthAfter)
}

func DocumentStatisticsFor(items []DocumentItem) DocumentStatistics {
	stats := DocumentStatistics{ItemCount: len(items), MaterialCount: map[string]int{}}
	for _, item := range items {
		stats.TotalPages += item.PageCount
		stats.MaterialCount[item.Material]++
	}
	return stats
}
