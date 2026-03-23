package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nikitaaldaev/bani/internal/domain"
)

func generateJWT(userID uuid.UUID, role domain.UserRole, secret []byte, ttl time.Duration) (string, error) {
	return generateJWTWithSession(userID, role, uuid.Nil, secret, ttl)
}

func generateJWTWithSession(userID uuid.UUID, role domain.UserRole, sessionID uuid.UUID, secret []byte, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    string(role),
		"exp":     time.Now().Add(ttl).Unix(),
	}
	if sessionID != uuid.Nil {
		claims["session_id"] = sessionID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
