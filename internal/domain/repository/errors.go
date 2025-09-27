package repository

import "errors"

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrAssetNotFound   = errors.New("asset not found")
)
