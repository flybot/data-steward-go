package token

import (
	"errors"
	"time"
)

const (
	TokenTypeRegular = "regular"
	TokenTypeRefresh = "refresh"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
	TokenType string    `json:"type"`
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayload(username string, uid int, duration time.Duration, tokenType string) (*Payload, error) {
	payload := &Payload{
		ID:        uid,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
		TokenType: tokenType,
	}
	return payload, nil
}

// Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
