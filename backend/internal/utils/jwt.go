package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"lims-backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims holds the JWT token claims.
type Claims struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	DeptID      *uint    `json:"dept_id"`
	IsAdmin     bool     `json:"is_admin"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

func generateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateToken creates a JWT token for the given user.
func GenerateToken(cfg *config.JWTConfig, userID uint, username string, deptID *uint, isAdmin bool, permissions []string) (string, error) {
	now := time.Now()
	exp := now.Add(time.Duration(cfg.ExpireHour) * time.Hour)
	claims := Claims{
		UserID:      userID,
		Username:    username,
		DeptID:      deptID,
		IsAdmin:     isAdmin,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        generateJTI(),
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken validates and parses a JWT token.
func ParseToken(cfg *config.JWTConfig, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
