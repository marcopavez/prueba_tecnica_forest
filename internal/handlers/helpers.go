package handlers

import (
	"encoding/json"
	"net/http"
)

func respond(response http.ResponseWriter, status int, v any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	json.NewEncoder(response).Encode(v)
}

func respondError(response http.ResponseWriter, status int, msg string) {
	respond(response, status, map[string]string{"error": msg})
}

func decode(request *http.Request, v any) error {
	return json.NewDecoder(request.Body).Decode(v)
}
