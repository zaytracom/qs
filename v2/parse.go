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

// WithAllowDots returns a copy of opts with AllowDots set.
func (opts ParseOptions) WithAllowDots(v bool) ParseOptions {
	opts.AllowDots = v
	return opts
}

// WithArrayLimit returns a copy of opts with ArrayLimit set.
func (opts ParseOptions) WithArrayLimit(v int) ParseOptions {
	opts.ArrayLimit = v
	return opts
}

// WithCharset returns a copy of opts with Charset set.
func (opts ParseOptions) WithCharset(v Charset) ParseOptions {
	opts.Charset = v
	return opts
}

// WithComma returns a copy of opts with Comma set.
func (opts ParseOptions) WithComma(v bool) ParseOptions {
	opts.Comma = v
	return opts
}

// WithDelimiter returns a copy of opts with Delimiter set.
func (opts ParseOptions) WithDelimiter(v string) ParseOptions {
	opts.Delimiter = v
	return opts
}

// WithDepth returns a copy of opts with Depth set.
func (opts ParseOptions) WithDepth(v int) ParseOptions {
	opts.Depth = v
	return opts
}

// WithDuplicates returns a copy of opts with Duplicates set.
func (opts ParseOptions) WithDuplicates(v DuplicateHandling) ParseOptions {
	opts.Duplicates = v
	return opts
}

// WithIgnoreQueryPrefix returns a copy of opts with IgnoreQueryPrefix set.
func (opts ParseOptions) WithIgnoreQueryPrefix(v bool) ParseOptions {
	opts.IgnoreQueryPrefix = v
	return opts
}

// WithParameterLimit returns a copy of opts with ParameterLimit set.
func (opts ParseOptions) WithParameterLimit(v int) ParseOptions {
	opts.ParameterLimit = v
	return opts
}

// WithParseArrays returns a copy of opts with ParseArrays set.
func (opts ParseOptions) WithParseArrays(v bool) ParseOptions {
	opts.ParseArrays = v
	return opts
}

// WithStrictNullHandling returns a copy of opts with StrictNullHandling set.
func (opts ParseOptions) WithStrictNullHandling(v bool) ParseOptions {
	opts.StrictNullHandling = v
	return opts
}
