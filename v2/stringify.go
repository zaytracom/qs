// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ArrayFormat specifies how arrays should be serialized in the query string.
type ArrayFormat string

const (
	// ArrayFormatIndices serializes arrays with indices: a[0]=b&a[1]=c
	ArrayFormatIndices ArrayFormat = "indices"
	// ArrayFormatBrackets serializes arrays with empty brackets: a[]=b&a[]=c
	ArrayFormatBrackets ArrayFormat = "brackets"
	// ArrayFormatRepeat serializes arrays by repeating the key: a=b&a=c
	ArrayFormatRepeat ArrayFormat = "repeat"
	// ArrayFormatComma serializes arrays as comma-separated values: a=b,c
	ArrayFormatComma ArrayFormat = "comma"
)

// EncoderFunc is a custom encoder function signature.
// Parameters:
//   - str: the string to encode
//   - charset: the charset being used
//   - kind: either "key" or "value" indicating what is being encoded
//   - format: the format (RFC1738 or RFC3986)
//
// Returns the encoded string.
type EncoderFunc func(str string, charset Charset, kind string, format Format) string

// SerializeDateFunc is a function that serializes a time.Time to a string.
type SerializeDateFunc func(t time.Time) string

// FilterFunc is a function that filters/transforms values during stringification.
// It receives the key (or prefix) and the value, and returns the transformed value.
// Return nil to skip this key.
type FilterFunc func(prefix string, value any) any

// SortFunc is a function for sorting keys.
// It returns true if key a should come before key b.
type SortFunc func(a, b string) bool

// StringifyOptions configures the behavior of the Stringify function.
type StringifyOptions struct {
	// AddQueryPrefix adds a leading ? to the output.
	// Default: false
	AddQueryPrefix bool

	// AllowDots uses dot notation instead of brackets for nested objects.
	// e.g., {a: {b: "c"}} → "a.b=c" instead of "a[b]=c"
	// Default: false
	AllowDots bool

	// allowDotsSet tracks if AllowDots was explicitly set (internal use only)
	allowDotsSet bool

	// AllowEmptyArrays outputs key[] for empty arrays.
	// When false, empty arrays produce no output.
	// Default: false
	AllowEmptyArrays bool

	// ArrayFormat specifies how arrays are serialized.
	// Default: ArrayFormatIndices
	ArrayFormat ArrayFormat

	// Charset specifies the character encoding to use.
	// Default: CharsetUTF8
	Charset Charset

	// CharsetSentinel adds a utf8=✓ parameter for charset indication.
	// Default: false
	CharsetSentinel bool

	// CommaRoundTrip ensures single-element arrays round-trip with comma format.
	// When true, [a] becomes "key[]=a" instead of "key=a" with comma format.
	// Default: false
	CommaRoundTrip bool

	// Delimiter is the string used to join key-value pairs.
	// Default: "&"
	Delimiter string

	// Encode enables URL encoding of keys and values.
	// Default: true
	Encode bool

	// EncodeDotInKeys encodes . as %2E in keys when using dot notation.
	// Default: false
	EncodeDotInKeys bool

	// Encoder is a custom function for encoding strings.
	// If nil, the default encoder is used.
	// Default: nil (uses built-in Encode function)
	Encoder EncoderFunc

	// EncodeValuesOnly only encodes values, not keys.
	// Default: false
	EncodeValuesOnly bool

	// Filter can be a function or a slice of strings.
	// If a function, it filters/transforms values during stringification.
	// If a slice of strings, only those keys are included.
	// Default: nil
	Filter any // FilterFunc or []string

	// Format specifies the RFC encoding format (RFC1738 or RFC3986).
	// Default: FormatRFC3986
	Format Format

	// Formatter is the function that applies format-specific encoding.
	// If nil, determined by Format option.
	// Default: nil (uses Formatters[Format])
	Formatter FormatterFunc

	// SerializeDate is a function for serializing time.Time values.
	// Default: time.Time.Format(time.RFC3339)
	SerializeDate SerializeDateFunc

	// SkipNulls skips keys with null/nil values.
	// Default: false
	SkipNulls bool

	// Sort is a function for sorting keys before stringification.
	// Default: nil (no sorting)
	Sort SortFunc

	// SortArrayIndices when true, sorts array indices as strings along with object keys.
	// This is needed for compatibility with JS qs library which sorts all keys including
	// array indices when sort option is provided (e.g., "0", "1", "10", "2" order).
	// Default: false (arrays preserve natural numeric order)
	SortArrayIndices bool

	// StrictNullHandling serializes null values without = sign.
	// e.g., {a: null} → "a" instead of "a="
	// Default: false
	StrictNullHandling bool
}

// Default values for StringifyOptions
const (
	DefaultStringifyDelimiter = "&"
)


// Validation errors for StringifyOptions
var (
	ErrInvalidStringifyAllowEmptyArrays = errors.New("allowEmptyArrays option must be a boolean")
	ErrInvalidEncodeDotInKeys           = errors.New("encodeDotInKeys option must be a boolean")
	ErrInvalidEncoder                   = errors.New("encoder must be a function")
	ErrInvalidStringifyCharset          = errors.New("charset must be utf-8 or iso-8859-1")
	ErrInvalidFormat                    = errors.New("unknown format option provided")
	ErrInvalidCommaRoundTrip            = errors.New("commaRoundTrip must be a boolean, or absent")
	ErrInvalidArrayFormat               = errors.New("arrayFormat must be indices, brackets, repeat, or comma")
	ErrCyclicReference                  = errors.New("cyclic object value")
)

// defaultSerializeDate is the default date serialization function.
func defaultSerializeDate(t time.Time) string {
	return t.Format(time.RFC3339)
}

// DefaultStringifyOptions returns StringifyOptions with default values.
func DefaultStringifyOptions() StringifyOptions {
	return StringifyOptions{
		AddQueryPrefix:     false,
		AllowDots:          false,
		AllowEmptyArrays:   false,
		ArrayFormat:        ArrayFormatIndices,
		Charset:            CharsetUTF8,
		CharsetSentinel:    false,
		CommaRoundTrip:     false,
		Delimiter:          DefaultStringifyDelimiter,
		Encode:             true,
		EncodeDotInKeys:    false,
		Encoder:            nil,
		EncodeValuesOnly:   false,
		Filter:             nil,
		Format:             DefaultFormat,
		Formatter:          nil,
		SerializeDate:      defaultSerializeDate,
		SkipNulls:          false,
		Sort:               nil,
		StrictNullHandling: false,
	}
}

// normalizeStringifyOptions validates and fills in defaults for StringifyOptions.
// It returns a new StringifyOptions with all fields properly set.
func normalizeStringifyOptions(opts *StringifyOptions) (StringifyOptions, error) {
	if opts == nil {
		return DefaultStringifyOptions(), nil
	}

	result := *opts

	// Validate and set charset
	if result.Charset == "" {
		result.Charset = CharsetUTF8
	} else if result.Charset != CharsetUTF8 && result.Charset != CharsetISO88591 {
		return result, ErrInvalidStringifyCharset
	}

	// Validate and set format
	if result.Format == "" {
		result.Format = DefaultFormat
	} else if result.Format != FormatRFC1738 && result.Format != FormatRFC3986 {
		return result, ErrInvalidFormat
	}

	// Set formatter based on format
	if result.Formatter == nil {
		result.Formatter = GetFormatter(result.Format)
	}

	// Validate array format
	if result.ArrayFormat == "" {
		result.ArrayFormat = ArrayFormatIndices
	} else if result.ArrayFormat != ArrayFormatIndices &&
		result.ArrayFormat != ArrayFormatBrackets &&
		result.ArrayFormat != ArrayFormatRepeat &&
		result.ArrayFormat != ArrayFormatComma {
		return result, ErrInvalidArrayFormat
	}

	// Validate filter (must be FilterFunc or []string or nil)
	if result.Filter != nil {
		switch result.Filter.(type) {
		case FilterFunc, func(string, any) any, []string:
			// Valid
		default:
			return result, errors.New("filter must be a function or array of strings")
		}
	}

	// Set default delimiter
	if result.Delimiter == "" {
		result.Delimiter = DefaultStringifyDelimiter
	}

	// Set default date serializer
	if result.SerializeDate == nil {
		result.SerializeDate = defaultSerializeDate
	}

	// If EncodeDotInKeys is true and AllowDots was NOT explicitly set, set AllowDots to true
	// JS: allowDots = typeof opts.allowDots === 'undefined'
	//     ? opts.encodeDotInKeys === true ? true : defaults.allowDots
	//     : !!opts.allowDots;
	if result.EncodeDotInKeys && !result.allowDotsSet {
		result.AllowDots = true
	}

	return result, nil
}

// StringifyOption is a functional option for configuring StringifyOptions.
type StringifyOption func(*StringifyOptions)

// WithStringifyAddQueryPrefix adds a leading ? to the output.
func WithStringifyAddQueryPrefix(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.AddQueryPrefix = v
	}
}

// WithStringifyAllowDots uses dot notation instead of brackets.
func WithStringifyAllowDots(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.AllowDots = v
		o.allowDotsSet = true
	}
}

// WithStringifyAllowEmptyArrays outputs key[] for empty arrays.
func WithStringifyAllowEmptyArrays(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.AllowEmptyArrays = v
	}
}

// WithStringifyArrayFormat sets how arrays are serialized.
func WithStringifyArrayFormat(v ArrayFormat) StringifyOption {
	return func(o *StringifyOptions) {
		o.ArrayFormat = v
	}
}

// WithStringifyCharset sets the character encoding to use.
func WithStringifyCharset(v Charset) StringifyOption {
	return func(o *StringifyOptions) {
		o.Charset = v
	}
}

// WithStringifyCharsetSentinel adds a utf8=✓ parameter for charset indication.
func WithStringifyCharsetSentinel(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.CharsetSentinel = v
	}
}

// WithStringifyCommaRoundTrip ensures single-element arrays round-trip with comma format.
func WithStringifyCommaRoundTrip(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.CommaRoundTrip = v
	}
}

// WithStringifyDelimiter sets the string used to join key-value pairs.
func WithStringifyDelimiter(v string) StringifyOption {
	return func(o *StringifyOptions) {
		o.Delimiter = v
	}
}

// WithStringifyEncode enables or disables URL encoding.
func WithStringifyEncode(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.Encode = v
	}
}

// WithStringifyEncodeDotInKeys encodes . as %2E in keys.
func WithStringifyEncodeDotInKeys(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.EncodeDotInKeys = v
		// Note: AllowDots is set in normalizeStringifyOptions if not explicitly set
	}
}

// WithStringifyEncoder sets a custom encoder function.
func WithStringifyEncoder(v EncoderFunc) StringifyOption {
	return func(o *StringifyOptions) {
		o.Encoder = v
	}
}

// WithStringifyEncodeValuesOnly only encodes values, not keys.
func WithStringifyEncodeValuesOnly(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.EncodeValuesOnly = v
	}
}

// WithStringifyFilter sets a filter function or array of allowed keys.
func WithStringifyFilter(v any) StringifyOption {
	return func(o *StringifyOptions) {
		o.Filter = v
	}
}

// WithStringifyFormat sets the RFC encoding format.
func WithStringifyFormat(v Format) StringifyOption {
	return func(o *StringifyOptions) {
		o.Format = v
		o.Formatter = GetFormatter(v)
	}
}

// WithStringifyFormatter sets a custom formatter function.
func WithStringifyFormatter(v FormatterFunc) StringifyOption {
	return func(o *StringifyOptions) {
		o.Formatter = v
	}
}

// WithStringifySerializeDate sets a custom date serialization function.
func WithStringifySerializeDate(v SerializeDateFunc) StringifyOption {
	return func(o *StringifyOptions) {
		o.SerializeDate = v
	}
}

// WithStringifySkipNulls skips keys with null/nil values.
func WithStringifySkipNulls(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.SkipNulls = v
	}
}

// WithStringifySort sets a function for sorting keys.
func WithStringifySort(v SortFunc) StringifyOption {
	return func(o *StringifyOptions) {
		o.Sort = v
	}
}

// WithStringifySortArrayIndices enables sorting array indices as strings.
// When true, array indices are sorted lexicographically like object keys,
// producing output like "a[0], a[1], a[10], a[2]" instead of "a[0], a[1], a[2], a[10]".
// This is needed for compatibility with JS qs library's sort behavior.
func WithStringifySortArrayIndices(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.SortArrayIndices = v
	}
}

// WithStringifyStrictNullHandling serializes null values without = sign.
func WithStringifyStrictNullHandling(v bool) StringifyOption {
	return func(o *StringifyOptions) {
		o.StrictNullHandling = v
	}
}

// applyStringifyOptions applies functional options to a StringifyOptions struct.
func applyStringifyOptions(opts ...StringifyOption) StringifyOptions {
	o := DefaultStringifyOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// arrayPrefixGenerators holds functions that generate the key prefix for array items.
var arrayPrefixGenerators = map[ArrayFormat]func(prefix string, key string) string{
	ArrayFormatBrackets: func(prefix, key string) string { return prefix + "[]" },
	ArrayFormatIndices:  func(prefix, key string) string { return prefix + "[" + key + "]" },
	ArrayFormatRepeat:   func(prefix, key string) string { return prefix },
	ArrayFormatComma:    nil, // Special case, handled separately
}

// isNonNullishPrimitive checks if a value is a non-nil primitive type (string, number, bool).
func isNonNullishPrimitive(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return true
	}
	return false
}

// sideChannel is used for cyclic reference detection.
// It tracks which objects have been seen during traversal using reflect to get pointer addresses.
type sideChannel struct {
	// Track seen objects by their reflect.Value pointer
	seen map[uintptr]int
}

func newSideChannel() *sideChannel {
	return &sideChannel{
		seen: make(map[uintptr]int),
	}
}

// getValuePtr returns a unique identifier for a value based on its memory address.
// Returns 0 for non-reference types (primitives).
// Note: Empty slices in Go may share the same backing array address (runtime optimization),
// so we return 0 for empty slices to avoid false cyclic reference detection.
func getValuePtr(v any) uintptr {
	if v == nil {
		return 0
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		if rv.IsNil() || rv.Len() == 0 {
			return 0
		}
		return rv.Pointer()
	case reflect.Slice:
		// Empty slices may share the same backing array pointer in Go,
		// so we skip cyclic detection for them to avoid false positives.
		if rv.IsNil() || rv.Len() == 0 {
			return 0
		}
		return rv.Pointer()
	case reflect.Ptr:
		if rv.IsNil() {
			return 0
		}
		return rv.Pointer()
	}
	return 0
}

func (sc *sideChannel) set(key any, step int) {
	ptr := getValuePtr(key)
	if ptr != 0 {
		sc.seen[ptr] = step
	}
}

func (sc *sideChannel) get(key any) (int, bool) {
	ptr := getValuePtr(key)
	if ptr == 0 {
		return 0, false
	}
	v, ok := sc.seen[ptr]
	return v, ok
}

func (sc *sideChannel) child() *sideChannel {
	child := newSideChannel()
	// Copy parent seen objects
	for k, v := range sc.seen {
		child.seen[k] = v
	}
	return child
}

// stringify is the internal recursive function that stringifies values.
func stringify(
	object any,
	prefix string,
	generateArrayPrefix func(string, string) string,
	commaRoundTrip bool,
	allowEmptyArrays bool,
	strictNullHandling bool,
	skipNulls bool,
	encodeDotInKeys bool,
	encoder func(string, Charset, string, Format) string,
	filter any,
	sort SortFunc,
	sortArrayIndices bool,
	allowDots bool,
	serializeDate SerializeDateFunc,
	format Format,
	formatter FormatterFunc,
	encodeValuesOnly bool,
	charset Charset,
	sideChannel *sideChannel,
	step int,
) ([]string, error) {
	obj := object

	// Cyclic reference detection - check if we've seen this object before
	if obj != nil && !isNonNullishPrimitive(obj) {
		if _, exists := sideChannel.get(obj); exists {
			return nil, ErrCyclicReference
		}
		// Mark this object as seen
		sideChannel.set(obj, step)
	}

	// Apply filter function if provided
	// Note: In JS, returning undefined from filter skips the key.
	// In Go, we use a special convention: filter should return the special
	// value SkipValue (a sentinel) to skip, not nil.
	// For backwards compatibility and JS parity, we skip only if:
	// 1. Filter returns nil AND original value was NOT nil (filter explicitly removed it)
	// 2. Filter returns SkipValue marker
	if filterFunc, ok := filter.(FilterFunc); ok {
		origObj := obj
		obj = filterFunc(prefix, obj)
		// Skip only if filter explicitly returns nil for non-nil input (JS undefined behavior)
		// If input was nil and output is nil, it's "keep null", not "skip"
		if obj == nil && origObj != nil {
			return []string{}, nil
		}
	} else if fn, ok := filter.(func(string, any) any); ok {
		origObj := obj
		obj = fn(prefix, obj)
		if obj == nil && origObj != nil {
			return []string{}, nil
		}
	}

	// Handle time.Time
	if t, ok := obj.(time.Time); ok {
		obj = serializeDate(t)
	}

	// Handle comma format with arrays - serialize dates in array first
	if generateArrayPrefix == nil && isSlice(obj) {
		obj = MaybeMap(obj, func(v any) any {
			if t, ok := v.(time.Time); ok {
				return serializeDate(t)
			}
			return v
		})
	}

	// Handle nil/null
	if obj == nil || IsExplicitNull(obj) {
		if strictNullHandling {
			if encoder != nil && !encodeValuesOnly {
				return []string{formatter(encoder(prefix, charset, "key", format))}, nil
			}
			return []string{formatter(prefix)}, nil
		}
		obj = ""
	}

	// Handle primitives
	if isNonNullishPrimitive(obj) {
		if encoder != nil {
			var keyValue string
			if encodeValuesOnly {
				keyValue = prefix
			} else {
				keyValue = encoder(prefix, charset, "key", format)
			}
			valStr := toString(obj)
			return []string{formatter(keyValue) + "=" + formatter(encoder(valStr, charset, "value", format))}, nil
		}
		return []string{formatter(prefix) + "=" + formatter(toString(obj))}, nil
	}

	var values []string

	// Handle objects and arrays
	var objKeys []any

	if generateArrayPrefix == nil && isSlice(obj) {
		// Comma format - join elements
		slice := toSlice(obj)
		if encodeValuesOnly && encoder != nil {
			// Encode each element
			encodedSlice := make([]string, len(slice))
			for i, v := range slice {
				if s, ok := v.(string); ok {
					encodedSlice[i] = encoder(s, charset, "value", format)
				} else {
					encodedSlice[i] = toString(v)
				}
			}
			if len(encodedSlice) > 0 {
				joined := strings.Join(encodedSlice, ",")
				// JS: obj.join(',') || null - if result is empty string, use null
				// Use ExplicitNullValue (not nil) so it's not skipped as sparse array
				if joined == "" {
					objKeys = []any{map[string]any{"value": ExplicitNullValue}}
				} else {
					objKeys = []any{map[string]any{"value": joined}}
				}
			}
		} else {
			if len(slice) > 0 {
				strSlice := make([]string, len(slice))
				for i, v := range slice {
					strSlice[i] = toString(v)
				}
				joined := strings.Join(strSlice, ",")
				// JS: obj.join(',') || null - if result is empty string, use null
				// Use ExplicitNullValue (not nil) so it's not skipped as sparse array
				if joined == "" {
					objKeys = []any{map[string]any{"value": ExplicitNullValue}}
				} else {
					objKeys = []any{map[string]any{"value": joined}}
				}
			}
		}
	} else if filterSlice, ok := filter.([]string); ok {
		// Filter is array of keys - applies to both arrays and objects
		// For arrays: filter acts as array of indices (strings that can convert to int)
		// For objects: filter acts as array of keys
		objKeys = make([]any, len(filterSlice))
		for i, k := range filterSlice {
			objKeys[i] = k
		}
	} else {
		// No filter array - get keys from object
		switch v := obj.(type) {
		case map[string]any:
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			if sort != nil {
				sortStrings(keys, sort)
			}
			objKeys = make([]any, len(keys))
			for i, k := range keys {
				objKeys[i] = k
			}
		case []any:
			if sortArrayIndices && sort != nil {
				// Convert indices to strings and sort them lexicographically
				// This matches JS qs behavior where sort applies to all keys including array indices
				keys := make([]string, len(v))
				for i := range v {
					keys[i] = strconv.Itoa(i)
				}
				sortStrings(keys, sort)
				objKeys = make([]any, len(keys))
				for i, k := range keys {
					objKeys[i] = k // Keep as string for sorted order
				}
			} else {
				// Normal numeric order
				objKeys = make([]any, len(v))
				for i := range v {
					objKeys[i] = i
				}
			}
		}
	}

	// Encode prefix dots if needed
	encodedPrefix := prefix
	if encodeDotInKeys {
		encodedPrefix = strings.ReplaceAll(prefix, ".", "%2E")
	}

	// Handle commaRoundTrip for single element arrays
	adjustedPrefix := encodedPrefix
	if commaRoundTrip && isSlice(obj) && len(toSlice(obj)) == 1 {
		adjustedPrefix = encodedPrefix + "[]"
	}

	// Handle empty arrays
	if allowEmptyArrays && isSlice(obj) && len(toSlice(obj)) == 0 {
		return []string{adjustedPrefix + "[]"}, nil
	}

	// Iterate over keys
	for _, key := range objKeys {
		var value any
		var keyStr string

		// Get value based on key type
		if keyMap, ok := key.(map[string]any); ok {
			// Special case for comma format
			if v, exists := keyMap["value"]; exists {
				value = v
				keyStr = ""
			}
		} else {
			keyStr = toString(key)
			var keyExists bool
			switch v := obj.(type) {
			case map[string]any:
				value, keyExists = v[keyStr]
				if !keyExists {
					// Key doesn't exist in map - treat as undefined, skip
					continue
				}
			case []any:
				idx, ok := toInt(key)
				if ok && idx >= 0 && idx < len(v) {
					value = v[idx]
					keyExists = true
				}
				// If key is not a valid index, skip (undefined)
				if !keyExists {
					continue
				}
			}
		}

		// In Go: nil represents undefined (sparse array slot), ExplicitNullValue represents null
		// undefined in arrays is always skipped (JS behavior: typeof obj === 'undefined' returns [])
		if value == nil && isSlice(obj) {
			continue
		}
		// Skip nulls if skipNulls is requested
		if skipNulls && (value == nil || IsExplicitNull(value)) {
			continue
		}

		// Generate key prefix
		var keyPrefix string

		// Check if this is a comma format special case (keyStr="" for joined value)
		if keyMap, ok := key.(map[string]any); ok && keyStr == "" {
			if _, hasValue := keyMap["value"]; hasValue {
				// Comma format - value is already joined, use adjustedPrefix directly
				keyPrefix = adjustedPrefix
			}
		}

		// If keyPrefix wasn't set by comma format handling, generate it normally
		if keyPrefix == "" && key != nil {
			encodedKey := keyStr
			if allowDots && encodeDotInKeys {
				encodedKey = strings.ReplaceAll(keyStr, ".", "%2E")
			}

			if isSlice(obj) {
				if generateArrayPrefix != nil {
					keyPrefix = generateArrayPrefix(adjustedPrefix, encodedKey)
				} else {
					keyPrefix = adjustedPrefix
				}
			} else {
				if allowDots {
					keyPrefix = adjustedPrefix + "." + encodedKey
				} else {
					keyPrefix = adjustedPrefix + "[" + encodedKey + "]"
				}
			}
		}

		// Create child side channel for recursive call
		childSc := sideChannel.child()

		// Determine encoder for recursive call
		var childEncoder func(string, Charset, string, Format) string
		if generateArrayPrefix == nil && encodeValuesOnly && isSlice(obj) {
			childEncoder = nil
		} else {
			childEncoder = encoder
		}

		// Recurse
		childValues, err := stringify(
			value,
			keyPrefix,
			generateArrayPrefix,
			commaRoundTrip,
			allowEmptyArrays,
			strictNullHandling,
			skipNulls,
			encodeDotInKeys,
			childEncoder,
			filter,
			sort,
			sortArrayIndices,
			allowDots,
			serializeDate,
			format,
			formatter,
			encodeValuesOnly,
			charset,
			childSc,
			step+1,
		)
		if err != nil {
			return nil, err
		}

		values = append(values, childValues...)
	}

	return values, nil
}

// Stringify converts a map or value to a URL query string.
//
// Example:
//
//	str, err := qs.Stringify(map[string]any{"a": "b", "c": "d"})
//	// str = "a=b&c=d"
//
//	str, err := qs.Stringify(map[string]any{"a": map[string]any{"b": "c"}})
//	// str = "a%5Bb%5D=c"  (a[b]=c URL encoded)
//
//	str, err := qs.Stringify(map[string]any{"a": []any{"b", "c"}})
//	// str = "a%5B0%5D=b&a%5B1%5D=c"  (a[0]=b&a[1]=c URL encoded)
func Stringify(obj any, opts ...StringifyOption) (string, error) {
	options := applyStringifyOptions(opts...)

	// Normalize options
	normalizedOpts, err := normalizeStringifyOptions(&options)
	if err != nil {
		return "", err
	}

	var filter any = normalizedOpts.Filter
	var objKeys []string

	// Handle filter
	if filterFunc, ok := filter.(FilterFunc); ok {
		obj = filterFunc("", obj)
	} else if fn, ok := filter.(func(string, any) any); ok {
		obj = fn("", obj)
	} else if filterSlice, ok := filter.([]string); ok {
		objKeys = filterSlice
	}

	// Handle non-object input
	if obj == nil {
		return "", nil
	}
	objMap, isMap := obj.(map[string]any)
	if !isMap {
		return "", nil
	}

	// Get array prefix generator
	generateArrayPrefix := arrayPrefixGenerators[normalizedOpts.ArrayFormat]
	commaRoundTrip := generateArrayPrefix == nil && normalizedOpts.CommaRoundTrip

	// Get keys if not filtered
	if objKeys == nil {
		objKeys = make([]string, 0, len(objMap))
		for k := range objMap {
			objKeys = append(objKeys, k)
		}
	}

	// Sort keys if requested
	if normalizedOpts.Sort != nil {
		sortStrings(objKeys, normalizedOpts.Sort)
	}

	// Set up encoder
	var encoder func(string, Charset, string, Format) string
	if normalizedOpts.Encode {
		if normalizedOpts.Encoder != nil {
			encoder = normalizedOpts.Encoder
		} else {
			encoder = func(str string, charset Charset, kind string, format Format) string {
				return Encode(str, charset, format)
			}
		}
	}

	// Initialize side channel for cycle detection
	sideChannel := newSideChannel()

	var keys []string
	for _, key := range objKeys {
		value, exists := objMap[key]

		// Skip non-existent keys (undefined in JS)
		if !exists {
			continue
		}

		// Skip nulls if requested
		if normalizedOpts.SkipNulls && (value == nil || IsExplicitNull(value)) {
			continue
		}

		keyValues, err := stringify(
			value,
			key,
			generateArrayPrefix,
			commaRoundTrip,
			normalizedOpts.AllowEmptyArrays,
			normalizedOpts.StrictNullHandling,
			normalizedOpts.SkipNulls,
			normalizedOpts.EncodeDotInKeys,
			encoder,
			filter,
			normalizedOpts.Sort,
			normalizedOpts.SortArrayIndices,
			normalizedOpts.AllowDots,
			normalizedOpts.SerializeDate,
			normalizedOpts.Format,
			normalizedOpts.Formatter,
			normalizedOpts.EncodeValuesOnly,
			normalizedOpts.Charset,
			sideChannel,
			0,
		)
		if err != nil {
			return "", err
		}

		keys = append(keys, keyValues...)
	}

	joined := strings.Join(keys, normalizedOpts.Delimiter)
	prefix := ""

	if normalizedOpts.AddQueryPrefix {
		prefix = "?"
	}

	// Add charset sentinel
	if normalizedOpts.CharsetSentinel {
		if normalizedOpts.Charset == CharsetISO88591 {
			// encodeURIComponent('&#10003;'), the "numeric entity" representation of a checkmark
			prefix += "utf8=%26%2310003%3B&"
		} else {
			// encodeURIComponent('✓')
			prefix += "utf8=%E2%9C%93&"
		}
	}

	if len(joined) > 0 {
		return prefix + joined, nil
	}
	return "", nil
}

// Helper functions

// isSlice checks if a value is a slice.
func isSlice(v any) bool {
	if v == nil {
		return false
	}
	_, ok := v.([]any)
	return ok
}

// toSlice converts a value to []any if possible.
func toSlice(v any) []any {
	if slice, ok := v.([]any); ok {
		return slice
	}
	return nil
}

// toString converts a value to string.
func toString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

// toInt converts a value to int.
func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case string:
		i, err := strconv.Atoi(val)
		return i, err == nil
	}
	return 0, false
}

// sortStrings sorts a slice of strings using a custom comparison function.
func sortStrings(slice []string, less SortFunc) {
	// Simple insertion sort for small arrays (typical case)
	for i := 1; i < len(slice); i++ {
		for j := i; j > 0 && less(slice[j], slice[j-1]); j-- {
			slice[j], slice[j-1] = slice[j-1], slice[j]
		}
	}
}
