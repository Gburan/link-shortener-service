package handler

import (
	"encoding/json"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, status int, errorMsg string, err error) {
	w.WriteHeader(status)
	err_ := json.NewEncoder(w).Encode(map[string]string{
		"error":   errorMsg,
		"details": err.Error(),
	})
	if err_ != nil {
		return
	}
}
