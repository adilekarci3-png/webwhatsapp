package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	jwtauth "example.com/webwhatsapp/backend/internal/interfaces/http/auth"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	// User alanı eklemek istersen buraya koyarsın
}

func randomJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Burada gerçek kullanıcı doğrulamasını senin user repo/service’inle yapacaksın.
func Login(cfg jwtauth.Config /*, userSvc ... */) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		// TODO: gerçek doğrulama
		// uid, err := userSvc.Verify(req.Username, req.Password)
		uid := req.Username
		if uid == "" || req.Password == "" {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		access, err := jwtauth.NewAccessToken(cfg, uid)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		jti := randomJTI()
		refresh, err := jwtauth.NewRefreshToken(cfg, uid, jti)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		// Refresh cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			Path:     "/auth/refresh",
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			Domain:   cfg.CookieDomain,
			MaxAge:   int(cfg.RefreshTTL.Seconds()),
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LoginResponse{AccessToken: access})
	}
}

func Refresh(cfg jwtauth.Config /*, refreshStore ... */) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("refresh_token")
		if err != nil || c.Value == "" {
			http.Error(w, "missing refresh", http.StatusUnauthorized)
			return
		}

		rc, err := jwtauth.ParseRefresh(cfg, c.Value)
		if err != nil {
			http.Error(w, "invalid refresh", http.StatusUnauthorized)
			return
		}

		// TODO (önerilir): rc.JTI doğrula + rotation uygula (DB/Redis)
		access, err := jwtauth.NewAccessToken(cfg, rc.UID)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accessToken": access,
		})
	}
}

func Logout(cfg jwtauth.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// refresh cookie’yi expire et
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/auth/refresh",
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			Domain:   cfg.CookieDomain,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})
		w.WriteHeader(http.StatusNoContent)
	}
}
