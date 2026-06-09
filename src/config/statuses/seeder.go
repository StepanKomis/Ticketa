package statuses

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/StepanKomis/Ticketa/src/config"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

// upserter je podmnožina db.Queries potřebná pro Seed.
// db.Queries toto rozhraní implementuje automaticky.
type upserter interface {
	UpsertTicketStatusByPosition(ctx context.Context, arg db.UpsertTicketStatusByPositionParams) (db.TicketStatus, error)
}

// Seed upsertuje stavy tiketů z konfigurace do databáze.
// Position je nulově indexovaný index v poli (0 = otevřeno, poslední = vyřešeno).
// Upsert je idempotentní — opakované spuštění se stejnou konfigurací je bezpečné.
// Prázdná pole Color obdrží náhodnou CSS hex hodnotu.
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
