package expander_url

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"link-shortener-service/internal/handler"
	usecase_expander_url "link-shortener-service/internal/usecase/expander_url"

	"github.com/go-playground/validator/v10"
)

type urlHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func NewUrlHandler(usecase usecase, validator *validator.Validate) *urlHandler {
	return &urlHandler{
		usecase:   usecase,
		validator: validator,
	}
}

func (h *urlHandler) ExpanderURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var url ExpandToOriginalURL
	if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "failed to decode request", err)
		return
	}

	if err := h.validator.Struct(url); err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	ctx := context.TODO()
	result, err := h.usecase.Run(ctx, usecase_expander_url.In{
		ShortedURL: url.ShortedURL,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(map[string]string{
		"original_url": result.Original,
	}); err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "failed to encode response", err)
		return
	}
}

func handleUseCaseError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase_expander_url.ErrURLNotFound):
		statusCode = http.StatusNotFound
		errorMsg = "original URL does not exist"
	case errors.Is(err, usecase_expander_url.ErrURLRetrieval):
		errorMsg = "failed to get original URL"
	}

	handler.RespondWithError(w, statusCode, errorMsg, err)
}
