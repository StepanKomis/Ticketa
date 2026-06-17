package activity

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/sqlc-dev/pqtype"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ActivityLogger zapisuje audit log akcí současně do DB tabulky activity_log
// a do souboru ve formátu JSON Lines (rotovaného přes lumberjack). Nulový
// *ActivityLogger je platný no-op — handlery i testy ho mohou předávat jako
// nil bez podmínek na volající straně.
type ActivityLogger struct {
	queries *db.Queries
	file    io.Writer
	logger  *logs.Logger
}

func NewActivityLogger(q *db.Queries, cfg *config.Config, l *logs.Logger) *ActivityLogger {
	return &ActivityLogger{
		queries: q,
		file: &lumberjack.Logger{
			Filename: cfg.Activity.LogFile,
			MaxAge:   cfg.Activity.LogMaxDays,
		},
		logger: l,
	}
}

func (a *ActivityLogger) log(ctx context.Context, eventType EventType, actorID int32, targetType string, targetID int64, payload map[string]any) {
	if a == nil {
		return
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		a.logger.Warnf("activity log: nepodařilo se serializovat payload pro %s: %s", eventType, err)
		payloadJSON = []byte("{}")
	}

	if err := a.queries.CreateActivityLog(ctx, db.CreateActivityLogParams{
		EventType:  string(eventType),
		ActorID:    sql.NullInt32{Int32: actorID, Valid: actorID != 0},
		TargetType: sql.NullString{String: targetType, Valid: targetType != ""},
		TargetID:   sql.NullInt64{Int64: targetID, Valid: targetID != 0},
		Payload:    pqtype.NullRawMessage{RawMessage: payloadJSON, Valid: true},
	}); err != nil {
		a.logger.Warnf("activity log: nepodařilo se zapsat %s do DB: %s", eventType, err)
	}

	line, err := json.Marshal(map[string]any{
		"event_type":  eventType,
		"actor_id":    actorID,
		"target_type": targetType,
		"target_id":   targetID,
		"payload":     payload,
		"created_at":  time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		a.logger.Warnf("activity log: nepodařilo se serializovat řádek souboru pro %s: %s", eventType, err)
		return
	}
	if _, err := a.file.Write(append(line, '\n')); err != nil {
		a.logger.Warnf("activity log: nepodařilo se zapsat %s do souboru: %s", eventType, err)
	}
}

// LogTiketVytvoren zaznamená vznik nového tiketu.
func (a *ActivityLogger) LogTiketVytvoren(ctx context.Context, actorID int32, ticketID int64, title string) {
	a.log(ctx, EventTiketVytvoren, actorID, TargetTicket, ticketID, map[string]any{"title": title})
}

// LogTiketAktualizovan zaznamená úpravu obsahu tiketu (název/popis).
func (a *ActivityLogger) LogTiketAktualizovan(ctx context.Context, actorID int32, ticketID int64, changedFields []string) {
	a.log(ctx, EventTiketAktualizovan, actorID, TargetTicket, ticketID, map[string]any{"changed_fields": changedFields})
}

// LogStavZmenen zaznamená změnu stavu tiketu.
func (a *ActivityLogger) LogStavZmenen(ctx context.Context, actorID int32, ticketID int64, oldStatus, newStatus string) {
	a.log(ctx, EventTiketStavZmenen, actorID, TargetTicket, ticketID, map[string]any{
		"old_status": oldStatus,
		"new_status": newStatus,
	})
}

// LogTiketPrirazen zaznamená změnu přiřazeného řešitele (přiřazení i odebrání).
func (a *ActivityLogger) LogTiketPrirazen(ctx context.Context, actorID int32, ticketID int64, oldAssigneeID, newAssigneeID *int32) {
	a.log(ctx, EventTiketPrirazen, actorID, TargetTicket, ticketID, map[string]any{
		"old_assignee_id": oldAssigneeID,
		"new_assignee_id": newAssigneeID,
	})
}

// LogTiketSmazan zaznamená smazání tiketu.
func (a *ActivityLogger) LogTiketSmazan(ctx context.Context, actorID int32, ticketID int64, title string) {
	a.log(ctx, EventTiketSmazan, actorID, TargetTicket, ticketID, map[string]any{"title": title})
}

// LogKomentarVytvoren zaznamená vytvoření komentáře k tiketu.
func (a *ActivityLogger) LogKomentarVytvoren(ctx context.Context, actorID int32, commentID, ticketID int64) {
	a.log(ctx, EventKomentarVytvoren, actorID, TargetComment, commentID, map[string]any{"ticket_id": ticketID})
}

// LogKomentarAktualizovan zaznamená úpravu komentáře.
func (a *ActivityLogger) LogKomentarAktualizovan(ctx context.Context, actorID int32, commentID, ticketID int64) {
	a.log(ctx, EventKomentarAktualizovan, actorID, TargetComment, commentID, map[string]any{"ticket_id": ticketID})
}

// LogKomentarSmazan zaznamená smazání komentáře.
func (a *ActivityLogger) LogKomentarSmazan(ctx context.Context, actorID int32, commentID, ticketID int64) {
	a.log(ctx, EventKomentarSmazan, actorID, TargetComment, commentID, map[string]any{"ticket_id": ticketID})
}

// LogUzivatelRegistrovan zaznamená registraci nového uživatele (aktérem je uživatel samotný).
func (a *ActivityLogger) LogUzivatelRegistrovan(ctx context.Context, userID int32, email string) {
	a.log(ctx, EventUzivatelRegistrovan, userID, TargetUser, int64(userID), map[string]any{"email": email})
}

// LogUzivatelSchvalen zaznamená schválení čekajícího uživatele správcem.
func (a *ActivityLogger) LogUzivatelSchvalen(ctx context.Context, actorID, targetUserID int32, email string) {
	a.log(ctx, EventUzivatelSchvalen, actorID, TargetUser, int64(targetUserID), map[string]any{"email": email})
}

// LogUzivatelZamitnuv zaznamená zamítnutí čekajícího uživatele správcem.
func (a *ActivityLogger) LogUzivatelZamitnuv(ctx context.Context, actorID, targetUserID int32, email string) {
	a.log(ctx, EventUzivatelZamitnuv, actorID, TargetUser, int64(targetUserID), map[string]any{"email": email})
}

// LogUzivatelDeaktivovan zaznamená deaktivaci uživatelského účtu správcem.
func (a *ActivityLogger) LogUzivatelDeaktivovan(ctx context.Context, actorID, targetUserID int32, email string) {
	a.log(ctx, EventUzivatelDeaktivovan, actorID, TargetUser, int64(targetUserID), map[string]any{"email": email})
}
