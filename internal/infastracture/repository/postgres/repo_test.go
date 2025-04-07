package postgres

import (
	"context"
	"errors"
	"testing"

	rep "link-shortener-service/internal/infastracture/repository"
	mockdb "link-shortener-service/internal/infastracture/repository/postgres/mocks"
	"link-shortener-service/internal/model"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutURLPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqURL := model.URLPair{
		Original: "https://some.com/asdasd",
		Shorted:  "xHsvC_0NTU",
	}
	dbURL := urlRow{
		OriginalURL: "https://some.com/asdasd",
		ShortedURL:  "xHsvC_0NTU",
	}

	tests := []struct {
		name          string
		setupMock     func(*mockdb.MockDBQuery)
		expected      *model.URLPair
		expectedError error
	}{
		{
			name: "successful insertion",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				mockDB.EXPECT().
					Exec(gomock.Any(), gomock.Any(), reqURL.Original, reqURL.Shorted).
					Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
			},
			expected:      &reqURL,
			expectedError: nil,
		},
		{
			name: "duplicate original URL - return existing long-short URL pair",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				mockDB.EXPECT().
					Exec(gomock.Any(), gomock.Any(), reqURL.Original, reqURL.Shorted).
					Return(pgconn.NewCommandTag(""), &pgconn.PgError{
						Code:           duplicatePgSQLErrCode,
						ConstraintName: "urls_original_url_key",
					})

				rows := pgxmock.
					NewRows([]string{"original_url", "shorted_url"}).
					AddRow(dbURL.OriginalURL, dbURL.ShortedURL).
					Kind()

				mockDB.EXPECT().
					Query(gomock.Any(), gomock.Any(), reqURL.Original).
					Return(rows, nil)
			},
			expected:      &reqURL,
			expectedError: nil,
		},
		{
			name: "duplicate original URL - db error while GetByURL request happened",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				mockDB.EXPECT().
					Exec(gomock.Any(), gomock.Any(), reqURL.Original, reqURL.Shorted).
					Return(pgconn.NewCommandTag(""), &pgconn.PgError{
						Code:           duplicatePgSQLErrCode,
						ConstraintName: "urls_original_url_key",
					})

				rows := pgxmock.
					NewRows([]string{"original_url", "shorted_url"}).
					RowError(1, rep.ErrBuildQuery).
					Kind()

				mockDB.EXPECT().
					Query(gomock.Any(), gomock.Any(), reqURL.Original).
					Return(rows, nil)
			},
			expected:      nil,
			expectedError: rep.ErrOriginalURLExist,
		},
		{
			name: "duplicate short URL",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				mockDB.EXPECT().
					Exec(gomock.Any(), gomock.Any(), reqURL.Original, reqURL.Shorted).
					Return(pgconn.NewCommandTag(""), &pgconn.PgError{
						Code:           duplicatePgSQLErrCode,
						ConstraintName: "urls_shorted_url_key",
					})
			},
			expected:      &model.URLPair{},
			expectedError: rep.ErrShortedURLExist,
		},
		{
			name: "error db - execute error",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				mockDB.EXPECT().
					Exec(gomock.Any(), gomock.Any(), reqURL.Original, reqURL.Shorted).
					Return(pgconn.NewCommandTag(""), &pgconn.PgError{
						Code: "2281337", // some unexpected error
					})
			},
			expected:      nil,
			expectedError: rep.ErrExecuteQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mockdb.NewMockDBQuery(ctrl)
			repo := &repository{db: mockDB}

			tt.setupMock(mockDB)

			result, err := repo.PutURLPair(context.Background(), reqURL)

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
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

func TestGetByURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reqURL := model.URLPair{
		Original: "https://some.com/asdasd",
	}
	dbURL := urlRow{
		OriginalURL: "https://some.com/asdasd",
		ShortedURL:  "xHsvC_0NTU",
	}

	tests := []struct {
		name             string
		requestedURLType string
		setupMock        func(*mockdb.MockDBQuery)
		expected         *model.URLPair
		expectedError    error
	}{
		{
			name:             "successful get",
			requestedURLType: "original_url",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				rows := pgxmock.
					NewRows([]string{"original_url", "shorted_url"}).
					AddRow(dbURL.OriginalURL, dbURL.ShortedURL).
					Kind()

				mockDB.EXPECT().
					Query(gomock.Any(), gomock.Any(), reqURL.Original).
					Return(rows, nil)
			},
			expected: &model.URLPair{
				Original: dbURL.OriginalURL,
				Shorted:  dbURL.ShortedURL,
			},
			expectedError: nil,
		},
		{
			name:             "db.Query error",
			requestedURLType: "original_url",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				mockDB.EXPECT().
					Query(gomock.Any(), gomock.Any(), reqURL.Original).
					Return(nil, errors.New("query error"))
			},
			expected:      nil,
			expectedError: rep.ErrExecuteQuery,
		},
		{
			name:             "no URL associated with requested URL",
			requestedURLType: "original_url",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				rows := pgxmock.
					NewRows([]string{"original_url", "shorted_url"}).
					Kind()

				mockDB.EXPECT().
					Query(gomock.Any(), gomock.Any(), reqURL.Original).
					Return(rows, nil)
			},
			expected:      nil,
			expectedError: rep.ErrNotFound,
		},
		{
			name:             "pgx.CollectOneRow - got some wrong columns data from db",
			requestedURLType: "original_url",
			setupMock: func(mockDB *mockdb.MockDBQuery) {
				rows := pgxmock.
					NewRows([]string{"some_unexected", "some_unexected_2"}).
					AddRow("unexp_data", "unexp_data_2").
					Kind()

				mockDB.EXPECT().
					Query(gomock.Any(), gomock.Any(), reqURL.Original).
					Return(rows, nil)
			},
			expected:      nil,
			expectedError: rep.ErrScanResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mockdb.NewMockDBQuery(ctrl)
			repo := &repository{db: mockDB}

			tt.setupMock(mockDB)

			result, err := repo.GetByURL(context.Background(), tt.requestedURLType, reqURL.Original)

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
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
