package repository

import (
	"context"

	"link-shortener-service/internal/model"
)

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=repository URLRepository
type URLRepository interface {
	PutURLPair(ctx context.Context, urlPair model.URLPair) (*model.URLPair, error)
	GetByURL(ctx context.Context, urlType string, knownURL string) (*model.URLPair, error)
}
