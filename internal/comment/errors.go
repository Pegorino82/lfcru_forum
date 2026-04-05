package comment

import "errors"

var (
	ErrEmptyContent   = errors.New("comment: empty content")
	ErrContentTooLong = errors.New("comment: content too long")
	ErrParentNotFound = errors.New("comment: parent not found")
	ErrNewsNotFound   = errors.New("comment: news not found")
)
