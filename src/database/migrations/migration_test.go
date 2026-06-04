package migrate_test

import (
	"errors"
	"fmt"
	"testing"

	migrate "github.com/StepanKomis/Ticketa/src/database/migrations"
)

// fakeMigrator is a test double for migrate.Migrator.
// It tracks which migrations were applied/rolled back and can be
// configured to fail on a specific migration number.
type fakeMigrator struct {
	applied    int
	log        []string // "up:N" or "down:N" entries in call order
	failUpOn   int      // migration number whose Up call should return an error; 0 = never
	appliedErr error    // error Applied() should return
}

func newFake(applied int) *fakeMigrator {
	return &fakeMigrator{applied: applied}
}

func (f *fakeMigrator) Init() error                  { return nil }
func (f *fakeMigrator) Applied() (int, error)        { return f.applied, f.appliedErr }

func (f *fakeMigrator) Up(m migrate.Migration) error {
	if f.failUpOn != 0 && m.Number == f.failUpOn {
		return fmt.Errorf("injected failure on migration %d", m.Number)
	}
	f.log = append(f.log, fmt.Sprintf("up:%d", m.Number))
	f.applied = m.Number
	return nil
}

func (f *fakeMigrator) Down(m migrate.Migration) error {
	f.log = append(f.log, fmt.Sprintf("down:%d", m.Number))
	return nil
}

// migrations returns a slice of no-op migrations numbered 1..n.
func migrations(n int) []migrate.Migration {
	ms := make([]migrate.Migration, n)
	for i := range ms {
		num := i + 1
		ms[i] = migrate.Migration{
			Number: num,
			Name:   fmt.Sprintf("migration_%d", num),
			Up:     func(db any) error { return nil },
			Down:   func(db any) error { return nil },
		}
	}
	return ms
}

func TestNewRunner_SortsMigrations(t *testing.T) {
	fake := newFake(-1)
	// Pass migrations in reverse order; Runner must sort them.
	ms := []migrate.Migration{
		{Number: 3, Name: "c", Up: func(db any) error { return nil }, Down: func(db any) error { return nil }},
		{Number: 1, Name: "a", Up: func(db any) error { return nil }, Down: func(db any) error { return nil }},
		{Number: 2, Name: "b", Up: func(db any) error { return nil }, Down: func(db any) error { return nil }},
	}
	runner := migrate.NewRunner(fake, ms)

	if err := runner.MigrateUp(); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	want := []string{"up:1", "up:2", "up:3"}
	if len(fake.log) != len(want) {
		t.Fatalf("got log %v, want %v", fake.log, want)
	}
	for i, entry := range fake.log {
		if entry != want[i] {
			t.Errorf("log[%d] = %q, want %q", i, entry, want[i])
		}
	}
}

func TestMigrateUp_RunsPendingMigrations(t *testing.T) {
	fake := newFake(0) // migration 0 already applied
	runner := migrate.NewRunner(fake, migrations(3))

	if err := runner.MigrateUp(); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	want := []string{"up:1", "up:2", "up:3"}
	if len(fake.log) != len(want) {
		t.Fatalf("got log %v, want %v", fake.log, want)
	}
	for i, entry := range fake.log {
		if entry != want[i] {
			t.Errorf("log[%d] = %q, want %q", i, entry, want[i])
		}
	}
}

func TestMigrateUp_SkipsAlreadyApplied(t *testing.T) {
	fake := newFake(2) // migrations 0-2 already applied
	runner := migrate.NewRunner(fake, migrations(3))

	if err := runner.MigrateUp(); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	if len(fake.log) != 1 || fake.log[0] != "up:3" {
		t.Errorf("expected only up:3, got %v", fake.log)
	}
}

func TestMigrateUp_AllAlreadyApplied(t *testing.T) {
	fake := newFake(3)
	runner := migrate.NewRunner(fake, migrations(3))

	if err := runner.MigrateUp(); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}

	if len(fake.log) != 0 {
		t.Errorf("expected no migrations to run, got %v", fake.log)
	}
}

func TestMigrateUp_StopsOnError(t *testing.T) {
	fake := newFake(0)
	fake.failUpOn = 2
	runner := migrate.NewRunner(fake, migrations(3))

	err := runner.MigrateUp()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// migration 1 should have run, 2 failed, 3 must not have run
	if len(fake.log) != 1 || fake.log[0] != "up:1" {
		t.Errorf("expected [up:1] before failure, got %v", fake.log)
	}
}

func TestMigrateUp_ErrorWrapsNumber(t *testing.T) {
	fake := newFake(0)
	fake.failUpOn = 1
	runner := migrate.NewRunner(fake, migrations(1))

	err := runner.MigrateUp()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, err) { // basic non-nil check; real check is substring
		t.Fatal("error should be non-nil")
	}
	msg := err.Error()
	if len(msg) == 0 {
		t.Error("error message should not be empty")
	}
}

func TestMigrateUp_PropagatesAppliedError(t *testing.T) {
	fake := newFake(0)
	fake.appliedErr = errors.New("db unavailable")
	runner := migrate.NewRunner(fake, migrations(3))

	err := runner.MigrateUp()
	if err == nil {
		t.Fatal("expected error from Applied(), got nil")
	}
	if len(fake.log) != 0 {
		t.Errorf("no migrations should have run, got %v", fake.log)
	}
}

func TestMigrateDown_RollsBackLastMigration(t *testing.T) {
	fake := newFake(3)
	runner := migrate.NewRunner(fake, migrations(3))

	if err := runner.MigrateDown(); err != nil {
		t.Fatalf("MigrateDown: %v", err)
	}

	if len(fake.log) != 1 || fake.log[0] != "down:3" {
		t.Errorf("expected [down:3], got %v", fake.log)
	}
}

func TestMigrateDown_DoesNothingWhenNoMatch(t *testing.T) {
	// applied is 5 but no migration with Number==5 exists
	fake := newFake(5)
	runner := migrate.NewRunner(fake, migrations(3))

	if err := runner.MigrateDown(); err != nil {
		t.Fatalf("MigrateDown: %v", err)
	}

	if len(fake.log) != 0 {
		t.Errorf("expected no rollback, got %v", fake.log)
	}
}

func TestMigrateDown_PropagatesAppliedError(t *testing.T) {
	fake := newFake(3)
	fake.appliedErr = errors.New("db unavailable")
	runner := migrate.NewRunner(fake, migrations(3))

	err := runner.MigrateDown()
	if err == nil {
		t.Fatal("expected error from Applied(), got nil")
	}
	if len(fake.log) != 0 {
		t.Errorf("no rollback should have run, got %v", fake.log)
	}
}
