package repository

import "errors"

var (
	ErrBuildQuery       = errors.New("failed to build SQL query")
	ErrExecuteQuery     = errors.New("failed to execute query")
	ErrScanResult       = errors.New("failed to scan result")
	ErrNotFound         = errors.New("URL not found")
	ErrOriginalURLExist = errors.New("original URL already exists")
	ErrShortedURLExist  = errors.New("short URL already exists")
	ErrUnknownURLType   = errors.New("got unexpected URL type")
)
