package ctxkeys

// contextKey is a private type preventing collisions with other packages.
type contextKey string

// SessionContextKey is stored in the request context by auth middleware
// and read by any handler that needs the validated session.
const SessionContextKey contextKey = "session"
