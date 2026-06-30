package handlers

import (
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

type NotificationHandler struct {
	queries    *db.Queries
	httpLogger *logs.Logger
}

func NewNotificationHandler(q *db.Queries, l *logs.Logger) *NotificationHandler {
	return &NotificationHandler{queries: q, httpLogger: l}
}

func (h *NotificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/notifications":
		h.list(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/notifications/mark-viewed":
		h.markViewed(w, r)
	default:
		defaultResponse(w)
	}
}

func (h *NotificationHandler) list(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	userID := int32(session.UserID)
	ctx := r.Context()

	rows, err := h.queries.GetNotificationsForUser(ctx, userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst oznámení")
		return
	}
	unreadCount, err := h.queries.CountUnreadNotifications(ctx, userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se spočítat oznámení")
		return
	}

	items := make([]notificationResponse, len(rows))
	for i, row := range rows {
		items[i] = notificationResponse{
			ID:        row.ID,
			Type:      row.Type,
			Text:      row.Text,
			IsViewed:  row.IsViewed,
			CreatedAt: row.CreatedAt,
		}
		if row.TicketID.Valid {
			v := row.TicketID.Int64
			items[i].TicketID = &v
		}
	}
	writeJSON(w, http.StatusOK, notificationListResponse{
		Items:       items,
		UnreadCount: unreadCount,
	})
}

func (h *NotificationHandler) markViewed(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	if err := h.queries.MarkAllNotificationsViewed(r.Context(), int32(session.UserID)); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se označit oznámení jako přečtená")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
