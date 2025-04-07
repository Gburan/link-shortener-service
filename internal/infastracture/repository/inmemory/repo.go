package inmemory

import (
	"context"
	"sync"

	rep "link-shortener-service/internal/infastracture/repository"

	"link-shortener-service/internal/model"
)

type repository struct {
	mu        sync.RWMutex
	shortOrig map[string]string
	origShort map[string]string
}

func NewMapRepository() *repository {
	return &repository{
		shortOrig: make(map[string]string),
		origShort: make(map[string]string),
	}
}

func (r *repository) PutURLPair(_ context.Context, urlPair model.URLPair) (*model.URLPair, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingShortened, exists := r.origShort[urlPair.Original]; exists {
		return &model.URLPair{
			Original: urlPair.Original,
			Shorted:  existingShortened,
		}, rep.ErrOriginalURLExist
	}

	if _, exists := r.shortOrig[urlPair.Shorted]; exists {
		return &model.URLPair{}, rep.ErrShortedURLExist
	}

	r.origShort[urlPair.Original] = urlPair.Shorted
	r.shortOrig[urlPair.Shorted] = urlPair.Original

	return &urlPair, nil
}

func (r *repository) GetByURL(_ context.Context, typeOfURL, knownURL string) (*model.URLPair, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var url string
	var exists bool
	switch typeOfURL {
	case "original_url":
		url, exists = r.origShort[knownURL]
		if !exists {
			return nil, rep.ErrNotFound
		}
		return &model.URLPair{Original: knownURL, Shorted: url}, nil
	case "shorted_url":
		url, exists = r.shortOrig[knownURL]
		if !exists {
			return nil, rep.ErrNotFound
		}
		return &model.URLPair{Original: url, Shorted: knownURL}, nil
	default:
		return nil, rep.ErrUnknownURLType
	}
}
