package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// CheckFunc verifies the health of a single dependency (e.g., database, broker).
type CheckFunc func(ctx context.Context) error

// Response represents the standardised health check payload across NovaBank services.
type Response struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// Handler returns an http.HandlerFunc compatible with chi and net/http.
// It runs all registered dependency checks and returns 200 OK if all succeed,
// or 503 Service Unavailable if any check fails.
func Handler(checks map[string]CheckFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		resp := Response{
			Status: "UP",
			Checks: make(map[string]string, len(checks)),
		}
		statusCode := http.StatusOK

		for name, check := range checks {
			if err := check(ctx); err != nil {
				resp.Status = "DOWN"
				resp.Checks[name] = "DOWN"
				statusCode = http.StatusServiceUnavailable
			} else {
				resp.Checks[name] = "UP"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
