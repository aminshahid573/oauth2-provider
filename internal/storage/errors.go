package storage

import "errors"

// ErrNotFound is returned when a requested item is not found in the storage.
var ErrNotFound = errors.New("not found")
