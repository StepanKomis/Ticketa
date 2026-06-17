package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
)

type TicketHandler struct {
	queries    *db.Queries
	httpLogger *logs.Logger
}

func NewTicketHandler(q *db.Queries, l *logs.Logger) *TicketHandler {
	return &TicketHandler{queries: q, httpLogger: l}
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
	Title    *string `json:"title"`
	Body     *string `json:"body"`
	Priority *string `json:"priority"`
	Location *string `json:"location"`
	Category *string `json:"category"`
	StatusID *int32  `json:"status_id"`
}

type patchTicketRequest struct {
	AssignedTo *int32  `json:"assigned_to"`
	StatusID   *int32  `json:"status_id"`
	Priority   *string `json:"priority"`
	Location   *string `json:"location"`
	Category   *string `json:"category"`
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

	params := db.CreateTicketParams{
		Title:    body.Title,
		Body:     body.Body,
		AuthorID: int32(session.UserID),
		Priority: body.Priority,
		Location: body.Location,
		Category: body.Category,
	}
	if body.StatusID != nil {
		params.StatusID = sql.NullInt32{Int32: *body.StatusID, Valid: true}
	}
	// assigned_to lze nastavit pouze staff nebo admin
	if body.AssignedTo != nil {
		user, ok := userFromContext(w, r)
		if !ok {
			return
		}
		if user.UserType == db.UserTypeStaff || user.UserType == db.UserTypeAdmin {
			params.AssignedTo = sql.NullInt32{Int32: *body.AssignedTo, Valid: true}
		}
	}

	ticket, err := h.queries.CreateTicket(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se vytvořit tiket")
		return
	}
	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	h.logHistory(r.Context(), ticket.ID, int32(session.UserID), actorName, "created", "", "")
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

	rows, err := h.queries.ListTicketsFiltered(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tikety")
		return
	}

	countParams := db.CountTicketsFilteredParams{
		StatusID:   params.StatusID,
		Priority:   params.Priority,
		AssignedTo: params.AssignedTo,
		AuthorID:   params.AuthorID,
		Category:   params.Category,
		Q:          params.Q,
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
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
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
		params.Priority = sql.NullString{String: *body.Priority, Valid: true}
	}
	if body.Location != nil {
		params.Location = sql.NullString{String: *body.Location, Valid: true}
	}
	if body.Category != nil {
		params.Category = sql.NullString{String: *body.Category, Valid: true}
	}
	if body.StatusID != nil {
		params.StatusID = sql.NullInt32{Int32: *body.StatusID, Valid: true}
	}

	ticket, err := h.queries.UpdateTicket(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat tiket")
		return
	}
	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	if body.Title != nil || body.Body != nil {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "content_updated", "", "")
	}
	if body.StatusID != nil {
		oldStatus := h.resolveStatusTitle(r.Context(), existing.StatusID)
		newStatus := h.resolveStatusTitle(r.Context(), ticket.StatusID)
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "status_changed", oldStatus, newStatus)
	}
	if body.Priority != nil && existing.Priority != ticket.Priority {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "priority_changed", existing.Priority, ticket.Priority)
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

	if !canMetaUpdate(user) {
		WriteError(w, http.StatusForbidden, "přístup odepřen — vyžaduje roli staff nebo admin")
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

	var body patchTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if body.Priority != nil && !validPriorities[*body.Priority] {
		WriteError(w, http.StatusUnprocessableEntity, "neplatná priorita (povolené hodnoty: low, medium, high, urgent)")
		return
	}

	params := db.UpdateTicketMetaParams{ID: id}
	// assigned_to: null = odebrat přiřazení, číslo = přiřadit
	if body.AssignedTo != nil {
		params.AssignedTo = sql.NullInt32{Int32: *body.AssignedTo, Valid: true}
	}
	if body.StatusID != nil {
		params.StatusID = sql.NullInt32{Int32: *body.StatusID, Valid: true}
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

	ticket, err := h.queries.UpdateTicketMeta(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat tiket")
		return
	}
	actorName := resolveAuthorName(r.Context(), h.queries, int32(session.UserID))
	if body.StatusID != nil {
		oldStatus := h.resolveStatusTitle(r.Context(), existing.StatusID)
		newStatus := h.resolveStatusTitle(r.Context(), ticket.StatusID)
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "status_changed", oldStatus, newStatus)
	}
	if body.Priority != nil && existing.Priority != ticket.Priority {
		h.logHistory(r.Context(), id, int32(session.UserID), actorName, "priority_changed", existing.Priority, ticket.Priority)
	}
	if body.AssignedTo != nil {
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

// canDelete: autor (pokud tiket nebyl přiřazen) nebo admin.
func canDelete(session db.Session, ticket db.GetTicketRow, userType db.UserType) bool {
	if userType == db.UserTypeAdmin {
		return true
	}
	return int32(session.UserID) == ticket.AuthorID && !ticket.AssignedTo.Valid
}

func toTicketResponseFromRow(r db.ListTicketsFilteredRow) ticketResponse {
	tr := ticketResponse{
		ID:           r.ID,
		Title:        r.Title,
		Body:         r.Body,
		Priority:     r.Priority,
		Location:     r.Location,
		Category:     r.Category,
		AssignedToName: r.AssigneeName,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		AuthorID:     r.AuthorID,
		AuthorName:   r.AuthorName,
		StatusID:     nullInt32{Int32: r.StatusID.Int32, Valid: r.StatusID.Valid},
		VoteCount:    r.VoteCount,
		UserHasVoted: r.UserHasVoted,
	}
	if r.AssignedTo.Valid {
		v := r.AssignedTo.Int32
		tr.AssignedTo = &v
	}
	return tr
}

func toTicketResponseFromGetRow(r db.GetTicketRow, currentUserID int32) ticketResponse {
	tr := ticketResponse{
		ID:           r.ID,
		Title:        r.Title,
		Body:         r.Body,
		Priority:     r.Priority,
		Location:     r.Location,
		Category:     r.Category,
		AssignedToName: r.AssigneeName,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		AuthorID:     r.AuthorID,
		AuthorName:   r.AuthorName,
		StatusID:     nullInt32{Int32: r.StatusID.Int32, Valid: r.StatusID.Valid},
		VoteCount:    r.VoteCount,
		UserHasVoted: r.UserHasVoted,
	}
	if r.AssignedTo.Valid {
		v := r.AssignedTo.Int32
		tr.AssignedTo = &v
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
	}
	if t.AssignedTo.Valid {
		v := t.AssignedTo.Int32
		tr.AssignedTo = &v
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
