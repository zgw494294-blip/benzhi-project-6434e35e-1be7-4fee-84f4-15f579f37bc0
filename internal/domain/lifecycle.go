package domain

func CanEdit(status BatchStatus) bool { return status != StatusFrozen }
func RequiresCorrection(b RestorationBatch) bool {
	for _, t := range b.Trials {
		if len(t.DeviationCodes) > 0 && !HasClosedRetest(&b, t.TrialID) {
			return true
		}
	}
	return false
}
func ReadyForFreeze(b RestorationBatch) bool {
	return (b.Status == StatusApproved || b.Status == StatusFrozen) && len(b.Trials) > 0 && !RequiresCorrection(b)
}
