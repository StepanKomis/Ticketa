package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/activity"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
)

type TicketHandler struct {
	queries        *db.Queries
	httpLogger     *logs.Logger
	activityLogger *activity.ActivityLogger
}

func NewTicketHandler(q *db.Queries, l *logs.Logger, al *activity.ActivityLogger) *TicketHandler {
	return &TicketHandler{queries: q, httpLogger: l, activityLogger: al}
}

func (h *TicketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/tickets":
		h.create(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/tickets":
		h.list(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/history"):
		h.listHistory(w, r)
	case r.Method == http.MethodGet && matchesIDPath(r.URL.Path, "/api/tickets/") && !strings.HasSuffix(r.URL.Path, "/vote"):
		h.get(w, r)
	case r.Method == http.MethodPut && matchesIDPath(r.URL.Path, "/api/tickets/"):
		h.update(w, r)
	case r.Method == http.MethodPatch && matchesIDPath(r.URL.Path, "/api/tickets/"):
		h.patch(w, r)
	case r.Method == http.MethodDelete && matchesIDPath(r.URL.Path, "/api/tickets/") && !strings.HasSuffix(r.URL.Path, "/vote"):
		h.delete(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/vote"):
		h.vote(w, r)
	case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/vote"):
		h.unvote(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/approve-priority"):
		h.approvePriority(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/reject-priority"):
		h.rejectPriority(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/claim"):
		h.claim(w, r)
	default:
		defaultResponse(w)
	}
}

type createTicketRequest struct {
	Title      string `json:"title"`
	Body       string `json:"body"`
	Priority   string `json:"priority"`
	Location   string `json:"location"`
	Category   string `json:"category"`
	AssignedTo *int32 `json:"assigned_to"`
	StatusID   *int32 `json:"status_id"`
}

type updateTicketRequest struct {
	Title          *string `json:"title"`
	Body           *string `json:"body"`
	Priority       *string `json:"priority"`
	Location       *string `json:"location"`
	Category       *string `json:"category"`
	ResolutionNote *string `json:"resolution_note"`
	StatusID       *int32  `json:"status_id"`
}

type patchTicketRequest struct {
	AssignedTo     *int32  `json:"assigned_to"`
	StatusID       *int32  `json:"status_id"`
	Priority       *string `json:"priority"`
	Location       *string `json:"location"`
	Category       *string `json:"category"`
	ResolutionNote *string `json:"resolution_note"`
}

var validPriorities = map[string]bool{
	"low": true, "medium": true, "high": true, "urgent": true,
}

func (h *TicketHandler) create(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	var body createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if body.Title == "" {
		WriteError(w, http.StatusUnprocessableEntity, "pole title je povinné")
		return
	}
	if body.Body == "" {
		WriteError(w, http.StatusUnprocessableEntity, "pole body je povinné")
		return
	}
	if body.Priority == "" {
		body.Priority = "medium"
	}
	if !validPriorities[body.Priority] {
		WriteError(w, http.StatusUnprocessableEntity, "neplatná priorita (povolené hodnoty: low, medium, high, urgent)")
		return
	}

	user, ok := userFromContext(w, r)
	if !ok {
		return
	}

	params := db.CreateTicketParams{
		Title:    body.Title,
		Body:     body.Body,
		AuthorID: int32(session.UserID),
		Priority: body.Priority,
		Location: body.Location,
		Category: body.Category,
	}
	if requiresPriorityApproval(user.UserType, body.Priority) {
		// Dokud staff/admin žádost neschválí, zůstává efektivní priorita "high"
		// (druhá nejvyšší) — tiket se neztratí v běžném provozu, ale ani se
		// nezobrazuje jako urgentní bez kontroly.
		params.Priority = "high"
		params.RequestedPriority = sql.NullString{String: body.Priority, Valid: true}
	}
	if body.StatusID != nil {
		params.StatusID = sql.NullInt32{Int32: *body.StatusID, Valid: true}
	}
	// assigned_to lze nastavit pouze staff nebo admin, a jen na řešitele
	// (staff/maintainer/admin) — neplatnou kombinaci tiše ignorujeme, stejně
	// jako dnešní kontrolu role zadavatele.
	if body.AssignedTo != nil && (user.UserType == db.UserTypeStaff || user.UserType == db.UserTypeAdmin) &&
		h.isAssignableTarget(r.Context(), *body.AssignedTo) {
		params.AssignedTo = sql.NullInt32{Int32: *body.AssignedTo, Valid: true}
	}

	ticket, err := h.queries.CreateTicket(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se vytvořit tiket")
		return
	}
	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	h.logHistory(r.Context(), ticket.ID, int32(session.UserID), actorName, "created", "", "")
	h.activityLogger.LogTiketVytvoren(r.Context(), int32(session.UserID), ticket.ID, ticket.Title)
	if ticket.RequestedPriority.Valid {
		h.logHistory(r.Context(), ticket.ID, int32(session.UserID), actorName, "priority_approval_requested", "", ticket.RequestedPriority.String)
		h.activityLogger.LogTiketPrioritaKeSchvaleni(r.Context(), int32(session.UserID), ticket.ID, ticket.RequestedPriority.String)
	}
	writeJSON(w, http.StatusCreated, toTicketResponseFromTicket(ticket, "", int32(session.UserID)))
}

func (h *TicketHandler) list(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	q := r.URL.Query()
	limit := 20
	offset := 0
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	params := db.ListTicketsFilteredParams{
		CurrentUserID: int32(session.UserID),
		Q:             q.Get("q"),
		Lim:           int32(limit),
		Off:           int32(offset),
	}
	if v := q.Get("status_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.StatusID = sql.NullInt32{Int32: int32(n), Valid: true}
		}
	}
	if v := q.Get("priority"); v != "" && validPriorities[v] {
		params.Priority = sql.NullString{String: v, Valid: true}
	}
	if v := q.Get("assigned_to"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.AssignedTo = sql.NullInt32{Int32: int32(n), Valid: true}
		}
	}
	if v := q.Get("author_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.AuthorID = sql.NullInt32{Int32: int32(n), Valid: true}
		}
	}
	if v := q.Get("category"); v != "" {
		params.Category = sql.NullString{String: v, Valid: true}
	}
	if v := q.Get("pending_priority_approval"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			params.PendingPriorityApproval = sql.NullBool{Bool: b, Valid: true}
		}
	}
	if v := q.Get("unassigned"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			params.UnassignedOnly = sql.NullBool{Bool: b, Valid: true}
		}
	}
	if v := q.Get("closed"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			params.Closed = sql.NullBool{Bool: b, Valid: true}
		}
	}

	rows, err := h.queries.ListTicketsFiltered(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tikety")
		return
	}

	countParams := db.CountTicketsFilteredParams{
		StatusID:                params.StatusID,
		Priority:                params.Priority,
		AssignedTo:              params.AssignedTo,
		AuthorID:                params.AuthorID,
		Category:                params.Category,
		PendingPriorityApproval: params.PendingPriorityApproval,
		UnassignedOnly:          params.UnassignedOnly,
		Closed:                  params.Closed,
		Q:                       params.Q,
	}
	total, err := h.queries.CountTicketsFiltered(r.Context(), countParams)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se spočítat tikety")
		return
	}

	items := make([]ticketResponse, len(rows))
	for i, row := range rows {
		items[i] = toTicketResponseFromRow(row)
	}
	writeJSON(w, http.StatusOK, ticketListResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *TicketHandler) get(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	row, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{
		ID:            id,
		CurrentUserID: int32(session.UserID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}
	writeJSON(w, http.StatusOK, toTicketResponseFromGetRow(row, int32(session.UserID)))
}

func (h *TicketHandler) update(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	user, ok := userFromContext(w, r)
	if !ok {
		return
	}

	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	existing, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{
		ID:            id,
		CurrentUserID: int32(session.UserID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}

	if !canEditContent(session, existing, user.UserType) {
		WriteError(w, http.StatusForbidden, "přístup odepřen")
		return
	}

	var body updateTicketRequest
	presence, err := decodeWithPresence(r, &body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if body.Priority != nil && !validPriorities[*body.Priority] {
		WriteError(w, http.StatusUnprocessableEntity, "neplatná priorita (povolené hodnoty: low, medium, high, urgent)")
		return
	}

	params := db.UpdateTicketParams{ID: id}
	if body.Title != nil {
		params.Title = sql.NullString{String: *body.Title, Valid: true}
	}
	if body.Body != nil {
		params.Body = sql.NullString{String: *body.Body, Valid: true}
	}
	if body.Priority != nil {
		if requiresPriorityApproval(user.UserType, *body.Priority) {
			// Žádost o urgentní prioritu — efektivní priorita zůstává nezměněná,
			// dokud ji staff/admin neschválí.
			params.RequestedPriority = sql.NullString{String: *body.Priority, Valid: true}
		} else {
			params.Priority = sql.NullString{String: *body.Priority, Valid: true}
		}
	}
	if body.Location != nil {
		params.Location = sql.NullString{String: *body.Location, Valid: true}
	}
	if body.Category != nil {
		params.Category = sql.NullString{String: *body.Category, Valid: true}
	}
	if body.ResolutionNote != nil {
		params.ResolutionNote = sql.NullString{String: *body.ResolutionNote, Valid: true}
	}
	// touch_status_id: status_id se mění jen pokud byl v requestu skutečně
	// uveden — jinak by každá úprava title/body bez status_id vynulovala stav.
	if _, touched := presence["status_id"]; touched {
		params.TouchStatusID = true
		if body.StatusID != nil {
			params.StatusID = sql.NullInt32{Int32: *body.StatusID, Valid: true}
		}
	}

	ticket, err := h.queries.UpdateTicket(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat tiket")
		return
	}
	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	if body.Title != nil && existing.Title != ticket.Title {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "title_changed", existing.Title, ticket.Title)
	}
	if body.Body != nil && existing.Body != ticket.Body {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "content_updated", "", "tělo")
	}
	if body.Title != nil || body.Body != nil {
		changedFields := make([]string, 0, 2)
		if body.Title != nil {
			changedFields = append(changedFields, "title")
		}
		if body.Body != nil {
			changedFields = append(changedFields, "body")
		}
		h.activityLogger.LogTiketAktualizovan(r.Context(), int32(session.UserID), id, changedFields)
	}
	if body.StatusID != nil {
		oldStatus := h.resolveStatusTitle(r.Context(), existing.StatusID)
		newStatus := h.resolveStatusTitle(r.Context(), ticket.StatusID)
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "status_changed", oldStatus, newStatus)
		h.activityLogger.LogStavZmenen(r.Context(), int32(session.UserID), id, oldStatus, newStatus)
	}
	if body.Priority != nil && existing.Priority != ticket.Priority {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "priority_changed", existing.Priority, ticket.Priority)
	}
	if ticket.RequestedPriority.Valid && existing.RequestedPriority.String != ticket.RequestedPriority.String {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "priority_approval_requested", "", ticket.RequestedPriority.String)
		h.activityLogger.LogTiketPrioritaKeSchvaleni(r.Context(), int32(session.UserID), id, ticket.RequestedPriority.String)
	}
	if body.Location != nil && existing.Location != ticket.Location {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "location_changed", existing.Location, ticket.Location)
	}
	if body.Category != nil && existing.Category != ticket.Category {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "category_changed", existing.Category, ticket.Category)
	}
	if body.ResolutionNote != nil && existing.ResolutionNote.String != ticket.ResolutionNote.String {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "resolution_note_added", "", ticket.ResolutionNote.String)
	}
	writeJSON(w, http.StatusOK, toTicketResponseFromTicket(ticket, existing.AssigneeName, int32(session.UserID)))
}

func (h *TicketHandler) patch(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	user, ok := userFromContext(w, r)
	if !ok {
		return
	}

	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	existing, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{ID: id, CurrentUserID: int32(session.UserID)})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}

	// fullAccess (staff/admin) může měnit libovolné meta-pole. Údržbář s
	// vlastním přiřazeným tiketem (statusOnly) může měnit jen stav — školník
	// nemá správcovská práva na přiřazení/priority/lokaci/kategorii.
	fullAccess := canMetaUpdate(user)
	statusOnly := !fullAccess && canUpdateOwnStatus(user, existing)
	if !fullAccess && !statusOnly {
		WriteError(w, http.StatusForbidden, "přístup odepřen — vyžaduje roli staff nebo admin")
		return
	}

	var body patchTicketRequest
	presence, err := decodeWithPresence(r, &body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if statusOnly {
		_, assignedToTouched := presence["assigned_to"]
		_, priorityTouched := presence["priority"]
		_, locationTouched := presence["location"]
		_, categoryTouched := presence["category"]
		if assignedToTouched || priorityTouched || locationTouched || categoryTouched {
			WriteError(w, http.StatusForbidden, "údržbář může u přiřazeného tiketu měnit jen stav")
			return
		}
	}
	if body.Priority != nil && !validPriorities[*body.Priority] {
		WriteError(w, http.StatusUnprocessableEntity, "neplatná priorita (povolené hodnoty: low, medium, high, urgent)")
		return
	}
	if body.AssignedTo != nil && !h.isAssignableTarget(r.Context(), *body.AssignedTo) {
		WriteError(w, http.StatusUnprocessableEntity, "lze přiřadit jen zaměstnanci, údržbáři nebo správci")
		return
	}

	params := db.UpdateTicketMetaParams{ID: id}
	// touch_*: pole se mění jen pokud bylo v requestu skutečně uvedeno (i jako
	// null) — jinak by PATCH s jedním polem vynuloval to druhé.
	_, assignedToTouched := presence["assigned_to"]
	if assignedToTouched {
		params.TouchAssignedTo = true
		if body.AssignedTo != nil {
			params.AssignedTo = sql.NullInt32{Int32: *body.AssignedTo, Valid: true}
		}
	}
	_, statusIDTouched := presence["status_id"]
	if statusIDTouched {
		params.TouchStatusID = true
		if body.StatusID != nil {
			params.StatusID = sql.NullInt32{Int32: *body.StatusID, Valid: true}
		}
	}
	if body.Priority != nil {
		params.Priority = sql.NullString{String: *body.Priority, Valid: true}
	}
	if body.Location != nil {
		params.Location = sql.NullString{String: *body.Location, Valid: true}
	}
	if body.Category != nil {
		params.Category = sql.NullString{String: *body.Category, Valid: true}
	}
	if body.ResolutionNote != nil {
		params.ResolutionNote = sql.NullString{String: *body.ResolutionNote, Valid: true}
	}

	ticket, err := h.queries.UpdateTicketMeta(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat tiket")
		return
	}
	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	if statusIDTouched {
		oldStatus := h.resolveStatusTitle(r.Context(), existing.StatusID)
		newStatus := h.resolveStatusTitle(r.Context(), ticket.StatusID)
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "status_changed", oldStatus, newStatus)
		h.activityLogger.LogStavZmenen(r.Context(), int32(session.UserID), id, oldStatus, newStatus)
	}
	if body.Priority != nil && existing.Priority != ticket.Priority {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "priority_changed", existing.Priority, ticket.Priority)
	}
	if body.Location != nil && existing.Location != ticket.Location {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "location_changed", existing.Location, ticket.Location)
	}
	if body.Category != nil && existing.Category != ticket.Category {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "category_changed", existing.Category, ticket.Category)
	}
	if body.ResolutionNote != nil && existing.ResolutionNote.String != ticket.ResolutionNote.String {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "resolution_note_added", "", ticket.ResolutionNote.String)
	}
	if assignedToTouched {
		var oldAssigneeID, newAssigneeID *int32
		if existing.AssignedTo.Valid {
			v := existing.AssignedTo.Int32
			oldAssigneeID = &v
		}
		if ticket.AssignedTo.Valid {
			v := ticket.AssignedTo.Int32
			newAssigneeID = &v
		}
		h.activityLogger.LogTiketPrirazen(r.Context(), int32(session.UserID), id, oldAssigneeID, newAssigneeID)
		if existing.AssignedTo.Valid && !ticket.AssignedTo.Valid {
			h.logHistory(r.Context(), id, int32(session.UserID), actorName, "unassigned", existing.AssigneeName, "")
		} else if ticket.AssignedTo.Valid {
			newAssigneeName := resolveAuthorName(r.Context(), h.queries, ticket.AssignedTo.Int32)
			h.logHistory(r.Context(), id, int32(session.UserID), actorName, "assigned", "", newAssigneeName)
		}
	}
	// Po meta-update znovu načteme pro assignee jméno
	updated, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{ID: id, CurrentUserID: int32(session.UserID)})
	if err != nil {
		writeJSON(w, http.StatusOK, toTicketResponseFromTicket(ticket, "", int32(session.UserID)))
		return
	}
	writeJSON(w, http.StatusOK, toTicketResponseFromGetRow(updated, int32(session.UserID)))
}

// claim umožní údržbáři převzít nepřiřazený tiket bez zásahu staff/admina.
func (h *TicketHandler) claim(w http.ResponseWriter, r *http.Request) {
	id, ok := pathIDWithAction(r.URL.Path, "/api/tickets/", "/claim")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID tiketu")
		return
	}

	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	user, ok := userFromContext(w, r)
	if !ok {
		return
	}

	if !canClaim(user) {
		WriteError(w, http.StatusForbidden, "přístup odepřen — vyžaduje roli údržbáře")
		return
	}

	existing, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{ID: id, CurrentUserID: int32(session.UserID)})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}
	if existing.AssignedTo.Valid {
		WriteError(w, http.StatusConflict, "tiket je již přiřazen")
		return
	}

	ticket, err := h.queries.UpdateTicketMeta(r.Context(), db.UpdateTicketMetaParams{
		ID:              id,
		TouchAssignedTo: true,
		AssignedTo:      sql.NullInt32{Int32: int32(session.UserID), Valid: true},
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se převzít tiket")
		return
	}

	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	h.activityLogger.LogTiketPrirazen(r.Context(), int32(session.UserID), id, nil, &user.ID)
	h.logHistory(r.Context(), id, int32(session.UserID), actorName, "assigned", "", actorName)

	updated, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{ID: id, CurrentUserID: int32(session.UserID)})
	if err != nil {
		writeJSON(w, http.StatusOK, toTicketResponseFromTicket(ticket, actorName, int32(session.UserID)))
		return
	}
	writeJSON(w, http.StatusOK, toTicketResponseFromGetRow(updated, int32(session.UserID)))
}

func (h *TicketHandler) delete(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	user, ok := userFromContext(w, r)
	if !ok {
		return
	}

	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	existing, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{
		ID:            id,
		CurrentUserID: int32(session.UserID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}

	if !canDelete(session, existing, user.UserType) {
		WriteError(w, http.StatusForbidden, "přístup odepřen")
		return
	}

	if err := h.queries.DeleteTicket(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se smazat tiket")
		return
	}
	h.activityLogger.LogTiketSmazan(r.Context(), int32(session.UserID), id, existing.Title)
	w.WriteHeader(http.StatusNoContent)
}

func (h *TicketHandler) vote(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	id, ok := voteTicketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	if err := h.queries.VoteTicket(r.Context(), db.VoteTicketParams{
		TicketID: id,
		UserID:   int32(session.UserID),
	}); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se přidat hlas")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TicketHandler) unvote(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	id, ok := voteTicketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	if err := h.queries.UnvoteTicket(r.Context(), db.UnvoteTicketParams{
		TicketID: id,
		UserID:   int32(session.UserID),
	}); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se odebrat hlas")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// approvePriority schválí čekající žádost o nejvyšší prioritu (urgent) tiketu.
// Pouze staff/admin (gated v router.go přes staffAdmin middleware).
//
// @Summary      Schválit žádost o urgentní prioritu
// @Description  Nastaví prioritu tiketu na urgent a zapíše schvalovatele. Přístupné pro staff a admin.
// @Tags         tickets
// @Param        id   path      int  true  "ID tiketu"
// @Success      200  {object}  ticketResponse
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse "Tiket nenalezen nebo nemá žádost o schválení priority"
// @Failure      500  {object}  errorResponse
// @Security     cookieAuth
// @Router       /api/tickets/{id}/approve-priority [post]
func (h *TicketHandler) approvePriority(w http.ResponseWriter, r *http.Request) {
	id, ok := pathIDWithAction(r.URL.Path, "/api/tickets/", "/approve-priority")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID tiketu")
		return
	}

	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	existing, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{ID: id, CurrentUserID: int32(session.UserID)})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}

	ticket, err := h.queries.ApproveTicketPriority(r.Context(), db.ApproveTicketPriorityParams{
		ID:         id,
		ApprovedBy: int32(session.UserID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen nebo nemá žádost o schválení priority")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se schválit prioritu")
		return
	}

	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	h.logHistory(r.Context(), id, int32(session.UserID), actorName, "priority_changed", existing.Priority, ticket.Priority)
	h.activityLogger.LogTiketPrioritaSchvalena(r.Context(), int32(session.UserID), id, ticket.Priority)
	writeJSON(w, http.StatusOK, toTicketResponseFromTicket(ticket, existing.AssigneeName, int32(session.UserID)))
}

// rejectPriority zamítne čekající žádost o nejvyšší prioritu (urgent) tiketu —
// priorita zůstává na dosavadní (fallback) hodnotě. Pouze staff/admin.
//
// @Summary      Zamítnout žádost o urgentní prioritu
// @Description  Zamítne žádost o urgentní prioritu, priorita tiketu se nemění. Přístupné pro staff a admin.
// @Tags         tickets
// @Param        id   path      int  true  "ID tiketu"
// @Success      200  {object}  ticketResponse
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      404  {object}  errorResponse "Tiket nenalezen nebo nemá žádost o schválení priority"
// @Failure      500  {object}  errorResponse
// @Security     cookieAuth
// @Router       /api/tickets/{id}/reject-priority [post]
func (h *TicketHandler) rejectPriority(w http.ResponseWriter, r *http.Request) {
	id, ok := pathIDWithAction(r.URL.Path, "/api/tickets/", "/reject-priority")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID tiketu")
		return
	}

	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	existing, err := h.queries.GetTicket(r.Context(), db.GetTicketParams{ID: id, CurrentUserID: int32(session.UserID)})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}
	requestedPriority := existing.RequestedPriority.String

	ticket, err := h.queries.RejectTicketPriority(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen nebo nemá žádost o schválení priority")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se zamítnout žádost o prioritu")
		return
	}

	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	h.logHistory(r.Context(), id, int32(session.UserID), actorName, "priority_approval_rejected", requestedPriority, "")
	h.activityLogger.LogTiketPrioritaZamitnuta(r.Context(), int32(session.UserID), id, requestedPriority)
	writeJSON(w, http.StatusOK, toTicketResponseFromTicket(ticket, existing.AssigneeName, int32(session.UserID)))
}

// canEditContent: autor tiketu může měnit obsah jen pokud tiket ještě nebyl přiřazen.
// Admin může vždy.
func canEditContent(session db.Session, ticket db.GetTicketRow, userType db.UserType) bool {
	if userType == db.UserTypeAdmin {
		return true
	}
	return int32(session.UserID) == ticket.AuthorID && !ticket.AssignedTo.Valid
}

// canMetaUpdate: staff nebo admin může měnit meta-pole.
func canMetaUpdate(user db.User) bool {
	return user.UserType == db.UserTypeStaff || user.UserType == db.UserTypeAdmin
}

// canUpdateOwnStatus: údržbář může měnit stav tiketu, který je přiřazen jemu
// — nemá ale správcovská práva na přiřazení/prioritu/lokaci/kategorii.
func canUpdateOwnStatus(user db.User, ticket db.GetTicketRow) bool {
	return user.UserType == db.UserTypeMaintainer &&
		ticket.AssignedTo.Valid && ticket.AssignedTo.Int32 == user.ID
}

// canClaim: nepřiřazený tiket si může sám převzít jen údržbář.
func canClaim(user db.User) bool {
	return user.UserType == db.UserTypeMaintainer
}

// isAssignableUserType: jen tyto role mohou být řešitelem tiketu.
func isAssignableUserType(t db.UserType) bool {
	return t == db.UserTypeStaff || t == db.UserTypeMaintainer || t == db.UserTypeAdmin
}

// isAssignableTarget ověří, že cílový uživatel existuje a má roli, která
// může být řešitelem tiketu (staff/maintainer/admin).
func (h *TicketHandler) isAssignableTarget(ctx context.Context, userID int32) bool {
	u, err := h.queries.GetUserByID(ctx, userID)
	if err != nil {
		return false
	}
	return isAssignableUserType(u.UserType)
}

// decodeWithPresence dekóduje tělo požadavku do v a navíc vrátí mapu
// přítomných JSON klíčů — díky tomu lze rozlišit "pole v requestu vůbec
// nebylo" (nemá se měnit) od "pole bylo posláno jako null" (má se vynulovat).
func decodeWithPresence(r *http.Request, v any) (map[string]json.RawMessage, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return nil, err
	}
	var presence map[string]json.RawMessage
	if err := json.Unmarshal(raw, &presence); err != nil {
		return nil, err
	}
	return presence, nil
}

// requiresPriorityApproval: urgent je nejvyšší existující priorita a potřebuje
// schválení staff/admin, pokud ho nezadává přímo někdo z nich.
func requiresPriorityApproval(userType db.UserType, priority string) bool {
	return priority == "urgent" && userType != db.UserTypeStaff && userType != db.UserTypeAdmin
}

// canDelete: autor (pokud tiket nebyl přiřazen) nebo admin.
func canDelete(session db.Session, ticket db.GetTicketRow, userType db.UserType) bool {
	if userType == db.UserTypeAdmin {
		return true
	}
	return int32(session.UserID) == ticket.AuthorID && !ticket.AssignedTo.Valid
}

func toTicketResponseFromRow(r db.ListTicketsFilteredRow) ticketResponse {
	tr := ticketResponse{
		ID:             r.ID,
		Title:          r.Title,
		Body:           r.Body,
		Priority:       r.Priority,
		Location:       r.Location,
		Category:       r.Category,
		AssignedToName: r.AssigneeName,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		AuthorID:       r.AuthorID,
		AuthorName:     r.AuthorName,
		StatusID:       nullInt32{Int32: r.StatusID.Int32, Valid: r.StatusID.Valid},
		VoteCount:      r.VoteCount,
		UserHasVoted:   r.UserHasVoted,
		IsClosed:       r.IsClosed,
	}
	if r.AssignedTo.Valid {
		v := r.AssignedTo.Int32
		tr.AssignedTo = &v
	}
	if r.RequestedPriority.Valid {
		v := r.RequestedPriority.String
		tr.RequestedPriority = &v
	}
	if r.PriorityApprovedBy.Valid {
		v := r.PriorityApprovedBy.Int32
		tr.PriorityApprovedBy = &v
	}
	if r.ResolutionNote.Valid {
		v := r.ResolutionNote.String
		tr.ResolutionNote = &v
	}
	return tr
}

func toTicketResponseFromGetRow(r db.GetTicketRow, currentUserID int32) ticketResponse {
	tr := ticketResponse{
		ID:             r.ID,
		Title:          r.Title,
		Body:           r.Body,
		Priority:       r.Priority,
		Location:       r.Location,
		Category:       r.Category,
		AssignedToName: r.AssigneeName,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		AuthorID:       r.AuthorID,
		AuthorName:     r.AuthorName,
		StatusID:       nullInt32{Int32: r.StatusID.Int32, Valid: r.StatusID.Valid},
		VoteCount:      r.VoteCount,
		UserHasVoted:   r.UserHasVoted,
		IsClosed:       r.IsClosed,
	}
	if r.AssignedTo.Valid {
		v := r.AssignedTo.Int32
		tr.AssignedTo = &v
	}
	if r.RequestedPriority.Valid {
		v := r.RequestedPriority.String
		tr.RequestedPriority = &v
	}
	if r.PriorityApprovedBy.Valid {
		v := r.PriorityApprovedBy.Int32
		tr.PriorityApprovedBy = &v
	}
	if r.ResolutionNote.Valid {
		v := r.ResolutionNote.String
		tr.ResolutionNote = &v
	}
	return tr
}

func toTicketResponseFromTicket(t db.Ticket, assigneeName string, _ int32) ticketResponse {
	tr := ticketResponse{
		ID:             t.ID,
		Title:          t.Title,
		Body:           t.Body,
		Priority:       t.Priority,
		Location:       t.Location,
		Category:       t.Category,
		AssignedToName: assigneeName,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		AuthorID:       t.AuthorID,
		StatusID:       nullInt32{Int32: t.StatusID.Int32, Valid: t.StatusID.Valid},
		IsClosed:       t.IsClosed,
	}
	if t.AssignedTo.Valid {
		v := t.AssignedTo.Int32
		tr.AssignedTo = &v
	}
	if t.RequestedPriority.Valid {
		v := t.RequestedPriority.String
		tr.RequestedPriority = &v
	}
	if t.PriorityApprovedBy.Valid {
		v := t.PriorityApprovedBy.Int32
		tr.PriorityApprovedBy = &v
	}
	if t.ResolutionNote.Valid {
		v := t.ResolutionNote.String
		tr.ResolutionNote = &v
	}
	return tr
}

func sessionFromContext(w http.ResponseWriter, r *http.Request) (db.Session, bool) {
	v := r.Context().Value(ctxkeys.SessionContextKey)
	if v == nil {
		WriteError(w, http.StatusUnauthorized, "nepřihlášen")
		return db.Session{}, false
	}
	return v.(db.Session), true
}

func userFromContext(w http.ResponseWriter, r *http.Request) (db.User, bool) {
	v := r.Context().Value(ctxkeys.UserContextKey)
	if v == nil {
		WriteError(w, http.StatusUnauthorized, "nepřihlášen")
		return db.User{}, false
	}
	return v.(db.User), true
}

func ticketIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	// Strip potential suffix like "/vote" or "/history" before parsing ID
	// For paths like /api/tickets/123, extract 123
	trimmed := strings.TrimSuffix(strings.TrimSuffix(path, "/vote"), "/history")
	id, ok := pathID(trimmed, "/api/tickets/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID tiketu")
	}
	return id, ok
}

func voteTicketIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	// /api/tickets/123/vote → extract 123
	trimmed := strings.TrimSuffix(path, "/vote")
	id, ok := pathID(trimmed, "/api/tickets/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID tiketu")
	}
	return id, ok
}

// logHistory zapíše záznam do ticket_history. Chyby jsou ignorovány — audit log nesmí blokovat hlavní operaci.
func (h *TicketHandler) logHistory(ctx context.Context, ticketID int64, actorID int32, actorName, event, oldVal, newVal string) {
	_ = h.queries.InsertTicketHistory(ctx, db.InsertTicketHistoryParams{
		TicketID:  ticketID,
		ActorID:   actorID,
		ActorName: actorName,
		Event:     event,
		OldVal:    sql.NullString{String: oldVal, Valid: oldVal != ""},
		NewVal:    sql.NullString{String: newVal, Valid: newVal != ""},
	})
}

// resolveStatusTitle vrátí název stavu tiketu dle ID, nebo prázdný řetězec při chybě.
func (h *TicketHandler) resolveStatusTitle(ctx context.Context, id sql.NullInt32) string {
	if !id.Valid {
		return ""
	}
	title, err := h.queries.GetStatusTitle(ctx, id.Int32)
	if err != nil {
		return ""
	}
	return title
}

func (h *TicketHandler) listHistory(w http.ResponseWriter, r *http.Request) {
	_, ok := sessionFromContext(w, r)
	if !ok {
		return
	}
	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}
	rows, err := h.queries.ListTicketHistory(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst historii tiketu")
		return
	}
	entries := make([]ticketHistoryEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, ticketHistoryEntry{
			ID:        row.ID,
			ActorName: row.ActorName,
			Event:     row.Event,
			OldVal:    row.OldVal.String,
			NewVal:    row.NewVal.String,
			CreatedAt: row.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, entries)
}

// resolveAuthorName vrátí zobrazitelné jméno uživatele.
// Pokud uživatel nemá vyplněné jméno ani příjmení, vrátí e-mailovou adresu jako zálohu.
func resolveAuthorName(ctx context.Context, q *db.Queries, authorID int32) string {
	user, err := q.GetUserByID(ctx, authorID)
	if err != nil {
		return ""
	}
	parts := make([]string, 0, 2)
	if user.FirstName.Valid && user.FirstName.String != "" {
		parts = append(parts, user.FirstName.String)
	}
	if user.LastName.Valid && user.LastName.String != "" {
		parts = append(parts, user.LastName.String)
	}
	if len(parts) > 0 {
		return strings.Join(parts, " ")
	}
	return user.Email
}
