package mathutil

import "errors"

var (
	ErrEmptyMatrix     = errors.New("matrix is empty")
	ErrNotSquareMatrix = errors.New("matrix is not square")
	ErrSingularMatrix  = errors.New("matrix is singular")
)
