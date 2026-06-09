package ctxkeys

// contextKey je privátní typ zabraňující kolizím s ostatními balíčky.
type contextKey string

// SessionContextKey je uložen do kontextu požadavku autentizačním middlewarem
// a čten libovolným handlerem, který potřebuje validovanou session.
const SessionContextKey contextKey = "session"
