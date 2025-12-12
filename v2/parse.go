// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
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

// Sentinel values indicating "not explicitly set" (used to distinguish from explicit 0)
// Using very negative values to avoid conflict with valid negative values (e.g., arrayLimit: -1 disables arrays)
const (
	notSetArrayLimit     = -999999
	notSetDepth          = -999999
	notSetParameterLimit = -999999
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

// Charset sentinel values for auto-detection
const (
	// isoSentinel is what browsers submit when ✓ appears in iso-8859-1 encoded form
	isoSentinel = "utf8=%26%2310003%3B" // encodeURIComponent('&#10003;')
	// charsetSentinel is the percent-encoded utf-8 octets for ✓
	charsetSentinel = "utf8=%E2%9C%93" // encodeURIComponent('✓')
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

	// Set defaults for numeric fields if they are not explicitly set (sentinel value)
	// This allows explicit 0 values to be preserved
	if result.ArrayLimit == notSetArrayLimit {
		result.ArrayLimit = DefaultArrayLimit
	}
	if result.Depth == notSetDepth {
		result.Depth = DefaultDepth
	}
	if result.ParameterLimit == notSetParameterLimit {
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
	// Start with defaults but use sentinel values for numeric fields
	// so we can distinguish "not set" from "explicitly set to 0"
	o := DefaultParseOptions()
	o.ArrayLimit = notSetArrayLimit
	o.Depth = notSetDepth
	o.ParameterLimit = notSetParameterLimit
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// interpretNumericEntities converts HTML numeric entities (&#NNN;) to characters.
// e.g., "&#9786;" → "☺"
func interpretNumericEntitiesFunc(str string) string {
	re := regexp.MustCompile(`&#(\d+);`)
	return re.ReplaceAllStringFunc(str, func(match string) string {
		// Extract the number between &# and ;
		numStr := match[2 : len(match)-1]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return match
		}
		return string(rune(num))
	})
}

// parseArrayValue handles comma-separated values and array limit checking.
// Returns the value as-is, split by comma, or error if limit exceeded.
func parseArrayValue(val string, opts *ParseOptions, currentArrayLength int) (any, error) {
	if val != "" && opts.Comma && strings.Contains(val, ",") {
		parts := strings.Split(val, ",")
		// Convert []string to []any for consistent type handling
		result := make([]any, len(parts))
		for i, p := range parts {
			result[i] = p
		}
		return result, nil
	}

	if opts.ThrowOnLimitExceeded && currentArrayLength >= opts.ArrayLimit {
		return nil, ErrArrayLimitExceeded
	}

	return val, nil
}

// splitByDelimiter splits a string by either a string delimiter or regexp.
// Unlike Go's SplitN (which keeps remainder in last element), this matches JS behavior:
// JS "a&b&c".split("&", 2) returns ["a", "b"] - exactly limit parts, remainder discarded.
func splitByDelimiter(str string, delimiter string, delimiterRegexp *regexp.Regexp, limit int) []string {
	if limit <= 0 {
		// No limit - split all
		if delimiterRegexp != nil {
			return delimiterRegexp.Split(str, -1)
		}
		return strings.Split(str, delimiter)
	}

	// Split with limit+1 to detect if there are more parts
	var parts []string
	if delimiterRegexp != nil {
		parts = delimiterRegexp.Split(str, limit+1)
	} else {
		parts = strings.SplitN(str, delimiter, limit+1)
	}

	// Truncate to limit (discard remainder like JS)
	if len(parts) > limit {
		parts = parts[:limit]
	}

	return parts
}

// prototypeKeys are JavaScript Object prototype properties that should be blocked.
var prototypeKeys = map[string]bool{
	"__proto__":   true,
	"constructor": true,
	"prototype":   true,
	// Common Object.prototype methods
	"toString":             true,
	"toLocaleString":       true,
	"valueOf":              true,
	"hasOwnProperty":       true,
	"isPrototypeOf":        true,
	"propertyIsEnumerable": true,
}

// isPrototypeKey checks if a key is a JavaScript prototype property.
func isPrototypeProp(key string) bool {
	return prototypeKeys[key]
}

// parseObject builds a nested object structure from a chain of keys.
// chain is like ["a", "[b]", "[c]"] and val is the leaf value.
// It builds from the leaf up: {c: val} -> {b: {c: val}} -> {a: {b: {c: val}}}
func parseObject(chain []string, val any, opts *ParseOptions, valuesParsed bool) any {
	if len(chain) == 0 {
		return val
	}

	leaf := val

	// Build from the end of chain backwards
	for i := len(chain) - 1; i >= 0; i-- {
		var obj any
		root := chain[i]

		if root == "[]" && opts.ParseArrays {
			// Empty brackets means array
			if opts.AllowEmptyArrays && (leaf == "" || (opts.StrictNullHandling && leaf == nil)) {
				obj = []any{}
			} else {
				obj = Combine([]any{}, leaf)
			}
		} else {
			// Object or indexed array
			objMap := make(map[string]any)

			// Clean the root - remove surrounding brackets if present
			cleanRoot := root
			if len(root) >= 2 && root[0] == '[' && root[len(root)-1] == ']' {
				cleanRoot = root[1 : len(root)-1]
			}

			// Decode dots in keys if enabled
			decodedRoot := cleanRoot
			if opts.DecodeDotInKeys {
				decodedRoot = strings.ReplaceAll(cleanRoot, "%2E", ".")
				decodedRoot = strings.ReplaceAll(decodedRoot, "%2e", ".")
			}

			// Try to parse as array index
			index, err := strconv.Atoi(decodedRoot)
			isValidIndex := err == nil && index >= 0 && strconv.Itoa(index) == decodedRoot && root != decodedRoot

			if !opts.ParseArrays && decodedRoot == "" {
				// When parseArrays is false and key is empty, use "0"
				obj = map[string]any{"0": leaf}
			} else if isValidIndex && opts.ParseArrays && index <= opts.ArrayLimit {
				// Create array with value at index
				arr := make([]any, index+1)
				arr[index] = leaf
				obj = arr
			} else if decodedRoot != "__proto__" {
				// Regular object key
				objMap[decodedRoot] = leaf
				obj = objMap
			} else if opts.AllowPrototypes {
				objMap[decodedRoot] = leaf
				obj = objMap
			} else {
				// Skip __proto__ key
				obj = objMap
			}
		}

		leaf = obj
	}

	return leaf
}

// parseKeys parses a key like "a[b][c]" into nested structure with value.
// It handles bracket notation, dot notation, depth limits, and prototype protection.
func parseKeys(givenKey string, val any, opts *ParseOptions, valuesParsed bool) (any, error) {
	if givenKey == "" {
		return nil, nil
	}

	// Transform dot notation to bracket notation if allowDots is enabled
	key := givenKey
	if opts.AllowDots {
		// Replace .foo with [foo], but not dots inside brackets
		// Pattern: \.([^.[]+) -> [$1]
		re := regexp.MustCompile(`\.([^.\[]+)`)
		key = re.ReplaceAllString(key, "[$1]")
	}

	// Regex to find bracket segments
	brackets := regexp.MustCompile(`(\[[^\[\]]*\])`)

	// Find first bracket segment
	segment := brackets.FindStringIndex(key)

	var parent string
	if opts.Depth > 0 && segment != nil {
		parent = key[:segment[0]]
	} else {
		parent = key
	}

	// Build the keys chain
	var keys []string

	if parent != "" {
		// Check prototype pollution for parent key
		if !opts.PlainObjects && isPrototypeProp(parent) {
			if !opts.AllowPrototypes {
				return nil, nil
			}
		}
		keys = append(keys, parent)
	}

	// Loop through bracket segments up to depth limit
	i := 0
	child := regexp.MustCompile(`(\[[^\[\]]*\])`)
	remaining := ""
	if segment != nil && opts.Depth > 0 {
		remaining = key[segment[0]:]
	}

	for opts.Depth > 0 && i < opts.Depth {
		match := child.FindStringIndex(remaining)
		if match == nil {
			break
		}

		seg := remaining[match[0]:match[1]]

		// Check prototype pollution for this segment
		innerKey := seg
		if len(seg) >= 2 {
			innerKey = seg[1 : len(seg)-1] // Remove brackets
		}
		if !opts.PlainObjects && isPrototypeProp(innerKey) {
			if !opts.AllowPrototypes {
				return nil, nil
			}
		}

		keys = append(keys, seg)
		remaining = remaining[match[1]:]
		i++
	}

	// If there's remaining content after depth limit
	if remaining != "" {
		if opts.StrictDepth {
			return nil, ErrDepthLimitExceeded
		}
		// Wrap remainder in extra brackets so it becomes a literal key
		// e.g., "[g]" becomes "[[g]]", which parseObject strips outer brackets
		// leaving "[g]" as the actual key name
		keys = append(keys, "["+remaining+"]")
	}

	return parseObject(keys, val, opts, valuesParsed), nil
}

// orderedResult holds parsed values with insertion order preserved.
// This is needed because Go map iteration is randomized, but JS Object.keys()
// returns keys in insertion order, which affects merge behavior.
type orderedResult struct {
	keys   []string       // keys in insertion order
	values map[string]any // key-value pairs
}

// parseValues parses a query string into a flat map of key-value pairs.
// This is the first stage of parsing before nested object reconstruction.
func parseValues(str string, opts *ParseOptions) (orderedResult, error) {
	result := orderedResult{
		keys:   make([]string, 0),
		values: make(map[string]any),
	}

	// Strip query prefix if requested
	cleanStr := str
	if opts.IgnoreQueryPrefix && len(cleanStr) > 0 && cleanStr[0] == '?' {
		cleanStr = cleanStr[1:]
	}

	// Decode URL-encoded brackets for easier parsing
	cleanStr = strings.ReplaceAll(cleanStr, "%5B", "[")
	cleanStr = strings.ReplaceAll(cleanStr, "%5b", "[")
	cleanStr = strings.ReplaceAll(cleanStr, "%5D", "]")
	cleanStr = strings.ReplaceAll(cleanStr, "%5d", "]")

	// Calculate limit for splitting
	limit := opts.ParameterLimit
	if opts.ThrowOnLimitExceeded {
		limit = opts.ParameterLimit + 1
	}

	// Split by delimiter
	parts := splitByDelimiter(cleanStr, opts.Delimiter, opts.DelimiterRegexp, limit)

	// Check parameter limit
	if opts.ThrowOnLimitExceeded && len(parts) > opts.ParameterLimit {
		return orderedResult{}, ErrParameterLimitExceeded
	}

	// Detect charset from sentinel
	charset := opts.Charset
	skipIndex := -1
	if opts.CharsetSentinel {
		for i, part := range parts {
			if strings.HasPrefix(part, "utf8=") {
				if part == charsetSentinel {
					charset = CharsetUTF8
				} else if part == isoSentinel {
					charset = CharsetISO88591
				}
				skipIndex = i
				break
			}
		}
	}

	// Default decoder function
	defaultDecoder := func(s string, cs Charset, kind string) (string, error) {
		return Decode(s, cs), nil
	}

	decoder := opts.Decoder
	if decoder == nil {
		decoder = defaultDecoder
	}

	// Parse each part
	for i, part := range parts {
		if i == skipIndex {
			continue
		}

		if part == "" {
			continue
		}

		// Find the = sign, handling bracket notation like "a[]=b"
		bracketEqualsPos := strings.Index(part, "]=")
		var pos int
		if bracketEqualsPos == -1 {
			pos = strings.Index(part, "=")
		} else {
			pos = bracketEqualsPos + 1
		}

		var key string
		var val any

		if pos == -1 {
			// No = sign, key only
			decoded, err := decoder(part, charset, "key")
			if err != nil {
				return orderedResult{}, err
			}
			key = decoded
			if opts.StrictNullHandling {
				val = nil
			} else {
				val = ""
			}
		} else {
			// Has = sign
			keyPart := part[:pos]
			valPart := part[pos+1:]

			decoded, err := decoder(keyPart, charset, "key")
			if err != nil {
				return orderedResult{}, err
			}
			key = decoded

			if key == "" {
				continue
			}

			// Get current array length for limit checking
			currentLen := 0
			if existing, ok := result.values[key]; ok {
				if arr, isArr := existing.([]any); isArr {
					currentLen = len(arr)
				}
			}

			// Handle comma-separated values and array limit
			parsedVal, err := parseArrayValue(valPart, opts, currentLen)
			if err != nil {
				return orderedResult{}, err
			}

			// Decode the value(s)
			val = MaybeMap(parsedVal, func(v any) any {
				if s, ok := v.(string); ok {
					decoded, _ := decoder(s, charset, "value")
					return decoded
				}
				return v
			})
		}

		// Interpret numeric entities if enabled
		if val != nil && opts.InterpretNumericEntities && charset == CharsetISO88591 {
			if s, ok := val.(string); ok {
				val = interpretNumericEntitiesFunc(s)
			} else if arr, ok := val.([]any); ok {
				for i, v := range arr {
					if s, ok := v.(string); ok {
						arr[i] = interpretNumericEntitiesFunc(s)
					}
				}
			}
		}

		// Handle []= (empty bracket) notation - wrap in array
		if strings.Contains(part, "[]=") {
			if arr, ok := val.([]any); ok {
				val = []any{arr}
			}
		}

		// Handle duplicate keys
		if key != "" {
			if existing, exists := result.values[key]; exists {
				switch opts.Duplicates {
				case DuplicateCombine:
					result.values[key] = Combine(existing, val)
				case DuplicateFirst:
					// Keep existing, do nothing
				case DuplicateLast:
					result.values[key] = val
				default:
					result.values[key] = Combine(existing, val)
				}
			} else {
				// New key - track insertion order
				result.keys = append(result.keys, key)
				result.values[key] = val
			}
		}
	}

	return result, nil
}

// Parse parses a URL query string into a map.
// It supports nested objects, arrays, and various encoding options.
//
// Example:
//
//	result, err := qs.Parse("a=b&c=d")
//	// result = map[string]any{"a": "b", "c": "d"}
//
//	result, err := qs.Parse("a[b]=c")
//	// result = map[string]any{"a": map[string]any{"b": "c"}}
//
//	result, err := qs.Parse("a.b.c=d", qs.WithAllowDots(true))
//	// result = map[string]any{"a": map[string]any{"b": map[string]any{"c": "d"}}}
func Parse(str string, opts ...ParseOption) (map[string]any, error) {
	options := applyParseOptions(opts...)

	// Normalize options
	normalizedOpts, err := normalizeParseOptions(&options)
	if err != nil {
		return nil, err
	}

	// Handle empty input
	if str == "" {
		return make(map[string]any), nil
	}

	// Parse the query string into flat key-value pairs
	tempObj, err := parseValues(str, &normalizedOpts)
	if err != nil {
		return nil, err
	}

	// Build nested structure using parseKeys
	// Iterate in insertion order (like JS Object.keys())
	result := make(map[string]any)

	for _, key := range tempObj.keys {
		val := tempObj.values[key]
		// Parse nested keys and build object structure
		newObj, err := parseKeys(key, val, &normalizedOpts, true)
		if err != nil {
			return nil, err
		}

		if newObj != nil {
			// Merge into result
			merged := Merge(result, newObj, normalizedOpts.AllowPrototypes)
			if m, ok := merged.(map[string]any); ok {
				result = m
			}
		}
	}

	// Compact sparse arrays if AllowSparse is false
	if !normalizedOpts.AllowSparse {
		compacted := Compact(result)
		if m, ok := compacted.(map[string]any); ok {
			return m, nil
		}
	}

	return result, nil
}
