// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Charset represents supported character sets for encoding/decoding.
type Charset string

const (
	// CharsetUTF8 is the default charset (UTF-8).
	CharsetUTF8 Charset = "utf-8"
	// CharsetISO88591 is the ISO-8859-1 (Latin-1) charset.
	CharsetISO88591 Charset = "iso-8859-1"
)

// hexTable is a pre-computed table of percent-encoded bytes.
// hexTable[i] returns the percent-encoded string for byte value i (e.g., hexTable[32] = "%20").
var hexTable = func() [256]string {
	var table [256]string
	for i := 0; i < 256; i++ {
		table[i] = "%" + strings.ToUpper(string("0123456789abcdef"[i>>4])+string("0123456789abcdef"[i&0x0F]))
	}
	return table
}()

// isUnreservedChar returns true if the character should not be percent-encoded.
// Unreserved characters per RFC 3986: A-Z a-z 0-9 - . _ ~
// RFC 1738 additionally allows ( ) to pass through unencoded.
func isUnreservedChar(c byte, format Format) bool {
	switch {
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	case c >= '0' && c <= '9':
		return true
	case c == '-', c == '.', c == '_', c == '~':
		return true
	case format == FormatRFC1738 && (c == '(' || c == ')'):
		return true
	}
	return false
}

// Encode encodes a string for use in a URL query string.
// It follows RFC 3986 percent-encoding, with optional RFC 1738 format.
//
// Parameters:
//   - str: the string to encode
//   - charset: the character set to use (CharsetUTF8 or CharsetISO88591)
//   - format: the encoding format (FormatRFC1738 or FormatRFC3986)
//
// For UTF-8 (default), multi-byte characters are encoded as multiple %XX sequences.
// For ISO-8859-1, characters outside the Latin-1 range are encoded as numeric entities (&#xxxx;).
func Encode(str string, charset Charset, format Format) string {
	if len(str) == 0 {
		return str
	}

	if charset == CharsetISO88591 {
		return encodeISO88591(str)
	}

	// UTF-8 encoding
	var result strings.Builder
	result.Grow(len(str) * 3) // Pre-allocate for worst case

	for i := 0; i < len(str); {
		c := str[i]

		// Check if it's an unreserved character (single byte ASCII)
		if isUnreservedChar(c, format) {
			result.WriteByte(c)
			i++
			continue
		}

		// Handle ASCII characters (0x00-0x7F)
		if c < 0x80 {
			result.WriteString(hexTable[c])
			i++
			continue
		}

		// Handle multi-byte UTF-8 sequences
		r, size := utf8.DecodeRuneInString(str[i:])
		if r == utf8.RuneError && size == 1 {
			// Invalid UTF-8, encode the byte as-is
			result.WriteString(hexTable[c])
			i++
			continue
		}

		// Encode each byte of the UTF-8 sequence
		for j := 0; j < size; j++ {
			result.WriteString(hexTable[str[i+j]])
		}
		i += size
	}

	return result.String()
}

// encodeISO88591 encodes a string using ISO-8859-1 charset.
// Characters outside the Latin-1 range (0x00-0xFF) are encoded as numeric HTML entities.
func encodeISO88591(str string) string {
	var result strings.Builder
	result.Grow(len(str) * 6) // Pre-allocate for entities

	for _, r := range str {
		if r <= 0xFF {
			// Character is in Latin-1 range
			if isUnreservedChar(byte(r), FormatRFC3986) {
				result.WriteRune(r)
			} else {
				result.WriteString(hexTable[byte(r)])
			}
		} else {
			// Character is outside Latin-1, encode as numeric entity: &#xxxx;
			// Percent-encoded form: %26%23{decimal-codepoint}%3B
			result.WriteString("%26%23")
			writeDecimal(&result, int(r))
			result.WriteString("%3B")
		}
	}

	return result.String()
}

// writeDecimal writes an integer as decimal string to the builder.
func writeDecimal(b *strings.Builder, n int) {
	if n == 0 {
		b.WriteByte('0')
		return
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	b.Write(digits[i:])
}

// Decode decodes a URL-encoded string.
// It handles both percent-encoding (%XX) and plus signs (+) as spaces.
//
// Parameters:
//   - str: the string to decode
//   - charset: the character set to use for interpretation
//
// For UTF-8 (default), percent-encoded bytes are interpreted as UTF-8 sequences.
// For ISO-8859-1, each percent-encoded byte is interpreted as a Latin-1 character.
// Invalid sequences are left as-is (graceful fallback).
func Decode(str string, charset Charset) string {
	if len(str) == 0 {
		return str
	}

	// Replace + with space first
	str = strings.ReplaceAll(str, "+", " ")

	if charset == CharsetISO88591 {
		return decodeISO88591(str)
	}

	// UTF-8 decoding using standard library with graceful fallback
	decoded, err := url.QueryUnescape(str)
	if err != nil {
		// Graceful fallback: return the string with + replaced by space
		return str
	}
	return decoded
}

// decodeISO88591 decodes a percent-encoded string as ISO-8859-1.
// Each %XX is interpreted as a single Latin-1 byte.
func decodeISO88591(str string) string {
	var result strings.Builder
	result.Grow(len(str))

	for i := 0; i < len(str); {
		if str[i] == '%' && i+2 < len(str) {
			// Try to decode %XX
			hi := unhex(str[i+1])
			lo := unhex(str[i+2])
			if hi >= 0 && lo >= 0 {
				result.WriteByte(byte(hi<<4 | lo))
				i += 3
				continue
			}
		}
		result.WriteByte(str[i])
		i++
	}

	return result.String()
}

// unhex returns the numeric value of a hex digit, or -1 if invalid.
func unhex(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	}
	return -1
}

// Merge performs a deep merge of source into target.
// It handles maps, slices, and primitive values according to qs library semantics.
//
// Rules:
//   - If source is nil, returns target unchanged
//   - If source is a primitive and target is a slice, appends source to target
//   - If source is a primitive and target is a map, sets target[source] = true
//   - If target is nil/primitive and source is a map/slice, returns [target, source...]
//   - If both are slices, merges by index (objects are recursively merged)
//   - If both are maps, recursively merges keys
func Merge(target, source any, allowPrototypes bool) any {
	if source == nil {
		return target
	}

	// Check if source is a primitive (not map or slice)
	sourceMap, sourceIsMap := source.(map[string]any)
	sourceSlice, sourceIsSlice := source.([]any)
	sourceIsPrimitive := !sourceIsMap && !sourceIsSlice

	if sourceIsPrimitive {
		// Source is primitive
		if targetSlice, ok := target.([]any); ok {
			return append(targetSlice, source)
		}
		if targetMap, ok := target.(map[string]any); ok {
			key, isString := source.(string)
			if isString && (allowPrototypes || !isPrototypeKey(key)) {
				targetMap[key] = true
			}
			return targetMap
		}
		// target is nil or primitive
		return []any{target, source}
	}

	// Source is map or slice
	if target == nil {
		if sourceIsSlice {
			return sourceSlice
		}
		return sourceMap
	}

	targetMap, targetIsMap := target.(map[string]any)
	targetSlice, targetIsSlice := target.([]any)

	// If target is primitive, convert to slice and concat
	if !targetIsMap && !targetIsSlice {
		if sourceIsSlice {
			return append([]any{target}, sourceSlice...)
		}
		return []any{target, source}
	}

	// Both are slices
	if targetIsSlice && sourceIsSlice {
		for i, item := range sourceSlice {
			if i < len(targetSlice) {
				targetItem := targetSlice[i]
				_, targetItemIsMap := targetItem.(map[string]any)
				_, itemIsMap := item.(map[string]any)
				if targetItemIsMap && itemIsMap {
					targetSlice[i] = Merge(targetItem, item, allowPrototypes)
				} else {
					targetSlice = append(targetSlice, item)
				}
			} else {
				// Extend slice to accommodate index
				for len(targetSlice) <= i {
					targetSlice = append(targetSlice, nil)
				}
				targetSlice[i] = item
			}
		}
		return targetSlice
	}

	// Target is slice but source is map - convert target to map
	var mergeTarget map[string]any
	if targetIsSlice {
		mergeTarget = ArrayToObject(targetSlice)
	} else {
		mergeTarget = targetMap
	}

	// Source is a map, merge into target
	if sourceIsMap {
		for key, value := range sourceMap {
			if existing, exists := mergeTarget[key]; exists {
				mergeTarget[key] = Merge(existing, value, allowPrototypes)
			} else {
				mergeTarget[key] = value
			}
		}
	}

	return mergeTarget
}

// isPrototypeKey returns true if the key is a JavaScript prototype pollution key.
// In Go this is less relevant, but we maintain compatibility.
func isPrototypeKey(key string) bool {
	return key == "__proto__" || key == "constructor" || key == "prototype"
}

// ArrayToObject converts a slice to a map with string indices as keys.
// Nil values in the slice are skipped.
func ArrayToObject(source []any) map[string]any {
	result := make(map[string]any)
	for i, v := range source {
		if v != nil {
			result[strconv.Itoa(i)] = v
		}
	}
	return result
}

// Compact removes nil/"undefined" holes from nested slices.
// It traverses the entire object graph and compacts any sparse arrays.
func Compact(value any) any {
	if value == nil {
		return nil
	}

	// For top-level slices, compact directly
	if slice, ok := value.([]any); ok {
		return compactSlice(slice)
	}

	// For maps, recursively compact all nested slices
	if m, ok := value.(map[string]any); ok {
		compactMap(m)
		return m
	}

	return value
}

// compactSlice removes nil values from a slice and recursively compacts nested structures.
func compactSlice(slice []any) []any {
	result := make([]any, 0, len(slice))
	for _, v := range slice {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case []any:
			result = append(result, compactSlice(val))
		case map[string]any:
			compactMap(val)
			result = append(result, val)
		default:
			result = append(result, v)
		}
	}
	return result
}

// compactMap recursively compacts all nested slices within a map.
func compactMap(m map[string]any) {
	for k, v := range m {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case []any:
			m[k] = compactSlice(val)
		case map[string]any:
			compactMap(val)
		}
	}
}

// Combine concatenates two values into a slice.
// If either value is already a slice, its elements are flattened into the result.
func Combine(a, b any) []any {
	var result []any

	if aSlice, ok := a.([]any); ok {
		result = append(result, aSlice...)
	} else if a != nil {
		result = append(result, a)
	}

	if bSlice, ok := b.([]any); ok {
		result = append(result, bSlice...)
	} else if b != nil {
		result = append(result, b)
	}

	return result
}

// MaybeMap applies a function to a value.
// If the value is a slice, the function is applied to each element.
// Otherwise, the function is applied to the value directly.
func MaybeMap(val any, fn func(any) any) any {
	if slice, ok := val.([]any); ok {
		result := make([]any, len(slice))
		for i, v := range slice {
			result[i] = fn(v)
		}
		return result
	}
	return fn(val)
}

// IsRegExp checks if the value is a compiled regular expression.
func IsRegExp(obj any) bool {
	_, ok := obj.(*regexp.Regexp)
	return ok
}

// Assign copies all key-value pairs from source to target.
// It returns the modified target map.
func Assign(target, source map[string]any) map[string]any {
	for k, v := range source {
		target[k] = v
	}
	return target
}
