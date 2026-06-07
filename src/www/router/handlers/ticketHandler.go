package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

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
	case r.Method == http.MethodGet && matchesIDPath(r.URL.Path, "/api/tickets/"):
		h.get(w, r)
	case r.Method == http.MethodPut && matchesIDPath(r.URL.Path, "/api/tickets/"):
		h.update(w, r)
	case r.Method == http.MethodDelete && matchesIDPath(r.URL.Path, "/api/tickets/"):
		h.delete(w, r)
	default:
		defaultResponse(w)
	}
}

type createTicketRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	StatusID *int32 `json:"status_id"`
}

func (h *TicketHandler) create(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	var body createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == "" {
		WriteError(w, http.StatusUnprocessableEntity, "title is required")
		return
	}
	if body.Body == "" {
		WriteError(w, http.StatusUnprocessableEntity, "body is required")
		return
	}

	params := db.CreateTicketParams{
		Title:    body.Title,
		Body:     body.Body,
		AuthorID: int32(session.UserID),
	}
	if body.StatusID != nil {
		params.StatusID = sql.NullInt32{Int32: *body.StatusID, Valid: true}
	}

	ticket, err := h.queries.CreateTicket(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not create ticket")
		return
	}
	writeJSON(w, http.StatusCreated, ticket)
}

func (h *TicketHandler) list(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.queries.ListTickets(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not list tickets")
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (h *TicketHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	ticket, err := h.queries.GetTicket(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "ticket not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not fetch ticket")
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

type updateTicketRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	StatusID *int32 `json:"status_id"`
}

func (h *TicketHandler) update(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	existing, err := h.queries.GetTicket(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "ticket not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not fetch ticket")
		return
	}

	if !canModifyTicket(session, existing) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	var body updateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params := db.UpdateTicketParams{
		ID:    id,
		Title: body.Title,
		Body:  body.Body,
	}
	if body.StatusID != nil {
		params.StatusID = sql.NullInt32{Int32: *body.StatusID, Valid: true}
	}

	ticket, err := h.queries.UpdateTicket(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not update ticket")
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

func (h *TicketHandler) delete(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	existing, err := h.queries.GetTicket(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "ticket not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not fetch ticket")
		return
	}

	if !canModifyTicket(session, existing) {
		WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.queries.DeleteTicket(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not delete ticket")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// sessionFromContext extracts the validated session from the request context.
// It writes a 401 and returns false if the session is absent.
func sessionFromContext(w http.ResponseWriter, r *http.Request) (db.Session, bool) {
	v := r.Context().Value(ctxkeys.SessionContextKey)
	if v == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return db.Session{}, false
	}
	return v.(db.Session), true
}

// canModifyTicket returns true when the session user is the ticket author or a maintainer.
func canModifyTicket(session db.Session, ticket db.Ticket) bool {
	return int32(session.UserID) == ticket.AuthorID
}

func ticketIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	id, ok := pathID(path, "/api/tickets/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid ticket id")
	}
	return id, ok
}
