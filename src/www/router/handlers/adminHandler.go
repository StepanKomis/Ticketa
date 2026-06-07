package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

type AdminHandler struct {
	queries    *db.Queries
	cfgStore   *config.Store
	httpLogger *logs.Logger
}

func NewAdminHandler(q *db.Queries, cfgStore *config.Store, l *logs.Logger) *AdminHandler {
	return &AdminHandler{queries: q, cfgStore: cfgStore, httpLogger: l}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/admin/config":
		h.getConfig(w, r)
	case r.Method == http.MethodPatch && r.URL.Path == "/api/admin/config":
		h.patchConfig(w, r)

	case r.Method == http.MethodGet && r.URL.Path == "/api/admin/ticket-statuses":
		h.listStatuses(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/admin/ticket-statuses":
		h.createStatus(w, r)
	case r.Method == http.MethodPut && matchesIDPath(r.URL.Path, "/api/admin/ticket-statuses/"):
		h.updateStatus(w, r)
	case r.Method == http.MethodDelete && matchesIDPath(r.URL.Path, "/api/admin/ticket-statuses/"):
		h.deleteStatus(w, r)

	case r.Method == http.MethodGet && r.URL.Path == "/api/admin/users":
		h.listUsers(w, r)
	case r.Method == http.MethodGet && matchesIDPath(r.URL.Path, "/api/admin/users/"):
		h.getUser(w, r)
	case r.Method == http.MethodPatch && matchesIDPath(r.URL.Path, "/api/admin/users/"):
		h.patchUser(w, r)

	default:
		defaultResponse(w)
	}
}

// matchesIDPath returns true when path starts with prefix and has a non-empty suffix.
func matchesIDPath(path, prefix string) bool {
	return len(path) > len(prefix) && path[:len(prefix)] == prefix
}

func pathID(path, prefix string) (int64, bool) {
	s := path[len(prefix):]
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil
}

// ---- Config ----------------------------------------------------------------

func (h *AdminHandler) getConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.cfgStore.Get())
}

func (h *AdminHandler) patchConfig(w http.ResponseWriter, r *http.Request) {
	var patch config.Config
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.cfgStore.Update(func(c *config.Config) error {
		if patch.Logging.Level != "" {
			c.Logging.Level = patch.Logging.Level
		}
		if patch.Logging.Dir != "" {
			c.Logging.Dir = patch.Logging.Dir
		}
		if len(patch.TicketStatuses) > 0 {
			c.TicketStatuses = patch.TicketStatuses
		}
		return nil
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.cfgStore.Get())
}

// ---- Ticket statuses -------------------------------------------------------

func (h *AdminHandler) listStatuses(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.ListTicketStatuses(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not list statuses")
		return
	}
	if rows == nil {
		rows = []db.TicketStatus{}
	}
	writeJSON(w, http.StatusOK, rows)
}

type createStatusRequest struct {
	Title    string `json:"title"`
	Color    string `json:"color"`
	Position int32  `json:"position"`
}

func (h *AdminHandler) createStatus(w http.ResponseWriter, r *http.Request) {
	var body createStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == "" {
		WriteError(w, http.StatusUnprocessableEntity, "title is required")
		return
	}
	if body.Color == "" {
		body.Color = "#808080"
	}

	row, err := h.queries.CreateTicketStatus(r.Context(), db.CreateTicketStatusParams{
		Title:    body.Title,
		Color:    body.Color,
		Position: body.Position,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not create status")
		return
	}

	// Keep config in sync
	_ = h.cfgStore.Update(func(c *config.Config) error {
		c.TicketStatuses = append(c.TicketStatuses, config.StatusConfig{
			Title: row.Title,
			Color: row.Color,
		})
		return nil
	})

	writeJSON(w, http.StatusCreated, row)
}

type updateStatusRequest struct {
	Title string `json:"title"`
	Color string `json:"color"`
}

func (h *AdminHandler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathID(r.URL.Path, "/api/admin/ticket-statuses/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	id := int32(id64)

	var body updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	row, err := h.queries.UpdateTicketStatus(r.Context(), db.UpdateTicketStatusParams{
		ID:    id,
		Title: body.Title,
		Color: body.Color,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "status not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not update status")
		return
	}

	// Rebuild config statuses from DB to stay authoritative
	h.syncStatusesToConfig(r.Context())

	writeJSON(w, http.StatusOK, row)
}

func (h *AdminHandler) deleteStatus(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathID(r.URL.Path, "/api/admin/ticket-statuses/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.queries.DeleteTicketStatus(r.Context(), int32(id64)); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not delete status")
		return
	}

	h.syncStatusesToConfig(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

// syncStatusesToConfig refreshes the in-memory config and YAML file from the DB.
func (h *AdminHandler) syncStatusesToConfig(ctx context.Context) {
	rows, err := h.queries.ListTicketStatuses(ctx)
	if err != nil {
		return
	}
	_ = h.cfgStore.Update(func(c *config.Config) error {
		c.TicketStatuses = make([]config.StatusConfig, len(rows))
		for i, r := range rows {
			c.TicketStatuses[i] = config.StatusConfig{Title: r.Title, Color: r.Color}
		}
		return nil
	})
}

// ---- Users -----------------------------------------------------------------

func (h *AdminHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.ListUsers(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not list users")
		return
	}
	if rows == nil {
		rows = []db.User{}
	}
	writeJSON(w, http.StatusOK, rows)
}

func (h *AdminHandler) getUser(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathID(r.URL.Path, "/api/admin/users/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), int32(id64))
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not fetch user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type patchUserRequest struct {
	IsActive *bool   `json:"is_active"`
	UserType *string `json:"user_type"`
}

func (h *AdminHandler) patchUser(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathID(r.URL.Path, "/api/admin/users/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	id := int32(id64)

	var body patchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := r.Context()

	if body.IsActive != nil {
		if err := h.queries.SetUserIsActive(ctx, db.SetUserIsActiveParams{
			ID:       id,
			IsActive: *body.IsActive,
		}); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update user")
			return
		}
	}

	if body.UserType != nil {
		ut := db.UserType(*body.UserType)
		if _, err := h.queries.SetUserType(ctx, db.SetUserTypeParams{
			ID:       id,
			UserType: ut,
		}); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update user type")
			return
		}
	}

	user, err := h.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not fetch updated user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// writeJSON is a convenience helper used across handlers.
func writeJSON(w http.ResponseWriter, code int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(body) //nolint:errcheck
}
