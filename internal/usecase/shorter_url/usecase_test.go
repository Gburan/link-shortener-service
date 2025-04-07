package shorter_url

import (
	"context"
	"errors"
	"strings"
	"testing"

	rep "link-shortener-service/internal/infastracture/repository"
	"link-shortener-service/internal/model"
	mockstorage "link-shortener-service/internal/usecase/contract/repository/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutURLPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	leftURLPart := "https://some.com/"
	URLLen := 10

	reqURL := In{
		OriginalURL: "https://some.com/asdasd",
	}

	tests := []struct {
		name          string
		req           In
		setupMock     func(*mockstorage.MockURLRepository)
		expected      *model.URLPair
		expectedError error
	}{
		{
			name: "successful put",
			req:  reqURL,
			setupMock: func(mockRepo *mockstorage.MockURLRepository) {
				mockRepo.EXPECT().
					PutURLPair(gomock.Any(), gomock.AssignableToTypeOf(model.URLPair{})).
					DoAndReturn(func(_ context.Context, in model.URLPair) (*model.URLPair, error) {
						return &model.URLPair{
							Original: in.Original,
							Shorted:  in.Shorted,
						}, nil
					})
			},
			expected: &model.URLPair{
				Original: reqURL.OriginalURL,
			},
		},
		{
			name: "retry on shorted URL conflict",
			req:  reqURL,
			setupMock: func(mockRepo *mockstorage.MockURLRepository) {
				gomock.InOrder(
					mockRepo.EXPECT().
						PutURLPair(gomock.Any(), gomock.AssignableToTypeOf(model.URLPair{})).
						Return(&model.URLPair{}, rep.ErrShortedURLExist),
					mockRepo.EXPECT().
						PutURLPair(gomock.Any(), gomock.AssignableToTypeOf(model.URLPair{})).
						DoAndReturn(func(_ context.Context, in model.URLPair) (*model.URLPair, error) {
							return &model.URLPair{
								Original: in.Original,
								Shorted:  in.Shorted,
							}, nil
						}),
				)
			},
			expected: &model.URLPair{
				Original: reqURL.OriginalURL,
			},
		},
		{
			name: "original URL already exists",
			req:  reqURL,
			setupMock: func(mockRepo *mockstorage.MockURLRepository) {
				mockRepo.EXPECT().
					PutURLPair(gomock.Any(), gomock.AssignableToTypeOf(model.URLPair{})).
					DoAndReturn(func(_ context.Context, in model.URLPair) (*model.URLPair, error) {
						return &model.URLPair{
							Original: in.Original,
							Shorted:  in.Shorted,
						}, rep.ErrOriginalURLExist
					})
			},
			expected: &model.URLPair{
				Original: reqURL.OriginalURL,
			},
		},
		{
			name: "unexpected error on insert",
			req:  reqURL,
			setupMock: func(mockRepo *mockstorage.MockURLRepository) {
				mockRepo.EXPECT().
					PutURLPair(gomock.Any(), gomock.AssignableToTypeOf(model.URLPair{})).
					Return(&model.URLPair{}, errors.New("db is down"))
			},
			expectedError: ErrCheckExistingURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mockstorage.NewMockURLRepository(ctrl)
			tt.setupMock(mockRepo)

			u := NewUsecase(mockRepo, leftURLPart, URLLen)
			result, err := u.Run(context.Background(), tt.req)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.expected != nil {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Original, result.Original)
				assert.True(t, strings.HasPrefix(result.Shorted, leftURLPart))
				assert.Equal(t, URLLen, len(strings.TrimPrefix(result.Shorted, leftURLPart)))
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
