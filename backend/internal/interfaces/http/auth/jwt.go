package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	CookieSecure  bool   // prod: true (https)
	CookieDomain  string // opsiyonel
}

type AccessClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UID string `json:"uid"`
	JTI string `json:"jti"` // refresh rotation için faydalı
	jwt.RegisteredClaims
}

func NewAccessToken(cfg Config, uid string) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uid,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(cfg.AccessSecret)
}

func NewRefreshToken(cfg Config, uid, jti string) (string, error) {
	now := time.Now()
	claims := RefreshClaims{
		UID: uid,
		JTI: jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uid,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.RefreshTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(cfg.RefreshSecret)
}

func ParseAccess(cfg Config, raw string) (*AccessClaims, error) {
	tok, err := jwt.ParseWithClaims(raw, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return cfg.AccessSecret, nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid access token")
	}
	return tok.Claims.(*AccessClaims), nil
}

func ParseRefresh(cfg Config, raw string) (*RefreshClaims, error) {
	tok, err := jwt.ParseWithClaims(raw, &RefreshClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return cfg.RefreshSecret, nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid refresh token")
	}
	return tok.Claims.(*RefreshClaims), nil
}
