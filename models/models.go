package models

type HealthResponse struct {
	Time   string `json:"time"`
	Status string `json:"status"`
}

type PairingResponse struct {
	Status string `json:"status"`
}
