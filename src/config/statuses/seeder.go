package statuses

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/StepanKomis/Ticketa/src/config"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

// upserter is the subset of db.Queries needed by Seed.
// db.Queries satisfies this interface automatically.
type upserter interface {
	UpsertTicketStatusByPosition(ctx context.Context, arg db.UpsertTicketStatusByPositionParams) (db.TicketStatus, error)
}

// Seed upserts ticket statuses from config into the database.
// Position is the zero-based slice index (0 = open state, last = solved state).
// The upsert is idempotent — re-running with the same config is safe.
// Empty Color fields receive a random CSS hex value.
func Seed(ctx context.Context, q upserter, statuses []config.StatusConfig) error {
	if len(statuses) < 3 {
		return fmt.Errorf("statuses: at least 3 ticket statuses required, got %d", len(statuses))
	}
	for i, s := range statuses {
		color := s.Color
		if color == "" {
			color = randomHex()
		}
		_, err := q.UpsertTicketStatusByPosition(ctx, db.UpsertTicketStatusByPositionParams{
			Title:    s.Title,
			Color:    color,
			Position: int32(i),
		})
		if err != nil {
			return fmt.Errorf("statuses: upsert position %d (%q): %w", i, s.Title, err)
		}
	}
	return nil
}

func randomHex() string {
	return fmt.Sprintf("#%06x", rand.IntN(0x1000000))
}
