package auth

import (
	"strings"
	"testing"
	"time"
)

func TestTokenManager_IssueAndVerify(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef" // 32 bytes
	tm := NewTokenManager(secret, time.Hour)

	tok, err := tm.Issue(42, "seller@example.com", "seller")
	if err != nil {
		t.Fatalf("Issue failed: %v", err)
	}

	// A JWT must have 3 dot-separated parts: header.payload.signature
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}
	t.Logf("issued JWT: %s", tok)

	// 1. Valid token verifies cleanly
	claims, err := tm.Verify(tok)
	if err != nil {
		t.Fatalf("Verify failed on valid token: %v", err)
	}
	if claims.UserID != 42 || claims.Email != "seller@example.com" || claims.Role != "seller" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	// 2. Wrong secret must fail
	otherTM := NewTokenManager("fedcba9876543210fedcba9876543210", time.Hour)
	if _, err := otherTM.Verify(tok); err == nil {
		t.Errorf("expected Verify to fail with wrong secret")
	}

	// 3. Expired token must fail
	expiredTM := NewTokenManager(secret, -time.Minute)
	expiredTok, _ := expiredTM.Issue(42, "seller@example.com", "seller")
	if _, err := tm.Verify(expiredTok); err == nil {
		t.Errorf("expected Verify to fail on expired token")
	}
}
