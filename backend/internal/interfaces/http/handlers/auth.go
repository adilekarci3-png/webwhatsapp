package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	authsvc "example.com/webwhatsapp/backend/internal/application/usecases/auth"
	jwtauth "example.com/webwhatsapp/backend/internal/interfaces/http/auth"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserPublic struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LoginResponse struct {
	AccessToken string      `json:"accessToken"`
	User        *UserPublic `json:"user"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	AccessToken string      `json:"accessToken"`
	User        *UserPublic `json:"user,omitempty"`
}

func randomJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// JSON body oku, UTF-8 BOM varsa temizle, Body'yi yeniden set et
func readBodyAndReset(r *http.Request) ([]byte, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	r.Body = io.NopCloser(bytes.NewBuffer(raw))
	return raw, nil
}

func writeRefreshCookie(w http.ResponseWriter, cfg jwtauth.Config, refresh string) {
	ck := &http.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(cfg.RefreshTTL.Seconds()),
	}

	// Domain boşsa set etme (localde gerek yok)
	if d := strings.TrimSpace(cfg.CookieDomain); d != "" && d != `""` {
		ck.Domain = d
	}
	http.SetCookie(w, ck)
}

func Login(cfg jwtauth.Config, userSvc *authsvc.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if userSvc == nil {
			http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
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
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		access, err := jwtauth.NewAccessToken(cfg, u.ID)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		refresh, err := jwtauth.NewRefreshToken(cfg, u.ID, randomJTI())
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		writeRefreshCookie(w, cfg, refresh)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LoginResponse{
			AccessToken: access,
			User: &UserPublic{
				ID:   u.ID,
				Name: u.Name,
			},
		})
	}
}

func Register(cfg jwtauth.Config, userSvc *authsvc.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if userSvc == nil {
			http.Error(w, "auth service unavailable", http.StatusServiceUnavailable)
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
		log.Printf("REGISTER OK id=%s email=%s", u.ID, u.Email)
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

		refresh, err := jwtauth.NewRefreshToken(cfg, u.ID, randomJTI())
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		writeRefreshCookie(w, cfg, refresh)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(RegisterResponse{
			AccessToken: access,
			User: &UserPublic{
				ID:   u.ID,
				Name: u.Name,
			},
		})
	}
}

func Refresh(cfg jwtauth.Config) http.HandlerFunc {
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
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ck := &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		}
		if d := strings.TrimSpace(cfg.CookieDomain); d != "" && d != `""` {
			ck.Domain = d
		}
		http.SetCookie(w, ck)

		w.WriteHeader(http.StatusNoContent)
	}
}
