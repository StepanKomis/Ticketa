package notifications

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

const (
	TypeTicketResolved   = "ticket_resolved"
	TypeTicketDeleted    = "ticket_deleted"
	TypeTicketAssigned   = "ticket_assigned"
	TypeRoleApproved     = "role_approved"
	TypeRoleRejected     = "role_rejected"
	TypePriorityApproved = "priority_approved"
)

// Notifier zapisuje oznámení do DB. Nulový *Notifier je platný no-op —
// handlery i testy ho mohou předávat jako nil bez podmínek na volající straně.
type Notifier struct {
	queries *db.Queries
	logger  *logs.Logger
}

func NewNotifier(q *db.Queries, l *logs.Logger) *Notifier {
	return &Notifier{queries: q, logger: l}
}

func (n *Notifier) send(ctx context.Context, userID int32, notifType, text string, ticketID *int64) {
	if n == nil {
		return
	}
	var tid sql.NullInt64
	if ticketID != nil {
		tid = sql.NullInt64{Int64: *ticketID, Valid: true}
	}
	if _, err := n.queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:   userID,
		Type:     notifType,
		Text:     text,
		TicketID: tid,
	}); err != nil {
		n.logger.Debugf("notifications: CreateNotification selhalo (user=%d type=%s): %s", userID, notifType, err)
	}
}

func (n *Notifier) NotifyTicketResolved(ctx context.Context, authorID int32, ticketID int64, title string) {
	n.send(ctx, authorID, TypeTicketResolved, fmt.Sprintf("Váš ticket byl vyřešen: %s", title), &ticketID)
}

func (n *Notifier) NotifyTicketDeleted(ctx context.Context, authorID int32, ticketID int64, title string) {
	n.send(ctx, authorID, TypeTicketDeleted, fmt.Sprintf("Váš ticket byl smazán: %s", title), &ticketID)
}

func (n *Notifier) NotifyTicketAssigned(ctx context.Context, assigneeID int32, ticketID int64, title string) {
	n.send(ctx, assigneeID, TypeTicketAssigned, fmt.Sprintf("Byl vám přidělen tiket: %s", title), &ticketID)
}

func czechRoleLabel(role string) string {
	switch role {
	case "staff":
		return "Učitel"
	case "maintainer":
		return "Údržbář"
	case "admin":
		return "Správce"
	default:
		return role
	}
}

func (n *Notifier) NotifyRoleApproved(ctx context.Context, userID int32, role string) {
	n.send(ctx, userID, TypeRoleApproved, fmt.Sprintf("Vaše žádost o roli %s byla schválena.", czechRoleLabel(role)), nil)
}

func (n *Notifier) NotifyRoleRejected(ctx context.Context, userID int32, role string) {
	n.send(ctx, userID, TypeRoleRejected, fmt.Sprintf("Vaše žádost o roli %s byla zamítnuta.", czechRoleLabel(role)), nil)
}

func (n *Notifier) NotifyPriorityApproved(ctx context.Context, authorID int32, ticketID int64, title string) {
	n.send(ctx, authorID, TypePriorityApproved, fmt.Sprintf("Urgentní priorita vašeho ticketu byla schválena: %s", title), &ticketID)
}
