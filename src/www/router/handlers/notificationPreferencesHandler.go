package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/notifications"
)

type NotificationPreferencesHandler struct {
	queries    *db.Queries
	httpLogger *logs.Logger
}

func NewNotificationPreferencesHandler(q *db.Queries, l *logs.Logger) *NotificationPreferencesHandler {
	return &NotificationPreferencesHandler{queries: q, httpLogger: l}
}

func (h *NotificationPreferencesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/notifications/preferences":
		h.get(w, r)
	case r.Method == http.MethodPut && r.URL.Path == "/api/notifications/preferences":
		h.put(w, r)
	default:
		defaultResponse(w)
	}
}

type notificationPreferencesResponse struct {
	EmailOptOuts []string `json:"emailOptOuts"`
}

type updatePreferencesRequest struct {
	EmailOptOuts []string `json:"emailOptOuts"`
}

// allowedOptOutTypes jsou typy, u kterých může uživatel vypnout email.
// urgent_ticket_broadcast nelze vypnout.
var allowedOptOutTypes = map[string]bool{
	notifications.TypeTicketResolved:   true,
	notifications.TypeTicketDeleted:    true,
	notifications.TypeTicketAssigned:   true,
	notifications.TypeRoleApproved:     true,
	notifications.TypeRoleRejected:     true,
	notifications.TypePriorityApproved: true,
}

func (h *NotificationPreferencesHandler) get(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	userID := int32(session.UserID)

	optOuts, err := h.queries.GetEmailOptOuts(r.Context(), userID)
	if err != nil {
		h.httpLogger.Debugf("get: GetEmailOptOuts selhalo (user=%d): %s", userID, err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst nastavení oznámení")
		return
	}
	if optOuts == nil {
		optOuts = []string{}
	}
	writeJSON(w, http.StatusOK, notificationPreferencesResponse{EmailOptOuts: optOuts})
}

func (h *NotificationPreferencesHandler) put(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	userID := int32(session.UserID)
	ctx := r.Context()

	var body updatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}

	// Sestavíme množinu požadovaných opt-outů (jen povolené typy).
	desired := make(map[string]bool)
	for _, t := range body.EmailOptOuts {
		if allowedOptOutTypes[t] {
			desired[t] = true
		}
	}

	// Načteme současné opt-outy a rozdílově aktualizujeme.
	current, err := h.queries.GetEmailOptOuts(ctx, userID)
	if err != nil {
		h.httpLogger.Debugf("put: GetEmailOptOuts selhalo (user=%d): %s", userID, err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst nastavení oznámení")
		return
	}
	currentSet := make(map[string]bool, len(current))
	for _, t := range current {
		currentSet[t] = true
	}

	for t := range desired {
		if !currentSet[t] {
			if err := h.queries.UpsertEmailOptOut(ctx, db.UpsertEmailOptOutParams{UserID: userID, Type: t}); err != nil {
				h.httpLogger.Debugf("put: UpsertEmailOptOut selhalo (user=%d, type=%s): %s", userID, t, err)
				WriteError(w, http.StatusInternalServerError, "nepodařilo se uložit nastavení")
				return
			}
		}
	}
	for t := range currentSet {
		if !desired[t] {
			if err := h.queries.DeleteEmailOptOut(ctx, db.DeleteEmailOptOutParams{UserID: userID, Type: t}); err != nil {
				h.httpLogger.Debugf("put: DeleteEmailOptOut selhalo (user=%d, type=%s): %s", userID, t, err)
				WriteError(w, http.StatusInternalServerError, "nepodařilo se uložit nastavení")
				return
			}
		}
	}

	// Vrátíme aktuální stav.
	updated, err := h.queries.GetEmailOptOuts(ctx, userID)
	if err != nil {
		h.httpLogger.Debugf("put: GetEmailOptOuts (po update) selhalo (user=%d): %s", userID, err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst nastavení oznámení")
		return
	}
	if updated == nil {
		updated = []string{}
	}
	writeJSON(w, http.StatusOK, notificationPreferencesResponse{EmailOptOuts: updated})
}
