package shorter_url

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	shorter_url "link-shortener-service/internal/handler/shorter_url/mocks"
	"link-shortener-service/internal/model"
	usecase_shorter_url "link-shortener-service/internal/usecase/shorter_url"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShorterURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validator.New(validator.WithRequiredStructEnabled())

	reqDTO := ShortFromOriginalURL{
		OriginalURL: "https://some.com/asdasd",
	}
	usecaseIn := usecase_shorter_url.In{
		OriginalURL: "https://some.com/asdasd",
	}
	usecaseOut := model.URLPair{
		Original: "https://some.com/asdasd",
		Shorted:  "xHsvC_0NTU",
	}

	tests := []struct {
		name          string
		setupMock     func(*shorter_url.Mockusecase)
		reqBody       string
		expectedCode  int
		expected      string
		expectedError string
	}{
		{
			name: "successful shorten",
			setupMock: func(mockUsecase *shorter_url.Mockusecase) {
				mockUsecase.EXPECT().
					Run(context.TODO(), usecaseIn).
					Return(&usecaseOut, nil)
			},
			reqBody:      fmt.Sprintf(`{"original_url":"%s"}`, reqDTO.OriginalURL),
			expectedCode: http.StatusOK,
			expected:     usecaseOut.Shorted,
		},
		{
			name:          "empty body",
			setupMock:     func(mockUsecase *shorter_url.Mockusecase) {},
			reqBody:       "",
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to decode request",
		},
		{
			name:          "validator error",
			setupMock:     func(mockUsecase *shorter_url.Mockusecase) {},
			reqBody:       fmt.Sprintf(`{"original_url":"%s"}`, "bad_url"),
			expectedCode:  http.StatusBadRequest,
			expectedError: "validation failed",
		},
		{
			name: "usecase.Run error",
			setupMock: func(mockUsecase *shorter_url.Mockusecase) {
				mockUsecase.EXPECT().
					Run(context.TODO(), usecaseIn).
					Return(nil, usecase_shorter_url.ErrCheckExistingURL)
			},
			reqBody:       fmt.Sprintf(`{"original_url":"%s"}`, reqDTO.OriginalURL),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed getting short URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := shorter_url.NewMockusecase(ctrl)
			handler := NewUrlHandler(mockUsecase, valid)

			tt.setupMock(mockUsecase)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(
				"POST",
				"/",
				strings.NewReader(tt.reqBody),
			)
			req.Header.Set("Content-Type", "application/json")

			handler.ShorterURL(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expected != "" {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, response["short_url"])
			}

			if tt.expectedError != "" {
				var errorResponse map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&errorResponse)
				require.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.expectedError)
			}
		})
	}
}
