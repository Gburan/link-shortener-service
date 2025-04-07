package handler

import (
	"encoding/json"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, status int, errorMsg string, err error) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errorMsg,
		"details": err.Error(),
	})
}
