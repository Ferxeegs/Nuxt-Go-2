// Package jwtutil issues and validates access tokens, and generates opaque
// refresh tokens (stored hashed in DB by the auth module).
package jwtutil

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// AccessClaims is the JWT payload for access tokens.
// ImpersonatorID is set when a SuperAdmin is acting as another user.
// ActiveRoleID/ActiveScopeID identify the selected role assignment context.
type AccessClaims struct {
	UserID         string   `json:"user_id"`
	Role           string   `json:"role"`
	Permissions    []string `json:"permissions"`
	ActiveRoleID   uint     `json:"active_role_id,omitempty"`
	ActiveScopeID  uint     `json:"active_scope_id,omitempty"`
	ImpersonatorID *string  `json:"impersonator_id,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret    []byte
	accessTTL time.Duration
}

func NewManager(secret string, accessTTL time.Duration) *Manager {
	return &Manager{secret: []byte(secret), accessTTL: accessTTL}
}

type AccessTokenInput struct {
	UserID         string
	Role           string
	Permissions    []string
	ActiveRoleID   uint
	ActiveScopeID  uint
	ImpersonatorID *string
}

// GenerateAccessToken issues a signed JWT embedding user id, active role,
// scope, and permissions so downstream handlers don't need a DB round-trip.
func (m *Manager) GenerateAccessToken(in AccessTokenInput) (token string, expiresAt time.Time, jti string, err error) {
	expiresAt = time.Now().Add(m.accessTTL)
	jti, err = GenerateRefreshToken()
	if err != nil {
		return "", time.Time{}, "", err
	}
	claims := AccessClaims{
		UserID:         in.UserID,
		Role:           in.Role,
		Permissions:    in.Permissions,
		ActiveRoleID:   in.ActiveRoleID,
		ActiveScopeID:  in.ActiveScopeID,
		ImpersonatorID: in.ImpersonatorID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, "", err
	}
	return signed, expiresAt, jti, nil
}

// ParseAccessToken validates signature and expiry and returns the claims.
func (m *Manager) ParseAccessToken(tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GenerateRefreshToken returns a cryptographically random opaque token. The
// caller is responsible for hashing it before persisting.
func GenerateRefreshToken() (string, error) {
	buf := make([]byte, 48)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
