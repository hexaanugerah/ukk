package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("jwtutil: invalid or expired token")

type Claims struct {
	UserID string   `json:"uid"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
}

func NewManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration, issuer string) *Manager {
	return &Manager{
		accessSecret: []byte(accessSecret), refreshSecret: []byte(refreshSecret),
		accessTTL: accessTTL, refreshTTL: refreshTTL, issuer: issuer,
	}
}

func (m *Manager) GenerateAccessToken(userID, email string, roles []string) (string, time.Time, error) {
	return m.generate(userID, email, roles, m.accessSecret, m.accessTTL)
}

func (m *Manager) GenerateRefreshToken(userID, email string, roles []string) (string, time.Time, error) {
	return m.generate(userID, email, roles, m.refreshSecret, m.refreshTTL)
}

func (m *Manager) generate(userID, email string, roles []string, secret []byte, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	claims := Claims{
		UserID: userID, Email: email, Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: m.issuer, IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	return signed, expiresAt, err
}

func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.accessSecret)
}

func (m *Manager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.refreshSecret)
}

func (m *Manager) parse(tokenStr string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
