package curation

import "errors"

var (
	ErrNotFound          = errors.New("collection not found")
	ErrItemNotFound      = errors.New("item not found")
	ErrInvalidItemType   = errors.New("invalid item type")
	ErrInvalidItemRef    = errors.New("referenced entity does not exist")
	ErrDuplicateItem     = errors.New("item already exists in this collection")
	ErrCycleDetected     = errors.New("moving collection would create a cycle")
	ErrCannotContainSelf = errors.New("collection cannot contain itself")
)
