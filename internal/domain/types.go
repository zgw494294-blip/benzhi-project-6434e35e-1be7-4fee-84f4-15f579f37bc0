package domain

import "time"

type BatchStatus string

const (
	StatusDraft      BatchStatus = "draft"
	StatusSampling   BatchStatus = "sampling"
	StatusTrialing   BatchStatus = "trialing"
	StatusCorrection BatchStatus = "correction"
	StatusReview     BatchStatus = "review"
	StatusApproved   BatchStatus = "approved"
	StatusFrozen     BatchStatus = "frozen"
	StatusRejected   BatchStatus = "rejected"
)

type DocumentItem struct {
	ItemID     string `json:"itemID"`
	Title      string `json:"title"`
	CallNumber string `json:"callNumber,omitempty"`
	Material   string `json:"material,omitempty"`
	PageCount  int    `json:"pageCount"`
}
type DocumentStatistics struct {
	ItemCount     int            `json:"itemCount"`
	TotalPages    int            `json:"totalPages"`
	MaterialCount map[string]int `json:"materialCount"`
}
type Sample struct {
	SampleID        string    `json:"sampleID"`
	BatchID         string    `json:"batchID"`
	Material        string    `json:"material"`
	InitialPH       float64   `json:"initialPH"`
	InitialStrength float64   `json:"initialStrength"`
	Observer        string    `json:"observer"`
	ObservedAt      time.Time `json:"observedAt"`
}
type RestorationBatch struct {
	BatchID       string             `json:"batchID"`
	Title         string             `json:"title"`
	Institution   string             `json:"institution"`
	DocumentItems []DocumentItem     `json:"documentItems"`
	DocumentStats DocumentStatistics `json:"documentStats"`
	TargetProcess string             `json:"targetProcess"`
	Status        BatchStatus        `json:"status"`
	Version       int                `json:"version"`
	CreatedBy     string             `json:"createdBy"`
	CreatedAt     *time.Time         `json:"createdAt"`
	ReleasedAt    *time.Time         `json:"releasedAt,omitempty"`
	Samples       []Sample           `json:"samples"`
	Trials        []ProcessTrial     `json:"trials"`
	Retests       []CorrectionRetest `json:"retests"`
	Credential    *ReleaseCredential `json:"credential,omitempty"`
}
type ProcessTrial struct {
	TrialID             string   `json:"trialID"`
	BatchID             string   `json:"batchID"`
	SampleID            string   `json:"sampleID"`
	Reagent             string   `json:"reagent"`
	Concentration       float64  `json:"concentration"`
	TemperatureCelsius  float64  `json:"temperatureCelsius"`
	DurationMinutes     int      `json:"durationMinutes"`
	InitialPH           float64  `json:"initialPH"`
	FinalPH             float64  `json:"finalPH"`
	StrengthBefore      float64  `json:"strengthBefore"`
	StrengthAfter       float64  `json:"strengthAfter"`
	AnalysisStatus      string   `json:"analysisStatus"`
	MetricDeviations    []string `json:"metricDeviations"`
	ParameterDeviations []string `json:"parameterDeviations"`
	DeviationCodes      []string `json:"deviationCodes"`
	Version             int      `json:"version"`
}
type CorrectionRetest struct {
	RetestID     string             `json:"retestID"`
	BatchID      string             `json:"batchID"`
	TrialID      string             `json:"trialID"`
	Action       string             `json:"action"`
	Result       string             `json:"result"`
	SubmittedBy  string             `json:"submittedBy"`
	Observations map[string]float64 `json:"observations,omitempty"`
	SubmittedAt  time.Time          `json:"submittedAt"`
}
type ReleaseCredential struct {
	CredentialID     string    `json:"credentialID"`
	BatchID          string    `json:"batchID"`
	EvidenceDigest   string    `json:"evidenceDigest"`
	Decision         string    `json:"decision"`
	Reviewer         string    `json:"reviewer"`
	VerificationCode string    `json:"verificationCode"`
	IssuedAt         time.Time `json:"issuedAt"`
	AuditSequence    int64     `json:"auditSequence"`
}
type Metrics struct {
	PHChange, StrengthRetention float64
	Deviations                  []string
	Passed                      bool
}

func now() time.Time { return time.Now().UTC() }
