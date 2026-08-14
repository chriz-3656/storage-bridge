package storage

import "errors"

var (
	ErrNotFound         = errors.New("object not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrIsDirectory      = errors.New("is a directory")
	ErrNotDirectory     = errors.New("not a directory")
	ErrNotSupported     = errors.New("operation not supported")
)
