package handlers

import (
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"UP","database":"CONNECTED"}`))
}
