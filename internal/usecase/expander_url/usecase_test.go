package expander_url

import (
	"context"
	"errors"
	"fmt"
	"testing"

	rep "link-shortener-service/internal/infastracture/repository"
	"link-shortener-service/internal/model"
	mockstorage "link-shortener-service/internal/usecase/contract/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrimLeftPartURL(t *testing.T) {
	var leftPartURLs = []string{
		"http://example.com/",
		"https://example.com/",
		"http://example.com:8080/",
		"https://example.com:443/",
		"http://127.0.0.1/",
		"http://127.0.0.1:3000/",
		"http://localhost/",
		"http://localhost:8080/",
		"https://sub.example.com/",
		"https://www.example.com/",
		"http://[::1]/",
		"http://[::1]:8080/",
		"http://user:pass@example.com/",
	}

	shortedURL := "xHsvC_0NTU"
	u := usecase{}

	for _, leftPart := range leftPartURLs {
		fullURL := fmt.Sprintf("%s%s", leftPart, shortedURL)
		t.Run(fullURL, func(t *testing.T) {
			result := u.trimLeftPartURL(fullURL)
			assert.Equal(t, shortedURL, result, "failed trim for: %s", fullURL)
		})
	}
}

func TestPutURLPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	outURL := model.URLPair{
		Original: "https://some.com/asdasd",
		Shorted:  "xHsvC_0NTU",
	}
	reqURL := In{
		ShortedURL: "xHsvC_0NTU",
	}

	tests := []struct {
		name          string
		req           In
		setupMock     func(*mockstorage.MockURLRepository)
		expected      *model.URLPair
		expectedError error
	}{
		{
			name: "successful get",
			req:  reqURL,
			setupMock: func(mockDB *mockstorage.MockURLRepository) {
				mockDB.EXPECT().
					GetByURL(ctx, shortURLColumnName, "xHsvC_0NTU").
					Return(&outURL, nil)
			},
			expected:      &outURL,
			expectedError: nil,
		},
		{
			name: "no original URL",
			req:  reqURL,
			setupMock: func(mockDB *mockstorage.MockURLRepository) {
				mockDB.EXPECT().
					GetByURL(ctx, shortURLColumnName, "xHsvC_0NTU").
					Return(nil, rep.ErrNotFound)
			},
			expected:      nil,
			expectedError: ErrURLNotFound,
		},
		{
			name: "internal error from storage",
			req:  reqURL,
			setupMock: func(mockDB *mockstorage.MockURLRepository) {
				mockDB.EXPECT().
					GetByURL(ctx, shortURLColumnName, "xHsvC_0NTU").
					Return(nil, errors.New("some_int_error"))
			},
			expected:      nil,
			expectedError: ErrURLRetrieval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mockstorage.NewMockURLRepository(ctrl)
			tt.setupMock(mockStorage)

			u := NewUsecase(mockStorage)
			result, err := u.Run(context.Background(), reqURL)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.expected != nil {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
