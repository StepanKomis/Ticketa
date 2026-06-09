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

// create vytvoří nový tiket přihlášeného uživatele.
// Autor tiketu je určen ze session cookie — nelze vytvořit tiket za jiného uživatele.
//
// @Summary      Vytvořit tiket
// @Description  Vytvoří nový tiket. Autor je automaticky nastaven ze session. StatusID je volitelné — pokud není zadáno, tiket nemá přiřazený stav.
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        body  body      createTicketRequest  true  "Nový tiket"
// @Success      201   {object}  ticketResponse       "Vytvořený tiket"
// @Failure      400   {object}  errorResponse        "Neplatné tělo požadavku"
// @Failure      401   {object}  errorResponse        "Chybí nebo vypršel session cookie"
// @Failure      422   {object}  errorResponse        "Chybí povinné pole (title nebo body)"
// @Failure      500   {object}  errorResponse        "Interní chyba"
// @Security     cookieAuth
// @Router       /api/tickets [post]
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
		WriteError(w, http.StatusInternalServerError, "nepodařilo se vytvořit tiket")
		return
	}
	writeJSON(w, http.StatusCreated, ticket)
}

// list vrátí seznam všech tiketů seřazených od nejnovějšího.
// Prázdný výsledek vrátí [] (nikdy null).
//
// @Summary      Seznam tiketů
// @Description  Vrátí všechny tikety seřazené od nejnovějšího. Přístupné pro všechny přihlášené uživatele.
// @Tags         tickets
// @Produce      json
// @Success      200  {array}   ticketResponse  "Seznam tiketů"
// @Failure      401  {object}  errorResponse   "Chybí nebo vypršel session cookie"
// @Failure      500  {object}  errorResponse   "Interní chyba"
// @Security     cookieAuth
// @Router       /api/tickets [get]
func (h *TicketHandler) list(w http.ResponseWriter, r *http.Request) {
	tickets, err := h.queries.ListTickets(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tikety")
		return
	}
	if tickets == nil {
		tickets = []db.Ticket{}
	}
	writeJSON(w, http.StatusOK, tickets)
}

// get vrátí jeden tiket podle ID.
//
// @Summary      Získat tiket
// @Description  Vrátí jeden tiket podle jeho ID.
// @Tags         tickets
// @Produce      json
// @Param        id   path      int             true  "ID tiketu"
// @Success      200  {object}  ticketResponse  "Tiket"
// @Failure      400  {object}  errorResponse   "Neplatné ID"
// @Failure      401  {object}  errorResponse   "Chybí nebo vypršel session cookie"
// @Failure      404  {object}  errorResponse   "Tiket nenalezen"
// @Failure      500  {object}  errorResponse   "Interní chyba"
// @Security     cookieAuth
// @Router       /api/tickets/{id} [get]
func (h *TicketHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := ticketIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	ticket, err := h.queries.GetTicket(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

type updateTicketRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	StatusID *int32 `json:"status_id"`
}

// update aktualizuje tiket. Povoleno pouze autorovi tiketu.
// Všechna pole v těle jsou volitelná — uvádějte pouze to, co chcete změnit.
//
// @Summary      Aktualizovat tiket
// @Description  Aktualizuje tiket. Povoleno pouze autorovi. Nelze přepisovat tikety ostatních uživatelů.
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "ID tiketu"
// @Param        body  body      updateTicketRequest  true  "Aktualizovaná data tiketu"
// @Success      200   {object}  ticketResponse       "Aktualizovaný tiket"
// @Failure      400   {object}  errorResponse        "Neplatné ID nebo tělo požadavku"
// @Failure      401   {object}  errorResponse        "Chybí nebo vypršel session cookie"
// @Failure      403   {object}  errorResponse        "Nejste autor tohoto tiketu"
// @Failure      404   {object}  errorResponse        "Tiket nenalezen"
// @Failure      500   {object}  errorResponse        "Interní chyba"
// @Security     cookieAuth
// @Router       /api/tickets/{id} [put]
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
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}

	if !canModifyTicket(session, existing) {
		WriteError(w, http.StatusForbidden, "přístup odepřen")
		return
	}

	var body updateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
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
		WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat tiket")
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

// delete smaže tiket. Povoleno pouze autorovi tiketu.
//
// @Summary      Smazat tiket
// @Description  Smaže tiket. Povoleno pouze autorovi. Vrátí 204 bez těla odpovědi.
// @Tags         tickets
// @Param        id  path  int  true  "ID tiketu"
// @Success      204 "Tiket smazán"
// @Failure      400 {object}  errorResponse  "Neplatné ID"
// @Failure      401 {object}  errorResponse  "Chybí nebo vypršel session cookie"
// @Failure      403 {object}  errorResponse  "Nejste autor tohoto tiketu"
// @Failure      404 {object}  errorResponse  "Tiket nenalezen"
// @Failure      500 {object}  errorResponse  "Interní chyba"
// @Security     cookieAuth
// @Router       /api/tickets/{id} [delete]
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
			WriteError(w, http.StatusNotFound, "tiket nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst tiket")
		return
	}

	if !canModifyTicket(session, existing) {
		WriteError(w, http.StatusForbidden, "přístup odepřen")
		return
	}

	if err := h.queries.DeleteTicket(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se smazat tiket")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// sessionFromContext načte validovanou session z kontextu požadavku.
// Zapíše 401 a vrátí false pokud session chybí.
func sessionFromContext(w http.ResponseWriter, r *http.Request) (db.Session, bool) {
	v := r.Context().Value(ctxkeys.SessionContextKey)
	if v == nil {
		WriteError(w, http.StatusUnauthorized, "nepřihlášen")
		return db.Session{}, false
	}
	return v.(db.Session), true
}

// canModifyTicket vrátí true pokud je uživatel session autorem tiketu nebo maintainerem.
func canModifyTicket(session db.Session, ticket db.Ticket) bool {
	return int32(session.UserID) == ticket.AuthorID
}

func ticketIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	id, ok := pathID(path, "/api/tickets/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID tiketu")
	}
	return id, ok
}
