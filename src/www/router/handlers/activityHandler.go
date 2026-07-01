package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

const defaultActivityLimit = 50
const maxActivityLimit = 200

type ActivityHandler struct {
	queries    *db.Queries
	httpLogger *logs.Logger
}

func NewActivityHandler(q *db.Queries, l *logs.Logger) *ActivityHandler {
	return &ActivityHandler{queries: q, httpLogger: l}
}

func (h *ActivityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/activity":
		h.listGlobal(w, r)
	case r.Method == http.MethodGet && matchesIDActionPath(r.URL.Path, "/api/users/", "/activity"):
		h.listForUser(w, r)
	default:
		defaultResponse(w)
	}
}

// listGlobal vrátí stránkovaný seznam všech aktivit. Přístupné pouze pro správce.
//
// @Summary      Globální feed aktivit
// @Description  Stránkovaný seznam všech zaznamenaných aktivit s volitelnými filtry. Přístupné pouze pro správce.
// @Tags         activity
// @Produce      json
// @Param        event_type   query     string  false  "Typ aktivity"
// @Param        actor_id     query     int     false  "ID aktéra"
// @Param        target_type  query     string  false  "Typ cíle (ticket, comment, user)"
// @Param        target_id    query     int     false  "ID cíle"
// @Param        from         query     string  false  "Od data (RFC3339)"
// @Param        to           query     string  false  "Do data (RFC3339)"
// @Param        limit        query     int     false  "Počet výsledků (výchozí 50, max 200)"
// @Param        offset       query     int     false  "Posun od začátku"
// @Success      200  {object}  activityListResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Security     cookieAuth
// @Router       /api/activity [get]
func (h *ActivityHandler) listGlobal(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var eventType, targetType sql.NullString
	var actorID sql.NullInt32
	var targetID sql.NullInt64
	var fromTs, toTs sql.NullTime

	if v := q.Get("event_type"); v != "" {
		eventType = sql.NullString{String: v, Valid: true}
	}
	if v := q.Get("target_type"); v != "" {
		targetType = sql.NullString{String: v, Valid: true}
	}
	if v := q.Get("actor_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			actorID = sql.NullInt32{Int32: int32(n), Valid: true}
		}
	}
	if v := q.Get("target_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			targetID = sql.NullInt64{Int64: n, Valid: true}
		}
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			fromTs = sql.NullTime{Time: t, Valid: true}
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			toTs = sql.NullTime{Time: t, Valid: true}
		}
	}

	limit, offset := parseLimitOffset(q, defaultActivityLimit, maxActivityLimit)

	ctx := r.Context()
	params := db.ListActivityLogParams{
		EventType:  eventType,
		ActorID:    actorID,
		TargetType: targetType,
		TargetID:   targetID,
		FromTs:     fromTs,
		ToTs:       toTs,
		Lim:        int32(limit),
		Off:        int32(offset),
	}
	rows, err := h.queries.ListActivityLog(ctx, params)
	if err != nil {
		h.httpLogger.Debugf("listGlobal: ListActivityLog selhalo: %s", err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst aktivity")
		return
	}
	total, err := h.queries.CountActivityLog(ctx, db.CountActivityLogParams{
		EventType:  eventType,
		ActorID:    actorID,
		TargetType: targetType,
		TargetID:   targetID,
		FromTs:     fromTs,
		ToTs:       toTs,
	})
	if err != nil {
		h.httpLogger.Debugf("listGlobal: CountActivityLog selhalo: %s", err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se spočítat aktivity")
		return
	}

	writeJSON(w, http.StatusOK, activityListResponse{
		Items:  toActivityEntries(ctx, h.queries, rows),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// listForUser vrátí aktivity konkrétního uživatele. Vlastní feed může vidět kdokoliv
// přihlášený, cizí feed jen správce.
//
// @Summary      Feed aktivit uživatele
// @Description  Stránkovaný seznam aktivit daného uživatele. Neadmin uživatel může vidět jen svůj vlastní feed.
// @Tags         activity
// @Produce      json
// @Param        id           path      int     true   "ID uživatele"
// @Param        event_type   query     string  false  "Typ aktivity"
// @Param        target_type  query     string  false  "Typ cíle (ticket, comment, user)"
// @Param        target_id    query     int     false  "ID cíle"
// @Param        from         query     string  false  "Od data (RFC3339)"
// @Param        to           query     string  false  "Do data (RFC3339)"
// @Param        limit        query     int     false  "Počet výsledků (výchozí 50, max 200)"
// @Param        offset       query     int     false  "Posun od začátku"
// @Success      200  {object}  activityListResponse
// @Failure      400  {object}  errorResponse
// @Failure      401  {object}  errorResponse
// @Failure      403  {object}  errorResponse
// @Failure      500  {object}  errorResponse
// @Security     cookieAuth
// @Router       /api/users/{id}/activity [get]
func (h *ActivityHandler) listForUser(w http.ResponseWriter, r *http.Request) {
	id64, ok := pathIDWithAction(r.URL.Path, "/api/users/", "/activity")
	if !ok {
		WriteError(w, http.StatusBadRequest, "neplatné ID uživatele")
		return
	}
	id := int32(id64)

	user, ok := userFromContext(w, r)
	if !ok {
		return
	}
	if user.UserType != db.UserTypeAdmin && user.ID != id {
		WriteError(w, http.StatusForbidden, "přístup odepřen")
		return
	}

	q := r.URL.Query()

	var eventType, targetType sql.NullString
	var targetID sql.NullInt64
	var fromTs, toTs sql.NullTime

	if v := q.Get("event_type"); v != "" {
		eventType = sql.NullString{String: v, Valid: true}
	}
	if v := q.Get("target_type"); v != "" {
		targetType = sql.NullString{String: v, Valid: true}
	}
	if v := q.Get("target_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			targetID = sql.NullInt64{Int64: n, Valid: true}
		}
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			fromTs = sql.NullTime{Time: t, Valid: true}
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			toTs = sql.NullTime{Time: t, Valid: true}
		}
	}

	limit, offset := parseLimitOffset(q, defaultActivityLimit, maxActivityLimit)

	ctx := r.Context()
	rows, err := h.queries.ListActivityLogForUser(ctx, db.ListActivityLogForUserParams{
		ActorID:    id,
		EventType:  eventType,
		TargetType: targetType,
		TargetID:   targetID,
		FromTs:     fromTs,
		ToTs:       toTs,
		Lim:        int32(limit),
		Off:        int32(offset),
	})
	if err != nil {
		h.httpLogger.Debugf("listForUser: ListActivityLogForUser selhalo (user=%d): %s", id, err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst aktivity")
		return
	}
	total, err := h.queries.CountActivityLogForUser(ctx, db.CountActivityLogForUserParams{
		ActorID:    id,
		EventType:  eventType,
		TargetType: targetType,
		TargetID:   targetID,
		FromTs:     fromTs,
		ToTs:       toTs,
	})
	if err != nil {
		h.httpLogger.Debugf("listForUser: CountActivityLogForUser selhalo (user=%d): %s", id, err)
		WriteError(w, http.StatusInternalServerError, "nepodařilo se spočítat aktivity")
		return
	}

	writeJSON(w, http.StatusOK, activityListResponse{
		Items:  toActivityEntries(ctx, h.queries, rows),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// parseLimitOffset rozparsuje stránkovací parametry se zadaným výchozím a maximálním limitem.
func parseLimitOffset(q map[string][]string, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	if l := getFirst(q, "limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			if n > maxLimit {
				n = maxLimit
			}
			limit = n
		}
	}
	if o := getFirst(q, "offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}
	return limit, offset
}

func getFirst(q map[string][]string, key string) string {
	if vs, ok := q[key]; ok && len(vs) > 0 {
		return vs[0]
	}
	return ""
}

// toActivityEntries převede řádky z DB na API odpověď a dopočítá jméno aktéra.
func toActivityEntries(ctx context.Context, q *db.Queries, rows []db.ActivityLog) []activityEntry {
	entries := make([]activityEntry, 0, len(rows))
	nameCache := make(map[int32]string, len(rows))
	for _, row := range rows {
		entry := activityEntry{
			ID:        row.ID,
			EventType: row.EventType,
			CreatedAt: row.CreatedAt,
		}
		if row.ActorID.Valid {
			actorID := row.ActorID.Int32
			entry.ActorID = &actorID
			if _, ok := nameCache[actorID]; !ok {
				nameCache[actorID] = resolveAuthorName(ctx, q, actorID)
			}
			entry.ActorName = nameCache[actorID]
		}
		if row.TargetType.Valid {
			entry.TargetType = row.TargetType.String
		}
		if row.TargetID.Valid {
			targetID := row.TargetID.Int64
			entry.TargetID = &targetID
		}
		if row.Payload.Valid {
			entry.Payload = row.Payload.RawMessage
		}
		entries = append(entries, entry)
	}
	return entries
}
