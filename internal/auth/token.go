package auth

import "time"

type TokenPair struct {
	IDToken      string
	RefreshToken string
	ExpiresAt    time.Time
}

func (t *TokenPair) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(-60 * time.Second))
}
