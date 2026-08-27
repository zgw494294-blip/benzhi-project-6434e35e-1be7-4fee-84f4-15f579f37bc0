package domain

import "strings"

type MaterialRule struct {
	Name            string   `json:"name"`
	AllowedReagents []string `json:"allowedReagents"`
	MaxTemperature  float64  `json:"maxTemperature"`
	MinDuration     int      `json:"minDuration"`
}

var MaterialRules = []MaterialRule{{Name: "棉纸", AllowedReagents: []string{"碳酸镁", "碳酸钙"}, MaxTemperature: 40, MinDuration: 5}, {Name: "麻纸", AllowedReagents: []string{"碳酸镁"}, MaxTemperature: 45, MinDuration: 5}, {Name: "机制纸", AllowedReagents: []string{"碳酸钙"}, MaxTemperature: 35, MinDuration: 10}}

func RuleForMaterial(name string) MaterialRule {
	for _, r := range MaterialRules {
		if r.Name == name {
			return r
		}
	}
	return MaterialRule{Name: name, MaxTemperature: 60, MinDuration: 1}
}
func ValidateProcess(t ProcessTrial, material string) error {
	r := RuleForMaterial(material)
	if t.TemperatureCelsius > r.MaxTemperature || t.DurationMinutes < r.MinDuration {
		return ErrInvalidMetrics
	}
	if len(r.AllowedReagents) > 0 {
		ok := false
		for _, x := range r.AllowedReagents {
			if x == t.Reagent {
				ok = true
			}
		}
		if !ok {
			return ErrInvalidMetrics
		}
	}
	return nil
}
func NormalizeMaterial(name string) string { return strings.TrimSpace(name) }
func SupportedMaterials() []string {
	out := make([]string, 0, len(MaterialRules))
	for _, r := range MaterialRules {
		out = append(out, r.Name)
	}
	return out
}
func ReagentSupported(material, reagent string) bool {
	r := RuleForMaterial(material)
	if len(r.AllowedReagents) == 0 {
		return true
	}
	for _, x := range r.AllowedReagents {
		if x == reagent {
			return true
		}
	}
	return false
}
func TemperatureAllowed(material string, v float64) bool {
	return v >= 0 && v <= RuleForMaterial(material).MaxTemperature
}
func DurationAllowed(material string, v int) bool { return v >= RuleForMaterial(material).MinDuration }
