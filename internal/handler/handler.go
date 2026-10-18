package handler

import (
	"encoding/json"
	"net/http"
)

func GetJob(w http.ResponseWriter, r *http.Request) {
	job := map[string]string{
		"id":     "1",
		"title":  "Backend Engineer",
		"status": "active",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(job); err != nil {
		return
	}
}
