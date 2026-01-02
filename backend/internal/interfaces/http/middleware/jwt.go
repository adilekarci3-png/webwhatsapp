package middleware

import (
	"context"
	"net/http"
	"strings"

	jwtauth "example.com/webwhatsapp/backend/internal/interfaces/http/auth"
)

type ctxKey string

const CtxUID ctxKey = "uid"

func UIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(CtxUID)
	s, ok := v.(string)
	return s, ok
}

func RequireJWT(cfg jwtauth.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")

		claims, err := jwtauth.ParseAccess(cfg, raw)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), CtxUID, claims.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
