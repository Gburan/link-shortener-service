package expander_url

import (
	"context"

	"link-shortener-service/internal/model"
	"link-shortener-service/internal/usecase/expander_url"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=expander_url usecase
type usecase interface {
	Run(ctx context.Context, req expander_url.In) (*model.URLPair, error)
}

type ExpandToOriginalURL struct {
	ShortedURL string `json:"shorted_url" validate:"required,url"`
}
