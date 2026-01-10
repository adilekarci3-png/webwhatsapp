package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	authsvc "example.com/webwhatsapp/backend/internal/application/usecases/auth"
	jwtauth "example.com/webwhatsapp/backend/internal/interfaces/http/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	User        any    `json:"user,omitempty"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RegisterResponse struct {
	AccessToken string `json:"accessToken"`
	User        any    `json:"user,omitempty"`
}

func readBodyAndReset(r *http.Request) ([]byte, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// UTF-8 BOM strip
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	r.Body = io.NopCloser(bytes.NewBuffer(raw))
	return raw, nil
}

// LOGIN (email+password -> verify -> token)
func Login(cfg jwtauth.Config, userSvc *authsvc.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if _, err := readBodyAndReset(r); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		var req LoginRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				http.Error(w, "empty body", http.StatusBadRequest)
				return
			}
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		u, err := userSvc.Verify(r.Context(), req.Email, req.Password)
		if err != nil {
			if err == authsvc.ErrInvalidCredentials {
				http.Error(w, "invalid credentials", http.StatusUnauthorized)
				return
			}
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		access, err := jwtauth.NewAccessToken(cfg, u.ID) // UID = userID
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		jti := jwtauth.RandomJTI() // aşağıda not: istersen mevcut randomJTI’ni jwtauth içine al
		refresh, err := jwtauth.NewRefreshToken(cfg, u.ID, jti)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			Domain:   cfg.CookieDomain,
			MaxAge:   int(cfg.RefreshTTL.Seconds()),
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LoginResponse{
			AccessToken: access,
			User: map[string]any{
				"id":    u.ID,
				"name":  u.Name,
				"email": u.Email,
			},
		})
	}
}

// REGISTER (insert -> token)
func Register(cfg jwtauth.Config, userSvc *authsvc.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if _, err := readBodyAndReset(r); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}

		var req RegisterRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				http.Error(w, "empty body", http.StatusBadRequest)
				return
			}
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		u, err := userSvc.Register(r.Context(), req.Name, req.Email, req.Password)
		if err != nil {
			if err == authsvc.ErrEmailExists {
				http.Error(w, "email already exists", http.StatusConflict)
				return
			}
			if err.Error() == "missing fields" {
				http.Error(w, "missing fields", http.StatusBadRequest)
				return
			}
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		access, err := jwtauth.NewAccessToken(cfg, u.ID)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		jti := jwtauth.RandomJTI()
		refresh, err := jwtauth.NewRefreshToken(cfg, u.ID, jti)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refresh,
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			Domain:   cfg.CookieDomain,
			MaxAge:   int(cfg.RefreshTTL.Seconds()),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(RegisterResponse{
			AccessToken: access,
			User: map[string]any{
				"id":    u.ID,
				"name":  u.Name,
				"email": u.Email,
			},
		})
	}
}

// REFRESH
func Refresh(cfg jwtauth.Config /*, refreshStore ... */) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

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

		// TODO: JTI doğrulama + rotation (Redis / DB)
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

// LOGOUT
func Logout(cfg jwtauth.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
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
