package shorter_url

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"link-shortener-service/internal/handler"
	usecase_shorter_url "link-shortener-service/internal/usecase/shorter_url"

	"github.com/go-playground/validator/v10"
)

type urlHandler struct {
	usecase   usecase
	validator *validator.Validate
}

func New(IShortURLUseCase usecase, validator *validator.Validate) *urlHandler {
	return &urlHandler{
		usecase:   IShortURLUseCase,
		validator: validator,
	}
}

func (h *urlHandler) ShorterURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var url ShortFromOriginalURL
	if err := json.NewDecoder(r.Body).Decode(&url); err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "failed to decode request", err)
		return
	}

	if err := h.validator.Struct(url); err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	ctx := context.TODO()
	result, err := h.usecase.Run(ctx, usecase_shorter_url.In{
		OriginalURL: url.OriginalURL,
	})
	if err != nil {
		handleUseCaseError(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(map[string]string{
		"short_url": result.Shorted,
	}); err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "failed to encode response", err)
		return
	}
}

func handleUseCaseError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	errorMsg := "internal server error"

	switch {
	case errors.Is(err, usecase_shorter_url.ErrCheckExistingURL):
		errorMsg = "failed getting short URL"
	}

	handler.RespondWithError(w, statusCode, errorMsg, err)
}
