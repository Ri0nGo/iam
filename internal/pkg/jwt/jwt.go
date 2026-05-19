package jwtpkg

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	issuer string
	secret []byte
	expire time.Duration
}

type Claims struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	Status   int      `json:"status"`
	TokenUse string   `json:"token_use"`
	ClientID string   `json:"client_id,omitempty"`
	Scope    string   `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

type SignOptions struct {
	TokenUse string
	ClientID string
	Scope    string
}

const (
	TokenUseConsole = "console"
	TokenUseOAuth2  = "oauth2"
)

func NewManager(issuer string, secret string, expireSeconds int64) *Manager {
	return &Manager{issuer: issuer, secret: []byte(secret), expire: time.Duration(expireSeconds) * time.Second}
}

func (m *Manager) Sign(userID uint64, username string, roles []string, status int, tokenID string) (string, int64, error) {
	return m.SignWithOptions(userID, username, roles, status, tokenID, SignOptions{TokenUse: TokenUseConsole})
}

func (m *Manager) SignWithOptions(userID uint64, username string, roles []string, status int, tokenID string, options SignOptions) (string, int64, error) {
	now := time.Now()
	exp := now.Add(m.expire)
	claims := Claims{
		Username: username,
		Roles:    roles,
		Status:   status,
		TokenUse: options.TokenUse,
		ClientID: options.ClientID,
		Scope:    options.Scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        tokenID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(m.expire.Seconds()), nil
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.Issuer != m.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}
	return claims, nil
}
