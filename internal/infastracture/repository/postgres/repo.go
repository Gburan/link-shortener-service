package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	rep "link-shortener-service/internal/infastracture/repository"
	"link-shortener-service/internal/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableName          = "urls"
	origURLColumnName  = "original_url"
	shortURLColumnName = "shorted_url"

	duplicatePgSQLErrCode = "23505"
)

type urlRow struct {
	OriginalURL string `db:"original_url"`
	ShortedURL  string `db:"shorted_url"`
}

type repository struct {
	db DBQuery
}

func NewDBRepository(pool *pgxpool.Pool) *repository {
	return &repository{
		db: pool,
	}
}

func (r *repository) PutURLPair(ctx context.Context, urlPair model.URLPair) (*model.URLPair, error) {
	queryBuilder := squirrel.Insert(tableName).
		PlaceholderFormat(squirrel.Dollar).
		Columns(origURLColumnName, shortURLColumnName).
		Values(urlPair.Original, urlPair.Shorted)

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", rep.ErrBuildQuery, err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err == nil {
		return &urlPair, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case duplicatePgSQLErrCode:
			if strings.Contains(pgErr.ConstraintName, origURLColumnName) {
				// if long URLPair already in db - need to find long URLPair
				URLPair, err_ := r.GetByURL(ctx, origURLColumnName, urlPair.Original)
				if err_ != nil {
					return nil, fmt.Errorf("%w: %v", rep.ErrOriginalURLExist, errors.Join(err_, err))
				}
				return &model.URLPair{
					Original: urlPair.Original,
					Shorted:  URLPair.Shorted,
				}, nil
			} else if strings.Contains(pgErr.ConstraintName, shortURLColumnName) {
				// no matter which data refers to existing short URLPair in db
				return &model.URLPair{}, fmt.Errorf("%w: %v", rep.ErrShortedURLExist, err)
			}
		}
	}
	return nil, fmt.Errorf("%w: %v", rep.ErrExecuteQuery, err)
}

func (r *repository) GetByURL(ctx context.Context, urlType string, knownURL string) (*model.URLPair, error) {
	queryBuilder := squirrel.Select(origURLColumnName, shortURLColumnName).
		PlaceholderFormat(squirrel.Dollar).
		From(tableName).
		Where(squirrel.Eq{urlType: knownURL})

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", rep.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", rep.ErrExecuteQuery, err)
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[urlRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", rep.ErrNotFound, knownURL)
		}
		return nil, fmt.Errorf("%w: %v", rep.ErrScanResult, err)
	}

	return &model.URLPair{
		Original: result.OriginalURL,
		Shorted:  result.ShortedURL,
	}, nil
}
