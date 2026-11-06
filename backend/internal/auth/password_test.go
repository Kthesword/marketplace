package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	plain := "super-secret-123"

	hash1, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword unexpected error: %v", err)
	}
	hash2, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword unexpected error: %v", err)
	}

	// Because bcrypt embeds a random salt each time, two hashes of the same
	// password must never be identical.
	if hash1 == hash2 {
		t.Fatalf("expected random salts to produce different hashes, got identical %q", hash1)
	}
	t.Logf("sample bcrypt hash: %s", hash1)

	if !CheckPassword(hash1, plain) {
		t.Errorf("CheckPassword returned false for the correct password")
	}
	if CheckPassword(hash1, "wrong-password") {
		t.Errorf("CheckPassword returned true for an incorrect password")
	}
}
