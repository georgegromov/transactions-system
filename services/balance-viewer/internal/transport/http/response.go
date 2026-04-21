package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteResponse(logger *slog.Logger, w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to encode response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
