package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/mailer"
)

type ServerSettingsHandler struct {
	queries *db.Queries
	mailer  *mailer.Mailer
	logger  *logs.Logger
}

func NewServerSettingsHandler(q *db.Queries, m *mailer.Mailer, l *logs.Logger) *ServerSettingsHandler {
	return &ServerSettingsHandler{queries: q, mailer: m, logger: l}
}

func (h *ServerSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/setup/smtp/test":
		h.testSMTP(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/setup/db/test":
		h.testDB(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/setup/complete":
		h.completeWizard(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/admin/server-settings":
		h.getServerSettings(w, r)
	case r.Method == http.MethodPatch && r.URL.Path == "/api/admin/server-settings/smtp":
		h.patchSMTP(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/api/admin/server-settings/smtp/test":
		h.testSMTP(w, r)
	default:
		WriteError(w, http.StatusNotFound, "endpoint nenalezen")
	}
}

// smtpTestRequest je tělo požadavku pro testování SMTP připojení.
type smtpTestRequest struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// smtpPatchRequest je tělo pro PATCH /api/admin/server-settings/smtp.
type smtpPatchRequest struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

// serverSettingField je jeden klíč v odpovědi server-settings.
type serverSettingField struct {
	Value   string `json:"value"`
	FromEnv bool   `json:"from_env"`
}

// serverSettingsResponse je odpověď GET /api/admin/server-settings.
type serverSettingsResponse struct {
	SMTP smtpSettings `json:"smtp"`
	DB   dbSettings   `json:"db"`
}

type smtpSettings struct {
	Host     serverSettingField `json:"host"`
	Port     serverSettingField `json:"port"`
	Username serverSettingField `json:"username"`
	Password serverSettingField `json:"password"`
	From     serverSettingField `json:"from"`
}

type dbSettings struct {
	Host     serverSettingField `json:"host"`
	Port     serverSettingField `json:"port"`
	User     serverSettingField `json:"user"`
	Database serverSettingField `json:"database"`
	SSLMode  serverSettingField `json:"sslmode"`
}

func (h *ServerSettingsHandler) testSMTP(w http.ResponseWriter, r *http.Request) {
	var body smtpTestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}

	// Pokud frontend neposílá heslo (pole z prostředí jsou maskovaná), použijeme
	// živý mailer, který má reálné přihlašovací údaje načtené při startu.
	if body.Password == "" && h.mailer != nil {
		if err := h.mailer.Ping(); err != nil {
			WriteError(w, http.StatusBadGateway, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if body.Host == "" || body.Port == "" {
		WriteError(w, http.StatusBadRequest, "host a port jsou povinné")
		return
	}
	if err := mailer.TestCredentials(body.Host, body.Port, body.Username, body.Password); err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ServerSettingsHandler) testDB(w http.ResponseWriter, r *http.Request) {
	// DB je již připojená, testujeme aktuální spojení.
	if _, err := h.queries.CountUsers(r.Context()); err != nil {
		WriteError(w, http.StatusBadGateway, "Databáze není dostupná: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ServerSettingsHandler) completeWizard(w http.ResponseWriter, r *http.Request) {
	err := h.queries.UpsertServerSetting(r.Context(), db.UpsertServerSettingParams{
		Key:     "wizard_completed",
		Value:   "true",
		FromEnv: false,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se dokončit wizard")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ServerSettingsHandler) getServerSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.GetAllServerSettings(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "nepodařilo se načíst nastavení")
		return
	}
	m := make(map[string]db.ServerSetting)
	for _, row := range rows {
		m[row.Key] = row
	}

	field := func(key string) serverSettingField {
		if s, ok := m[key]; ok {
			val := s.Value
			// Zobrazíme *** pro hesla uložená z env, prázdný string pro neuložená.
			if (key == "smtp_password" || key == "pg_password") && s.FromEnv {
				val = ""
			}
			return serverSettingField{Value: val, FromEnv: s.FromEnv}
		}
		return serverSettingField{}
	}

	res := serverSettingsResponse{
		SMTP: smtpSettings{
			Host:     field("smtp_host"),
			Port:     field("smtp_port"),
			Username: field("smtp_user"),
			Password: field("smtp_password"),
			From:     field("smtp_from"),
		},
		DB: dbSettings{
			Host:     field("pg_host"),
			Port:     field("pg_port"),
			User:     field("pg_user"),
			Database: field("pg_database"),
			SSLMode:  field("pg_sslmode"),
		},
	}

	jsonRes, err := json.Marshal(res)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes) //nolint:errcheck
}

func (h *ServerSettingsHandler) patchSMTP(w http.ResponseWriter, r *http.Request) {
	var body smtpPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "neplatné tělo požadavku")
		return
	}
	if body.Host == "" || body.Port == "" {
		WriteError(w, http.StatusBadRequest, "host a port jsou povinné")
		return
	}

	ctx := r.Context()
	settings := []struct {
		key string
		val string
	}{
		{"smtp_host", body.Host},
		{"smtp_port", body.Port},
		{"smtp_user", body.Username},
		{"smtp_password", body.Password},
		{"smtp_from", body.From},
	}
	for _, s := range settings {
		if err := h.queries.UpsertServerSetting(ctx, db.UpsertServerSettingParams{
			Key:     s.key,
			Value:   s.val,
			FromEnv: false,
		}); err != nil {
			WriteError(w, http.StatusInternalServerError, "nepodařilo se uložit nastavení")
			return
		}
	}

	// Reload maileru za běhu pokud byl inicializován. Pokud je nil (SMTP nebylo
	// nakonfigurováno při startu), nastavení se uloží do DB a aktivuje po restartu.
	if h.mailer != nil {
		h.mailer.Reload(body.Host, body.Port, body.Username, body.Password, body.From)
	} else {
		h.logger.Info("SMTP nastaveno v DB; aktivuje se po restartu serveru.")
	}

	w.WriteHeader(http.StatusNoContent)
}
