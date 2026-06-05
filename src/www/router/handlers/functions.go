package handlers

import (
	"encoding/json"
	"net/http"
)

// writeError writes a JSON error body with the shape
// {"code": <int32>, "status": "<text>", "msg": "<msg>"}.
// It mirrors the signature of http.Error so call sites read the same way.
func writeError(w http.ResponseWriter, code int, msg string) {
	body, err := json.Marshal(errorResponse{
		Code:   int32(code),
		Status: http.StatusText(code),
		Msg:    msg,
	})
	if err != nil {
		http.Error(w, msg, code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	w.Write(body)
}

func defaultResponse(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}