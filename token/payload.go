package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token has been expired")
	ErrInvalidToken = errors.New("invalid token")
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
	Role      string    `json:"role"`
}

func NewPayload(username string, role string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(time.Duration(duration)),
		Role:      role,
	}

	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrTokenExpired
	}
	return nil
}

func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: p.ExpiredAt}, nil
}

// GetNotBefore implements the Claims interface.
func (p *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{p.IssuedAt}, nil
}

// GetIssuedAt implements the Claims interface.
func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{p.IssuedAt}, nil
}

// GetAudience implements the Claims interface.
func (p *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return []string{}, nil
}

// GetIssuer implements the Claims interface.
func (p *Payload) GetIssuer() (string, error) {
	return "",nil
}

// GetSubject implements the Claims interface.
func (p *Payload) GetSubject() (string, error) {
	return "", nil
}
