package error

import "errors"

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrMissingWhereClause = errors.New("missing where clause")
	ErrShouldUseFieldNameToSpecifyColumn = errors.New("should use field name to specify column")
)