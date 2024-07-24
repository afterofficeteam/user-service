package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Payload struct {
	Email  string
	UserID string
	Role   string
	jwt.RegisteredClaims
}

func NewPayload(email string, userID string, role string, duration time.Duration) (*Payload, error) {
	usrEmail, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	timeNow := time.Now()
	payload := &Payload{
		Email:  email,
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(timeNow.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(timeNow),
			NotBefore: jwt.NewNumericDate(timeNow),
			Issuer:    "user_login",
			Subject:   "shopifun",
			ID:        usrEmail.String(),
		},
	}
	return payload, nil
}
