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

	// Strict mode errors
	ErrUnclosedBracket        = errors.New("unclosed bracket")
	ErrUnmatchedCloseBracket  = errors.New("unmatched closing bracket")
	ErrEmptyKey               = errors.New("empty key")
	ErrInvalidPercentEncoding = errors.New("invalid percent-encoding sequence")
	ErrEmptyBracketContent    = errors.New("empty bracket content in non-array context")
	ErrConsecutiveDots        = errors.New("consecutive dots in key")
	ErrLeadingDot             = errors.New("leading dot in key")
	ErrTrailingDot            = errors.New("trailing dot in key")
)
