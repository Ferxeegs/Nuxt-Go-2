package turnstile

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-cp-boilerplate/internal/pkg/apperror"
)

func TestNoop_SkipsVerification(t *testing.T) {
	v := New("")
	if v.Enabled() {
		t.Fatal("expected empty secret to disable turnstile")
	}
	if err := v.Verify(context.Background(), "", ""); err != nil {
		t.Fatalf("noop verify should succeed: %v", err)
	}
}

func TestClient_VerifySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.Form.Get("secret"); got != "test-secret" {
			t.Fatalf("secret = %q", got)
		}
		if got := r.Form.Get("response"); got != "token-abc" {
			t.Fatalf("response = %q", got)
		}
		if got := r.Form.Get("remoteip"); got != "203.0.113.10" {
			t.Fatalf("remoteip = %q", got)
		}
		_ = json.NewEncoder(w).Encode(siteverifyResponse{Success: true})
	}))
	defer server.Close()

	prev := siteverifyURL
	siteverifyURL = server.URL
	defer func() { siteverifyURL = prev }()

	client := New("test-secret")
	if err := client.Verify(context.Background(), "token-abc", "203.0.113.10"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestClient_RejectsEmptyToken(t *testing.T) {
	err := New("test-secret").Verify(context.Background(), "  ", "")
	assertStatus(t, err, http.StatusBadRequest)
}

func TestClient_MapsDuplicateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(siteverifyResponse{
			Success:    false,
			ErrorCodes: []string{"timeout-or-duplicate"},
		})
	}))
	defer server.Close()

	prev := siteverifyURL
	siteverifyURL = server.URL
	defer func() { siteverifyURL = prev }()

	err := New("test-secret").Verify(context.Background(), "used-token", "")
	assertStatus(t, err, http.StatusBadRequest)
}

func assertStatus(t *testing.T, err error, want int) {
	t.Helper()
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %v", err)
	}
	if appErr.HTTPStatus != want {
		t.Fatalf("status = %d, want %d (%s)", appErr.HTTPStatus, want, appErr.Message)
	}
}
