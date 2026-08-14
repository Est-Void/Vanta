package daemon

import (
	"encoding/json"
	"net/http"

	"github.com/Est-Void/Vanta/api"
)

func writeError(w http.ResponseWriter, code api.Code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(api.StatusCode(code))
	json.NewEncoder(w).Encode(api.Error{Code: code, Message: msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
