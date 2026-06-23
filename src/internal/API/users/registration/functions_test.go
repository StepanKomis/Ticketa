package userregistration_test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	userregistration "github.com/StepanKomis/Ticketa/src/internal/API/users/registration"
)

// ---------------------------------------------------------------------------
// ValidatePassword
// ---------------------------------------------------------------------------

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name    string
		pw      string
		wantErr bool
	}{
		{"too short", "Ab1!", true},
		{"exactly min length, valid", "Secret1!", false},
		{"too long (73 chars)", string(make([]byte, 73)), true},
		{"exactly max length, valid", buildPassword(72), false},
		{"no digit", "Secret!!!", true},
		{"no special character", "Secret123", true},
		{"valid with multiple special chars", "S3cr3t!@#", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := userregistration.ValidatePassword(tc.pw)
			if tc.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// buildPassword returns a string of length n that satisfies all validation
// rules (contains a digit and a special character).
func buildPassword(n int) string {
	if n < 3 {
		panic("n too small")
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	b[0] = '1'
	b[1] = '!'
	return string(b)
}

// ---------------------------------------------------------------------------
// Minimal SQL mock driver for RegisterNewLocalUser
// ---------------------------------------------------------------------------

var (
	regDriverOnce sync.Once
	regDriver     = &scriptDriver{ch: make(chan *connScript, 16)}
)

func init() {
	regDriverOnce.Do(func() {
		sql.Register("mock_userregistration", regDriver)
	})
}

type scriptDriver struct{ ch chan *connScript }

func (d *scriptDriver) Open(_ string) (driver.Conn, error) {
	select {
	case s := <-d.ch:
		return &mockConn{s: s}, nil
	default:
		return nil, fmt.Errorf("no script queued")
	}
}

type connScript struct {
	queries   []queryResult
	execs     []error
	commitErr error
	qi, ei    int
}

type queryResult struct {
	cols []string
	rows [][]driver.Value
	err  error
}

func newMockDB(t *testing.T, s *connScript) *sql.DB {
	t.Helper()
	regDriver.ch <- s
	db, err := sql.Open("mock_userregistration", "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	return db
}

type mockConn struct{ s *connScript }

func (c *mockConn) Prepare(_ string) (driver.Stmt, error) { return &mockStmt{c: c}, nil }
func (c *mockConn) Close() error                          { return nil }
func (c *mockConn) Begin() (driver.Tx, error)             { return &mockTx{c: c}, nil }

type mockTx struct{ c *mockConn }

func (t *mockTx) Commit() error   { return t.c.s.commitErr }
func (t *mockTx) Rollback() error { return nil }

type mockStmt struct{ c *mockConn }

func (s *mockStmt) Close() error  { return nil }
func (s *mockStmt) NumInput() int { return -1 }

func (s *mockStmt) Exec(_ []driver.Value) (driver.Result, error) {
	sc := s.c.s
	if sc.ei >= len(sc.execs) {
		return nil, fmt.Errorf("unexpected Exec call (index %d)", sc.ei)
	}
	err := sc.execs[sc.ei]
	sc.ei++
	return mockResult{}, err
}

func (s *mockStmt) Query(_ []driver.Value) (driver.Rows, error) {
	sc := s.c.s
	if sc.qi >= len(sc.queries) {
		return nil, fmt.Errorf("unexpected Query call (index %d)", sc.qi)
	}
	r := sc.queries[sc.qi]
	sc.qi++
	if r.err != nil {
		return nil, r.err
	}
	return &mockRows{cols: r.cols, data: r.rows}, nil
}

type mockResult struct{}

func (mockResult) LastInsertId() (int64, error) { return 0, nil }
func (mockResult) RowsAffected() (int64, error) { return 1, nil }

type mockRows struct {
	cols []string
	data [][]driver.Value
	i    int
}

func (r *mockRows) Columns() []string { return r.cols }
func (r *mockRows) Close() error      { return nil }
func (r *mockRows) Next(dest []driver.Value) error {
	if r.i >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.i])
	r.i++
	return nil
}

// userRow returns a driver row matching the column order used by sqlc's
// CreateUser query:
// id, email, first_name, last_name, user_type, provider, is_active,
// created_at, last_login_at, requested_role, approved_by
func userRow(id int64, email string) queryResult {
	return queryResult{
		cols: []string{
			"id", "email", "first_name", "last_name", "user_type", "provider",
			"is_active", "created_at", "last_login_at", "requested_role", "approved_by",
		},
		rows: [][]driver.Value{
			{id, email, "John", "Doe", "student", "local", true, time.Now(), nil, nil, nil},
		},
	}
}

// countUsersRow odpovídá CountUsers — RegisterNewLocalUser ho volá jako
// úplně první dotaz (zjištění, jestli je registrující se uživatel první v
// systému).
func countUsersRow(n int64) queryResult {
	return queryResult{cols: []string{"count"}, rows: [][]driver.Value{{n}}}
}

// ---------------------------------------------------------------------------
// RegisterNewLocalUser
// ---------------------------------------------------------------------------

func TestRegisterNewLocalUser_Success(t *testing.T) {
	script := &connScript{
		queries: []queryResult{countUsersRow(5), userRow(42, "jane@example.com")},
		execs:   []error{nil}, // CreateLocalLogin succeeds; student nepotřebuje SetRequestedRole
	}
	db := newMockDB(t, script)

	req := userregistration.RegistrationRequest{
		Email:     "jane@example.com",
		Password:  "Secret1!",
		FirstName: "Jane",
		LastName:  "Doe",
		UserType:  "student",
	}

	id, err := userregistration.RegisterNewLocalUser(req, db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected id 42, got %d", id)
	}
}

func TestRegisterNewLocalUser_CreateUserError(t *testing.T) {
	script := &connScript{
		queries: []queryResult{countUsersRow(5), {err: fmt.Errorf("duplicate email")}},
	}
	db := newMockDB(t, script)

	req := userregistration.RegistrationRequest{
		Email:    "jane@example.com",
		Password: "Secret1!",
	}

	_, err := userregistration.RegisterNewLocalUser(req, db)
	if err == nil {
		t.Fatal("expected error from CreateUser, got nil")
	}
}

func TestRegisterNewLocalUser_CreateLocalLoginError(t *testing.T) {
	script := &connScript{
		queries: []queryResult{countUsersRow(5), userRow(1, "jane@example.com")},
		execs:   []error{fmt.Errorf("local login insert failed")},
	}
	db := newMockDB(t, script)

	req := userregistration.RegistrationRequest{
		Email:    "jane@example.com",
		Password: "Secret1!",
	}

	_, err := userregistration.RegisterNewLocalUser(req, db)
	if err == nil {
		t.Fatal("expected error from CreateLocalLogin, got nil")
	}
}

func TestRegisterNewLocalUser_CommitError(t *testing.T) {
	script := &connScript{
		queries:   []queryResult{countUsersRow(5), userRow(1, "jane@example.com")},
		execs:     []error{nil},
		commitErr: fmt.Errorf("commit failed"),
	}
	db := newMockDB(t, script)

	req := userregistration.RegistrationRequest{
		Email:    "jane@example.com",
		Password: "Secret1!",
	}

	_, err := userregistration.RegisterNewLocalUser(req, db)
	if err == nil {
		t.Fatal("expected error from Commit, got nil")
	}
}

// Student se aktivuje okamžitě — funkce nesmí volat SetRequestedRole (jediný
// dostupný exec slot je pro CreateLocalLogin; kdyby se navíc volal
// SetRequestedRole, vrátila by se chyba "unexpected Exec call").
func TestRegisterNewLocalUser_Student_SkipsApproval(t *testing.T) {
	script := &connScript{
		queries: []queryResult{countUsersRow(5), userRow(2, "student@example.com")},
		execs:   []error{nil},
	}
	db := newMockDB(t, script)

	req := userregistration.RegistrationRequest{
		Email:    "student@example.com",
		Password: "Secret1!",
		UserType: "student",
	}

	if _, err := userregistration.RegisterNewLocalUser(req, db); err != nil {
		t.Fatalf("student registration should not require SetRequestedRole, got error: %v", err)
	}
}

// Staff/maintainer čekají na schválení — funkce MUSÍ volat SetRequestedRole
// jako druhý exec. S jen jedním dostupným exec slotem (CreateLocalLogin)
// musí vrátit chybu, protože druhé volání narazí na "unexpected Exec call".
func TestRegisterNewLocalUser_StaffAndMaintainer_RequireApproval(t *testing.T) {
	for _, userType := range []string{"staff", "maintainer"} {
		t.Run(userType, func(t *testing.T) {
			script := &connScript{
				queries: []queryResult{countUsersRow(5), userRow(3, "pending@example.com")},
				execs:   []error{nil}, // jen CreateLocalLogin — SetRequestedRole nemá slot
			}
			db := newMockDB(t, script)

			req := userregistration.RegistrationRequest{
				Email:    "pending@example.com",
				Password: "Secret1!",
				UserType: userType,
			}

			if _, err := userregistration.RegisterNewLocalUser(req, db); err == nil {
				t.Fatalf("%s registration should call SetRequestedRole and fail with only one exec slot scripted", userType)
			}

			// Se dvěma exec sloty (CreateLocalLogin + SetRequestedRole) musí uspět.
			script2 := &connScript{
				queries: []queryResult{countUsersRow(5), userRow(3, "pending@example.com")},
				execs:   []error{nil, nil},
			}
			db2 := newMockDB(t, script2)
			if _, err := userregistration.RegisterNewLocalUser(req, db2); err != nil {
				t.Fatalf("%s registration with SetRequestedRole scripted should succeed, got: %v", userType, err)
			}
		})
	}
}
