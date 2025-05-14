package database

import "errors"

// Common database errors
var (
	ErrRecordNotFound = errors.New("record not found")
)
