// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"errors"
	"regexp"
)

// DuplicateHandling specifies how duplicate keys should be handled during parsing.
type DuplicateHandling string

const (
	// DuplicateCombine combines duplicate values into an array (default).
	DuplicateCombine DuplicateHandling = "combine"
	// DuplicateFirst keeps only the first value for duplicate keys.
	DuplicateFirst DuplicateHandling = "first"
	// DuplicateLast keeps only the last value for duplicate keys.
	DuplicateLast DuplicateHandling = "last"
)

// DecoderFunc is a custom decoder function signature.
// Parameters:
//   - str: the string to decode
//   - charset: the charset being used
//   - kind: either "key" or "value" indicating what is being decoded
//
// Returns the decoded string and any error.
type DecoderFunc func(str string, charset Charset, kind string) (string, error)

// ParseOptions configures the behavior of the Parse function.
type ParseOptions struct {
	// AllowDots enables dot notation parsing (e.g., "a.b.c" → {a: {b: {c: ...}}}).
	// Default: false
	AllowDots bool

	// AllowEmptyArrays allows empty arrays when value is empty string.
	// Only applies when the key ends with [].
	// Default: false
	AllowEmptyArrays bool

	// AllowPrototypes allows keys that would overwrite Object prototype properties.
	// In Go this is less relevant but maintained for JS compatibility.
	// Default: false
	AllowPrototypes bool

	// AllowSparse preserves sparse arrays without compacting them.
	// When false, arrays like [a, undefined, b] become [a, b].
	// Default: false
	AllowSparse bool

	// ArrayLimit is the maximum index for array parsing.
	// Indices above this limit will cause the array to be converted to an object.
	// Default: 20
	ArrayLimit int

	// Charset specifies the character encoding to use.
	// Default: CharsetUTF8
	Charset Charset

	// CharsetSentinel enables automatic charset detection via utf8=✓ parameter.
	// Default: false
	CharsetSentinel bool

	// Comma enables parsing comma-separated values as arrays.
	// e.g., "a=1,2,3" → {a: ["1", "2", "3"]}
	// Default: false
	Comma bool

	// DecodeDotInKeys decodes %2E as . in keys.
	// Default: false
	DecodeDotInKeys bool

	// Decoder is a custom function for decoding strings.
	// If nil, the default decoder is used.
	// Default: nil (uses built-in Decode function)
	Decoder DecoderFunc

	// Delimiter is the string or regexp used to split key-value pairs.
	// Default: "&"
	Delimiter string

	// DelimiterRegexp is used instead of Delimiter when regex splitting is needed.
	// If set, Delimiter is ignored.
	// Default: nil
	DelimiterRegexp *regexp.Regexp

	// Depth is the maximum depth for nested object parsing.
	// Set to 0 to disable nested parsing entirely.
	// Default: 5
	Depth int

	// Duplicates specifies how to handle duplicate keys.
	// Default: DuplicateCombine
	Duplicates DuplicateHandling

	// IgnoreQueryPrefix strips a leading ? from the input string.
	// Default: false
	IgnoreQueryPrefix bool

	// InterpretNumericEntities converts HTML numeric entities (&#NNN;) to characters.
	// Only applies when Charset is ISO-8859-1.
	// Default: false
	InterpretNumericEntities bool

	// ParameterLimit is the maximum number of parameters to parse.
	// Parameters beyond this limit are ignored.
	// Default: 1000
	ParameterLimit int

	// ParseArrays enables array parsing (e.g., "a[0]=b" or "a[]=b").
	// When false, brackets are preserved as literal characters in keys.
	// Default: true
	ParseArrays bool

	// PlainObjects creates objects without prototype (Go: less relevant).
	// Maintained for JS compatibility.
	// Default: false
	PlainObjects bool

	// StrictDepth returns an error when input depth exceeds Depth option.
	// When false, excess depth is preserved as a literal key.
	// Default: false
	StrictDepth bool

	// StrictNullHandling treats keys without values as null instead of empty string.
	// e.g., "a" → {a: null} instead of {a: ""}
	// Default: false
	StrictNullHandling bool

	// ThrowOnLimitExceeded returns an error when ParameterLimit or ArrayLimit is exceeded.
	// When false, excess parameters/elements are silently ignored.
	// Default: false
	ThrowOnLimitExceeded bool
}

// Default values for ParseOptions
const (
	DefaultArrayLimit     = 20
	DefaultDepth          = 5
	DefaultParameterLimit = 1000
	DefaultDelimiter      = "&"
)

// DefaultParseOptions returns ParseOptions with default values.
func DefaultParseOptions() ParseOptions {
	return ParseOptions{
		AllowDots:                false,
		AllowEmptyArrays:         false,
		AllowPrototypes:          false,
		AllowSparse:              false,
		ArrayLimit:               DefaultArrayLimit,
		Charset:                  CharsetUTF8,
		CharsetSentinel:          false,
		Comma:                    false,
		DecodeDotInKeys:          false,
		Decoder:                  nil,
		Delimiter:                DefaultDelimiter,
		DelimiterRegexp:          nil,
		Depth:                    DefaultDepth,
		Duplicates:               DuplicateCombine,
		IgnoreQueryPrefix:        false,
		InterpretNumericEntities: false,
		ParameterLimit:           DefaultParameterLimit,
		ParseArrays:              true,
		PlainObjects:             false,
		StrictDepth:              false,
		StrictNullHandling:       false,
		ThrowOnLimitExceeded:     false,
	}
}

// Validation errors
var (
	ErrInvalidAllowEmptyArrays = errors.New("allowEmptyArrays option must be a boolean")
	ErrInvalidDecodeDotInKeys  = errors.New("decodeDotInKeys option must be a boolean")
	ErrInvalidDecoder          = errors.New("decoder must be a function")
	ErrInvalidCharset          = errors.New("charset must be utf-8 or iso-8859-1")
	ErrInvalidDuplicates       = errors.New("duplicates must be combine, first, or last")
	ErrInvalidThrowOnLimit     = errors.New("throwOnLimitExceeded option must be a boolean")
	ErrParameterLimitExceeded  = errors.New("parameter limit exceeded")
	ErrArrayLimitExceeded      = errors.New("array limit exceeded")
	ErrDepthLimitExceeded      = errors.New("depth limit exceeded")
)

// normalizeParseOptions validates and fills in defaults for ParseOptions.
// It returns a new ParseOptions with all fields properly set.
func normalizeParseOptions(opts *ParseOptions) (ParseOptions, error) {
	if opts == nil {
		return DefaultParseOptions(), nil
	}

	result := *opts

	// Validate and set charset
	if result.Charset == "" {
		result.Charset = CharsetUTF8
	} else if result.Charset != CharsetUTF8 && result.Charset != CharsetISO88591 {
		return result, ErrInvalidCharset
	}

	// Validate duplicates
	if result.Duplicates == "" {
		result.Duplicates = DuplicateCombine
	} else if result.Duplicates != DuplicateCombine &&
		result.Duplicates != DuplicateFirst &&
		result.Duplicates != DuplicateLast {
		return result, ErrInvalidDuplicates
	}

	// Set defaults for numeric fields if they are zero
	if result.ArrayLimit == 0 {
		result.ArrayLimit = DefaultArrayLimit
	}
	if result.Depth == 0 {
		result.Depth = DefaultDepth
	}
	if result.ParameterLimit == 0 {
		result.ParameterLimit = DefaultParameterLimit
	}

	// Set default delimiter
	if result.Delimiter == "" && result.DelimiterRegexp == nil {
		result.Delimiter = DefaultDelimiter
	}

	// If DecodeDotInKeys is true, AllowDots should also be true
	if result.DecodeDotInKeys && !result.AllowDots {
		result.AllowDots = true
	}

	return result, nil
}

// ParseOption is a functional option for configuring ParseOptions.
type ParseOption func(*ParseOptions)

// WithAllowDots enables dot notation parsing (e.g., "a.b.c" → {a: {b: {c: ...}}}).
func WithAllowDots(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.AllowDots = v
	}
}

// WithAllowEmptyArrays allows empty arrays when value is empty string.
func WithAllowEmptyArrays(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.AllowEmptyArrays = v
	}
}

// WithAllowPrototypes allows keys that would overwrite Object prototype properties.
func WithAllowPrototypes(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.AllowPrototypes = v
	}
}

// WithAllowSparse preserves sparse arrays without compacting them.
func WithAllowSparse(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.AllowSparse = v
	}
}

// WithArrayLimit sets the maximum index for array parsing.
func WithArrayLimit(v int) ParseOption {
	return func(o *ParseOptions) {
		o.ArrayLimit = v
	}
}

// WithCharset sets the character encoding to use.
func WithCharset(v Charset) ParseOption {
	return func(o *ParseOptions) {
		o.Charset = v
	}
}

// WithCharsetSentinel enables automatic charset detection via utf8=✓ parameter.
func WithCharsetSentinel(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.CharsetSentinel = v
	}
}

// WithComma enables parsing comma-separated values as arrays.
func WithComma(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.Comma = v
	}
}

// WithDecodeDotInKeys decodes %2E as . in keys.
func WithDecodeDotInKeys(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.DecodeDotInKeys = v
		if v {
			o.AllowDots = true
		}
	}
}

// WithDecoder sets a custom decoder function.
func WithDecoder(v DecoderFunc) ParseOption {
	return func(o *ParseOptions) {
		o.Decoder = v
	}
}

// WithDelimiter sets the string used to split key-value pairs.
func WithDelimiter(v string) ParseOption {
	return func(o *ParseOptions) {
		o.Delimiter = v
		o.DelimiterRegexp = nil
	}
}

// WithDelimiterRegexp sets a regexp used to split key-value pairs.
func WithDelimiterRegexp(v *regexp.Regexp) ParseOption {
	return func(o *ParseOptions) {
		o.DelimiterRegexp = v
		o.Delimiter = ""
	}
}

// WithDepth sets the maximum depth for nested object parsing.
func WithDepth(v int) ParseOption {
	return func(o *ParseOptions) {
		o.Depth = v
	}
}

// WithDuplicates sets how to handle duplicate keys.
func WithDuplicates(v DuplicateHandling) ParseOption {
	return func(o *ParseOptions) {
		o.Duplicates = v
	}
}

// WithIgnoreQueryPrefix strips a leading ? from the input string.
func WithIgnoreQueryPrefix(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.IgnoreQueryPrefix = v
	}
}

// WithInterpretNumericEntities converts HTML numeric entities to characters.
func WithInterpretNumericEntities(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.InterpretNumericEntities = v
	}
}

// WithParameterLimit sets the maximum number of parameters to parse.
func WithParameterLimit(v int) ParseOption {
	return func(o *ParseOptions) {
		o.ParameterLimit = v
	}
}

// WithParseArrays enables or disables array parsing.
func WithParseArrays(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.ParseArrays = v
	}
}

// WithPlainObjects creates objects without prototype.
func WithPlainObjects(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.PlainObjects = v
	}
}

// WithStrictDepth returns an error when input depth exceeds Depth option.
func WithStrictDepth(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.StrictDepth = v
	}
}

// WithStrictNullHandling treats keys without values as null instead of empty string.
func WithStrictNullHandling(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.StrictNullHandling = v
	}
}

// WithThrowOnLimitExceeded returns an error when limits are exceeded.
func WithThrowOnLimitExceeded(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.ThrowOnLimitExceeded = v
	}
}

// applyParseOptions applies functional options to a ParseOptions struct.
func applyParseOptions(opts ...ParseOption) ParseOptions {
	o := DefaultParseOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
