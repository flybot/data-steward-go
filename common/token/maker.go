package token

import (
	"time"
)

type TokenResponse struct {
	Token string `json:"accessToken"`
}
type TokensResponse struct {
	Token        string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(username string, uid int, duration time.Duration, tokenType string) (string, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}
