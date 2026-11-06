package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the payload stored inside our JWTs.
// It embeds jwt.RegisteredClaims (which gives us standard fields like
// Subject "sub", IssuedAt "iat", and ExpiresAt "exp") plus our custom fields.
type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"` // "seller" | "buyer"
	jwt.RegisteredClaims
}

// TokenManager issues and verifies HS256 JWTs using a shared secret.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// Issue creates a signed JWT for the given user.
func (tm *TokenManager) Issue(userID int64, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

// Verify parses a JWT string, checks its HS256 signature and expiration time,
// and returns the typed Claims if valid.
func (tm *TokenManager) Verify(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	// WithValidMethods pins HS256 so an attacker cannot forge a token with "alg":"none".
	token, err := jwt.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwt.Token) (any, error) { return tm.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}
