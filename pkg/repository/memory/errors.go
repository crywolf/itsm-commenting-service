package memory

import "errors"

// ErrNotFound represents the error when object is not found in the storage
var ErrNotFound = errors.New("record was not found")
