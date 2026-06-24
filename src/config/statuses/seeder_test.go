package statuses_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/StepanKomis/Ticketa/src/config"
	"github.com/StepanKomis/Ticketa/src/config/statuses"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

// fakeUpserter records calls and optionally fails on a given call index.
type fakeUpserter struct {
	calls   []db.UpsertTicketStatusByPositionParams
	failOn  int // 1-based; 0 = never fail
	failErr error
}

func (f *fakeUpserter) UpsertTicketStatusByPosition(_ context.Context, arg db.UpsertTicketStatusByPositionParams) (db.TicketStatus, error) {
	f.calls = append(f.calls, arg)
	if f.failOn != 0 && len(f.calls) == f.failOn {
		return db.TicketStatus{}, f.failErr
	}
	return db.TicketStatus{
		Title:    arg.Title,
		Color:    arg.Color,
		Position: arg.Position,
		IsClosed: arg.IsClosed,
	}, nil
}

func threeStatuses() []config.StatusConfig {
	return []config.StatusConfig{
		{Title: "Otevřeno", Color: "#3498db"},
		{Title: "Probíhá", Color: "#f39c12"},
		{Title: "Vyřešeno", Color: "#2ecc71", IsClosed: true},
	}
}

func TestSeed_CallsUpsertForEachStatus(t *testing.T) {
	fake := &fakeUpserter{}
	if err := statuses.Seed(context.Background(), fake, threeStatuses()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.calls) != 3 {
		t.Fatalf("expected 3 upsert calls, got %d", len(fake.calls))
	}
	for i, call := range fake.calls {
		if int(call.Position) != i {
			t.Errorf("call[%d]: position = %d, want %d", i, call.Position, i)
		}
	}
	if fake.calls[0].Title != "Otevřeno" || fake.calls[2].Title != "Vyřešeno" {
		t.Errorf("titles do not match config: %+v", fake.calls)
	}
}

func TestSeed_EmptyColor_GeneratesValidHex(t *testing.T) {
	cfg := []config.StatusConfig{
		{Title: "Otevřeno", Color: ""},
		{Title: "Probíhá", Color: ""},
		{Title: "Vyřešeno", Color: "", IsClosed: true},
	}
	fake := &fakeUpserter{}
	if err := statuses.Seed(context.Background(), fake, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hexRe := regexp.MustCompile(`^#[0-9a-f]{6}$`)
	for i, call := range fake.calls {
		if !hexRe.MatchString(call.Color) {
			t.Errorf("call[%d]: color %q does not match #rrggbb", i, call.Color)
		}
	}
}

func TestSeed_TooFewStatuses_ReturnsError(t *testing.T) {
	fake := &fakeUpserter{}
	err := statuses.Seed(context.Background(), fake, []config.StatusConfig{
		{Title: "Jen jeden", Color: "#fff"},
		{Title: "Jen dva", Color: "#000"},
	})
	if err == nil {
		t.Fatal("expected error for < 3 statuses, got nil")
	}
	if len(fake.calls) != 0 {
		t.Errorf("expected no DB calls on validation error, got %d", len(fake.calls))
	}
}

func TestSeed_NoClosedStatus_ReturnsError(t *testing.T) {
	fake := &fakeUpserter{}
	err := statuses.Seed(context.Background(), fake, []config.StatusConfig{
		{Title: "Otevřeno", Color: "#3498db"},
		{Title: "Probíhá", Color: "#f39c12"},
		{Title: "Vyřešeno", Color: "#2ecc71"}, // chybí IsClosed
	})
	if err == nil {
		t.Fatal("expected error when no status has is_closed=true, got nil")
	}
	if len(fake.calls) != 0 {
		t.Errorf("expected no DB calls on validation error, got %d", len(fake.calls))
	}
}

func TestSeed_DBError_PropagatesAndStops(t *testing.T) {
	injected := errors.New("connection refused")
	fake := &fakeUpserter{failOn: 2, failErr: injected}

	err := statuses.Seed(context.Background(), fake, threeStatuses())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, injected) {
		t.Errorf("error should wrap injected error; got: %v", err)
	}
	// First call succeeded, second failed, third must not have run.
	if len(fake.calls) != 2 {
		t.Errorf("expected 2 calls before stop, got %d", len(fake.calls))
	}
}
