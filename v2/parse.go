// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/zaytracom/qs/v2/lang"
)

// Pre-compiled regular expressions for performance.
var (
	numericEntityRe = regexp.MustCompile(`&#(\d+);`)
	dotNotationRe   = regexp.MustCompile(`\.([^.\[]+)`)
	bracketRe       = regexp.MustCompile(`(\[[^\[\]]*\])`)
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

	// StrictMode enables strict syntax validation.
	// When true, returns errors for:
	// - Unclosed or unmatched brackets
	// - Empty keys
	// - Invalid percent-encoding sequences
	// - Leading/trailing/consecutive dots (when AllowDots is true)
	// Default: false
	StrictMode bool
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

// Strict mode errors (re-exported from lang package)
var (
	ErrUnclosedBracket        = lang.ErrUnclosedBracket
	ErrUnmatchedCloseBracket  = lang.ErrUnmatchedCloseBracket
	ErrEmptyKey               = lang.ErrEmptyKey
	ErrInvalidPercentEncoding = lang.ErrInvalidPercentEncoding
	ErrConsecutiveDots        = lang.ErrConsecutiveDots
	ErrLeadingDot             = lang.ErrLeadingDot
	ErrTrailingDot            = lang.ErrTrailingDot
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

// WithParseAllowDots enables dot notation parsing (e.g., "a.b.c" → {a: {b: {c: ...}}}).
func WithParseAllowDots(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.AllowDots = v
	}
}

// WithParseAllowEmptyArrays allows empty arrays when value is empty string.
func WithParseAllowEmptyArrays(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.AllowEmptyArrays = v
	}
}

// WithParseAllowSparse preserves sparse arrays without compacting them.
func WithParseAllowSparse(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.AllowSparse = v
	}
}

// WithParseArrayLimit sets the maximum index for array parsing.
func WithParseArrayLimit(v int) ParseOption {
	return func(o *ParseOptions) {
		o.ArrayLimit = v
	}
}

// WithParseCharset sets the character encoding to use.
func WithParseCharset(v Charset) ParseOption {
	return func(o *ParseOptions) {
		o.Charset = v
	}
}

// WithParseCharsetSentinel enables automatic charset detection via utf8=✓ parameter.
func WithParseCharsetSentinel(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.CharsetSentinel = v
	}
}

// WithParseComma enables parsing comma-separated values as arrays.
func WithParseComma(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.Comma = v
	}
}

// WithParseDecodeDotInKeys decodes %2E as . in keys.
func WithParseDecodeDotInKeys(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.DecodeDotInKeys = v
		if v {
			o.AllowDots = true
		}
	}
}

// WithParseDecoder sets a custom decoder function.
func WithParseDecoder(v DecoderFunc) ParseOption {
	return func(o *ParseOptions) {
		o.Decoder = v
	}
}

// WithParseDelimiter sets the string used to split key-value pairs.
func WithParseDelimiter(v string) ParseOption {
	return func(o *ParseOptions) {
		o.Delimiter = v
		o.DelimiterRegexp = nil
	}
}

// WithParseDelimiterRegexp sets a regexp used to split key-value pairs.
func WithParseDelimiterRegexp(v *regexp.Regexp) ParseOption {
	return func(o *ParseOptions) {
		o.DelimiterRegexp = v
		o.Delimiter = ""
	}
}

// WithParseDepth sets the maximum depth for nested object parsing.
func WithParseDepth(v int) ParseOption {
	return func(o *ParseOptions) {
		o.Depth = v
	}
}

// WithParseDuplicates sets how to handle duplicate keys.
func WithParseDuplicates(v DuplicateHandling) ParseOption {
	return func(o *ParseOptions) {
		o.Duplicates = v
	}
}

// WithParseIgnoreQueryPrefix strips a leading ? from the input string.
func WithParseIgnoreQueryPrefix(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.IgnoreQueryPrefix = v
	}
}

// WithParseInterpretNumericEntities converts HTML numeric entities to characters.
func WithParseInterpretNumericEntities(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.InterpretNumericEntities = v
	}
}

// WithParseParameterLimit sets the maximum number of parameters to parse.
func WithParseParameterLimit(v int) ParseOption {
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

// WithParseStrictDepth returns an error when input depth exceeds Depth option.
func WithParseStrictDepth(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.StrictDepth = v
	}
}

// WithParseStrictNullHandling treats keys without values as null instead of empty string.
func WithParseStrictNullHandling(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.StrictNullHandling = v
	}
}

// WithParseThrowOnLimitExceeded returns an error when limits are exceeded.
func WithParseThrowOnLimitExceeded(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.ThrowOnLimitExceeded = v
	}
}

// WithParseStrictMode enables strict syntax validation.
// When true, returns errors for unclosed brackets, empty keys,
// invalid percent-encoding, and dot notation issues.
func WithParseStrictMode(v bool) ParseOption {
	return func(o *ParseOptions) {
		o.StrictMode = v
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
	return numericEntityRe.ReplaceAllStringFunc(str, func(match string) string {
		// Extract the number between &# and ;
		numStr := match[2 : len(match)-1]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return match
		}
		return string(rune(num))
	})
}

// decodeBrackets decodes URL-encoded brackets (%5B/%5D) in a single pass.
func decodeBrackets(s string) string {
	if !strings.Contains(s, "%5") {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		if i+2 < len(s) && s[i] == '%' && s[i+1] == '5' {
			switch s[i+2] {
			case 'B', 'b':
				b.WriteByte('[')
				i += 3
				continue
			case 'D', 'd':
				b.WriteByte(']')
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
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
			if opts.AllowEmptyArrays && (leaf == "" || IsExplicitNull(leaf)) {
				obj = []any{}
			} else if IsExplicitNull(leaf) {
				// With strictNullHandling, null is meaningful - include it explicitly
				obj = []any{ExplicitNullValue}
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
			} else {
				// Regular object key
				objMap[decodedRoot] = leaf
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
		key = dotNotationRe.ReplaceAllString(key, "[$1]")
	}

	// Find first bracket segment
	segment := bracketRe.FindStringIndex(key)

	var parent string
	if opts.Depth > 0 && segment != nil {
		parent = key[:segment[0]]
	} else {
		parent = key
	}

	// Build the keys chain
	var keys []string

	if parent != "" {
		keys = append(keys, parent)
	}

	// Loop through bracket segments up to depth limit
	i := 0
	remaining := ""
	if segment != nil && opts.Depth > 0 {
		remaining = key[segment[0]:]
	}

	for opts.Depth > 0 && i < opts.Depth {
		match := bracketRe.FindStringIndex(remaining)
		if match == nil {
			break
		}

		seg := remaining[match[0]:match[1]]
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

	// For regexp delimiter or multi-char string delimiter, fall back to split-based parsing
	if normalizedOpts.DelimiterRegexp != nil || len(normalizedOpts.Delimiter) > 1 {
		return parseWithRegexpDelimiter(str, &normalizedOpts)
	}

	// Build AST config from parse options
	cfg := buildLangConfig(&normalizedOpts)

	// Estimate arena size
	arena := lang.NewArena(estimateParams(str))

	// Parse directly with AST parser
	qs, detectedCharset, err := lang.Parse(arena, str, cfg)
	if err != nil {
		if err == lang.ErrParameterLimitExceeded {
			return nil, ErrParameterLimitExceeded
		}
		if err == lang.ErrDepthLimitExceeded {
			return nil, ErrDepthLimitExceeded
		}
		return nil, err
	}

	// Use detected charset (from sentinel) if charset sentinel is enabled
	charset := normalizedOpts.Charset
	if normalizedOpts.CharsetSentinel {
		charset = charsetFromLang(detectedCharset)
	}

	// Accumulate values by raw key, storing chain only once per unique key
	type accumulated struct {
		chain []string
		val   any
	}
	keyOrder := make([]string, 0, qs.ParamLen)
	keyData := make(map[string]*accumulated, qs.ParamLen)

	for i := uint16(0); i < qs.ParamLen; i++ {
		param := arena.Params[i]
		rawKey := arena.GetString(param.Key.Raw)

		if existing, exists := keyData[rawKey]; exists {
			// Key already seen - just accumulate value
			val, err := extractValue(arena, param, charset, &normalizedOpts)
			if err != nil {
				return nil, err
			}

			switch normalizedOpts.Duplicates {
			case DuplicateFirst:
				// Keep existing
			case DuplicateLast:
				existing.val = val
			default:
				if normalizedOpts.ThrowOnLimitExceeded {
					if arr, isArr := existing.val.([]any); isArr && len(arr) >= normalizedOpts.ArrayLimit {
						return nil, ErrArrayLimitExceeded
					}
				}
				existing.val = Combine(existing.val, val)
			}
		} else {
			// First occurrence - build full key info
			info, err := buildKeyInfo(arena, param, charset, &normalizedOpts)
			if err != nil {
				return nil, err
			}
			if info == nil {
				continue
			}
			keyOrder = append(keyOrder, rawKey)
			keyData[rawKey] = &accumulated{chain: info.chain, val: info.val}
		}
	}

	// Build nested structure from accumulated values
	result := make(map[string]any)
	for _, rawKey := range keyOrder {
		data := keyData[rawKey]
		newObj := parseObject(data.chain, data.val, &normalizedOpts, true)
		if newObj != nil {
			merged := Merge(result, newObj)
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
	} else {
		convertExplicitNulls(result)
	}

	return result, nil
}

// buildLangConfig converts ParseOptions to lang.Config.
func buildLangConfig(opts *ParseOptions) lang.Config {
	cfg := lang.DefaultConfig()

	// Delimiter (single byte only, regexp handled separately)
	if opts.Delimiter != "" && len(opts.Delimiter) == 1 {
		cfg.Delimiter = opts.Delimiter[0]
	}

	// Depth and limits
	if opts.Depth >= 0 {
		cfg.Depth = uint16(opts.Depth)
	}
	if opts.ArrayLimit >= 0 {
		cfg.ArrayLimit = uint16(opts.ArrayLimit)
	}
	if opts.ParameterLimit >= 0 {
		cfg.ParameterLimit = uint16(opts.ParameterLimit)
	}

	cfg.ParseArrays = opts.ParseArrays
	cfg.Charset = charsetToLang(opts.Charset)

	// Flags
	if opts.AllowDots {
		cfg.Flags |= lang.FlagAllowDots
	}
	if opts.Comma {
		cfg.Flags |= lang.FlagComma
	}
	if opts.StrictNullHandling {
		cfg.Flags |= lang.FlagStrictNullHandling
	}
	if opts.StrictDepth {
		cfg.Flags |= lang.FlagStrictDepth
	}
	if opts.IgnoreQueryPrefix {
		cfg.Flags |= lang.FlagIgnoreQueryPrefix
	}
	if opts.CharsetSentinel {
		cfg.Flags |= lang.FlagCharsetSentinel
	}
	if opts.DecodeDotInKeys {
		cfg.Flags |= lang.FlagDecodeDotInKeys
	}
	if opts.ThrowOnLimitExceeded {
		cfg.Flags |= lang.FlagThrowOnLimitExceeded
	}
	if opts.StrictMode {
		cfg.Flags |= lang.FlagStrictMode
	}

	return cfg
}

// keyInfoResult holds parsed key chain and value for accumulation.
type keyInfoResult struct {
	chain []string
	val   any
}

// getDecoder returns the decoder function from options or default.
func getDecoder(opts *ParseOptions) DecoderFunc {
	if opts.Decoder != nil {
		return opts.Decoder
	}
	return func(s string, cs Charset, kind string) (string, error) {
		return Decode(s, cs), nil
	}
}

// extractValue extracts only the value from a param (no chain building).
func extractValue(arena *lang.Arena, param lang.Param, charset Charset, opts *ParseOptions) (any, error) {
	decoder := getDecoder(opts)
	key := param.Key

	var val any
	if !param.HasEquals {
		if opts.StrictNullHandling {
			val = ExplicitNullValue
		} else {
			val = ""
		}
	} else if param.ValueIdx != 0xFFFF {
		v := arena.Values[param.ValueIdx]
		switch v.Kind {
		case lang.ValNull:
			if opts.StrictNullHandling {
				val = ExplicitNullValue
			} else {
				val = ""
			}
		case lang.ValComma:
			parts := make([]any, v.PartsLen)
			for j := uint8(0); j < v.PartsLen; j++ {
				partSpan := arena.ValueParts[int(v.PartsOff)+int(j)]
				decoded, err := decoder(arena.GetString(partSpan), charset, "value")
				if err != nil {
					return nil, err
				}
				parts[j] = decoded
			}
			val = parts
		default:
			decoded, err := decoder(arena.GetString(v.Raw), charset, "value")
			if err != nil {
				return nil, err
			}
			val = decoded
		}
	} else {
		val = ""
	}

	if opts.InterpretNumericEntities && charset == CharsetISO88591 {
		val = applyNumericEntities(val)
	}

	// Wrap comma-split array if key ends with []
	if key.SegLen > 0 {
		lastSeg := arena.Segments[int(key.SegStart)+int(key.SegLen)-1]
		if lastSeg.Kind == lang.SegEmpty {
			if arr, ok := val.([]any); ok {
				val = []any{arr}
			}
		}
	}

	return val, nil
}

// applyNumericEntities interprets numeric entities in value.
func applyNumericEntities(val any) any {
	if s, ok := val.(string); ok {
		return interpretNumericEntitiesFunc(s)
	}
	if arr, ok := val.([]any); ok {
		for i, v := range arr {
			if s, ok := v.(string); ok {
				arr[i] = interpretNumericEntitiesFunc(s)
			}
		}
	}
	return val
}

// buildKeyInfo extracts key chain and value from AST param.
func buildKeyInfo(arena *lang.Arena, param lang.Param, charset Charset, opts *ParseOptions) (*keyInfoResult, error) {
	key := param.Key
	if key.SegLen == 0 {
		return nil, nil
	}

	decoder := getDecoder(opts)

	// Get the value
	val, err := extractValue(arena, param, charset, opts)
	if err != nil {
		return nil, err
	}

	// Build chain of keys from segments
	chain := make([]string, 0, key.SegLen)
	for j := uint8(0); j < key.SegLen; j++ {
		seg := arena.Segments[int(key.SegStart)+int(j)]
		decoded, err := decoder(arena.GetString(seg.Span), charset, "key")
		if err != nil {
			return nil, err
		}

		switch seg.Kind {
		case lang.SegEmpty:
			chain = append(chain, "[]")
		case lang.SegLiteral:
			chain = append(chain, "["+decoded+"]")
		default: // SegIdent, SegIndex
			if seg.Notation == lang.NotationRoot {
				chain = append(chain, decoded)
			} else {
				chain = append(chain, "["+decoded+"]")
			}
		}
	}

	return &keyInfoResult{chain: chain, val: val}, nil
}

// parseWithRegexpDelimiter handles parsing when a regexp or multi-char delimiter is used.
// This falls back to the split-based approach since lang.Parse only supports single-byte delimiters.
func parseWithRegexpDelimiter(str string, opts *ParseOptions) (map[string]any, error) {
	// Strip query prefix if requested
	cleanStr := str
	if opts.IgnoreQueryPrefix && len(cleanStr) > 0 && cleanStr[0] == '?' {
		cleanStr = cleanStr[1:]
	}

	// Calculate limit for splitting
	limit := opts.ParameterLimit
	if opts.ThrowOnLimitExceeded {
		limit = opts.ParameterLimit + 1
	}

	// Split by regexp delimiter
	parts := splitByDelimiter(cleanStr, opts.Delimiter, opts.DelimiterRegexp, limit)

	// Check parameter limit
	if opts.ThrowOnLimitExceeded && len(parts) > opts.ParameterLimit {
		return nil, ErrParameterLimitExceeded
	}

	// Detect charset from sentinel
	charset := opts.Charset
	skipIndex := -1
	if opts.CharsetSentinel {
		for i, part := range parts {
			if strings.HasPrefix(part, "utf8=") {
				if part == charsetSentinel {
					charset = CharsetUTF8
					skipIndex = i
				} else if part == isoSentinel {
					charset = CharsetISO88591
					skipIndex = i
				}
				break
			}
		}
	}

	// Setup decoder
	decoder := opts.Decoder
	if decoder == nil {
		decoder = func(s string, cs Charset, kind string) (string, error) {
			return Decode(s, cs), nil
		}
	}

	// Parse each part
	result := make(map[string]any)
	for i, part := range parts {
		if i == skipIndex || part == "" {
			continue
		}

		// Find the = separator (respecting brackets)
		eqIdx := findEqualsOutsideBrackets(part)

		var key, val string
		hasEquals := false
		if eqIdx >= 0 {
			key = part[:eqIdx]
			val = part[eqIdx+1:]
			hasEquals = true
		} else {
			key = part
		}

		// Decode brackets in key
		key = decodeBrackets(key)

		// Decode key
		decodedKey, err := decoder(key, charset, "key")
		if err != nil {
			return nil, err
		}
		if decodedKey == "" {
			continue
		}

		// Handle value
		var parsedVal any
		if !hasEquals {
			if opts.StrictNullHandling {
				parsedVal = ExplicitNullValue
			} else {
				parsedVal = ""
			}
		} else {
			// Handle comma values
			if val != "" && opts.Comma && strings.Contains(val, ",") {
				valParts := strings.Split(val, ",")
				arr := make([]any, len(valParts))
				for j, p := range valParts {
					decoded, err := decoder(p, charset, "value")
					if err != nil {
						return nil, err
					}
					arr[j] = decoded
				}
				parsedVal = arr
			} else {
				decoded, err := decoder(val, charset, "value")
				if err != nil {
					return nil, err
				}
				parsedVal = decoded
			}

			// Interpret numeric entities if enabled
			if opts.InterpretNumericEntities && charset == CharsetISO88591 {
				if s, ok := parsedVal.(string); ok {
					parsedVal = interpretNumericEntitiesFunc(s)
				} else if arr, ok := parsedVal.([]any); ok {
					for j, v := range arr {
						if s, ok := v.(string); ok {
							arr[j] = interpretNumericEntitiesFunc(s)
						}
					}
				}
			}

			// Handle []= pattern
			if strings.Contains(part, "[]=") {
				if arr, ok := parsedVal.([]any); ok {
					parsedVal = []any{arr}
				}
			}
		}

		// Build nested structure
		newObj, err := parseKeys(decodedKey, parsedVal, opts, true)
		if err != nil {
			return nil, err
		}

		if newObj != nil {
			switch opts.Duplicates {
			case DuplicateFirst:
				result = mergeKeepFirst(result, newObj)
			case DuplicateLast:
				result = mergeKeepLast(result, newObj)
			default:
				merged := Merge(result, newObj)
				if m, ok := merged.(map[string]any); ok {
					result = m
				}
			}
		}
	}

	// Compact sparse arrays if AllowSparse is false
	if !opts.AllowSparse {
		compacted := Compact(result)
		if m, ok := compacted.(map[string]any); ok {
			return m, nil
		}
	} else {
		convertExplicitNulls(result)
	}

	return result, nil
}

// findEqualsOutsideBrackets finds the index of '=' that is not inside brackets.
func findEqualsOutsideBrackets(s string) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case '%':
			// Check for encoded brackets
			if i+2 < len(s) && s[i+1] == '5' {
				switch s[i+2] {
				case 'B', 'b':
					depth++
					i += 2
				case 'D', 'd':
					if depth > 0 {
						depth--
					}
					i += 2
				}
			}
		case '=':
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// estimateParams estimates the number of parameters in a query string.
func estimateParams(s string) int {
	if s == "" {
		return 0
	}
	n := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '&' {
			n++
		}
	}
	return n
}

// mergeKeepFirst merges newObj into result, keeping first values for duplicates.
func mergeKeepFirst(result map[string]any, newObj any) map[string]any {
	newMap, ok := newObj.(map[string]any)
	if !ok {
		return result
	}

	for key, newVal := range newMap {
		existingVal, exists := result[key]
		if !exists {
			result[key] = newVal
			continue
		}

		// Handle nested objects - recurse
		existingMap, existingIsMap := existingVal.(map[string]any)
		newValMap, newIsMap := newVal.(map[string]any)
		if existingIsMap && newIsMap {
			result[key] = mergeKeepFirst(existingMap, newValMap)
			continue
		}

		// For arrays, keep existing
		// For scalar values, keep existing
		// This is the "first" behavior
	}

	return result
}

// mergeKeepLast merges newObj into result, keeping last values for duplicates.
func mergeKeepLast(result map[string]any, newObj any) map[string]any {
	newMap, ok := newObj.(map[string]any)
	if !ok {
		return result
	}

	for key, newVal := range newMap {
		existingVal, exists := result[key]
		if !exists {
			result[key] = newVal
			continue
		}

		// Handle nested objects - recurse
		existingMap, existingIsMap := existingVal.(map[string]any)
		newValMap, newIsMap := newVal.(map[string]any)
		if existingIsMap && newIsMap {
			result[key] = mergeKeepLast(existingMap, newValMap)
			continue
		}

		// For non-map values, keep the new one (last wins)
		result[key] = newVal
	}

	return result
}

// charsetToLang converts qs.Charset to lang.Charset.
func charsetToLang(c Charset) lang.Charset {
	if c == CharsetISO88591 {
		return lang.CharsetISO88591
	}
	return lang.CharsetUTF8
}

// charsetFromLang converts lang.Charset to qs.Charset.
func charsetFromLang(c lang.Charset) Charset {
	if c == lang.CharsetISO88591 {
		return CharsetISO88591
	}
	return CharsetUTF8
}
