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
	case r.Method == http.MethodPost && matchesIDActionPath(r.URL.Path, "/api/admin/users/", "/approve"):
		h.approveUser(w, r)
	case r.Method == http.MethodPost && matchesIDActionPath(r.URL.Path, "/api/admin/users/", "/reject"):
		h.rejectUser(w, r)

	default:
		defaultResponse(w)
	}
}

// matchesIDPath vrátí true pokud cesta začíná prefixem a má neprázdný suffix.
func matchesIDPath(path, prefix string) bool {
	return len(path) > len(prefix) && path[:len(prefix)] == prefix
}

// matchesIDActionPath vrátí true pokud cesta odpovídá vzoru /prefix{id}/action.
func matchesIDActionPath(path, prefix, action string) bool {
	if !matchesIDPath(path, prefix) {
		return false
	}
	return len(path) > len(prefix)+len(action) && path[len(path)-len(action):] == action
}

// pathIDWithAction extrahuje číselné ID z cesty ve tvaru /prefix{id}/action.
func pathIDWithAction(path, prefix, action string) (int64, bool) {
	s := path[len(prefix) : len(path)-len(action)]
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil
}

func pathID(path, prefix string) (int64, bool) {
	s := path[len(prefix):]
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil
}

// ---- Config ----------------------------------------------------------------

// getConfig vrátí aktuální runtime konfiguraci systému.
// Konfigurace se čte z in-memory store (synchronizovaného s ticketa.yaml na disku).
//
// @Summary      Získat konfiguraci
// @Description  Vrátí aktuální runtime konfiguraci — logování a seznam stavů tiketů.
// @Tags         admin
// @Produce      json
// @Success      200  {object}  configResponse  "Aktuální konfigurace"
// @Failure      401  {object}  errorResponse   "Chybí nebo vypršel session cookie"
// @Failure      403  {object}  errorResponse   "Přístup pouze pro maintainer"
// @Security     cookieAuth
// @Router       /api/admin/config [get]
func (h *AdminHandler) getConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.cfgStore.Get())
}

// patchConfig aktualizuje runtime konfiguraci systému.
// Změny se zapisují atomicky do /config/ticketa.yaml a projeví se okamžitě bez restartu.
// Pokud je uveden TicketStatuses, musí obsahovat alespoň 3 položky.
//
// @Summary      Aktualizovat konfiguraci
// @Description  Aktualizuje runtime konfiguraci. Uvádějte pouze pole, která chcete změnit. Změny jsou zapsány atomicky (temp soubor + rename) do ticketa.yaml.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        body  body      patchConfigRequest  true  "Změny konfigurace (všechna pole volitelná)"
// @Success      200   {object}  configResponse      "Aktualizovaná konfigurace"
// @Failure      400   {object}  errorResponse       "Neplatné tělo nebo chyba zápisu na disk"
// @Failure      401   {object}  errorResponse       "Chybí nebo vypršel session cookie"
// @Failure      403   {object}  errorResponse       "Přístup pouze pro maintainer"
// @Security     cookieAuth
// @Router       /api/admin/config [patch]
func (h *AdminHandler) patchConfig(w http.ResponseWriter, r *http.Request) {
	var patch config.Config
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
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

// listStatuses vrátí seznam všech stavů tiketů seřazených podle pozice.
// Prázdný výsledek vrátí [] (nikdy null).
//
// @Summary      Seznam stavů tiketů
// @Description  Vrátí všechny stavy tiketů seřazené podle position. Používá se k naplnění výběru stavu při vytváření nebo editaci tiketu.
// @Tags         admin
// @Produce      json
// @Success      200  {array}   ticketStatusResponse  "Seznam stavů"
// @Failure      401  {object}  errorResponse         "Chybí nebo vypršel session cookie"
// @Failure      403  {object}  errorResponse         "Přístup pouze pro maintainer"
// @Failure      500  {object}  errorResponse         "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/ticket-statuses [get]
func (h *AdminHandler) listStatuses(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.ListTicketStatuses(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst stavy")
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

// createStatus vytvoří nový stav tiketu a přidá ho do YAML konfigurace.
// Výchozí barva je #808080 pokud není zadána.
//
// @Summary      Vytvořit stav tiketu
// @Description  Vytvoří nový stav tiketu. Nový stav je automaticky přidán do ticketa.yaml. Position musí být unikátní.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        body  body      createStatusRequest   true  "Nový stav"
// @Success      201   {object}  ticketStatusResponse  "Vytvořený stav"
// @Failure      400   {object}  errorResponse         "Neplatné tělo požadavku"
// @Failure      401   {object}  errorResponse         "Chybí nebo vypršel session cookie"
// @Failure      403   {object}  errorResponse         "Přístup pouze pro maintainer"
// @Failure      422   {object}  errorResponse         "Chybí povinné pole title"
// @Failure      500   {object}  errorResponse         "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/ticket-statuses [post]
func (h *AdminHandler) createStatus(w http.ResponseWriter, r *http.Request) {
	var body createStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if body.Title == "" {
		WriteError(w, http.StatusUnprocessableEntity, "pole title je povinné")
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
		WriteError(w, http.StatusInternalServerError, "nepodařilo se vytvořit stav")
		return
	}

	// Synchronizace konfigurace
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

// updateStatus aktualizuje existující stav tiketu a synchronizuje YAML konfiguraci.
//
// @Summary      Aktualizovat stav tiketu
// @Description  Aktualizuje název nebo barvu stavu. Změna je automaticky synchronizována do ticketa.yaml.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int                   true  "ID stavu"
// @Param        body  body      updateStatusRequest   true  "Aktualizovaná data stavu"
// @Success      200   {object}  ticketStatusResponse  "Aktualizovaný stav"
// @Failure      400   {object}  errorResponse         "Neplatné ID nebo tělo požadavku"
// @Failure      401   {object}  errorResponse         "Chybí nebo vypršel session cookie"
// @Failure      403   {object}  errorResponse         "Přístup pouze pro maintainer"
// @Failure      404   {object}  errorResponse         "Stav nenalezen"
// @Failure      500   {object}  errorResponse         "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/ticket-statuses/{id} [put]
func (h *AdminHandler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathID(r.URL.Path, "/api/admin/ticket-statuses/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID")
		return
	}
	id := int32(id64)

	var body updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}

	row, err := h.queries.UpdateTicketStatus(r.Context(), db.UpdateTicketStatusParams{
		ID:    id,
		Title: body.Title,
		Color: body.Color,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "stav nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat stav")
		return
	}

	// Přegenerování stavů konfigurace z DB jako autoritativního zdroje
	h.syncStatusesToConfig(r.Context())

	writeJSON(w, http.StatusOK, row)
}

// deleteStatus smaže stav tiketu a synchronizuje YAML konfiguraci.
// Tikety odkazující na tento stav budou mít status_id nastaven na null.
//
// @Summary      Smazat stav tiketu
// @Description  Smaže stav tiketu. Tikety odkazující na tento stav budou mít status_id nastaven na null. Konfigurace je automaticky synchronizována.
// @Tags         admin
// @Param        id  path  int  true  "ID stavu"
// @Success      204 "Stav smazán"
// @Failure      400 {object}  errorResponse  "Neplatné ID"
// @Failure      401 {object}  errorResponse  "Chybí nebo vypršel session cookie"
// @Failure      403 {object}  errorResponse  "Přístup pouze pro maintainer"
// @Failure      500 {object}  errorResponse  "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/ticket-statuses/{id} [delete]
func (h *AdminHandler) deleteStatus(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathID(r.URL.Path, "/api/admin/ticket-statuses/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID")
		return
	}

	if err := h.queries.DeleteTicketStatus(r.Context(), int32(id64)); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se smazat stav")
		return
	}

	h.syncStatusesToConfig(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

// syncStatusesToConfig aktualizuje in-memory konfiguraci a YAML soubor z DB.
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

// listUsers vrátí seznam všech uživatelů.
// Prázdný výsledek vrátí [] (nikdy null).
//
// @Summary      Seznam uživatelů
// @Description  Vrátí všechny uživatele systému. Přístupné pouze pro maintainer.
// @Tags         admin
// @Produce      json
// @Success      200  {array}   userResponse   "Seznam uživatelů"
// @Failure      401  {object}  errorResponse  "Chybí nebo vypršel session cookie"
// @Failure      403  {object}  errorResponse  "Přístup pouze pro maintainer"
// @Failure      500  {object}  errorResponse  "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/users [get]
func (h *AdminHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.ListUsers(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst uživatele")
		return
	}
	if rows == nil {
		rows = []db.User{}
	}
	writeJSON(w, http.StatusOK, rows)
}

// getUser vrátí jednoho uživatele podle ID.
//
// @Summary      Získat uživatele
// @Description  Vrátí uživatele podle jeho ID. Přístupné pouze pro maintainer.
// @Tags         admin
// @Produce      json
// @Param        id   path      int           true  "ID uživatele"
// @Success      200  {object}  userResponse  "Uživatel"
// @Failure      400  {object}  errorResponse "Neplatné ID"
// @Failure      401  {object}  errorResponse "Chybí nebo vypršel session cookie"
// @Failure      403  {object}  errorResponse "Přístup pouze pro maintainer"
// @Failure      404  {object}  errorResponse "Uživatel nenalezen"
// @Failure      500  {object}  errorResponse "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/users/{id} [get]
func (h *AdminHandler) getUser(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathID(r.URL.Path, "/api/admin/users/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), int32(id64))
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "uživatel nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst uživatele")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type patchUserRequest struct {
	IsActive *bool   `json:"is_active"`
	UserType *string `json:"user_type"`
}

// patchUser aktualizuje uživatele — aktivaci nebo roli.
// Lze měnit samostatně nebo obojí naráz. Neaktivní uživatelé se nemohou přihlásit.
//
// @Summary      Aktualizovat uživatele
// @Description  Aktualizuje is_active nebo user_type uživatele. Obě pole jsou volitelná. Neaktivní uživatelé se nemohou přihlásit — při deaktivaci je jejich existující session okamžitě zneplatněna.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int               true  "ID uživatele"
// @Param        body  body      patchUserRequest  true  "Aktualizovaná data uživatele"
// @Success      200   {object}  userResponse      "Aktualizovaný uživatel"
// @Failure      400   {object}  errorResponse     "Neplatné ID nebo tělo požadavku"
// @Failure      401   {object}  errorResponse     "Chybí nebo vypršel session cookie"
// @Failure      403   {object}  errorResponse     "Přístup pouze pro maintainer"
// @Failure      404   {object}  errorResponse     "Uživatel nenalezen"
// @Failure      500   {object}  errorResponse     "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/users/{id} [patch]
func (h *AdminHandler) patchUser(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathID(r.URL.Path, "/api/admin/users/")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID")
		return
	}
	id := int32(id64)

	var body patchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}

	ctx := r.Context()

	if body.IsActive != nil {
		if err := h.queries.SetUserIsActive(ctx, db.SetUserIsActiveParams{
			ID:       id,
			IsActive: *body.IsActive,
		}); err != nil {
			WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat uživatele")
			return
		}

		// Deaktivace musí mít okamžitý účinek — session uživatele se zneplatní hned,
		// ne až při jeho dalším požadavku.
		if !*body.IsActive {
			if err := h.queries.SoftDeleteSessionByUserID(ctx, int64(id)); err != nil {
				WriteError(w, http.StatusInternalServerError, "nepodařilo se zneplatnit session uživatele")
				return
			}
		}
	}

	if body.UserType != nil {
		ut := db.UserType(*body.UserType)
		if _, err := h.queries.SetUserType(ctx, db.SetUserTypeParams{
			ID:       id,
			UserType: ut,
		}); err != nil {
			WriteError(w, http.StatusInternalServerError, "nepodařilo se aktualizovat typ uživatele")
			return
		}
	}

	user, err := h.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "uživatel nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst aktualizovaného uživatele")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// approveUser schválí čekajícího uživatele — nastaví user_type = requested_role a zapíše kdo schválil.
//
// @Summary      Schválit čekajícího uživatele
// @Description  Nastaví user_type na requested_role a zapíše ID schvalovatele. Přístupné pro staff a admin.
// @Tags         admin
// @Param        id   path      int  true  "ID uživatele"
// @Success      200  {object}  userResponse  "Schválený uživatel"
// @Failure      400  {object}  errorResponse "Neplatné ID"
// @Failure      401  {object}  errorResponse "Nepřihlášen"
// @Failure      403  {object}  errorResponse "Přístup odepřen"
// @Failure      404  {object}  errorResponse "Uživatel nenalezen"
// @Failure      500  {object}  errorResponse "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/users/{id}/approve [post]
func (h *AdminHandler) approveUser(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathIDWithAction(r.URL.Path, "/api/admin/users/", "/approve")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID")
		return
	}
	id := int32(id64)

	approver, ok := userFromContext(w, r)
	if !ok {
		return
	}

	ctx := r.Context()

	if err := h.queries.ApprovePendingUser(ctx, db.ApprovePendingUserParams{
		ID:         id,
		ApprovedBy: sql.NullInt32{Int32: approver.ID, Valid: true},
	}); err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "uživatel nenalezen nebo nebyl ve stavu pending")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se schválit uživatele")
		return
	}

	user, err := h.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			WriteError(w, http.StatusNotFound, "uživatel nenalezen")
			return
		}
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst uživatele")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// rejectUser zamítne čekajícího uživatele — deaktivuje jeho účet a okamžitě zneplatní session.
//
// @Summary      Zamítnout čekajícího uživatele
// @Description  Deaktivuje účet a zneplatní session. Přístupné pro staff a admin.
// @Tags         admin
// @Param        id   path      int  true  "ID uživatele"
// @Success      204  "Zamítnuto"
// @Failure      400  {object}  errorResponse "Neplatné ID"
// @Failure      401  {object}  errorResponse "Nepřihlášen"
// @Failure      403  {object}  errorResponse "Přístup odepřen"
// @Failure      500  {object}  errorResponse "Interní chyba"
// @Security     cookieAuth
// @Router       /api/admin/users/{id}/reject [post]
func (h *AdminHandler) rejectUser(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathIDWithAction(r.URL.Path, "/api/admin/users/", "/reject")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID")
		return
	}
	id := int32(id64)

	ctx := r.Context()

	if err := h.queries.RejectPendingUser(ctx, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se zamítnout uživatele")
		return
	}
	if err := h.queries.SoftDeleteSessionByUserID(ctx, int64(id)); err != nil {
		h.httpLogger.Debugf("rejectUser: SoftDeleteSessionByUserID selhalo pro user_id=%d: %s", id, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeJSON je pomocná funkce pro zápis JSON odpovědi.
func writeJSON(w http.ResponseWriter, code int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(body) //nolint:errcheck
}
