package shorter_url

import (
	"context"

	"link-shortener-service/internal/model"
	"link-shortener-service/internal/usecase/shorter_url"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=shorter_url usecase
type usecase interface {
	Run(ctx context.Context, req shorter_url.In) (*model.URLPair, error)
}

type ShortFromOriginalURL struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
}
