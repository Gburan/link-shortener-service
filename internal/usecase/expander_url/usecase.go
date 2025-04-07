package expander_url

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	rep "link-shortener-service/internal/infastracture/repository"
	"link-shortener-service/internal/model"
	"link-shortener-service/internal/usecase/contract/repository"
)

const (
	shortURLColumnName = "shorted_url"
)

var (
	ErrURLNotFound  = errors.New("URLPair not found")
	ErrURLRetrieval = errors.New("failed to retrieve URLPair")

	leftURLPart = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^/]+/`)
)

type usecase struct {
	repo repository.URLRepository
}

func NewUsecase(repo repository.URLRepository) *usecase {
	return &usecase{repo: repo}
}

func (u *usecase) Run(ctx context.Context, req In) (*model.URLPair, error) {
	record, err := u.repo.GetByURL(ctx, shortURLColumnName, u.trimLeftPartURL(req.ShortedURL))
	if err != nil {
		if errors.Is(err, rep.ErrNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrURLNotFound, req.ShortedURL)
		}
		return nil, fmt.Errorf("%w: %v", ErrURLRetrieval, err)
	}
	return record, nil
}

func (u *usecase) trimLeftPartURL(fullURL string) string {
	return leftURLPart.ReplaceAllString(fullURL, "")
}
