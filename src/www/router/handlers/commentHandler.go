package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

type CommentHandler struct {
	queries    *db.Queries
	httpLogger *logs.Logger
}

func NewCommentHandler(q *db.Queries, l *logs.Logger) *CommentHandler {
	return &CommentHandler{queries: q, httpLogger: l}
}

func (h *CommentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && matchesTicketCommentsPath(r.URL.Path):
		h.create(w, r)
	case r.Method == http.MethodGet && matchesTicketCommentsPath(r.URL.Path):
		h.list(w, r)
	case r.Method == http.MethodPut && matchesIDPath(r.URL.Path, "/api/comments/"):
		h.update(w, r)
	case r.Method == http.MethodDelete && matchesIDPath(r.URL.Path, "/api/comments/"):
		h.softDelete(w, r)
	default:
		defaultResponse(w)
	}
}

type createCommentRequest struct {
	Body     string `json:"body"`
	ParentID *int64 `json:"parent_id"`
}

// create vytvoří nový komentář nebo odpověď na existující komentář pod tiketem.
// Autor je určen ze session cookie — nelze vytvořit komentář za jiného uživatele.
// Pole parent_id je volitelné; pokud je zadáno, komentář je odpovědí na daný komentář.
//
// @Summary      Vytvořit komentář
// @Description  Vytvoří komentář pod tiketem. Pokud je uveden parent_id, jde o odpověď na jiný komentář. Autor je nastaven automaticky ze session.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "ID tiketu"
// @Param        body  body      createCommentRequest  true  "Nový komentář"
// @Success      201   {object}  commentResponse       "Vytvořený komentář"
// @Failure      400   {object}  errorResponse         "Neplatné ID tiketu nebo tělo požadavku"
// @Failure      401   {object}  errorResponse         "Chybí nebo vypršel session cookie"
// @Failure      422   {object}  errorResponse         "Chybí povinné pole body"
// @Failure      500   {object}  errorResponse         "Interní chyba"
// @Security     cookieAuth
// @Router       /api/tickets/{id}/comments [post]
func (h *CommentHandler) create(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	ticketID, ok := ticketIDFromCommentPath(w, r.URL.Path)
	if !ok {
		return
	}

	var body createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if body.Body == "" {
		WriteError(w, http.StatusUnprocessableEntity, "pole body je povinné")
		return
	}

	params := db.CreateCommentParams{
		TicketID: ticketID,
		AuthorID: int32(session.UserID),
		Body:     body.Body,
	}
	if body.ParentID != nil {
		params.ParentID = sql.NullInt64{Int64: *body.ParentID, Valid: true}
	}

	comment, err := h.queries.CreateComment(r.Context(), params)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se vytvořit komentář")
		return
	}
	authorName := resolveAuthorName(r.Context(), h.queries, comment.AuthorID)
	writeJSON(w, http.StatusCreated, toCommentResponse(comment, authorName))
}

// list vrátí seznam komentářů tiketu seřazených od nejstaršího.
// Smazané komentáře nejsou zahrnuty. Prázdný výsledek vrátí [] (nikdy null).
//
// @Summary      Seznam komentářů tiketu
// @Description  Vrátí všechny aktivní (nesmazané) komentáře k danému tiketu, řazené od nejstaršího.
// @Tags         comments
// @Produce      json
// @Param        id   path      int              true  "ID tiketu"
// @Success      200  {array}   commentResponse  "Seznam komentářů"
// @Failure      400  {object}  errorResponse    "Neplatné ID tiketu"
// @Failure      401  {object}  errorResponse    "Chybí nebo vypršel session cookie"
// @Failure      500  {object}  errorResponse    "Interní chyba"
// @Security     cookieAuth
// @Router       /api/tickets/{id}/comments [get]
func (h *CommentHandler) list(w http.ResponseWriter, r *http.Request) {
	ticketID, ok := ticketIDFromCommentPath(w, r.URL.Path)
	if !ok {
		return
	}

	comments, err := h.queries.ListCommentsByTicket(r.Context(), ticketID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst komentáře")
		return
	}

	nameCache := make(map[int32]string, len(comments))
	result := make([]commentResponse, len(comments))
	for i, c := range comments {
		if _, ok := nameCache[c.AuthorID]; !ok {
			nameCache[c.AuthorID] = resolveAuthorName(r.Context(), h.queries, c.AuthorID)
		}
		result[i] = toCommentResponse(c, nameCache[c.AuthorID])
	}
	writeJSON(w, http.StatusOK, result)
}

type updateCommentRequest struct {
	Body string `json:"body"`
}

// update aktualizuje text komentáře. Povoleno pouze autorovi komentáře.
//
// @Summary      Aktualizovat komentář
// @Description  Aktualizuje text komentáře. Povoleno pouze autorovi. Nelze upravovat komentáře ostatních uživatelů.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "ID komentáře"
// @Param        body  body      updateCommentRequest  true  "Aktualizovaný text"
// @Success      200   {object}  commentResponse       "Aktualizovaný komentář"
// @Failure      400   {object}  errorResponse         "Neplatné ID nebo tělo požadavku"
// @Failure      401   {object}  errorResponse         "Chybí nebo vypršel session cookie"
// @Failure      403   {object}  errorResponse         "Nejste autor tohoto komentáře"
// @Failure      404   {object}  errorResponse         "Komentář nenalezen nebo smazán"
// @Failure      422   {object}  errorResponse         "Chybí povinné pole body"
// @Failure      500   {object}  errorResponse         "Interní chyba"
// @Security     cookieAuth
// @Router       /api/comments/{id} [put]
func (h *CommentHandler) update(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	id, ok := commentIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	existing, err := h.queries.GetComment(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "komentář nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst komentář")
		return
	}
	if existing.Deleted {
		WriteError(w, http.StatusNotFound, "komentář nenalezen")
		return
	}
	if int32(session.UserID) != existing.AuthorID {
		WriteError(w, http.StatusForbidden, "přístup odepřen")
		return
	}

	var body updateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if body.Body == "" {
		WriteError(w, http.StatusUnprocessableEntity, "pole body je povinné")
		return
	}

	comment, err := h.queries.UpdateComment(r.Context(), db.UpdateCommentParams{
		ID:   id,
		Body: body.Body,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat komentář")
		return
	}
	authorName := resolveAuthorName(r.Context(), h.queries, comment.AuthorID)
	writeJSON(w, http.StatusOK, toCommentResponse(comment, authorName))
}

// softDelete označí komentář jako smazaný (soft delete). Povoleno autorovi nebo uživateli s rolí staff/maintainer.
// Komentář zůstane v databázi, ale přestane se zobrazovat v seznamu.
//
// @Summary      Smazat komentář
// @Description  Provede soft delete komentáře. Povoleno autorovi, učitelům (staff) a správcům (maintainer). Vrátí 204 bez těla odpovědi.
// @Tags         comments
// @Param        id  path  int  true  "ID komentáře"
// @Success      204 "Komentář smazán"
// @Failure      400 {object}  errorResponse  "Neplatné ID"
// @Failure      401 {object}  errorResponse  "Chybí nebo vypršel session cookie"
// @Failure      403 {object}  errorResponse  "Přístup odepřen"
// @Failure      404 {object}  errorResponse  "Komentář nenalezen nebo již smazán"
// @Failure      500 {object}  errorResponse  "Interní chyba"
// @Security     cookieAuth
// @Router       /api/comments/{id} [delete]
func (h *CommentHandler) softDelete(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(w, r)
	if !ok {
		return
	}

	id, ok := commentIDFromPath(w, r.URL.Path)
	if !ok {
		return
	}

	existing, err := h.queries.GetComment(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "komentář nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst komentář")
		return
	}
	if existing.Deleted {
		WriteError(w, http.StatusNotFound, "komentář nenalezen")
		return
	}

	user, ok := userFromContext(w, r)
	if !ok {
		return
	}
	if !canDeleteComment(session, existing, user.UserType) {
		WriteError(w, http.StatusForbidden, "přístup odepřen")
		return
	}

	if err := h.queries.SoftDeleteComment(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se smazat komentář")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func canDeleteComment(session db.Session, comment db.TicketComment, userType db.UserType) bool {
	return int32(session.UserID) == comment.AuthorID ||
		userType == db.UserTypeStaff ||
		userType == db.UserTypeAdmin
}

func toCommentResponse(c db.TicketComment, authorName string) commentResponse {
	var parentID *int64
	if c.ParentID.Valid {
		parentID = &c.ParentID.Int64
	}
	return commentResponse{
		ID:         c.ID,
		TicketID:   c.TicketID,
		AuthorID:   c.AuthorID,
		AuthorName: authorName,
		ParentID:   parentID,
		Body:       c.Body,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

// matchesTicketCommentsPath vrátí true pro cesty ve tvaru /api/tickets/{n}/comments.
// Kontroluje prefix i suffix, aby se odlišily cesty jako /api/tickets/5/comments/3.
func matchesTicketCommentsPath(path string) bool {
	const pre, suf = "/api/tickets/", "/comments"
	return strings.HasPrefix(path, pre) && strings.HasSuffix(path, suf) &&
		len(path) > len(pre)+len(suf)
}

// ticketIDFromCommentPath extrahuje číslo tiketu z /api/tickets/{id}/comments.
func ticketIDFromCommentPath(w http.ResponseWriter, path string) (int64, bool) {
	const pre, suf = "/api/tickets/", "/comments"
	id, err := strconv.ParseInt(path[len(pre):len(path)-len(suf)], 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné ID tiketu")
	}
	return id, err == nil
}

func commentIDFromPath(w http.ResponseWriter, path string) (int64, bool) {
	id, ok := pathID(path, "/api/comments/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID komentáře")
	}
	return id, ok
}
