package httpapi

import "time"

type HealthResponse struct {
	Status  string    `json:"status"`
	Service string    `json:"service"`
	Time    time.Time `json:"time"`
}

func healthResponse() HealthResponse {
	return HealthResponse{Status: "ok", Service: "archive-deacidification", Time: time.Now().UTC()}
}
