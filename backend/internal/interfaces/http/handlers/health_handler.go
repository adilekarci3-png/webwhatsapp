package handlers

import "net/http"

// Health godoc
// @Summary Health check
// @Description Service is up
// @Tags system
// @Produce plain
// @Success 200 {string} string "ok"
// @Router /health [get]
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
