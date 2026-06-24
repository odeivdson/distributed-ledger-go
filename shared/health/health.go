package health

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

func Handler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := HealthResponse{
			Status:    "up",
			Timestamp: time.Now(),
			Service:   serviceName,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}
