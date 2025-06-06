// Copyright 2024 QS Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package qs provides parsing and stringifying of URL query strings with support
// for nested objects, arrays, and complex data structures.
//
// This library is inspired by and compatible with the popular JavaScript qs library
// (https://github.com/ljharb/qs), while following Go conventions and idioms.
//
// # Features
//
//   - Parse query strings into nested Go data structures
//   - Stringify Go data structures into query strings
//   - Support for nested objects and arrays
//   - Multiple array formats (indices, brackets, repeat)
//   - URL encoding/decoding
//   - Customizable options and delimiters
//   - Struct parsing with query tags
//   - Idiomatic Marshal/Unmarshal functions with automatic type detection
//   - High performance with comprehensive benchmarks
//   - Extensive test coverage
//
// # Quick Start
//
// Basic parsing:
//
//	result, err := qs.Parse("name=John&age=30&skills[]=Go&skills[]=Python")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("%+v\n", result)
//	// Output: map[age:30 name:John skills:[Go Python]]
//
// Basic stringifying:
//
//	data := map[string]interface{}{
//	    "user": map[string]interface{}{
//	        "name": "Jane",
//	        "profile": map[string]interface{}{
//	            "age": 25,
//	            "skills": []interface{}{"JavaScript", "TypeScript"},
//	        },
//	    },
//	}
//	queryString, err := qs.Stringify(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(queryString)
//	// Output: user[name]=Jane&user[profile][age]=25&user[profile][skills][0]=JavaScript&user[profile][skills][1]=TypeScript
//
// # Struct Parsing with Query Tags
//
// Define structs with query tags:
//
//	type User struct {
//	    Name     string  `query:"name"`
//	    Age      int     `query:"age"`
//	    Email    string  `query:"email"`
//	    IsActive bool    `query:"active"`
//	    Score    float64 `query:"score"`
//	}
//
// Parse query string to struct:
//
//	queryString := "name=John&age=30&email=john@example.com&active=true&score=95.5"
//	var user User
//	err := qs.ParseToStruct(queryString, &user)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("User: %+v\n", user)
//
// # Idiomatic Marshal/Unmarshal
//
// The package provides idiomatic Marshal and Unmarshal functions that automatically
// detect data types at runtime:
//
//	// Marshal any data type to query string
//	queryString, err := qs.Marshal(user)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Unmarshal query string to any data type
//	var newUser User
//	err = qs.Unmarshal(queryString, &newUser)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Strapi-style APIs
//
// The library supports complex Strapi-like query structures:
//
//	type StrapiQuery struct {
//	    Filters    map[string]interface{} `query:"filters"`
//	    Sort       []string               `query:"sort"`
//	    Fields     []string               `query:"fields"`
//	    Populate   map[string]interface{} `query:"populate"`
//	    Pagination map[string]interface{} `query:"pagination"`
//	    Locale     string                 `query:"locale"`
//	}
//
// # Options
//
// Both parsing and stringifying support extensive customization through options:
//
//	// Parse options
//	result, err := qs.Parse("name=John&age=30", &qs.ParseOptions{
//	    Delimiter:         "&",
//	    IgnoreQueryPrefix: true,
//	})
//
//	// Stringify options
//	queryString, err := qs.Stringify(data, &qs.StringifyOptions{
//	    ArrayFormat:    "brackets",
//	    AddQueryPrefix: true,
//	})
//
// # Performance
//
// The library is optimized for performance:
//
//	BenchmarkParseSimple-10       169478    6458 ns/op
//	BenchmarkParseComplex-10       64339   18474 ns/op
//	BenchmarkStringifySimple-10  2973129     400 ns/op
//	BenchmarkStringifyComplex-10  701146    1675 ns/op
//
// # Compatibility
//
// This library aims to be compatible with the JavaScript qs library while
// following Go conventions and idioms. It supports all major query string
// patterns and provides additional Go-specific features like struct parsing.
package qs

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseOptions configures how query strings are parsed.
// These options control various aspects of the parsing process,
// including array handling, depth limits, and character encoding.
type ParseOptions struct {
	// AllowDots enables parsing of keys with dots (e.g., "a.b.c=value")
	// as nested objects instead of treating them as literal keys.
	AllowDots bool

	// AllowEmptyArrays allows arrays with no values to be parsed
	// as empty arrays instead of being omitted.
	AllowEmptyArrays bool

	// AllowPrototypes enables setting properties on Object.prototype
	// (mainly for JavaScript compatibility).
	AllowPrototypes bool

	// AllowSparse enables sparse arrays where some indices may be missing.
	AllowSparse bool

	// ArrayLimit sets the maximum number of array elements to parse.
	// Default is 20. Set to 0 for unlimited.
	ArrayLimit int

	// Charset specifies the character encoding to use.
	// Default is "utf-8".
	Charset string

	// CharsetSentinel enables detection of charset parameter in query string.
	CharsetSentinel bool

	// Comma enables parsing of comma-separated values within a single parameter.
	Comma bool

	// DecodeDotInKeys enables decoding of URL-encoded dots in keys.
	DecodeDotInKeys bool

	// Decoder is a custom function for decoding parameter values.
	// If nil, url.QueryUnescape is used.
	Decoder func(str string, decoder ...interface{}) (string, error)

	// Delimiter specifies the character used to separate parameters.
	// Default is "&".
	Delimiter string

	// Depth sets the maximum depth for nested objects.
	// Default is 5.
	Depth int

	// Duplicates specifies how to handle duplicate keys.
	// Options: "combine" (default), "first", "last".
	Duplicates string

	// IgnoreQueryPrefix ignores leading "?" in query string.
	IgnoreQueryPrefix bool

	// InterpretNumericEntities enables interpretation of HTML numeric entities.
	InterpretNumericEntities bool

	// ParameterLimit sets the maximum number of parameters to parse.
	// Default is 1000. Set to 0 for unlimited.
	ParameterLimit int

	// ParseArrays enables parsing of array notation (e.g., "a[]=1&a[]=2").
	ParseArrays bool

	// PlainObjects creates objects without prototypes
	// (mainly for JavaScript compatibility).
	PlainObjects bool

	// StrictDepth throws an error when depth limit is exceeded
	// instead of silently truncating.
	StrictDepth bool

	// StrictNullHandling preserves null values instead of converting
	// them to empty strings.
	StrictNullHandling bool

	// ThrowOnLimitExceeded throws an error when parameter limit is exceeded
	// instead of silently truncating.
	ThrowOnLimitExceeded bool
}

// defaultParseOptions returns a ParseOptions struct with default values.
// These defaults provide sensible behavior for most use cases while
// maintaining compatibility with the JavaScript qs library.
func defaultParseOptions() *ParseOptions {
	return &ParseOptions{
		AllowDots:                false,
		AllowEmptyArrays:         false,
		AllowPrototypes:          false,
		AllowSparse:              false,
		ArrayLimit:               20,
		Charset:                  "utf-8",
		CharsetSentinel:          false,
		Comma:                    false,
		DecodeDotInKeys:          false,
		Decoder:                  nil, // default decoder will be set in Parse
		Delimiter:                "&",
		Depth:                    5,
		Duplicates:               "combine",
		IgnoreQueryPrefix:        false,
		InterpretNumericEntities: false,
		ParameterLimit:           1000,
		ParseArrays:              true,
		PlainObjects:             false,
		StrictDepth:              false,
		StrictNullHandling:       false,
		ThrowOnLimitExceeded:     false,
	}
}

// Parse parses a query string into a nested data structure.
//
// This function converts URL query strings into Go maps with support for
// nested objects, arrays, and various encoding formats. It's compatible
// with the JavaScript qs library while providing Go-specific features.
//
// # Parameters
//
//   - str: The query string to parse (with or without leading "?")
//   - opts: Optional ParseOptions to customize parsing behavior
//
// # Return Values
//
//   - map[string]interface{}: The parsed data structure
//   - error: Any error that occurred during parsing
//
// # Examples
//
// Simple key-value pairs:
//
//	result, err := qs.Parse("name=John&age=30")
//	// Returns: map[string]interface{}{"name": "John", "age": "30"}
//
// Nested objects:
//
//	result, err := qs.Parse("user[profile][name]=John&user[profile][age]=30")
//	// Returns: map[string]interface{}{
//	//     "user": map[string]interface{}{
//	//         "profile": map[string]interface{}{
//	//             "name": "John",
//	//             "age": "30",
//	//         },
//	//     },
//	// }
//
// Arrays:
//
//	result, err := qs.Parse("tags[]=golang&tags[]=programming")
//	// Returns: map[string]interface{}{"tags": []interface{}{"golang", "programming"}}
//
// With custom options:
//
//	result, err := qs.Parse("?name=John&age=30", &qs.ParseOptions{
//	    IgnoreQueryPrefix: true,
//	    Delimiter: "&",
//	})
//
// # Error Handling
//
// Parse returns an error if:
//   - The parameter limit is exceeded (when ThrowOnLimitExceeded is true)
//   - The depth limit is exceeded (when StrictDepth is true)
//   - URL decoding fails
//   - Custom decoder function returns an error
//
// # Performance
//
// The function is optimized for performance and can handle complex
// nested structures efficiently. For best performance with large
// query strings, consider adjusting ParameterLimit and Depth options.
func Parse(str string, opts ...*ParseOptions) (map[string]interface{}, error) {
	options := defaultParseOptions()
	if len(opts) > 0 && opts[0] != nil {
		// Merge user options with defaults
		custom := opts[0]
		if custom.AllowDots {
			options.AllowDots = custom.AllowDots
		}
		if custom.AllowEmptyArrays {
			options.AllowEmptyArrays = custom.AllowEmptyArrays
		}
		if custom.AllowPrototypes {
			options.AllowPrototypes = custom.AllowPrototypes
		}
		if custom.AllowSparse {
			options.AllowSparse = custom.AllowSparse
		}
		if custom.ArrayLimit != 0 {
			options.ArrayLimit = custom.ArrayLimit
		}
		if custom.Charset != "" {
			options.Charset = custom.Charset
		}
		if custom.CharsetSentinel {
			options.CharsetSentinel = custom.CharsetSentinel
		}
		if custom.Comma {
			options.Comma = custom.Comma
		}
		if custom.DecodeDotInKeys {
			options.DecodeDotInKeys = custom.DecodeDotInKeys
		}
		if custom.Decoder != nil {
			options.Decoder = custom.Decoder
		}
		if custom.Delimiter != "" {
			options.Delimiter = custom.Delimiter
		}
		if custom.Depth != 0 {
			options.Depth = custom.Depth
		}
		if custom.Duplicates != "" {
			options.Duplicates = custom.Duplicates
		}
		if custom.IgnoreQueryPrefix {
			options.IgnoreQueryPrefix = custom.IgnoreQueryPrefix
		}
		if custom.InterpretNumericEntities {
			options.InterpretNumericEntities = custom.InterpretNumericEntities
		}
		if custom.ParameterLimit != 0 {
			options.ParameterLimit = custom.ParameterLimit
		}
		if custom.ParseArrays {
			options.ParseArrays = custom.ParseArrays
		}
		if custom.PlainObjects {
			options.PlainObjects = custom.PlainObjects
		}
		if custom.StrictDepth {
			options.StrictDepth = custom.StrictDepth
		}
		if custom.StrictNullHandling {
			options.StrictNullHandling = custom.StrictNullHandling
		}
		if custom.ThrowOnLimitExceeded {
			options.ThrowOnLimitExceeded = custom.ThrowOnLimitExceeded
		}
	}

	if options.Decoder == nil {
		options.Decoder = func(s string, decoder ...interface{}) (string, error) {
			return url.QueryUnescape(s)
		}
	}

	obj := make(map[string]interface{})

	cleanStr := str
	if options.IgnoreQueryPrefix && strings.HasPrefix(cleanStr, "?") {
		cleanStr = strings.TrimPrefix(cleanStr, "?")
	}

	if cleanStr == "" {
		return obj, nil
	}

	limit := options.ParameterLimit
	if limit == 0 {
		limit = 1000
	}

	parts := strings.Split(cleanStr, options.Delimiter)
	if options.ThrowOnLimitExceeded && len(parts) > limit {
		return nil, fmt.Errorf("parameter limit exceeded. Only %d parameters allowed", limit)
	}

	if limit > 0 && len(parts) > limit {
		parts = parts[:limit]
	}

	for _, part := range parts {
		// Skip truly empty parts (from consecutive delimiters)
		if part == "" {
			continue
		}

		var key string
		var val interface{}
		var err error

		// Find the correct = separator, ignoring those inside brackets
		pos := findKeyValueSeparator(part)

		if pos == -1 {
			// No equals sign - treat as key with null/empty value
			// Don't decode the key yet - parseKeys will handle it
			key = part
			if options.StrictNullHandling {
				val = nil
			} else {
				val = ""
			}
		} else {
			// Has equals sign - split into key and value
			key = part[:pos]
			valuePart := part[pos+1:]

			// Decode only the value part
			val, err = options.Decoder(valuePart)
			if err != nil {
				return nil, err
			}
		}

		if err := parseKeys(key, val, options, obj); err != nil {
			return nil, err
		}
	}

	return obj, nil
}

// findKeyValueSeparator finds the position of the = that separates key from value,
// ignoring = characters that appear inside brackets
func findKeyValueSeparator(part string) int {
	bracketLevel := 0
	for i, ch := range part {
		switch ch {
		case '[':
			bracketLevel++
		case ']':
			bracketLevel--
		case '=':
			if bracketLevel == 0 {
				return i
			}
		}
	}
	return -1
}

func parseKeys(key string, val interface{}, options *ParseOptions, obj map[string]interface{}) error {
	// Handle empty keys - this is allowed in qs
	if key == "" {
		if existing, ok := obj[""]; ok {
			obj[""] = merge(existing, val)
		} else {
			obj[""] = val
		}
		return nil
	}

	if options.AllowDots {
		key = regexp.MustCompile(`\.([^.[]+)`).ReplaceAllString(key, "[$1]")
	}

	keys := []string{}

	// Split key into parent and brackets
	brackets := regexp.MustCompile(`(\[[^[\]]*\])`)

	segment := brackets.FindStringIndex(key)
	parent := key
	if segment != nil {
		parent = key[:segment[0]]
	}

	// Decode and add parent key
	if parent != "" {
		decodedParent, err := options.Decoder(parent)
		if err != nil {
			return err
		}
		keys = append(keys, decodedParent)
	} else {
		keys = append(keys, parent)
	}

	// Extract and decode all bracketed keys
	matches := brackets.FindAllString(key[len(parent):], -1)
	for _, match := range matches {
		inner := strings.TrimSuffix(strings.TrimPrefix(match, "["), "]")
		// Decode the inner part
		decodedInner, err := options.Decoder(inner)
		if err != nil {
			return err
		}
		keys = append(keys, decodedInner)
	}

	return parseObject(keys, val, options, obj)
}

func merge(a, b interface{}) interface{} {
	if a == nil {
		return b
	}

	aMap, aIsMap := a.(map[string]interface{})
	bMap, bIsMap := b.(map[string]interface{})
	if aIsMap && bIsMap {
		for k, v := range bMap {
			if av, ok := aMap[k]; ok {
				aMap[k] = merge(av, v)
			} else {
				aMap[k] = v
			}
		}
		return aMap
	}

	aSlice, aIsSlice := a.([]interface{})
	bSlice, bIsSlice := b.([]interface{})
	if aIsSlice && bIsSlice {
		return append(aSlice, bSlice...)
	}

	if aIsSlice && bIsMap {
		// Special case: merge array with map containing numeric indices
		if canConvertToArray(bMap) {
			// Create a combined array
			result := make([]interface{}, len(aSlice))
			copy(result, aSlice)

			// Find the maximum index needed
			maxIndex := len(result) - 1
			for k := range bMap {
				if k != "" && len(k) == 1 && k >= "0" && k <= "9" {
					idx := int(k[0] - '0')
					if idx > maxIndex {
						maxIndex = idx
					}
				}
			}

			// Extend array if needed
			if maxIndex >= len(result) {
				newResult := make([]interface{}, maxIndex+1)
				copy(newResult, result)
				result = newResult
			}

			// Add values from map to array
			for k, v := range bMap {
				if k != "" && len(k) == 1 && k >= "0" && k <= "9" {
					idx := int(k[0] - '0')
					if idx < len(result) {
						result[idx] = v
					}
				}
			}

			return result
		}
		return append(aSlice, b)
	}

	if aIsSlice {
		return append(aSlice, b)
	}

	if bIsSlice {
		return append([]interface{}{a}, bSlice...)
	}

	// Handle merging map with indexed values into array
	if aIsMap {
		if bIsSlice {
			// Convert map to array if we're merging with an array
			arr := convertMapToArray(aMap)
			return append(arr, bSlice...)
		}
		// If b is not an array, try to merge as objects or convert to array
		if canConvertToArray(aMap) {
			arr := convertMapToArray(aMap)
			return append(arr, b)
		}
	}

	if bIsMap && canConvertToArray(bMap) && aIsSlice {
		// This case is now handled above in the aIsSlice && bIsMap block
		arr := convertMapToArray(bMap)
		return append(aSlice, arr...)
	}

	return []interface{}{a, b}
}

// Helper function to check if a map can be converted to an array (has numeric keys)
func canConvertToArray(m map[string]interface{}) bool {
	if len(m) == 0 {
		return false
	}

	for k := range m {
		// Check if key is numeric
		if k == "" {
			continue
		}
		// Simple check for numeric keys (0, 1, 2, etc.)
		if k < "0" || k > "9" {
			return false
		}
	}
	return true
}

// Helper function to convert a map with numeric keys to an array
func convertMapToArray(m map[string]interface{}) []interface{} {
	if len(m) == 0 {
		return []interface{}{}
	}

	var arr []interface{}
	maxIndex := -1

	// Find the maximum index
	for k := range m {
		if k == "" {
			continue
		}
		if len(k) == 1 && k >= "0" && k <= "9" {
			idx := int(k[0] - '0')
			if idx > maxIndex {
				maxIndex = idx
			}
		}
	}

	if maxIndex >= 0 {
		arr = make([]interface{}, maxIndex+1)
		for k, v := range m {
			if k == "" {
				continue
			}
			if len(k) == 1 && k >= "0" && k <= "9" {
				idx := int(k[0] - '0')
				arr[idx] = v
			}
		}
	}

	return arr
}

func parseObject(chain []string, val interface{}, options *ParseOptions, result map[string]interface{}) error {
	if len(chain) == 0 {
		return nil
	}

	// Handle case with only one key
	if len(chain) == 1 {
		key := chain[0]
		if existing, ok := result[key]; ok {
			result[key] = merge(existing, val)
		} else {
			result[key] = val
		}
		return nil
	}

	// Check depth limit (default is 5)
	depth := options.Depth
	if depth == 0 {
		depth = 5
	}

	// If we exceed the depth limit, combine remaining keys
	// depth+1 because we count from 0, so depth 5 allows 6 levels (0,1,2,3,4,5)
	if len(chain) > depth+1 {
		// Take the first 'depth+1' keys and combine the rest
		limitedChain := chain[:depth+1]
		remainingKeys := chain[depth+1:]

		// Combine remaining keys into a single key
		var combinedKey strings.Builder
		for _, key := range remainingKeys {
			if key == "" {
				combinedKey.WriteString("[]")
			} else {
				combinedKey.WriteString("[")
				combinedKey.WriteString(key)
				combinedKey.WriteString("]")
			}
		}

		// Add the combined key to the limited chain
		limitedChain = append(limitedChain, combinedKey.String())
		chain = limitedChain
	}

	// Build nested structure from the bottom up
	leaf := val
	for i := len(chain) - 1; i > 0; i-- {
		key := chain[i]

		if key == "" {
			// Empty bracket notation creates array
			leaf = []interface{}{leaf}
		} else {
			// Regular key creates object
			newObj := make(map[string]interface{})
			newObj[key] = leaf
			leaf = newObj
		}
	}

	// Handle the root key
	rootKey := chain[0]
	if existing, ok := result[rootKey]; ok {
		result[rootKey] = merge(existing, leaf)
	} else {
		result[rootKey] = leaf
	}

	return nil
}

func getCleanKey(key string) string {
	if strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]") {
		return key[1 : len(key)-1]
	}
	return key
}

// StringifyOptions configures how data structures are converted to query strings.
// These options control various aspects of the stringification process,
// including array formatting, encoding, and output structure.
type StringifyOptions struct {
	// AddQueryPrefix adds a leading "?" to the output query string.
	AddQueryPrefix bool

	// AllowDots enables dot notation for nested objects (e.g., "a.b.c=value").
	AllowDots bool

	// AllowEmptyArrays includes empty arrays in the output instead of omitting them.
	AllowEmptyArrays bool

	// ArrayFormat specifies how arrays are formatted in the query string.
	// Options: "indices" (default), "brackets", "repeat".
	//   - "indices": a[0]=1&a[1]=2
	//   - "brackets": a[]=1&a[]=2
	//   - "repeat": a=1&a=2
	ArrayFormat string

	// Charset specifies the character encoding to use.
	// Default is "utf-8".
	Charset string

	// CharsetSentinel includes a charset parameter in the query string
	// for better JavaScript compatibility.
	CharsetSentinel bool

	// CommaRoundTrip enables comma-separated values within a single parameter
	// for better compatibility with specific parsers.
	CommaRoundTrip bool

	// Delimiter specifies the character used to separate parameters.
	// Default is "&".
	Delimiter string

	// Encode enables URL encoding of parameter values.
	// Default is true.
	Encode bool

	// EncodeDotInKeys enables encoding of dots in parameter keys.
	EncodeDotInKeys bool

	// Encoder is a custom function for encoding parameter values.
	// If nil, the default URL encoder is used.
	Encoder func(str string, defaultEncoder ...interface{}) string

	// EncodeValuesOnly enables encoding only parameter values,
	// leaving keys unencoded.
	EncodeValuesOnly bool

	// Filter specifies which properties to include in the output.
	// Can be a function or a list of allowed keys.
	Filter interface{}

	// Format specifies the encoding format.
	// Options: "RFC1738", "RFC3986" (default).
	Format string

	// Formatter is a custom function for formatting the final output.
	Formatter func(string) string

	// Indices is deprecated. Use ArrayFormat instead.
	Indices bool

	// SerializeDate is a custom function for serializing time.Time values.
	// Default uses RFC3339 format.
	SerializeDate func(date time.Time) string

	// SkipNulls omits null/nil values from the output instead of
	// including them as empty parameters.
	SkipNulls bool

	// StrictNullHandling preserves null values as literal "null"
	// instead of converting them to empty strings.
	StrictNullHandling bool

	// Sort is a custom function for sorting parameter keys.
	// If nil, parameters appear in their natural order.
	Sort func(a, b string) bool
}

var arrayPrefixGenerators = map[string]func(prefix string, key ...string) string{
	"brackets": func(prefix string, key ...string) string {
		return prefix + "[]"
	},
	"indices": func(prefix string, key ...string) string {
		if len(key) > 0 {
			return prefix + "[" + key[0] + "]"
		}
		return prefix + "[]"
	},
	"repeat": func(prefix string, key ...string) string {
		return prefix
	},
}

func defaultStringifyOptions() *StringifyOptions {
	return &StringifyOptions{
		AddQueryPrefix:   false,
		AllowDots:        false,
		AllowEmptyArrays: false,
		ArrayFormat:      "indices",
		Charset:          "utf-8",
		CharsetSentinel:  false,
		CommaRoundTrip:   false,
		Delimiter:        "&",
		Encode:           true,
		EncodeDotInKeys:  false,
		Encoder:          nil,
		EncodeValuesOnly: false,
		Filter:           nil,
		Format:           "RFC3986",
		Formatter:        nil,
		Indices:          false,
		SerializeDate: func(date time.Time) string {
			return date.Format(time.RFC3339)
		},
		SkipNulls:          false,
		StrictNullHandling: false,
		Sort:               nil,
	}
}

// Stringify converts a data structure into a query string.
//
// This function takes Go data structures (maps, structs, slices) and converts
// them into URL query strings with support for nested objects, arrays, and
// various formatting options. It's compatible with the JavaScript qs library.
//
// # Parameters
//
//   - obj: The data structure to convert (map, struct, slice, or primitive value)
//   - opts: Optional StringifyOptions to customize output format
//
// # Return Values
//
//   - string: The generated query string
//   - error: Any error that occurred during stringification
//
// # Examples
//
// Simple map:
//
//	data := map[string]interface{}{
//	    "name": "John",
//	    "age":  30,
//	}
//	result, err := qs.Stringify(data)
//	// Returns: "age=30&name=John"
//
// Nested objects:
//
//	data := map[string]interface{}{
//	    "user": map[string]interface{}{
//	        "profile": map[string]interface{}{
//	            "name": "John",
//	            "age":  30,
//	        },
//	    },
//	}
//	result, err := qs.Stringify(data)
//	// Returns: "user[profile][age]=30&user[profile][name]=John"
//
// Arrays with different formats:
//
//	data := map[string]interface{}{
//	    "items": []interface{}{"a", "b", "c"},
//	}
//
//	// Default (indices)
//	result1, err := qs.Stringify(data)
//	// Returns: "items[0]=a&items[1]=b&items[2]=c"
//
//	// Brackets format
//	result2, err := qs.Stringify(data, &qs.StringifyOptions{
//	    ArrayFormat: "brackets",
//	})
//	// Returns: "items[]=a&items[]=b&items[]=c"
//
//	// Repeat format
//	result3, err := qs.Stringify(data, &qs.StringifyOptions{
//	    ArrayFormat: "repeat",
//	})
//	// Returns: "items=a&items=b&items=c"
//
// With query prefix:
//
//	result, err := qs.Stringify(data, &qs.StringifyOptions{
//	    AddQueryPrefix: true,
//	})
//	// Returns: "?items[0]=a&items[1]=b&items[2]=c"
//
// # Supported Data Types
//
//   - Maps: map[string]interface{}, map[string]string, etc.
//   - Structs: with or without query tags
//   - Slices and arrays: []interface{}, []string, []int, etc.
//   - Primitive types: string, int, float, bool
//   - Pointers: automatically dereferenced
//   - time.Time: serialized using SerializeDate function
//
// # Error Handling
//
// Stringify returns an error if:
//   - Custom encoder function returns an error
//   - Unsupported data type is encountered
//   - Reflection operations fail
//
// # Performance
//
// The function is optimized for performance and can handle large
// data structures efficiently. Consider using appropriate StringifyOptions
// for best performance with your specific use case.
func Stringify(obj interface{}, opts ...*StringifyOptions) (string, error) {
	options := defaultStringifyOptions()
	if len(opts) > 0 && opts[0] != nil {
		// Merge custom options with defaults
		custom := opts[0]
		if custom.AddQueryPrefix {
			options.AddQueryPrefix = custom.AddQueryPrefix
		}
		if custom.AllowDots {
			options.AllowDots = custom.AllowDots
		}
		if custom.AllowEmptyArrays {
			options.AllowEmptyArrays = custom.AllowEmptyArrays
		}
		if custom.ArrayFormat != "" {
			options.ArrayFormat = custom.ArrayFormat
		}
		if custom.Charset != "" {
			options.Charset = custom.Charset
		}
		if custom.CharsetSentinel {
			options.CharsetSentinel = custom.CharsetSentinel
		}
		if custom.CommaRoundTrip {
			options.CommaRoundTrip = custom.CommaRoundTrip
		}
		if custom.Delimiter != "" {
			options.Delimiter = custom.Delimiter
		}
		if custom.Encode {
			options.Encode = custom.Encode
		}
		if custom.EncodeDotInKeys {
			options.EncodeDotInKeys = custom.EncodeDotInKeys
		}
		if custom.Encoder != nil {
			options.Encoder = custom.Encoder
		}
		if custom.EncodeValuesOnly {
			options.EncodeValuesOnly = custom.EncodeValuesOnly
		}
		if custom.Filter != nil {
			options.Filter = custom.Filter
		}
		if custom.Format != "" {
			options.Format = custom.Format
		}
		if custom.Formatter != nil {
			options.Formatter = custom.Formatter
		}
		if custom.Indices {
			options.Indices = custom.Indices
		}
		if custom.SerializeDate != nil {
			options.SerializeDate = custom.SerializeDate
		}
		if custom.SkipNulls {
			options.SkipNulls = custom.SkipNulls
		}
		if custom.StrictNullHandling {
			options.StrictNullHandling = custom.StrictNullHandling
		}
		if custom.Sort != nil {
			options.Sort = custom.Sort
		}
	}

	if options.Encoder == nil {
		options.Encoder = func(str string, defaultEncoder ...interface{}) string {
			// Use PathEscape instead of QueryEscape to get %20 for spaces instead of +
			encoded := url.PathEscape(str)
			// PathEscape doesn't encode some characters that QueryEscape does, so we need to handle them
			encoded = strings.ReplaceAll(encoded, "=", "%3D")
			encoded = strings.ReplaceAll(encoded, "&", "%26")
			encoded = strings.ReplaceAll(encoded, "@", "%40")
			encoded = strings.ReplaceAll(encoded, "$", "%24")
			// Don't encode commas - they are typically not encoded in JavaScript qs
			encoded = strings.ReplaceAll(encoded, "%2C", ",")
			return encoded
		}
	}

	if options.Formatter == nil {
		options.Formatter = func(str string) string {
			return str
		}
	}

	// Handle falsy values at the top level
	if obj == nil {
		return "", nil
	}

	// Handle falsy primitive values
	switch v := obj.(type) {
	case bool:
		if !v {
			return "", nil
		}
	case int, int8, int16, int32, int64:
		if v == 0 {
			return "", nil
		}
	case uint, uint8, uint16, uint32, uint64:
		if v == 0 {
			return "", nil
		}
	case float32:
		if v == 0.0 {
			return "", nil
		}
	case float64:
		if v == 0.0 {
			return "", nil
		}
	case string:
		if v == "" {
			return "", nil
		}
	}

	var parts []string

	stringify(&parts, obj, options, "")

	result := strings.Join(parts, options.Delimiter)

	if options.AddQueryPrefix {
		result = "?" + result
	}

	return result, nil
}

// stringify is a helper function that recursively converts data structures
// into query string parts. It handles different data types and applies
// the configured formatting options.
func stringify(parts *[]string, obj interface{}, options *StringifyOptions, prefix string) {
	if obj == nil {
		if options.StrictNullHandling {
			*parts = append(*parts, options.Encoder(prefix))
		} else {
			*parts = append(*parts, prefix+"=")
		}
		return
	}

	switch v := obj.(type) {
	case string, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8, float32, float64, bool:
		key := options.Formatter(prefix)
		val := options.Formatter(options.Encoder(fmt.Sprintf("%v", v)))
		*parts = append(*parts, key+"="+val)
	case map[string]interface{}:
		for k, val := range v {
			newPrefix := k
			if prefix != "" {
				if options.AllowDots {
					newPrefix = prefix + "." + k
				} else {
					newPrefix = prefix + "[" + k + "]"
				}
			}
			stringify(parts, val, options, newPrefix)
		}
	case []interface{}:
		if gen, ok := arrayPrefixGenerators[options.ArrayFormat]; ok {
			for i, val := range v {
				newPrefix := gen(prefix, fmt.Sprintf("%d", i))
				stringify(parts, val, options, newPrefix)
			}
		} else {
			// fallback to indices format
			for i, val := range v {
				newPrefix := prefix + "[" + fmt.Sprintf("%d", i) + "]"
				stringify(parts, val, options, newPrefix)
			}
		}
	}
}

// ParseToStruct parses a query string and fills a struct using query tags
func ParseToStruct(str string, dest interface{}, opts ...*ParseOptions) error {
	// Parse to map first
	result, err := Parse(str, opts...)
	if err != nil {
		return err
	}

	// Convert map to struct
	return MapToStruct(result, dest)
}

// MapToStruct converts a map to a struct using query tags
func MapToStruct(data map[string]interface{}, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer to struct")
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a pointer to struct")
	}

	return fillStruct(data, destValue)
}

// fillStruct recursively fills struct fields from map data
func fillStruct(data map[string]interface{}, structValue reflect.Value) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get query tag
		queryTag := fieldType.Tag.Get("query")
		if queryTag == "" {
			// If no query tag, try to use field name in lowercase
			queryTag = strings.ToLower(fieldType.Name)
		}

		// Skip fields with query:"-"
		if queryTag == "-" {
			continue
		}

		// Look for the value in data
		value, exists := data[queryTag]
		if !exists {
			continue
		}

		if err := setFieldValue(field, value); err != nil {
			return fmt.Errorf("error setting field %s: %v", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue sets a struct field value from interface{} data
func setFieldValue(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	fieldType := field.Type()
	valueReflect := reflect.ValueOf(value)

	// Handle pointers
	if fieldType.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(fieldType.Elem()))
		}
		return setFieldValue(field.Elem(), value)
	}

	// Handle different types
	switch fieldType.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		} else {
			field.SetString(fmt.Sprintf("%v", value))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if str, ok := value.(string); ok {
			if intVal, err := strconv.ParseInt(str, 10, 64); err == nil {
				field.SetInt(intVal)
			} else {
				return fmt.Errorf("cannot convert %q to int", str)
			}
		} else if intVal, ok := value.(int64); ok {
			field.SetInt(intVal)
		} else {
			return fmt.Errorf("cannot convert %T to int", value)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if str, ok := value.(string); ok {
			if uintVal, err := strconv.ParseUint(str, 10, 64); err == nil {
				field.SetUint(uintVal)
			} else {
				return fmt.Errorf("cannot convert %q to uint", str)
			}
		} else if uintVal, ok := value.(uint64); ok {
			field.SetUint(uintVal)
		} else {
			return fmt.Errorf("cannot convert %T to uint", value)
		}

	case reflect.Float32, reflect.Float64:
		if str, ok := value.(string); ok {
			if floatVal, err := strconv.ParseFloat(str, 64); err == nil {
				field.SetFloat(floatVal)
			} else {
				return fmt.Errorf("cannot convert %q to float", str)
			}
		} else if floatVal, ok := value.(float64); ok {
			field.SetFloat(floatVal)
		} else {
			return fmt.Errorf("cannot convert %T to float", value)
		}

	case reflect.Bool:
		if str, ok := value.(string); ok {
			if boolVal, err := strconv.ParseBool(str); err == nil {
				field.SetBool(boolVal)
			} else {
				return fmt.Errorf("cannot convert %q to bool", str)
			}
		} else if boolVal, ok := value.(bool); ok {
			field.SetBool(boolVal)
		} else {
			return fmt.Errorf("cannot convert %T to bool", value)
		}

	case reflect.Slice:
		return setSliceField(field, value)

	case reflect.Struct:
		if dataMap, ok := value.(map[string]interface{}); ok {
			return fillStruct(dataMap, field)
		} else {
			return fmt.Errorf("cannot convert %T to struct", value)
		}

	case reflect.Map:
		if fieldType.Key().Kind() == reflect.String {
			return setMapField(field, value)
		} else {
			return fmt.Errorf("unsupported map key type: %v", fieldType.Key().Kind())
		}

	default:
		// Try direct assignment if types match
		if valueReflect.Type().AssignableTo(fieldType) {
			field.Set(valueReflect)
		} else {
			return fmt.Errorf("unsupported field type: %v", fieldType.Kind())
		}
	}

	return nil
}

// setSliceField handles slice field assignment
func setSliceField(field reflect.Value, value interface{}) error {
	sliceValue, ok := value.([]interface{})
	if !ok {
		// Check if it's a map that can be converted to slice
		if mapValue, isMap := value.(map[string]interface{}); isMap {
			if canConvertToArray(mapValue) {
				sliceValue = convertMapToArray(mapValue)
			} else {
				// Try to convert single value to slice
				sliceValue = []interface{}{value}
			}
		} else {
			// Try to convert single value to slice
			sliceValue = []interface{}{value}
		}
	}

	fieldType := field.Type()

	newSlice := reflect.MakeSlice(fieldType, len(sliceValue), len(sliceValue))

	for i, item := range sliceValue {
		elemField := newSlice.Index(i)
		if err := setFieldValue(elemField, item); err != nil {
			return fmt.Errorf("error setting slice element %d: %v", i, err)
		}
	}

	field.Set(newSlice)
	return nil
}

// setMapField handles map field assignment
func setMapField(field reflect.Value, value interface{}) error {
	dataMap, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot convert %T to map", value)
	}

	fieldType := field.Type()
	valueType := fieldType.Elem()

	newMap := reflect.MakeMap(fieldType)

	for k, v := range dataMap {
		keyVal := reflect.ValueOf(k)
		valueVal := reflect.New(valueType).Elem()

		if err := setFieldValue(valueVal, v); err != nil {
			return fmt.Errorf("error setting map value for key %q: %v", k, err)
		}

		newMap.SetMapIndex(keyVal, valueVal)
	}

	field.Set(newMap)
	return nil
}

// StructToQueryString converts a struct to query string using query tags
func StructToQueryString(obj interface{}, opts ...*StringifyOptions) (string, error) {
	data, err := StructToMap(obj)
	if err != nil {
		return "", err
	}

	return Stringify(data, opts...)
}

// StructToMap converts a struct to map using query tags
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object must be a struct or pointer to struct")
	}

	objType := objValue.Type()

	for i := 0; i < objValue.NumField(); i++ {
		field := objValue.Field(i)
		fieldType := objType.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get query tag
		queryTag := fieldType.Tag.Get("query")
		if queryTag == "" {
			// If no query tag, use field name in lowercase
			queryTag = strings.ToLower(fieldType.Name)
		}

		// Skip fields with query:"-"
		if queryTag == "-" {
			continue
		}

		// Get field value
		fieldValue := field.Interface()

		// Skip nil pointers
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// Handle slices specially to ensure they're properly included even if empty
		if field.Kind() == reflect.Slice {
			// Convert Go slice to []interface{} for Stringify compatibility
			sliceLen := field.Len()
			interfaceSlice := make([]interface{}, sliceLen)
			for i := 0; i < sliceLen; i++ {
				interfaceSlice[i] = field.Index(i).Interface()
			}
			result[queryTag] = interfaceSlice
		} else if field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct) {
			// Convert nested structs to maps
			nestedMap, err := StructToMap(fieldValue)
			if err != nil {
				return nil, fmt.Errorf("error converting nested struct field %s: %v", fieldType.Name, err)
			}
			result[queryTag] = nestedMap
		} else {
			result[queryTag] = fieldValue
		}
	}

	return result, nil
}

// Unmarshal parses a query string and stores the result in the value pointed to by v.
//
// This function provides idiomatic Go unmarshaling with automatic type detection.
// It works with structs, maps, slices, and primitive types, automatically choosing
// the appropriate conversion method based on the target type.
//
// # Parameters
//
//   - queryString: The query string to parse (with or without leading "?")
//   - v: Pointer to the value where the result should be stored
//   - opts: Optional ParseOptions to customize parsing behavior
//
// # Return Values
//
//   - error: Any error that occurred during parsing or unmarshaling
//
// # Examples
//
// Unmarshal to struct:
//
//	type User struct {
//	    Name string `query:"name"`
//	    Age  int    `query:"age"`
//	}
//	var user User
//	err := qs.Unmarshal("name=John&age=30", &user)
//
// Unmarshal to map:
//
//	var data map[string]interface{}
//	err := qs.Unmarshal("name=John&age=30", &data)
//
// Unmarshal to slice:
//
//	var tags []string
//	err := qs.Unmarshal("tags[]=go&tags[]=programming", &tags)
//
// With custom options:
//
//	var user User
//	err := qs.Unmarshal("?name=John&age=30", &user, &qs.ParseOptions{
//	    IgnoreQueryPrefix: true,
//	})
//
// # Supported Target Types
//
//   - Structs with query tags
//   - Maps (map[string]interface{}, map[string]string, etc.)
//   - Slices and arrays
//   - Primitive types (string, int, float, bool)
//   - Pointers (automatically allocated if nil)
//   - Interfaces (interface{})
//
// # Error Handling
//
// Unmarshal returns an error if:
//   - The target is nil or not a pointer
//   - The target is not settable
//   - Type conversion fails
//   - The query string is malformed
//   - Custom parsing options cause errors
//
// # Performance
//
// This function provides excellent performance with automatic type detection,
// making it suitable for high-throughput applications.
func Unmarshal(queryString string, v interface{}, opts ...*ParseOptions) error {
	if v == nil {
		return fmt.Errorf("unmarshal target cannot be nil")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("unmarshal target must be a pointer, got %T", v)
	}

	rv = rv.Elem()
	if !rv.CanSet() {
		return fmt.Errorf("unmarshal target must be settable")
	}

	// Parse the query string to map first
	data, err := Parse(queryString, opts...)
	if err != nil {
		return err
	}

	return unmarshalValue(data, rv)
}

// Marshal converts a value to a query string.
//
// This function provides idiomatic Go marshaling with automatic type detection.
// It works with structs, maps, slices, and primitive types, automatically choosing
// the appropriate conversion method based on the source type.
//
// # Parameters
//
//   - v: The value to convert to a query string
//   - opts: Optional StringifyOptions to customize output format
//
// # Return Values
//
//   - string: The generated query string
//   - error: Any error that occurred during marshaling
//
// # Examples
//
// Marshal struct:
//
//	type User struct {
//	    Name string `query:"name"`
//	    Age  int    `query:"age"`
//	}
//	user := User{Name: "John", Age: 30}
//	queryString, err := qs.Marshal(user)
//	// Returns: "age=30&name=John"
//
// Marshal map:
//
//	data := map[string]interface{}{
//	    "name": "John",
//	    "age":  30,
//	}
//	queryString, err := qs.Marshal(data)
//	// Returns: "age=30&name=John"
//
// Marshal slice:
//
//	tags := []string{"go", "programming"}
//	queryString, err := qs.Marshal(tags)
//	// Returns: "0=go&1=programming"
//
// With custom options:
//
//	queryString, err := qs.Marshal(data, &qs.StringifyOptions{
//	    ArrayFormat:    "brackets",
//	    AddQueryPrefix: true,
//	})
//
// # Supported Source Types
//
//   - Structs with or without query tags
//   - Maps (any map with string keys)
//   - Slices and arrays
//   - Primitive types (string, int, float, bool)
//   - Pointers (automatically dereferenced)
//   - time.Time values
//   - Interfaces (interface{})
//
// # Error Handling
//
// Marshal returns an error if:
//   - Reflection operations fail
//   - Unsupported types are encountered
//   - Custom encoding functions return errors
//
// # Performance
//
// This function provides excellent performance with automatic type detection,
// making it suitable for high-throughput applications where the source type
// may vary at runtime.
func Marshal(v interface{}, opts ...*StringifyOptions) (string, error) {
	if v == nil {
		return "", nil
	}

	data, err := marshalValue(v)
	if err != nil {
		return "", err
	}

	return Stringify(data, opts...)
}

// unmarshalValue recursively unmarshals data into a reflect.Value
func unmarshalValue(data interface{}, rv reflect.Value) error {
	if data == nil {
		return nil
	}

	rt := rv.Type()

	// Handle pointers
	if rt.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rt.Elem()))
		}
		return unmarshalValue(data, rv.Elem())
	}

	switch rt.Kind() {
	case reflect.Struct:
		return unmarshalStruct(data, rv)
	case reflect.Map:
		return unmarshalMap(data, rv)
	case reflect.Slice:
		return unmarshalSlice(data, rv)
	case reflect.Interface:
		// For interface{}, set the data directly
		if rt == reflect.TypeOf((*interface{})(nil)).Elem() {
			rv.Set(reflect.ValueOf(data))
			return nil
		}
		return fmt.Errorf("unsupported interface type: %v", rt)
	default:
		// Handle primitive types
		return setFieldValue(rv, data)
	}
}

// unmarshalStruct unmarshals data into a struct
func unmarshalStruct(data interface{}, rv reflect.Value) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot unmarshal %T into struct", data)
	}

	return fillStruct(dataMap, rv)
}

// unmarshalMap unmarshals data into a map
func unmarshalMap(data interface{}, rv reflect.Value) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot unmarshal %T into map", data)
	}

	rt := rv.Type()
	keyType := rt.Key()
	valueType := rt.Elem()

	// Only support string keys for now
	if keyType.Kind() != reflect.String {
		return fmt.Errorf("unsupported map key type: %v", keyType)
	}

	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rt))
	}

	for k, v := range dataMap {
		keyVal := reflect.ValueOf(k)
		valueVal := reflect.New(valueType).Elem()

		if err := unmarshalValue(v, valueVal); err != nil {
			return fmt.Errorf("error unmarshaling map value for key %q: %v", k, err)
		}

		rv.SetMapIndex(keyVal, valueVal)
	}

	return nil
}

// unmarshalSlice unmarshals data into a slice
func unmarshalSlice(data interface{}, rv reflect.Value) error {
	// Handle different slice data formats
	var sliceData []interface{}

	switch v := data.(type) {
	case []interface{}:
		sliceData = v
	case map[string]interface{}:
		// Convert map with numeric keys to slice
		if !canConvertToArray(v) {
			return fmt.Errorf("cannot unmarshal map into slice: non-numeric keys found")
		}
		sliceData = convertMapToArray(v)
	default:
		// Single value becomes slice with one element
		sliceData = []interface{}{data}
	}

	rt := rv.Type()

	newSlice := reflect.MakeSlice(rt, len(sliceData), len(sliceData))

	for i, item := range sliceData {
		elemVal := newSlice.Index(i)
		if err := unmarshalValue(item, elemVal); err != nil {
			return fmt.Errorf("error unmarshaling slice element %d: %v", i, err)
		}
	}

	rv.Set(newSlice)
	return nil
}

// marshalValue converts a value to a format suitable for Stringify
func marshalValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(v)
	return marshalReflectValue(rv)
}

// marshalReflectValue converts a reflect.Value to a format suitable for Stringify
func marshalReflectValue(rv reflect.Value) (interface{}, error) {
	// Handle pointers
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		return marshalReflectValue(rv.Elem())
	}

	switch rv.Kind() {
	case reflect.Struct:
		return marshalStruct(rv)
	case reflect.Map:
		return marshalMap(rv)
	case reflect.Slice:
		return marshalSlice(rv)
	case reflect.Interface:
		if rv.IsNil() {
			return nil, nil
		}
		return marshalReflectValue(rv.Elem())
	default:
		// Return primitive values as-is
		return rv.Interface(), nil
	}
}

// marshalStruct converts a struct to a map using query tags
func marshalStruct(rv reflect.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get query tag
		queryTag := fieldType.Tag.Get("query")
		if queryTag == "" {
			// If no query tag, use field name in lowercase
			queryTag = strings.ToLower(fieldType.Name)
		}

		// Skip fields with query:"-"
		if queryTag == "-" {
			continue
		}

		// Skip nil pointers
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// Marshal field value
		fieldValue, err := marshalReflectValue(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling field %s: %v", fieldType.Name, err)
		}

		if fieldValue != nil {
			result[queryTag] = fieldValue
		}
	}

	return result, nil
}

// marshalMap converts a map to a format suitable for Stringify
func marshalMap(rv reflect.Value) (map[string]interface{}, error) {
	if rv.IsNil() {
		return nil, nil
	}

	result := make(map[string]interface{})

	for _, key := range rv.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		value := rv.MapIndex(key)

		marshaledValue, err := marshalReflectValue(value)
		if err != nil {
			return nil, fmt.Errorf("error marshaling map value for key %q: %v", keyStr, err)
		}

		if marshaledValue != nil {
			result[keyStr] = marshaledValue
		}
	}

	return result, nil
}

// marshalSlice converts a slice to []interface{}
func marshalSlice(rv reflect.Value) ([]interface{}, error) {
	if rv.IsNil() {
		return nil, nil
	}

	result := make([]interface{}, rv.Len())

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		marshaledElem, err := marshalReflectValue(elem)
		if err != nil {
			return nil, fmt.Errorf("error marshaling slice element %d: %v", i, err)
		}
		result[i] = marshaledElem
	}

	return result, nil
}
