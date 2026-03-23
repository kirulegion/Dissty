package token

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type IncompleteClaims struct {
	AccountID string `json:"account_id"`
	Status    string `json:"status"`
	jwt.RegisteredClaims
}

type Claims struct {
	AccountID string `json:"account_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	jwt.RegisteredClaims
}

type TokenService interface {
	GenerateIncompleteToken(accountID uuid.UUID) (string, error)
	GenerateCompleteToken(accountID, userID uuid.UUID) (string, error)
	ValidateToken(token string) (*Claims, error)
}

type jwtService struct {
	secret string
}

func NewTokenService() TokenService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET env var not set")
	}
	return &jwtService{secret: secret}
}

func (s *jwtService) GenerateIncompleteToken(accountID uuid.UUID) (string, error) {
	claims := &IncompleteClaims{
		AccountID: accountID.String(),
		Status:    "incomplete",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *jwtService) GenerateCompleteToken(accountID, userID uuid.UUID) (string, error) {
	claims := &Claims{
		AccountID: accountID.String(),
		UserID:    userID.String(),
		Status:    "complete",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *jwtService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
