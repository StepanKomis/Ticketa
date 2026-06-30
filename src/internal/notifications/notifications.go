package notifications

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/mailer"
)

const (
	TypeTicketResolved   = "ticket_resolved"
	TypeTicketDeleted    = "ticket_deleted"
	TypeTicketAssigned   = "ticket_assigned"
	TypeRoleApproved     = "role_approved"
	TypeRoleRejected     = "role_rejected"
	TypePriorityApproved = "priority_approved"
)

// Notifier zapisuje oznámení do DB a odesílá emaily. Nulový *Notifier je platný no-op —
// handlery i testy ho mohou předávat jako nil bez podmínek na volající straně.
type Notifier struct {
	queries *db.Queries
	logger  *logs.Logger
	mailer  *mailer.Mailer
}

func NewNotifier(q *db.Queries, l *logs.Logger, m *mailer.Mailer) *Notifier {
	return &Notifier{queries: q, logger: l, mailer: m}
}

// emailForUser vrátí e-mailovou adresu uživatele, nebo "" při chybě.
func (n *Notifier) emailForUser(ctx context.Context, userID int32) string {
	if n == nil || n.mailer == nil {
		return ""
	}
	u, err := n.queries.GetUserByID(ctx, userID)
	if err != nil {
		return ""
	}
	return u.Email
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
	text := fmt.Sprintf("Váš ticket byl vyřešen: %s", title)
	n.send(ctx, authorID, TypeTicketResolved, text, &ticketID)
	if email := n.emailForUser(ctx, authorID); email != "" {
		go n.mailer.Send(email, "Ticket vyřešen", text)
	}
}

func (n *Notifier) NotifyTicketDeleted(ctx context.Context, authorID int32, ticketID int64, title string) {
	text := fmt.Sprintf("Váš ticket byl smazán: %s", title)
	n.send(ctx, authorID, TypeTicketDeleted, text, &ticketID)
	if email := n.emailForUser(ctx, authorID); email != "" {
		go n.mailer.Send(email, "Ticket smazán", text)
	}
}

func (n *Notifier) NotifyTicketAssigned(ctx context.Context, assigneeID int32, ticketID int64, title string) {
	text := fmt.Sprintf("Byl vám přidělen tiket: %s", title)
	n.send(ctx, assigneeID, TypeTicketAssigned, text, &ticketID)
	if email := n.emailForUser(ctx, assigneeID); email != "" {
		go n.mailer.Send(email, "Přidělení tiketu", text)
	}
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
	text := fmt.Sprintf("Vaše žádost o roli %s byla schválena.", czechRoleLabel(role))
	n.send(ctx, userID, TypeRoleApproved, text, nil)
	if email := n.emailForUser(ctx, userID); email != "" {
		go n.mailer.Send(email, "Žádost o roli schválena", text)
	}
}

func (n *Notifier) NotifyRoleRejected(ctx context.Context, userID int32, role string) {
	text := fmt.Sprintf("Vaše žádost o roli %s byla zamítnuta.", czechRoleLabel(role))
	n.send(ctx, userID, TypeRoleRejected, text, nil)
	if email := n.emailForUser(ctx, userID); email != "" {
		go n.mailer.Send(email, "Žádost o roli zamítnuta", text)
	}
}

func (n *Notifier) NotifyPriorityApproved(ctx context.Context, authorID int32, ticketID int64, title string) {
	text := fmt.Sprintf("Urgentní priorita vašeho ticketu byla schválena: %s", title)
	n.send(ctx, authorID, TypePriorityApproved, text, &ticketID)
	if email := n.emailForUser(ctx, authorID); email != "" {
		go n.mailer.Send(email, "Priorita schválena", text)
	}
}
