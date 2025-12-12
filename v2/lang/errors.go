package lang

import "errors"

var (
	ErrNilArena = errors.New("nil arena")

	ErrParameterLimitExceeded = errors.New("parameter limit exceeded")
	ErrDepthLimitExceeded     = errors.New("depth limit exceeded")

	ErrTooManyParams      = errors.New("too many params for AST encoding")
	ErrTooManySegments    = errors.New("too many segments for AST encoding")
	ErrTooManyValues      = errors.New("too many values for AST encoding")
	ErrTooManyValueParts  = errors.New("too many value parts for AST encoding")
	ErrSpanTooLarge       = errors.New("span too large for AST encoding")
	ErrInvalidPercentCode = errors.New("invalid percent-encoding")
)
