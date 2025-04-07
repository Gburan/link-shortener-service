package shorter_url

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	rep "link-shortener-service/internal/infastracture/repository"
	"link-shortener-service/internal/model"
	"link-shortener-service/internal/usecase/contract/repository"
)

var (
	ErrCheckExistingURL = errors.New("failed to get short URLPair")

	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890_")
)

type usecase struct {
	repo        repository.URLRepository
	leftURLPart string
	urlLength   int
}

func NewUsecase(repo repository.URLRepository, leftURLPart string, urlLength int) *usecase {
	return &usecase{
		repo:        repo,
		leftURLPart: leftURLPart,
		urlLength:   urlLength,
	}
}

func (u *usecase) Run(ctx context.Context, req In) (*model.URLPair, error) {
	var shortedURL string
	urlPair := model.URLPair{
		Original: req.OriginalURL,
		Shorted:  shortedURL,
	}

	for {
		urlPair.Shorted = u.generateShortURL()
		record, err := u.repo.PutURLPair(ctx, urlPair)
		record.Shorted = fmt.Sprintf("%s%s", u.leftURLPart, record.Shorted)
		if err == nil {
			return record, nil
		}

		if errors.Is(err, rep.ErrShortedURLExist) {
			continue
		}
		if errors.Is(err, rep.ErrOriginalURLExist) {
			return record, nil
		}
		return nil, fmt.Errorf("%w: %v", ErrCheckExistingURL, err)
	}
}

func (u *usecase) generateShortURL() string {
	cntSymbols := len(letterRunes)
	b := make([]rune, u.urlLength)

	for i := range b {
		b[i] = letterRunes[rand.Intn(cntSymbols)]
	}
	return string(b)
}
