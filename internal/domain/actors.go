package domain

import "strings"

func ValidateActor(actor string) error {
	if strings.TrimSpace(actor) == "" {
		return ErrUnauthorized
	}
	return nil
}
func IsReviewer(actor string) bool {
	return strings.Contains(actor, "负责人") || strings.Contains(actor, "review")
}
