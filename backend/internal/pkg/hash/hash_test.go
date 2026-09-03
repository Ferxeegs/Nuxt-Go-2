package hash

import "testing"

func TestHashAndVerify(t *testing.T) {
	hashed, err := Hash("correct-password", 10)
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}
	if hashed == "correct-password" {
		t.Fatal("expected hashed value to differ from plaintext")
	}
	if !Verify("correct-password", hashed) {
		t.Error("expected Verify to succeed with the correct password")
	}
	if Verify("wrong-password", hashed) {
		t.Error("expected Verify to fail with an incorrect password")
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	a := HashToken("some-refresh-token")
	b := HashToken("some-refresh-token")
	if a != b {
		t.Error("expected HashToken to be deterministic for the same input")
	}
	if a == HashToken("a-different-token") {
		t.Error("expected different inputs to produce different digests")
	}
}
