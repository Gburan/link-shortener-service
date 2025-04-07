package expander_url

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	expander_url "link-shortener-service/internal/handler/expander_url/mocks"
	"link-shortener-service/internal/model"
	usecase_expander_url "link-shortener-service/internal/usecase/expander_url"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShorterURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	valid := validator.New(validator.WithRequiredStructEnabled())

	reqDTO := ExpandToOriginalURL{
		ShortedURL: "https://some.com/xHsvC_0NTU",
	}
	usecaseIn := usecase_expander_url.In{
		ShortedURL: "https://some.com/xHsvC_0NTU",
	}
	usecaseOut := model.URLPair{
		Original: "https://some.com/asdasd",
		Shorted:  "xHsvC_0NTU",
	}

	tests := []struct {
		name          string
		setupMock     func(*expander_url.Mockusecase)
		reqBody       string
		expectedCode  int
		expected      string
		expectedError string
	}{
		{
			name: "successful expand",
			setupMock: func(mockUsecase *expander_url.Mockusecase) {
				mockUsecase.EXPECT().
					Run(context.TODO(), usecaseIn).
					Return(&usecaseOut, nil)
			},
			reqBody:      fmt.Sprintf(`{"shorted_url":"%s"}`, reqDTO.ShortedURL),
			expectedCode: http.StatusOK,
			expected:     usecaseOut.Original,
		},
		{
			name:          "empty body",
			setupMock:     func(mockUsecase *expander_url.Mockusecase) {},
			reqBody:       "",
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to decode request",
		},
		{
			name:          "validator error",
			setupMock:     func(mockUsecase *expander_url.Mockusecase) {},
			reqBody:       fmt.Sprintf(`{"shorted_url":"%s"}`, "bad_url"),
			expectedCode:  http.StatusBadRequest,
			expectedError: "validation failed",
		},
		{
			name: "usecase.Run error - not found existing short URL associating",
			setupMock: func(mockUsecase *expander_url.Mockusecase) {
				mockUsecase.EXPECT().
					Run(context.TODO(), usecaseIn).
					Return(nil, usecase_expander_url.ErrURLNotFound)
			},
			reqBody:       fmt.Sprintf(`{"shorted_url":"%s"}`, reqDTO.ShortedURL),
			expectedCode:  http.StatusNotFound,
			expectedError: "original URL does not exist",
		},
		{
			name: "usecase.Run error - error from storage",
			setupMock: func(mockUsecase *expander_url.Mockusecase) {
				mockUsecase.EXPECT().
					Run(context.TODO(), usecaseIn).
					Return(nil, usecase_expander_url.ErrURLRetrieval)
			},
			reqBody:       fmt.Sprintf(`{"shorted_url":"%s"}`, reqDTO.ShortedURL),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get original URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := expander_url.NewMockusecase(ctrl)
			handler := New(mockUsecase, valid)

			tt.setupMock(mockUsecase)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(
				"GET",
				"/",
				strings.NewReader(tt.reqBody),
			)
			req.Header.Set("Content-Type", "application/json")

			handler.ExpanderURL(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expected != "" {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expected, response["original_url"])
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
