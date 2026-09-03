package jwtutil

import (
	"testing"
	"time"
)

func TestGenerateAndParseAccessToken(t *testing.T) {
	m := NewManager("test-secret", time.Minute)
	userID := "11111111-1111-1111-1111-111111111111"

	token, expiresAt, jti, err := m.GenerateAccessToken(AccessTokenInput{
		UserID:        userID,
		Role:          "Admin",
		Permissions:   []string{"setting.view"},
		ActiveRoleID:  2,
		ActiveScopeID: 1,
	})
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if token == "" || jti == "" {
		t.Fatal("expected non-empty token and jti")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("expected expiresAt to be in the future")
	}

	claims, err := m.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Role != "Admin" {
		t.Errorf("expected Role Admin, got %s", claims.Role)
	}
	if claims.ActiveRoleID != 2 || claims.ActiveScopeID != 1 {
		t.Errorf("unexpected active assignment: role=%d scope=%d", claims.ActiveRoleID, claims.ActiveScopeID)
	}
	if len(claims.Permissions) != 1 || claims.Permissions[0] != "setting.view" {
		t.Errorf("unexpected permissions: %v", claims.Permissions)
	}
	if claims.ID != jti {
		t.Errorf("expected claims.ID %s to match returned jti %s", claims.ID, jti)
	}
	if claims.ImpersonatorID != nil {
		t.Fatal("expected nil ImpersonatorID for normal token")
	}
}

func TestGenerateAccessToken_WithImpersonator(t *testing.T) {
	m := NewManager("test-secret", time.Minute)
	userID := "11111111-1111-1111-1111-111111111111"
	actor := "22222222-2222-2222-2222-222222222222"

	token, _, _, err := m.GenerateAccessToken(AccessTokenInput{
		UserID:         userID,
		Role:           "Editor",
		Permissions:    []string{"setting.view"},
		ImpersonatorID: &actor,
	})
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := m.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if claims.ImpersonatorID == nil || *claims.ImpersonatorID != actor {
		t.Fatalf("expected ImpersonatorID %s, got %v", actor, claims.ImpersonatorID)
	}
}

func TestParseAccessToken_InvalidSignature(t *testing.T) {
	m1 := NewManager("secret-one", time.Minute)
	m2 := NewManager("secret-two", time.Minute)

	token, _, _, err := m1.GenerateAccessToken(AccessTokenInput{
		UserID: "11111111-1111-1111-1111-111111111111",
		Role:   "Editor",
	})
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	if _, err := m2.ParseAccessToken(token); err == nil {
		t.Fatal("expected error parsing token signed with a different secret")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	m := NewManager("test-secret", -time.Minute)

	token, _, _, err := m.GenerateAccessToken(AccessTokenInput{
		UserID: "11111111-1111-1111-1111-111111111111",
		Role:   "Editor",
	})
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	if _, err := m.ParseAccessToken(token); err == nil {
		t.Fatal("expected error parsing an expired token")
	}
}

func TestGenerateRefreshToken_Unique(t *testing.T) {
	a, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	b, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	if a == b {
		t.Fatal("expected two generated refresh tokens to differ")
	}
}
