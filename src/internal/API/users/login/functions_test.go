package login_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	login "github.com/StepanKomis/Ticketa/src/internal/API/users/login"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// bcryptHash generates a hash with MinCost to keep tests fast.
// Production uses cost 12 (security.bcryptCost); CheckPassword works for any cost.
func bcryptHash(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(h)
}

func userRow(t *testing.T, id int32, email, plainPw string) db.GetUserWithLocalLoginRow {
	t.Helper()
	return db.GetUserWithLocalLoginRow{
		ID:           id,
		Email:        email,
		UserType:     db.UserTypeStudent,
		IsActive:     true,
		CreatedAt:    time.Now(),
		PasswordHash: bcryptHash(t, plainPw),
	}
}

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockQuerier struct {
	row db.GetUserWithLocalLoginRow
	err error
}

func (m *mockQuerier) GetUserWithLocalLogin(_ context.Context, _ string) (db.GetUserWithLocalLoginRow, error) {
	return m.row, m.err
}

type mockStore struct {
	token string
	err   error
}

func (m *mockStore) Create(_ context.Context, _ int64, _ *http.Request) (db.Session, error) {
	if m.err != nil {
		return db.Session{}, m.err
	}
	now := time.Now()
	return db.Session{Token: m.token, CreatedAt: now, ExpiresAt: now.Add(time.Hour), LastSeenAt: now}, nil
}

// ---------------------------------------------------------------------------
// ToJson
// ---------------------------------------------------------------------------

func TestToJson_RoundTrip(t *testing.T) {
	lr := &login.LoginRequest{Email: "user@example.com", Password: "s3cr3t!"}
	data := lr.ToJson()
	if data == nil {
		t.Fatal("expected non-nil JSON")
	}
	var out login.LoginRequest
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Email != lr.Email || out.Password != lr.Password {
		t.Errorf("round-trip mismatch: got %+v", out)
	}
}

// ---------------------------------------------------------------------------
// Validate
// ---------------------------------------------------------------------------

func TestValidate_Success(t *testing.T) {
	const pw = "s3cr3t!"
	q := &mockQuerier{row: userRow(t, 42, "user@example.com", pw)}
	store := &mockStore{token: "tok123"}
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)

	lr := &login.LoginRequest{Email: "user@example.com", Password: pw}
	token, err := lr.Validate(q, store, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "tok123" {
		t.Errorf("expected token %q, got %q", "tok123", token)
	}
}

// TestValidate_WrongPassword is the critical security path: a valid user email
// with the wrong password must be rejected and must not produce a session token.
func TestValidate_WrongPassword(t *testing.T) {
	q := &mockQuerier{row: userRow(t, 42, "user@example.com", "correct-pw")}
	store := &mockStore{token: "should-not-be-issued"}
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)

	lr := &login.LoginRequest{Email: "user@example.com", Password: "wrong-pw"}
	token, err := lr.Validate(q, store, req)
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if token != "" {
		t.Errorf("expected empty token on failure, got %q", token)
	}
}

func TestValidate_UserNotFound(t *testing.T) {
	q := &mockQuerier{err: errors.New("sql: no rows in result set")}
	store := &mockStore{}
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)

	lr := &login.LoginRequest{Email: "missing@example.com", Password: "s3cr3t!"}
	_, err := lr.Validate(q, store, req)
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}

func TestValidate_SessionCreateError(t *testing.T) {
	const pw = "s3cr3t!"
	q := &mockQuerier{row: userRow(t, 7, "user@example.com", pw)}
	store := &mockStore{err: errors.New("session insert failed")}
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)

	lr := &login.LoginRequest{Email: "user@example.com", Password: pw}
	_, err := lr.Validate(q, store, req)
	if err == nil {
		t.Fatal("expected error from session creation, got nil")
	}
}
