package inmemory

import (
	"context"
	"testing"

	rep "link-shortener-service/internal/infastracture/repository"
	"link-shortener-service/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestPutURLPair(t *testing.T) {
	pair := model.URLPair{
		Original: "https://some.com/",
		Shorted:  "https://somedomain.su/xHsvC_0NTU",
	}

	tests := []struct {
		name           string
		preData        *model.URLPair
		urlPair        model.URLPair
		expectedError  error
		expectedResult *model.URLPair
	}{
		{
			name:           "add of new URL pair",
			urlPair:        pair,
			expectedError:  nil,
			expectedResult: &pair,
		},
		{
			name:           "add URL with existing original URL",
			preData:        &model.URLPair{Original: pair.Original},
			urlPair:        pair,
			expectedError:  rep.ErrOriginalURLExist,
			expectedResult: nil,
		},
		{
			name:           "add URL with existing shortened URL",
			preData:        &model.URLPair{Shorted: pair.Shorted},
			urlPair:        pair,
			expectedError:  rep.ErrShortedURLExist,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMapRepository()

			if tt.preData != nil {
				repo.origShort[tt.preData.Original] = tt.preData.Shorted
				repo.shortOrig[tt.preData.Shorted] = tt.preData.Original
			}

			result, err := repo.PutURLPair(context.Background(), tt.urlPair)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestGetByURL(t *testing.T) {
	repo := NewMapRepository()
	urlPair := model.URLPair{
		Original: "https://some.com/",
		Shorted:  "http://short.ly/xHsvC_0NTU",
	}
	repo.PutURLPair(context.Background(), urlPair)

	tests := []struct {
		name           string
		typeOfURL      string
		knownURL       string
		expectedResult *model.URLPair
		expectedError  error
	}{
		{
			name:           "get URL by original URL",
			typeOfURL:      "original_url",
			knownURL:       "https://some.com/",
			expectedResult: &urlPair,
			expectedError:  nil,
		},
		{
			name:           "get URL by shortened URL",
			typeOfURL:      "shorted_url",
			knownURL:       "http://short.ly/xHsvC_0NTU",
			expectedResult: &urlPair,
			expectedError:  nil,
		},
		{
			name:           "get URL by non-existent URL - original",
			typeOfURL:      "original_url",
			expectedResult: nil,
			knownURL:       "https://somesomesome",
			expectedError:  rep.ErrNotFound,
		},
		{
			name:           "get URL by non-existent URL - short",
			typeOfURL:      "shorted_url",
			expectedResult: nil,
			knownURL:       "https://somesomesome",
			expectedError:  rep.ErrNotFound,
		},
		{
			name:           "get URL by non-existent URL - short",
			typeOfURL:      "unexpected",
			expectedResult: nil,
			knownURL:       "https://somesomesome",
			expectedError:  rep.ErrUnknownURLType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByURL(context.Background(), tt.typeOfURL, tt.knownURL)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
