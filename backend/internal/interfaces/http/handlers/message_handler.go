package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/webwhatsapp/backend/internal/application/usecases/messaging"
)

func MessagesList(svc *messaging.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// ✅ Degraded mode guard: svc yoksa process asla düşmesin
		if svc == nil {
			http.Error(w, "messaging service unavailable", http.StatusServiceUnavailable)
			return
		}

		conv := r.URL.Query().Get("conversationId")
		if conv == "" {
			conv = r.URL.Query().Get("room") // UI kolaylığı
		}
		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil {
				limit = v
			}
		}

		out, err := svc.ListMessages(r.Context(), conv, limit)
		if err != nil {
			// İsterseniz burada invalid input için 400, diğerleri için 500 ayırabiliriz.
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
