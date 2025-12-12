// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

// This file contains tests ported from the original JavaScript qs library
// to ensure compatibility with the JS implementation.

package qs

import (
	"strings"
	"testing"
	"time"
)

// The tests below combine:
// - compatibility tests ported from the original JavaScript qs library, and
// - additional Go-specific unit tests for this package.

func TestStringifyBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			name:     "empty object",
			input:    map[string]any{},
			expected: "",
		},
		{
			name:     "simple key-value",
			input:    map[string]any{"a": "b"},
			expected: "a=b",
		},
		{
			name:     "multiple keys",
			input:    map[string]any{"a": "b", "c": "d"},
			opts:     []StringifyOption{WithSort(func(a, b string) bool { return a < b })},
			expected: "a=b&c=d",
		},
		{
			name:     "numeric value",
			input:    map[string]any{"a": 123},
			expected: "a=123",
		},
		{
			name:     "boolean true",
			input:    map[string]any{"a": true},
			expected: "a=true",
		},
		{
			name:     "boolean false",
			input:    map[string]any{"a": false},
			expected: "a=false",
		},
		{
			name:     "empty string value",
			input:    map[string]any{"a": ""},
			expected: "a=",
		},
		{
			name:     "special characters in value",
			input:    map[string]any{"a": "hello world"},
			expected: "a=hello%20world",
		},
		{
			name:     "special characters in key",
			input:    map[string]any{"hello world": "a"},
			expected: "hello%20world=a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStringifyAddQueryPrefix(t *testing.T) {
	result, err := Stringify(map[string]any{"a": "b"}, WithStringifyAddQueryPrefix(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "?a=b"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStringifyNested(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			name:     "nested object - bracket notation",
			input:    map[string]any{"a": map[string]any{"b": "c"}},
			expected: "a%5Bb%5D=c",
		},
		{
			name:     "nested object - dot notation",
			input:    map[string]any{"a": map[string]any{"b": "c"}},
			opts:     []StringifyOption{WithStringifyAllowDots(true)},
			expected: "a.b=c",
		},
		{
			name:     "deeply nested",
			input:    map[string]any{"a": map[string]any{"b": map[string]any{"c": "d"}}},
			expected: "a%5Bb%5D%5Bc%5D=d",
		},
		{
			name:     "deeply nested - dot notation",
			input:    map[string]any{"a": map[string]any{"b": map[string]any{"c": "d"}}},
			opts:     []StringifyOption{WithStringifyAllowDots(true)},
			expected: "a.b.c=d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStringifyArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			name:     "array - indices format (default)",
			input:    map[string]any{"a": []any{"b", "c"}},
			expected: "a%5B0%5D=b&a%5B1%5D=c",
		},
		{
			name:     "array - brackets format",
			input:    map[string]any{"a": []any{"b", "c"}},
			opts:     []StringifyOption{WithArrayFormat(ArrayFormatBrackets)},
			expected: "a%5B%5D=b&a%5B%5D=c",
		},
		{
			name:     "array - repeat format",
			input:    map[string]any{"a": []any{"b", "c"}},
			opts:     []StringifyOption{WithArrayFormat(ArrayFormatRepeat)},
			expected: "a=b&a=c",
		},
		{
			name:     "array - comma format",
			input:    map[string]any{"a": []any{"b", "c"}},
			opts:     []StringifyOption{WithArrayFormat(ArrayFormatComma)},
			expected: "a=b%2Cc",
		},
		{
			name:     "single element array - indices",
			input:    map[string]any{"a": []any{"b"}},
			expected: "a%5B0%5D=b",
		},
		{
			name:     "empty array",
			input:    map[string]any{"a": []any{}},
			expected: "",
		},
		{
			name:     "empty array with allowEmptyArrays",
			input:    map[string]any{"a": []any{}},
			opts:     []StringifyOption{WithStringifyAllowEmptyArrays(true)},
			expected: "a[]", // Note: empty array brackets aren't encoded in key position
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStringifyCommaRoundTrip(t *testing.T) {
	// With commaRoundTrip, single element arrays use [] to preserve array type
	result, err := Stringify(
		map[string]any{"a": []any{"b"}},
		WithArrayFormat(ArrayFormatComma),
		WithCommaRoundTrip(true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a%5B%5D=b"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStringifyEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			name:     "RFC3986 (default) - space as %20",
			input:    map[string]any{"a": "b c"},
			expected: "a=b%20c",
		},
		{
			name:     "RFC1738 - space as +",
			input:    map[string]any{"a": "b c"},
			opts:     []StringifyOption{WithFormat(FormatRFC1738)},
			expected: "a=b+c",
		},
		{
			name:     "encode disabled",
			input:    map[string]any{"a": "b c"},
			opts:     []StringifyOption{WithEncode(false)},
			expected: "a=b c",
		},
		{
			name:     "encodeValuesOnly",
			input:    map[string]any{"a b": "c d"},
			opts:     []StringifyOption{WithEncodeValuesOnly(true)},
			expected: "a b=c%20d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStringifyNullHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			name:     "nil value - default",
			input:    map[string]any{"a": nil},
			expected: "a=",
		},
		{
			name:     "nil value - strictNullHandling",
			input:    map[string]any{"a": nil},
			opts:     []StringifyOption{WithStringifyStrictNullHandling(true)},
			expected: "a",
		},
		{
			name:     "nil value - skipNulls",
			input:    map[string]any{"a": nil, "b": "c"},
			opts:     []StringifyOption{WithSkipNulls(true)},
			expected: "b=c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStringifyFilter(t *testing.T) {
	t.Run("filter with array of keys", func(t *testing.T) {
		input := map[string]any{"a": "1", "b": "2", "c": "3"}
		result, err := Stringify(input, WithFilter([]string{"a", "c"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Order depends on filter array order
		if result != "a=1&c=3" {
			t.Errorf("expected 'a=1&c=3', got %q", result)
		}
	})

	t.Run("filter with function", func(t *testing.T) {
		input := map[string]any{"a": "1", "b": "2"}
		filter := FilterFunc(func(prefix string, value any) any {
			if prefix == "b" {
				return nil // Remove 'b' - in JS, undefined is skipped entirely
			}
			return value
		})
		result, err := Stringify(input, WithFilter(filter), WithSort(func(a, b string) bool { return a < b }))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// In JS, returning undefined from filter skips the key entirely
		// (see lib/stringify.js:137-139: "if (typeof obj === 'undefined') return values")
		if result != "a=1" {
			t.Errorf("expected 'a=1', got %q", result)
		}
	})
}

func TestStringifySort(t *testing.T) {
	input := map[string]any{"c": "1", "a": "2", "b": "3"}
	result, err := Stringify(input, WithSort(func(a, b string) bool {
		return a < b
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=2&b=3&c=1"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStringifySerializeDate(t *testing.T) {
	date := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	input := map[string]any{"date": date}

	t.Run("default date serialization", func(t *testing.T) {
		result, err := Stringify(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Default uses RFC3339
		expected := "date=2023-06-15T12%3A00%3A00Z"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("custom date serialization", func(t *testing.T) {
		result, err := Stringify(input, WithSerializeDate(func(t time.Time) string {
			return t.Format("2006-01-02")
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "date=2023-06-15"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

func TestStringifyDelimiter(t *testing.T) {
	input := map[string]any{"a": "1", "b": "2"}
	result, err := Stringify(input, WithStringifyDelimiter(";"), WithSort(func(a, b string) bool { return a < b }))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=1;b=2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStringifyCharsetSentinel(t *testing.T) {
	t.Run("UTF-8 charset sentinel", func(t *testing.T) {
		result, err := Stringify(map[string]any{"a": "b"}, WithStringifyCharsetSentinel(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "utf8=%E2%9C%93&a=b"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("ISO-8859-1 charset sentinel", func(t *testing.T) {
		result, err := Stringify(
			map[string]any{"a": "b"},
			WithStringifyCharsetSentinel(true),
			WithStringifyCharset(CharsetISO88591),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "utf8=%26%2310003%3B&a=b"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

func TestStringifyEncodeDotInKeys(t *testing.T) {
	input := map[string]any{"a.b": map[string]any{"c.d": "e"}}
	result, err := Stringify(input, WithEncodeDotInKeys(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With encodeDotInKeys, dots in keys are encoded as %2E, then % is encoded as %25
	// So we get %252E (matches JS behavior)
	expected := "a%252Eb.c%252Ed=e"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStringifyCyclicReference(t *testing.T) {
	// Create a cyclic structure
	a := map[string]any{"b": nil}
	a["b"] = a // Self-reference

	_, err := Stringify(a)
	if err != ErrCyclicReference {
		t.Errorf("expected ErrCyclicReference, got %v", err)
	}
}

func TestStringifyCustomEncoder(t *testing.T) {
	encoder := func(str string, charset Charset, kind string, format Format) string {
		// Custom encoder that uppercases values
		if kind == "value" {
			return Encode(str+"!", charset, format)
		}
		return Encode(str, charset, format)
	}

	result, err := Stringify(map[string]any{"a": "b"}, WithEncoder(encoder))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=b%21"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestStringifyNilInput(t *testing.T) {
	result, err := Stringify(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestStringifyOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := DefaultStringifyOptions()
		if opts.Delimiter != "&" {
			t.Errorf("expected default delimiter '&', got %q", opts.Delimiter)
		}
		if opts.Encode != true {
			t.Error("expected Encode to be true by default")
		}
		if opts.ArrayFormat != ArrayFormatIndices {
			t.Errorf("expected ArrayFormatIndices, got %v", opts.ArrayFormat)
		}
		if opts.Format != FormatRFC3986 {
			t.Errorf("expected FormatRFC3986, got %v", opts.Format)
		}
	})

	t.Run("invalid charset", func(t *testing.T) {
		opts := StringifyOptions{Charset: "invalid"}
		_, err := normalizeStringifyOptions(&opts)
		if err != ErrInvalidStringifyCharset {
			t.Errorf("expected ErrInvalidStringifyCharset, got %v", err)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		opts := StringifyOptions{Format: "invalid"}
		_, err := normalizeStringifyOptions(&opts)
		if err != ErrInvalidFormat {
			t.Errorf("expected ErrInvalidFormat, got %v", err)
		}
	})

	t.Run("invalid array format", func(t *testing.T) {
		opts := StringifyOptions{ArrayFormat: "invalid"}
		_, err := normalizeStringifyOptions(&opts)
		if err != ErrInvalidArrayFormat {
			t.Errorf("expected ErrInvalidArrayFormat, got %v", err)
		}
	})
}

// TestStringifyRoundTrip tests that parsing and stringifying produces equivalent results
func TestStringifyRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		query string
		opts  []StringifyOption
	}{
		{
			name:  "simple",
			query: "a=b",
		},
		{
			name:  "nested bracket",
			query: "a%5Bb%5D=c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query string
			parsed, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			// Stringify it back
			result, err := Stringify(parsed, tt.opts...)
			if err != nil {
				t.Fatalf("stringify error: %v", err)
			}

			if result != tt.query {
				t.Errorf("round-trip failed: expected %q, got %q", tt.query, result)
			}
		})
	}
}

// TestJSStringifyQuerystringObject tests basic object stringification
func TestJSStringifyQuerystringObject(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{"simple key-value", map[string]any{"a": "b"}, "a=b"},
		{"numeric value", map[string]any{"a": 1}, "a=1"},
		{"underscore in value", map[string]any{"a": "A_Z"}, "a=A_Z"},
		{"euro sign", map[string]any{"a": "€"}, "a=%E2%82%AC"},
		{"hebrew aleph", map[string]any{"a": "א"}, "a=%D7%90"},
		{"surrogate pair", map[string]any{"a": "𐐷"}, "a=%F0%90%90%B7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyFalsyValues tests stringification of falsy values
func TestJSStringifyFalsyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{"nil input", nil, nil, ""},
		{"nil with strictNullHandling", nil, []StringifyOption{WithStringifyStrictNullHandling(true)}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyEncodeDotInKeys tests encoding of dots in keys
func TestJSStringifyEncodeDotInKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			"allowDots false, encodeDotInKeys false",
			map[string]any{"name.obj": map[string]any{"first": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(false),
				WithEncodeDotInKeys(false),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name.obj%5Bfirst%5D=John&name.obj%5Blast%5D=Doe",
		},
		{
			"allowDots true, encodeDotInKeys false",
			map[string]any{"name.obj": map[string]any{"first": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(true),
				WithEncodeDotInKeys(false),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name.obj.first=John&name.obj.last=Doe",
		},
		{
			"allowDots false, encodeDotInKeys true",
			map[string]any{"name.obj": map[string]any{"first": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(false),
				WithEncodeDotInKeys(true),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name%252Eobj%5Bfirst%5D=John&name%252Eobj%5Blast%5D=Doe",
		},
		{
			"allowDots true, encodeDotInKeys true",
			map[string]any{"name.obj": map[string]any{"first": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(true),
				WithEncodeDotInKeys(true),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name%252Eobj.first=John&name%252Eobj.last=Doe",
		},
		{
			"nested with multiple dots - allowDots false, encodeDotInKeys false",
			map[string]any{"name.obj.subobject": map[string]any{"first.godly.name": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(false),
				WithEncodeDotInKeys(false),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name.obj.subobject%5Bfirst.godly.name%5D=John&name.obj.subobject%5Blast%5D=Doe",
		},
		{
			"nested with multiple dots - allowDots true, encodeDotInKeys false",
			map[string]any{"name.obj.subobject": map[string]any{"first.godly.name": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(true),
				WithEncodeDotInKeys(false),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name.obj.subobject.first.godly.name=John&name.obj.subobject.last=Doe",
		},
		{
			"nested with multiple dots - allowDots false, encodeDotInKeys true",
			map[string]any{"name.obj.subobject": map[string]any{"first.godly.name": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(false),
				WithEncodeDotInKeys(true),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name%252Eobj%252Esubobject%5Bfirst.godly.name%5D=John&name%252Eobj%252Esubobject%5Blast%5D=Doe",
		},
		{
			"nested with multiple dots - allowDots true, encodeDotInKeys true",
			map[string]any{"name.obj.subobject": map[string]any{"first.godly.name": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(true),
				WithEncodeDotInKeys(true),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name%252Eobj%252Esubobject.first%252Egodly%252Ename=John&name%252Eobj%252Esubobject.last=Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyEncodeDotInKeysAutoAllowDots tests that encodeDotInKeys auto-sets allowDots
func TestJSStringifyEncodeDotInKeysAutoAllowDots(t *testing.T) {
	input := map[string]any{"name.obj.subobject": map[string]any{"first.godly.name": "John", "last": "Doe"}}
	result, err := Stringify(input,
		WithEncodeDotInKeys(true),
		WithSort(func(a, b string) bool { return a < b }),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "name%252Eobj%252Esubobject.first%252Egodly%252Ename=John&name%252Eobj%252Esubobject.last=Doe"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyEncodeDotInKeysWithEncodeValuesOnly tests encodeDotInKeys with encodeValuesOnly
func TestJSStringifyEncodeDotInKeysWithEncodeValuesOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			"simple object",
			map[string]any{"name.obj": map[string]any{"first": "John", "last": "Doe"}},
			[]StringifyOption{
				WithEncodeDotInKeys(true),
				WithStringifyAllowDots(true),
				WithEncodeValuesOnly(true),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name%2Eobj.first=John&name%2Eobj.last=Doe",
		},
		{
			"nested object",
			map[string]any{"name.obj.subobject": map[string]any{"first.godly.name": "John", "last": "Doe"}},
			[]StringifyOption{
				WithStringifyAllowDots(true),
				WithEncodeDotInKeys(true),
				WithEncodeValuesOnly(true),
				WithSort(func(a, b string) bool { return a < b }),
			},
			"name%2Eobj%2Esubobject.first%2Egodly%2Ename=John&name%2Eobj%2Esubobject.last=Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyAddQueryPrefix tests adding query prefix
func TestJSStringifyAddQueryPrefix(t *testing.T) {
	result, err := Stringify(map[string]any{"a": "b"}, WithStringifyAddQueryPrefix(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "?a=b" {
		t.Errorf("expected '?a=b', got %q", result)
	}
}

// TestJSStringifyAddQueryPrefixEmptyObject tests query prefix with empty object
func TestJSStringifyAddQueryPrefixEmptyObject(t *testing.T) {
	result, err := Stringify(map[string]any{}, WithStringifyAddQueryPrefix(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected '', got %q", result)
	}
}

// TestJSStringifyNestedFalsyValues tests nested falsy values
func TestJSStringifyNestedFalsyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			"nested null - default",
			map[string]any{"a": map[string]any{"b": map[string]any{"c": nil}}},
			nil,
			"a%5Bb%5D%5Bc%5D=",
		},
		{
			"nested null - strictNullHandling",
			map[string]any{"a": map[string]any{"b": map[string]any{"c": nil}}},
			[]StringifyOption{WithStringifyStrictNullHandling(true)},
			"a%5Bb%5D%5Bc%5D",
		},
		{
			"nested false",
			map[string]any{"a": map[string]any{"b": map[string]any{"c": false}}},
			nil,
			"a%5Bb%5D%5Bc%5D=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyNestedObject tests nested object stringification
func TestJSStringifyNestedObject(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			"simple nested",
			map[string]any{"a": map[string]any{"b": "c"}},
			"a%5Bb%5D=c",
		},
		{
			"deeply nested",
			map[string]any{"a": map[string]any{"b": map[string]any{"c": map[string]any{"d": "e"}}}},
			"a%5Bb%5D%5Bc%5D%5Bd%5D=e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyAllowDots tests dot notation for nested objects
func TestJSStringifyAllowDots(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			"simple nested with dots",
			map[string]any{"a": map[string]any{"b": "c"}},
			"a.b=c",
		},
		{
			"deeply nested with dots",
			map[string]any{"a": map[string]any{"b": map[string]any{"c": map[string]any{"d": "e"}}}},
			"a.b.c.d=e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, WithStringifyAllowDots(true))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyArrayValue tests array value stringification
func TestJSStringifyArrayValue(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			"indices format",
			map[string]any{"a": []any{"b", "c", "d"}},
			[]StringifyOption{WithArrayFormat(ArrayFormatIndices)},
			"a%5B0%5D=b&a%5B1%5D=c&a%5B2%5D=d",
		},
		{
			"brackets format",
			map[string]any{"a": []any{"b", "c", "d"}},
			[]StringifyOption{WithArrayFormat(ArrayFormatBrackets)},
			"a%5B%5D=b&a%5B%5D=c&a%5B%5D=d",
		},
		{
			"comma format",
			map[string]any{"a": []any{"b", "c", "d"}},
			[]StringifyOption{WithArrayFormat(ArrayFormatComma)},
			"a=b%2Cc%2Cd",
		},
		{
			"comma format with commaRoundTrip",
			map[string]any{"a": []any{"b", "c", "d"}},
			[]StringifyOption{WithArrayFormat(ArrayFormatComma), WithCommaRoundTrip(true)},
			"a=b%2Cc%2Cd",
		},
		{
			"default format",
			map[string]any{"a": []any{"b", "c", "d"}},
			nil,
			"a%5B0%5D=b&a%5B1%5D=c&a%5B2%5D=d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifySkipNulls tests skipping null values
func TestJSStringifySkipNulls(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			"omits nulls when asked",
			map[string]any{"a": "b", "c": nil},
			[]StringifyOption{WithSkipNulls(true), WithSort(func(a, b string) bool { return a < b })},
			"a=b",
		},
		{
			"omits nested nulls when asked",
			map[string]any{"a": map[string]any{"b": "c", "d": nil}},
			[]StringifyOption{WithSkipNulls(true), WithSort(func(a, b string) bool { return a < b })},
			"a%5Bb%5D=c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyOmitArrayIndices tests omitting array indices
func TestJSStringifyOmitArrayIndices(t *testing.T) {
	result, err := Stringify(map[string]any{"a": []any{"b", "c", "d"}}, WithArrayFormat(ArrayFormatRepeat))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=b&a=c&a=d"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyOmitEmptyArray tests omitting empty arrays
func TestJSStringifyOmitEmptyArray(t *testing.T) {
	result, err := Stringify(map[string]any{"a": []any{}, "b": "zz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "b=zz"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyAllowEmptyArrays tests allowing empty arrays
func TestJSStringifyAllowEmptyArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			"default - omit empty arrays",
			map[string]any{"a": []any{}, "b": "zz"},
			nil,
			"b=zz",
		},
		{
			"allowEmptyArrays false",
			map[string]any{"a": []any{}, "b": "zz"},
			[]StringifyOption{WithStringifyAllowEmptyArrays(false)},
			"b=zz",
		},
		{
			"allowEmptyArrays true",
			map[string]any{"a": []any{}, "b": "zz"},
			[]StringifyOption{WithStringifyAllowEmptyArrays(true), WithSort(func(a, b string) bool { return a < b })},
			"a[]&b=zz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyAllowEmptyArraysStrictNullHandling tests empty arrays with strictNullHandling
func TestJSStringifyAllowEmptyArraysStrictNullHandling(t *testing.T) {
	result, err := Stringify(
		map[string]any{"testEmptyArray": []any{}},
		WithStringifyStrictNullHandling(true),
		WithStringifyAllowEmptyArrays(true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "testEmptyArray[]"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyArraySingleVsMultipleItems tests single vs multiple array items
func TestJSStringifyArraySingleVsMultipleItems(t *testing.T) {
	t.Run("non-array item", func(t *testing.T) {
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"indices", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices)}, "a=c"},
			{"brackets", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets)}, "a=c"},
			{"comma", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma)}, "a=c"},
			{"default", []StringifyOption{WithEncodeValuesOnly(true)}, "a=c"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(map[string]any{"a": "c"}, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})

	t.Run("array with single item", func(t *testing.T) {
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"indices", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices)}, "a[0]=c"},
			{"brackets", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets)}, "a[]=c"},
			{"comma", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma)}, "a=c"},
			{"comma with commaRoundTrip", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma), WithCommaRoundTrip(true)}, "a[]=c"},
			{"default", []StringifyOption{WithEncodeValuesOnly(true)}, "a[0]=c"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(map[string]any{"a": []any{"c"}}, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})

	t.Run("array with multiple items", func(t *testing.T) {
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"indices", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices)}, "a[0]=c&a[1]=d"},
			{"brackets", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets)}, "a[]=c&a[]=d"},
			{"comma", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma)}, "a=c,d"},
			{"comma with commaRoundTrip", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma), WithCommaRoundTrip(true)}, "a=c,d"},
			{"default", []StringifyOption{WithEncodeValuesOnly(true)}, "a[0]=c&a[1]=d"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(map[string]any{"a": []any{"c", "d"}}, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})

	t.Run("array with items containing comma", func(t *testing.T) {
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"comma encodeValuesOnly", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma)}, "a=c%2Cd,e"},
			{"comma", []StringifyOption{WithArrayFormat(ArrayFormatComma)}, "a=c%2Cd%2Ce"},
			{"comma encodeValuesOnly commaRoundTrip", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma), WithCommaRoundTrip(true)}, "a=c%2Cd,e"},
			{"comma commaRoundTrip", []StringifyOption{WithArrayFormat(ArrayFormatComma), WithCommaRoundTrip(true)}, "a=c%2Cd%2Ce"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(map[string]any{"a": []any{"c,d", "e"}}, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})
}

// TestJSStringifyNestedArrayValue tests nested array value
func TestJSStringifyNestedArrayValue(t *testing.T) {
	tests := []struct {
		name     string
		opts     []StringifyOption
		expected string
	}{
		{"indices", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices)}, "a[b][0]=c&a[b][1]=d"},
		{"brackets", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets)}, "a[b][]=c&a[b][]=d"},
		{"comma", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma)}, "a[b]=c,d"},
		{"default", []StringifyOption{WithEncodeValuesOnly(true)}, "a[b][0]=c&a[b][1]=d"},
	}

	input := map[string]any{"a": map[string]any{"b": []any{"c", "d"}}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyCommaAndEmptyArrayValues tests comma and empty array values
func TestJSStringifyCommaAndEmptyArrayValues(t *testing.T) {
	input := map[string]any{"a": []any{",", "", "c,d%"}}

	tests := []struct {
		name     string
		opts     []StringifyOption
		expected string
	}{
		{"encode false, indices", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatIndices)}, "a[0]=,&a[1]=&a[2]=c,d%"},
		{"encode false, brackets", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatBrackets)}, "a[]=,&a[]=&a[]=c,d%"},
		{"encode false, comma", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatComma)}, "a=,,,c,d%"},
		{"encode false, repeat", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatRepeat)}, "a=,&a=&a=c,d%"},
		{"encodeValuesOnly, indices", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices)}, "a[0]=%2C&a[1]=&a[2]=c%2Cd%25"},
		{"encodeValuesOnly, brackets", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets)}, "a[]=%2C&a[]=&a[]=c%2Cd%25"},
		{"encodeValuesOnly, comma", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma)}, "a=%2C,,c%2Cd%25"},
		{"encodeValuesOnly, repeat", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatRepeat)}, "a=%2C&a=&a=c%2Cd%25"},
		{"encode all, indices", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(false), WithArrayFormat(ArrayFormatIndices)}, "a%5B0%5D=%2C&a%5B1%5D=&a%5B2%5D=c%2Cd%25"},
		{"encode all, brackets", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(false), WithArrayFormat(ArrayFormatBrackets)}, "a%5B%5D=%2C&a%5B%5D=&a%5B%5D=c%2Cd%25"},
		{"encode all, comma", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(false), WithArrayFormat(ArrayFormatComma)}, "a=%2C%2C%2Cc%2Cd%25"},
		{"encode all, repeat", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(false), WithArrayFormat(ArrayFormatRepeat)}, "a=%2C&a=&a=c%2Cd%25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyCommaAndEmptyNonArrayValues tests comma and empty non-array values
func TestJSStringifyCommaAndEmptyNonArrayValues(t *testing.T) {
	input := map[string]any{"a": ",", "b": "", "c": "c,d%"}

	tests := []struct {
		name     string
		opts     []StringifyOption
		expected string
	}{
		{"encode false, indices", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatIndices), WithSort(func(a, b string) bool { return a < b })}, "a=,&b=&c=c,d%"},
		{"encode false, brackets", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatBrackets), WithSort(func(a, b string) bool { return a < b })}, "a=,&b=&c=c,d%"},
		{"encode false, comma", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatComma), WithSort(func(a, b string) bool { return a < b })}, "a=,&b=&c=c,d%"},
		{"encode false, repeat", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatRepeat), WithSort(func(a, b string) bool { return a < b })}, "a=,&b=&c=c,d%"},
		{"encodeValuesOnly, indices", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices), WithSort(func(a, b string) bool { return a < b })}, "a=%2C&b=&c=c%2Cd%25"},
		{"encodeValuesOnly, brackets", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets), WithSort(func(a, b string) bool { return a < b })}, "a=%2C&b=&c=c%2Cd%25"},
		{"encodeValuesOnly, comma", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma), WithSort(func(a, b string) bool { return a < b })}, "a=%2C&b=&c=c%2Cd%25"},
		{"encodeValuesOnly, repeat", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatRepeat), WithSort(func(a, b string) bool { return a < b })}, "a=%2C&b=&c=c%2Cd%25"},
		{"encode all, indices", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(false), WithArrayFormat(ArrayFormatIndices), WithSort(func(a, b string) bool { return a < b })}, "a=%2C&b=&c=c%2Cd%25"},
		{"encode all, brackets", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(false), WithArrayFormat(ArrayFormatBrackets), WithSort(func(a, b string) bool { return a < b })}, "a=%2C&b=&c=c%2Cd%25"},
		{"encode all, comma", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(false), WithArrayFormat(ArrayFormatComma), WithSort(func(a, b string) bool { return a < b })}, "a=%2C&b=&c=c%2Cd%25"},
		{"encode all, repeat", []StringifyOption{WithEncode(true), WithEncodeValuesOnly(false), WithArrayFormat(ArrayFormatRepeat), WithSort(func(a, b string) bool { return a < b })}, "a=%2C&b=&c=c%2Cd%25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyNestedArrayValueWithDots tests nested array value with dots
func TestJSStringifyNestedArrayValueWithDots(t *testing.T) {
	tests := []struct {
		name     string
		opts     []StringifyOption
		expected string
	}{
		{"indices", []StringifyOption{WithStringifyAllowDots(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices)}, "a.b[0]=c&a.b[1]=d"},
		{"brackets", []StringifyOption{WithStringifyAllowDots(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets)}, "a.b[]=c&a.b[]=d"},
		{"comma", []StringifyOption{WithStringifyAllowDots(true), WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatComma)}, "a.b=c,d"},
		{"default", []StringifyOption{WithStringifyAllowDots(true), WithEncodeValuesOnly(true)}, "a.b[0]=c&a.b[1]=d"},
	}

	input := map[string]any{"a": map[string]any{"b": []any{"c", "d"}}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyObjectInsideArray tests object inside array
func TestJSStringifyObjectInsideArray(t *testing.T) {
	t.Run("simple object in array", func(t *testing.T) {
		input := map[string]any{"a": []any{map[string]any{"b": "c"}}}
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"indices", []StringifyOption{WithArrayFormat(ArrayFormatIndices), WithEncodeValuesOnly(true)}, "a[0][b]=c"},
			{"repeat", []StringifyOption{WithArrayFormat(ArrayFormatRepeat), WithEncodeValuesOnly(true)}, "a[b]=c"},
			{"brackets", []StringifyOption{WithArrayFormat(ArrayFormatBrackets), WithEncodeValuesOnly(true)}, "a[][b]=c"},
			{"default", []StringifyOption{WithEncodeValuesOnly(true)}, "a[0][b]=c"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(input, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})

	t.Run("nested object in array", func(t *testing.T) {
		input := map[string]any{"a": []any{map[string]any{"b": map[string]any{"c": []any{1}}}}}
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"indices", []StringifyOption{WithArrayFormat(ArrayFormatIndices), WithEncodeValuesOnly(true)}, "a[0][b][c][0]=1"},
			{"repeat", []StringifyOption{WithArrayFormat(ArrayFormatRepeat), WithEncodeValuesOnly(true)}, "a[b][c]=1"},
			{"brackets", []StringifyOption{WithArrayFormat(ArrayFormatBrackets), WithEncodeValuesOnly(true)}, "a[][b][c][]=1"},
			{"default", []StringifyOption{WithEncodeValuesOnly(true)}, "a[0][b][c][0]=1"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(input, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})
}

// TestJSStringifyArrayWithMixedObjectsAndPrimitives tests arrays with mixed content
func TestJSStringifyArrayWithMixedObjectsAndPrimitives(t *testing.T) {
	input := map[string]any{"a": []any{map[string]any{"b": 1}, 2, 3}}
	tests := []struct {
		name     string
		opts     []StringifyOption
		expected string
	}{
		{"indices", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices)}, "a[0][b]=1&a[1]=2&a[2]=3"},
		{"brackets", []StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets)}, "a[][b]=1&a[]=2&a[]=3"},
		{"default", []StringifyOption{WithEncodeValuesOnly(true)}, "a[0][b]=1&a[1]=2&a[2]=3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyObjectInsideArrayWithDots tests object inside array with dots notation
func TestJSStringifyObjectInsideArrayWithDots(t *testing.T) {
	t.Run("simple object in array with dots", func(t *testing.T) {
		input := map[string]any{"a": []any{map[string]any{"b": "c"}}}
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"indices", []StringifyOption{WithStringifyAllowDots(true), WithEncode(false), WithArrayFormat(ArrayFormatIndices)}, "a[0].b=c"},
			{"brackets", []StringifyOption{WithStringifyAllowDots(true), WithEncode(false), WithArrayFormat(ArrayFormatBrackets)}, "a[].b=c"},
			{"default", []StringifyOption{WithStringifyAllowDots(true), WithEncode(false)}, "a[0].b=c"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(input, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})

	t.Run("nested object in array with dots", func(t *testing.T) {
		input := map[string]any{"a": []any{map[string]any{"b": map[string]any{"c": []any{1}}}}}
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"indices", []StringifyOption{WithStringifyAllowDots(true), WithEncode(false), WithArrayFormat(ArrayFormatIndices)}, "a[0].b.c[0]=1"},
			{"brackets", []StringifyOption{WithStringifyAllowDots(true), WithEncode(false), WithArrayFormat(ArrayFormatBrackets)}, "a[].b.c[]=1"},
			{"default", []StringifyOption{WithStringifyAllowDots(true), WithEncode(false)}, "a[0].b.c[0]=1"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(input, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})
}

// TestJSStringifyDoesNotOmitObjectKeysWhenIndicesFalse tests that object keys are preserved
func TestJSStringifyDoesNotOmitObjectKeysWhenIndicesFalse(t *testing.T) {
	result, err := Stringify(map[string]any{"a": []any{map[string]any{"b": "c"}}}, WithArrayFormat(ArrayFormatRepeat))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a%5Bb%5D=c"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyIndicesTrue tests indices notation with indices=true
func TestJSStringifyIndicesTrue(t *testing.T) {
	result, err := Stringify(map[string]any{"a": []any{"b", "c"}}, WithArrayFormat(ArrayFormatIndices))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a%5B0%5D=b&a%5B1%5D=c"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyDefaultArrayFormat tests default array format
func TestJSStringifyDefaultArrayFormat(t *testing.T) {
	result, err := Stringify(map[string]any{"a": []any{"b", "c"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a%5B0%5D=b&a%5B1%5D=c"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyRepeatFormat tests repeat notation
func TestJSStringifyRepeatFormat(t *testing.T) {
	result, err := Stringify(map[string]any{"a": []any{"b", "c"}}, WithArrayFormat(ArrayFormatRepeat))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=b&a=c"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyBracketsFormat tests brackets notation
func TestJSStringifyBracketsFormat(t *testing.T) {
	result, err := Stringify(map[string]any{"a": []any{"b", "c"}}, WithArrayFormat(ArrayFormatBrackets))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a%5B%5D=b&a%5B%5D=c"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyComplicatedObject tests complicated object
func TestJSStringifyComplicatedObject(t *testing.T) {
	result, err := Stringify(map[string]any{"a": map[string]any{"b": "c", "d": "e"}}, WithSort(func(a, b string) bool { return a < b }))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a%5Bb%5D=c&a%5Bd%5D=e"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyEmptyValue tests empty value
func TestJSStringifyEmptyValue(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{"empty string", map[string]any{"a": ""}, nil, "a="},
		{"null strictNullHandling", map[string]any{"a": nil}, []StringifyOption{WithStringifyStrictNullHandling(true)}, "a"},
		{"two empty strings", map[string]any{"a": "", "b": ""}, []StringifyOption{WithSort(func(a, b string) bool { return a < b })}, "a=&b="},
		{"null and empty strictNullHandling", map[string]any{"a": nil, "b": ""}, []StringifyOption{WithStringifyStrictNullHandling(true), WithSort(func(a, b string) bool { return a < b })}, "a&b="},
		{"nested empty string", map[string]any{"a": map[string]any{"b": ""}}, nil, "a%5Bb%5D="},
		{"nested null strictNullHandling", map[string]any{"a": map[string]any{"b": nil}}, []StringifyOption{WithStringifyStrictNullHandling(true)}, "a%5Bb%5D"},
		{"nested null no strictNullHandling", map[string]any{"a": map[string]any{"b": nil}}, []StringifyOption{WithStringifyStrictNullHandling(false)}, "a%5Bb%5D="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyEmptyArrayDifferentFormats tests empty array in different formats
func TestJSStringifyEmptyArrayDifferentFormats(t *testing.T) {
	// Use ExplicitNullValue to represent JS null (not undefined/sparse slot)
	// In Go: nil = undefined (sparse), ExplicitNullValue = null
	input := map[string]any{"a": []any{}, "b": []any{ExplicitNullValue}, "c": "c"}

	tests := []struct {
		name     string
		opts     []StringifyOption
		expected string
	}{
		{"encode false default", []StringifyOption{WithEncode(false), WithSort(func(a, b string) bool { return a < b })}, "b[0]=&c=c"},
		{"encode false indices", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatIndices), WithSort(func(a, b string) bool { return a < b })}, "b[0]=&c=c"},
		{"encode false brackets", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatBrackets), WithSort(func(a, b string) bool { return a < b })}, "b[]=&c=c"},
		{"encode false repeat", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatRepeat), WithSort(func(a, b string) bool { return a < b })}, "b=&c=c"},
		{"encode false comma", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatComma), WithSort(func(a, b string) bool { return a < b })}, "b=&c=c"},
		{"encode false comma commaRoundTrip", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatComma), WithCommaRoundTrip(true), WithSort(func(a, b string) bool { return a < b })}, "b[]=&c=c"},
		{"encode false indices strictNullHandling", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatIndices), WithStringifyStrictNullHandling(true), WithSort(func(a, b string) bool { return a < b })}, "b[0]&c=c"},
		{"encode false brackets strictNullHandling", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatBrackets), WithStringifyStrictNullHandling(true), WithSort(func(a, b string) bool { return a < b })}, "b[]&c=c"},
		{"encode false repeat strictNullHandling", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatRepeat), WithStringifyStrictNullHandling(true), WithSort(func(a, b string) bool { return a < b })}, "b&c=c"},
		{"encode false comma strictNullHandling", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatComma), WithStringifyStrictNullHandling(true), WithSort(func(a, b string) bool { return a < b })}, "b&c=c"},
		{"encode false comma strictNullHandling commaRoundTrip", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatComma), WithStringifyStrictNullHandling(true), WithCommaRoundTrip(true), WithSort(func(a, b string) bool { return a < b })}, "b[]&c=c"},
		{"encode false indices skipNulls", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatIndices), WithSkipNulls(true), WithSort(func(a, b string) bool { return a < b })}, "c=c"},
		{"encode false brackets skipNulls", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatBrackets), WithSkipNulls(true), WithSort(func(a, b string) bool { return a < b })}, "c=c"},
		{"encode false repeat skipNulls", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatRepeat), WithSkipNulls(true), WithSort(func(a, b string) bool { return a < b })}, "c=c"},
		{"encode false comma skipNulls", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatComma), WithSkipNulls(true), WithSort(func(a, b string) bool { return a < b })}, "c=c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyInvalidInput tests invalid input returns empty string
func TestJSStringifyInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
	}{
		{"nil", nil},
		{"empty map", map[string]any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != "" {
				t.Errorf("expected '', got %q", result)
			}
		})
	}
}

// TestJSStringifyDropsUndefined tests dropping undefined values
func TestJSStringifyDropsUndefined(t *testing.T) {
	// In Go, we use nil to represent undefined
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			"nested undefined and null strictNullHandling",
			map[string]any{"a": map[string]any{"c": nil}},
			[]StringifyOption{WithStringifyStrictNullHandling(true)},
			"a%5Bc%5D",
		},
		{
			"nested undefined and null no strictNullHandling",
			map[string]any{"a": map[string]any{"c": nil}},
			[]StringifyOption{WithStringifyStrictNullHandling(false)},
			"a%5Bc%5D=",
		},
		{
			"nested undefined and empty string",
			map[string]any{"a": map[string]any{"c": ""}},
			nil,
			"a%5Bc%5D=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyURLEncodesValues tests URL encoding of values
func TestJSStringifyURLEncodesValues(t *testing.T) {
	result, err := Stringify(map[string]any{"a": "b c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=b%20c"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyDate tests date stringification
func TestJSStringifyDate(t *testing.T) {
	now := time.Now().UTC()
	expected := "a=" + strings.ReplaceAll(now.Format(time.RFC3339), ":", "%3A")
	result, err := Stringify(map[string]any{"a": now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyWeirdObject tests weird object from qs
func TestJSStringifyWeirdObject(t *testing.T) {
	result, err := Stringify(map[string]any{"my weird field": "~q1!2\"'w$5&7/z8)?"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "my%20weird%20field=~q1%212%22%27w%245%267%2Fz8%29%3F"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyBooleanValues tests boolean values
func TestJSStringifyBooleanValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{"true", map[string]any{"a": true}, "a=true"},
		{"nested true", map[string]any{"a": map[string]any{"b": true}}, "a%5Bb%5D=true"},
		{"false", map[string]any{"b": false}, "b=false"},
		{"nested false", map[string]any{"b": map[string]any{"c": false}}, "b%5Bc%5D=false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyAlternativeDelimiter tests alternative delimiter
func TestJSStringifyAlternativeDelimiter(t *testing.T) {
	result, err := Stringify(map[string]any{"a": "b", "c": "d"}, WithStringifyDelimiter(";"), WithSort(func(a, b string) bool { return a < b }))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=b;c=d"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyCyclicReferences tests cyclic reference detection
func TestJSStringifyCyclicReferences(t *testing.T) {
	// Create a cyclic structure
	a := map[string]any{}
	a["b"] = a

	_, err := Stringify(map[string]any{"foo[bar]": "baz", "foo[baz]": a})
	if err != ErrCyclicReference {
		t.Errorf("expected ErrCyclicReference, got %v", err)
	}

	circular := map[string]any{"a": "value"}
	circular["a"] = circular
	_, err = Stringify(circular)
	if err != ErrCyclicReference {
		t.Errorf("expected ErrCyclicReference, got %v", err)
	}
}

// TestJSStringifyNonCircularDuplicatedReferences tests non-circular duplicated references
func TestJSStringifyNonCircularDuplicatedReferences(t *testing.T) {
	hourOfDay := map[string]any{"function": "hour_of_day"}

	p1 := map[string]any{
		"function":  "gte",
		"arguments": []any{hourOfDay, 0},
	}
	p2 := map[string]any{
		"function":  "lte",
		"arguments": []any{hourOfDay, 23},
	}

	input := map[string]any{"filters": map[string]any{"$and": []any{p1, p2}}}

	// Sort to ensure deterministic order: "arguments" < "function"
	sortAsc := func(a, b string) bool { return a < b }

	t.Run("indices", func(t *testing.T) {
		result, err := Stringify(input, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices), WithSort(sortAsc))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// With sort: arguments comes before function
		expected := "filters[$and][0][arguments][0][function]=hour_of_day&filters[$and][0][arguments][1]=0&filters[$and][0][function]=gte&filters[$and][1][arguments][0][function]=hour_of_day&filters[$and][1][arguments][1]=23&filters[$and][1][function]=lte"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("brackets", func(t *testing.T) {
		result, err := Stringify(input, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets), WithSort(sortAsc))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// With sort: arguments comes before function
		expected := "filters[$and][][arguments][][function]=hour_of_day&filters[$and][][arguments][]=0&filters[$and][][function]=gte&filters[$and][][arguments][][function]=hour_of_day&filters[$and][][arguments][]=23&filters[$and][][function]=lte"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("repeat", func(t *testing.T) {
		result, err := Stringify(input, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatRepeat), WithSort(sortAsc))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// With sort: arguments comes before function
		expected := "filters[$and][arguments][function]=hour_of_day&filters[$and][arguments]=0&filters[$and][function]=gte&filters[$and][arguments][function]=hour_of_day&filters[$and][arguments]=23&filters[$and][function]=lte"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestJSStringifyFilterArray tests filter with array of keys
func TestJSStringifyFilterArray(t *testing.T) {
	t.Run("simple filter", func(t *testing.T) {
		result, err := Stringify(map[string]any{"a": "b"}, WithFilter([]string{"a"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a=b" {
			t.Errorf("expected 'a=b', got %q", result)
		}
	})

	t.Run("empty filter", func(t *testing.T) {
		result, err := Stringify(map[string]any{"a": 1}, WithFilter([]string{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "" {
			t.Errorf("expected '', got %q", result)
		}
	})

	t.Run("complex filter indices", func(t *testing.T) {
		result, err := Stringify(
			map[string]any{"a": map[string]any{"b": []any{1, 2, 3, 4}, "c": "d"}, "c": "f"},
			WithFilter([]string{"a", "b", "0", "2"}),
			WithArrayFormat(ArrayFormatIndices),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a%5Bb%5D%5B0%5D=1&a%5Bb%5D%5B2%5D=3"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("complex filter brackets", func(t *testing.T) {
		result, err := Stringify(
			map[string]any{"a": map[string]any{"b": []any{1, 2, 3, 4}, "c": "d"}, "c": "f"},
			WithFilter([]string{"a", "b", "0", "2"}),
			WithArrayFormat(ArrayFormatBrackets),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a%5Bb%5D%5B%5D=1&a%5Bb%5D%5B%5D=3"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("complex filter default", func(t *testing.T) {
		result, err := Stringify(
			map[string]any{"a": map[string]any{"b": []any{1, 2, 3, 4}, "c": "d"}, "c": "f"},
			WithFilter([]string{"a", "b", "0", "2"}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a%5Bb%5D%5B0%5D=1&a%5Bb%5D%5B2%5D=3"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestJSStringifyFilterFunction tests filter with function
func TestJSStringifyFilterFunction(t *testing.T) {
	calls := 0
	obj := map[string]any{"a": "b", "c": "d", "e": map[string]any{"f": time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)}}

	filter := FilterFunc(func(prefix string, value any) any {
		calls++
		if calls == 1 {
			if prefix != "" {
				t.Errorf("first call prefix expected '', got %q", prefix)
			}
		}
		if prefix == "c" {
			return nil
		}
		if t, ok := value.(time.Time); ok {
			return t.UnixMilli()
		}
		return value
	})

	result, err := Stringify(obj, WithFilter(filter), WithSort(func(a, b string) bool { return a < b }))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=b&e%5Bf%5D=1257894000000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyDisableEncoding tests disabling URI encoding
func TestJSStringifyDisableEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{"simple", map[string]any{"a": "b"}, []StringifyOption{WithEncode(false)}, "a=b"},
		{"nested", map[string]any{"a": map[string]any{"b": "c"}}, []StringifyOption{WithEncode(false)}, "a[b]=c"},
		{"null strictNullHandling", map[string]any{"a": "b", "c": nil}, []StringifyOption{WithStringifyStrictNullHandling(true), WithEncode(false), WithSort(func(a, b string) bool { return a < b })}, "a=b&c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifySort tests sorting keys
func TestJSStringifySort(t *testing.T) {
	sort := func(a, b string) bool {
		return a < b
	}

	t.Run("simple sort", func(t *testing.T) {
		result, err := Stringify(map[string]any{"a": "c", "z": "y", "b": "f"}, WithSort(sort))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a=c&b=f&z=y"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("nested sort", func(t *testing.T) {
		result, err := Stringify(map[string]any{"a": "c", "z": map[string]any{"j": "a", "i": "b"}, "b": "f"}, WithSort(sort))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a=c&b=f&z%5Bi%5D=b&z%5Bj%5D=a"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestJSStringifySortDeep tests sorting keys at depth 3 or more
func TestJSStringifySortDeep(t *testing.T) {
	sort := func(a, b string) bool {
		return a < b
	}

	input := map[string]any{
		"a": "a",
		"z": map[string]any{
			"zj": map[string]any{"zjb": "zjb", "zja": "zja"},
			"zi": map[string]any{"zib": "zib", "zia": "zia"},
		},
		"b": "b",
	}

	t.Run("with sort", func(t *testing.T) {
		result, err := Stringify(input, WithSort(sort), WithEncode(false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a=a&b=b&z[zi][zia]=zia&z[zi][zib]=zib&z[zj][zja]=zja&z[zj][zjb]=zjb"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestJSStringifySerializeDate tests serializeDate option
func TestJSStringifySerializeDate(t *testing.T) {
	date := time.Now().UTC()

	t.Run("default is ISO string", func(t *testing.T) {
		result, err := Stringify(map[string]any{"a": date})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a=" + strings.ReplaceAll(date.Format(time.RFC3339), ":", "%3A")
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("custom serializeDate", func(t *testing.T) {
		specificDate := time.Unix(0, 6*1000000)
		result, err := Stringify(
			map[string]any{"a": specificDate},
			WithSerializeDate(func(d time.Time) string {
				return toString(d.UnixNano() / 1000000 * 7)
			}),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a=42"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("with arrayFormat comma", func(t *testing.T) {
		result, err := Stringify(
			map[string]any{"a": []any{date}},
			WithSerializeDate(func(d time.Time) string {
				return toString(d.UnixMilli())
			}),
			WithArrayFormat(ArrayFormatComma),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a=" + toString(date.UnixMilli())
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("with arrayFormat comma commaRoundTrip", func(t *testing.T) {
		result, err := Stringify(
			map[string]any{"a": []any{date}},
			WithSerializeDate(func(d time.Time) string {
				return toString(d.UnixMilli())
			}),
			WithArrayFormat(ArrayFormatComma),
			WithCommaRoundTrip(true),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "a%5B%5D=" + toString(date.UnixMilli())
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestJSStringifyRFC1738 tests RFC 1738 serialization
func TestJSStringifyRFC1738(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{"space in value", map[string]any{"a": "b c"}, "a=b+c"},
		{"space in key and value", map[string]any{"a b": "c d"}, "a+b=c+d"},
		{"parentheses", map[string]any{"foo(ref)": "bar"}, "foo(ref)=bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, WithFormat(FormatRFC1738))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyRFC3986 tests RFC 3986 spaces serialization
func TestJSStringifyRFC3986(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{"space in value", map[string]any{"a": "b c"}, "a=b%20c"},
		{"space in key and value", map[string]any{"a b": "c d"}, "a%20b=c%20d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, WithFormat(FormatRFC3986))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyBackwardCompatibility tests backward compatibility to RFC 3986
func TestJSStringifyBackwardCompatibility(t *testing.T) {
	result, err := Stringify(map[string]any{"a": "b c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=b%20c"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyInvalidFormat tests invalid format throws error
func TestJSStringifyInvalidFormat(t *testing.T) {
	invalidFormats := []Format{"UFO1234", ""}
	for _, format := range invalidFormats {
		if format == "" {
			continue // empty format is valid (defaults to RFC3986)
		}
		_, err := Stringify(map[string]any{"a": "b c"}, WithFormat(format))
		if err != ErrInvalidFormat {
			t.Errorf("expected ErrInvalidFormat for format %q, got %v", format, err)
		}
	}
}

// TestJSStringifyEncodeValuesOnly tests encodeValuesOnly option
func TestJSStringifyEncodeValuesOnly(t *testing.T) {
	sortAsc := func(a, b string) bool { return a < b }
	tests := []struct {
		name     string
		input    map[string]any
		opts     []StringifyOption
		expected string
	}{
		{
			"encodeValuesOnly indices",
			map[string]any{"a": "b", "c": []any{"d", "e=f"}, "f": []any{[]any{"g"}, []any{"h"}}},
			[]StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices), WithSort(sortAsc)},
			"a=b&c[0]=d&c[1]=e%3Df&f[0][0]=g&f[1][0]=h",
		},
		{
			"encodeValuesOnly brackets",
			map[string]any{"a": "b", "c": []any{"d", "e=f"}, "f": []any{[]any{"g"}, []any{"h"}}},
			[]StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets), WithSort(sortAsc)},
			"a=b&c[]=d&c[]=e%3Df&f[][]=g&f[][]=h",
		},
		{
			"encodeValuesOnly repeat",
			map[string]any{"a": "b", "c": []any{"d", "e=f"}, "f": []any{[]any{"g"}, []any{"h"}}},
			[]StringifyOption{WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatRepeat), WithSort(sortAsc)},
			"a=b&c=d&c=e%3Df&f=g&f=h",
		},
		{
			"no encodeValuesOnly indices",
			map[string]any{"a": "b", "c": []any{"d", "e"}, "f": []any{[]any{"g"}, []any{"h"}}},
			[]StringifyOption{WithArrayFormat(ArrayFormatIndices), WithSort(sortAsc)},
			"a=b&c%5B0%5D=d&c%5B1%5D=e&f%5B0%5D%5B0%5D=g&f%5B1%5D%5B0%5D=h",
		},
		{
			"no encodeValuesOnly brackets",
			map[string]any{"a": "b", "c": []any{"d", "e"}, "f": []any{[]any{"g"}, []any{"h"}}},
			[]StringifyOption{WithArrayFormat(ArrayFormatBrackets), WithSort(sortAsc)},
			"a=b&c%5B%5D=d&c%5B%5D=e&f%5B%5D%5B%5D=g&f%5B%5D%5B%5D=h",
		},
		{
			"no encodeValuesOnly repeat",
			map[string]any{"a": "b", "c": []any{"d", "e"}, "f": []any{[]any{"g"}, []any{"h"}}},
			[]StringifyOption{WithArrayFormat(ArrayFormatRepeat), WithSort(sortAsc)},
			"a=b&c=d&c=e&f=g&f=h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Stringify(tt.input, tt.opts...)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestJSStringifyEncodeValuesOnlyStrictNullHandling tests encodeValuesOnly with strictNullHandling
func TestJSStringifyEncodeValuesOnlyStrictNullHandling(t *testing.T) {
	result, err := Stringify(
		map[string]any{"a": map[string]any{"b": nil}},
		WithEncodeValuesOnly(true),
		WithStringifyStrictNullHandling(true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a[b]"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyInvalidCharset tests invalid charset throws error
func TestJSStringifyInvalidCharset(t *testing.T) {
	_, err := Stringify(map[string]any{"a": "b"}, WithStringifyCharset("foobar"))
	if err != ErrInvalidStringifyCharset {
		t.Errorf("expected ErrInvalidStringifyCharset, got %v", err)
	}
}

// TestJSStringifyISO88591 tests ISO-8859-1 charset
func TestJSStringifyISO88591(t *testing.T) {
	result, err := Stringify(map[string]any{"æ": "æ"}, WithStringifyCharset(CharsetISO88591))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "%E6=%E6"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyISO88591NumericEntities tests unrepresentable chars as numeric entities
func TestJSStringifyISO88591NumericEntities(t *testing.T) {
	result, err := Stringify(map[string]any{"a": "☺"}, WithStringifyCharset(CharsetISO88591))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=%26%239786%3B"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyUTF8Explicit tests explicit UTF-8 charset
func TestJSStringifyUTF8Explicit(t *testing.T) {
	result, err := Stringify(map[string]any{"a": "æ"}, WithStringifyCharset(CharsetUTF8))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a=%C3%A6"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyCharsetSentinel tests charsetSentinel option
func TestJSStringifyCharsetSentinel(t *testing.T) {
	t.Run("UTF-8 sentinel", func(t *testing.T) {
		result, err := Stringify(map[string]any{"a": "æ"}, WithStringifyCharsetSentinel(true), WithStringifyCharset(CharsetUTF8))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "utf8=%E2%9C%93&a=%C3%A6"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("ISO-8859-1 sentinel", func(t *testing.T) {
		result, err := Stringify(map[string]any{"a": "æ"}, WithStringifyCharsetSentinel(true), WithStringifyCharset(CharsetISO88591))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "utf8=%26%2310003%3B&a=%E6"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestJSStringifyStrictNullHandlingWithCustomFilter tests strictNullHandling with custom filter
func TestJSStringifyStrictNullHandlingWithCustomFilter(t *testing.T) {
	filter := FilterFunc(func(prefix string, value any) any {
		return value
	})

	result, err := Stringify(map[string]any{"key": nil}, WithStringifyStrictNullHandling(true), WithFilter(filter))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "key"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyStrictNullHandlingWithNullSerializeDate tests strictNullHandling with null serializeDate
func TestJSStringifyStrictNullHandlingWithNullSerializeDate(t *testing.T) {
	serializeDate := func(d time.Time) string {
		return ""
	}

	date := time.Now()
	result, err := Stringify(map[string]any{"key": date}, WithStringifyStrictNullHandling(true), WithSerializeDate(serializeDate))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "key="
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyEncoderKeyValue tests encoding keys and values differently
func TestJSStringifyEncoderKeyValue(t *testing.T) {
	encoder := func(str string, charset Charset, kind string, format Format) string {
		encoded := Encode(str, charset, format)
		if kind == "key" {
			return strings.ToLower(encoded)
		}
		if kind == "value" {
			return strings.ToUpper(encoded)
		}
		return encoded
	}

	result, err := Stringify(map[string]any{"KeY": "vAlUe"}, WithEncoder(encoder))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "key=VALUE"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestJSStringifyObjectsInsideArrays tests objects inside arrays
func TestJSStringifyObjectsInsideArrays(t *testing.T) {
	obj := map[string]any{"a": map[string]any{"b": map[string]any{"c": "d", "e": "f"}}}
	withArray := map[string]any{"a": map[string]any{"b": []any{map[string]any{"c": "d", "e": "f"}}}}

	t.Run("no array", func(t *testing.T) {
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"no arrayFormat", []StringifyOption{WithEncode(false), WithSort(func(a, b string) bool { return a < b })}, "a[b][c]=d&a[b][e]=f"},
			{"brackets", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatBrackets), WithSort(func(a, b string) bool { return a < b })}, "a[b][c]=d&a[b][e]=f"},
			{"indices", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatIndices), WithSort(func(a, b string) bool { return a < b })}, "a[b][c]=d&a[b][e]=f"},
			{"repeat", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatRepeat), WithSort(func(a, b string) bool { return a < b })}, "a[b][c]=d&a[b][e]=f"},
			{"comma", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatComma), WithSort(func(a, b string) bool { return a < b })}, "a[b][c]=d&a[b][e]=f"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(obj, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})

	t.Run("with array", func(t *testing.T) {
		tests := []struct {
			name     string
			opts     []StringifyOption
			expected string
		}{
			{"no arrayFormat", []StringifyOption{WithEncode(false), WithSort(func(a, b string) bool { return a < b })}, "a[b][0][c]=d&a[b][0][e]=f"},
			{"brackets", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatBrackets), WithSort(func(a, b string) bool { return a < b })}, "a[b][][c]=d&a[b][][e]=f"},
			{"indices", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatIndices), WithSort(func(a, b string) bool { return a < b })}, "a[b][0][c]=d&a[b][0][e]=f"},
			{"repeat", []StringifyOption{WithEncode(false), WithArrayFormat(ArrayFormatRepeat), WithSort(func(a, b string) bool { return a < b })}, "a[b][c]=d&a[b][e]=f"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := Stringify(withArray, tt.opts...)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			})
		}
	})
}

// TestJSStringifyEmptyKeys tests stringifying empty keys
func TestJSStringifyEmptyKeys(t *testing.T) {
	type emptyKeyTestCase struct {
		input           string
		withEmptyKeys   map[string]any
		stringifyOutput map[string]string
	}

	testCases := []emptyKeyTestCase{
		{
			input:         "&=",
			withEmptyKeys: map[string]any{"": ""},
			stringifyOutput: map[string]string{
				"brackets": "=",
				"indices":  "=",
				"repeat":   "=",
			},
		},
		{
			input:         "=",
			withEmptyKeys: map[string]any{"": ""},
			stringifyOutput: map[string]string{
				"brackets": "=",
				"indices":  "=",
				"repeat":   "=",
			},
		},
		{
			input:         "=a",
			withEmptyKeys: map[string]any{"": "a"},
			stringifyOutput: map[string]string{
				"brackets": "=a",
				"indices":  "=a",
				"repeat":   "=a",
			},
		},
		{
			input:         "a==a",
			withEmptyKeys: map[string]any{"a": "=a"},
			stringifyOutput: map[string]string{
				"brackets": "a==a",
				"indices":  "a==a",
				"repeat":   "a==a",
			},
		},
	}

	for _, tc := range testCases {
		t.Run("test case: "+tc.input, func(t *testing.T) {
			for format, expected := range tc.stringifyOutput {
				t.Run(format, func(t *testing.T) {
					var opts []StringifyOption
					opts = append(opts, WithEncode(false))
					switch format {
					case "indices":
						opts = append(opts, WithArrayFormat(ArrayFormatIndices))
					case "brackets":
						opts = append(opts, WithArrayFormat(ArrayFormatBrackets))
					case "repeat":
						opts = append(opts, WithArrayFormat(ArrayFormatRepeat))
					}

					result, err := Stringify(tc.withEmptyKeys, opts...)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
						return
					}
					if result != expected {
						t.Errorf("expected %q, got %q", expected, result)
					}
				})
			}
		})
	}
}

// TestJSStringifyEmptyKeysEdgeCases tests edge cases with empty keys
func TestJSStringifyEmptyKeysEdgeCases(t *testing.T) {
	t.Run("empty string key with nested empty array", func(t *testing.T) {
		result, err := Stringify(map[string]any{"": map[string]any{"": []any{2, 3}}}, WithEncode(false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "[][0]=2&[][1]=3"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("empty string key with nested array and value", func(t *testing.T) {
		result, err := Stringify(map[string]any{"": map[string]any{"": []any{2, 3}, "a": 2}}, WithEncode(false), WithSort(func(a, b string) bool { return a < b }))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "[][0]=2&[][1]=3&[a]=2"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("empty string key with nested empty array indices", func(t *testing.T) {
		result, err := Stringify(map[string]any{"": map[string]any{"": []any{2, 3}}}, WithEncode(false), WithArrayFormat(ArrayFormatIndices))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "[][0]=2&[][1]=3"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("empty string key with nested array and value indices", func(t *testing.T) {
		result, err := Stringify(map[string]any{"": map[string]any{"": []any{2, 3}, "a": 2}}, WithEncode(false), WithArrayFormat(ArrayFormatIndices), WithSort(func(a, b string) bool { return a < b }))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "[][0]=2&[][1]=3&[a]=2"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestJSStringifyLongString tests encoding a very long string
func TestJSStringifyLongString(t *testing.T) {
	var chars []string
	var expected []string
	for i := 0; i < 5000; i++ {
		chars = append(chars, " "+toString(i))
		expected = append(expected, "%20"+toString(i))
	}

	obj := map[string]any{"foo": strings.Join(chars, "")}
	result, err := Stringify(obj, WithArrayFormat(ArrayFormatBrackets), WithStringifyCharset(CharsetUTF8))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedResult := "foo=" + strings.Join(expected, "")
	if result != expectedResult {
		t.Errorf("result length mismatch: expected %d, got %d", len(expectedResult), len(result))
	}
}

// TestJSStringifySparseArrays tests stringifying sparse arrays
// JS: t.test('stringifies sparse arrays', ...)
func TestJSStringifySparseArrays(t *testing.T) {
	// In Go, we represent sparse arrays with nil elements
	// [, '2', , , '1'] becomes []any{nil, "2", nil, nil, "1"}

	t.Run("simple sparse array with indices", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, '2', , , '1'] }, { encodeValuesOnly: true, arrayFormat: 'indices' }), 'a[1]=2&a[4]=1');
		result, err := Stringify(map[string]any{"a": []any{nil, "2", nil, nil, "1"}},
			WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[1]=2&a[4]=1" {
			t.Errorf("expected 'a[1]=2&a[4]=1', got %q", result)
		}
	})

	t.Run("simple sparse array with brackets", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, '2', , , '1'] }, { encodeValuesOnly: true, arrayFormat: 'brackets' }), 'a[]=2&a[]=1');
		result, err := Stringify(map[string]any{"a": []any{nil, "2", nil, nil, "1"}},
			WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[]=2&a[]=1" {
			t.Errorf("expected 'a[]=2&a[]=1', got %q", result)
		}
	})

	t.Run("simple sparse array with repeat", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, '2', , , '1'] }, { encodeValuesOnly: true, arrayFormat: 'repeat' }), 'a=2&a=1');
		result, err := Stringify(map[string]any{"a": []any{nil, "2", nil, nil, "1"}},
			WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatRepeat))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a=2&a=1" {
			t.Errorf("expected 'a=2&a=1', got %q", result)
		}
	})

	t.Run("nested sparse array with object indices", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, { b: [, , { c: '1' }] }] }, { encodeValuesOnly: true, arrayFormat: 'indices' }), 'a[1][b][2][c]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, map[string]any{"b": []any{nil, nil, map[string]any{"c": "1"}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[1][b][2][c]=1" {
			t.Errorf("expected 'a[1][b][2][c]=1', got %q", result)
		}
	})

	t.Run("nested sparse array with object brackets", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, { b: [, , { c: '1' }] }] }, { encodeValuesOnly: true, arrayFormat: 'brackets' }), 'a[][b][][c]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, map[string]any{"b": []any{nil, nil, map[string]any{"c": "1"}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[][b][][c]=1" {
			t.Errorf("expected 'a[][b][][c]=1', got %q", result)
		}
	})

	t.Run("nested sparse array with object repeat", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, { b: [, , { c: '1' }] }] }, { encodeValuesOnly: true, arrayFormat: 'repeat' }), 'a[b][c]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, map[string]any{"b": []any{nil, nil, map[string]any{"c": "1"}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatRepeat))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[b][c]=1" {
			t.Errorf("expected 'a[b][c]=1', got %q", result)
		}
	})

	t.Run("deeply nested sparse arrays indices", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, [, , [, , , { c: '1' }]]] }, { encodeValuesOnly: true, arrayFormat: 'indices' }), 'a[1][2][3][c]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, []any{nil, nil, []any{nil, nil, nil, map[string]any{"c": "1"}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[1][2][3][c]=1" {
			t.Errorf("expected 'a[1][2][3][c]=1', got %q", result)
		}
	})

	t.Run("deeply nested sparse arrays brackets", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, [, , [, , , { c: '1' }]]] }, { encodeValuesOnly: true, arrayFormat: 'brackets' }), 'a[][][][c]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, []any{nil, nil, []any{nil, nil, nil, map[string]any{"c": "1"}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[][][][c]=1" {
			t.Errorf("expected 'a[][][][c]=1', got %q", result)
		}
	})

	t.Run("deeply nested sparse arrays repeat", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, [, , [, , , { c: '1' }]]] }, { encodeValuesOnly: true, arrayFormat: 'repeat' }), 'a[c]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, []any{nil, nil, []any{nil, nil, nil, map[string]any{"c": "1"}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatRepeat))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[c]=1" {
			t.Errorf("expected 'a[c]=1', got %q", result)
		}
	})

	t.Run("deeply nested sparse arrays with sparse value indices", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, [, , [, , , { c: [, '1'] }]]] }, { encodeValuesOnly: true, arrayFormat: 'indices' }), 'a[1][2][3][c][1]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, []any{nil, nil, []any{nil, nil, nil, map[string]any{"c": []any{nil, "1"}}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatIndices))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[1][2][3][c][1]=1" {
			t.Errorf("expected 'a[1][2][3][c][1]=1', got %q", result)
		}
	})

	t.Run("deeply nested sparse arrays with sparse value brackets", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, [, , [, , , { c: [, '1'] }]]] }, { encodeValuesOnly: true, arrayFormat: 'brackets' }), 'a[][][][c][]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, []any{nil, nil, []any{nil, nil, nil, map[string]any{"c": []any{nil, "1"}}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatBrackets))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[][][][c][]=1" {
			t.Errorf("expected 'a[][][][c][]=1', got %q", result)
		}
	})

	t.Run("deeply nested sparse arrays with sparse value repeat", func(t *testing.T) {
		// st.equal(qs.stringify({ a: [, [, , [, , , { c: [, '1'] }]]] }, { encodeValuesOnly: true, arrayFormat: 'repeat' }), 'a[c]=1');
		result, err := Stringify(map[string]any{
			"a": []any{nil, []any{nil, nil, []any{nil, nil, nil, map[string]any{"c": []any{nil, "1"}}}}},
		}, WithEncodeValuesOnly(true), WithArrayFormat(ArrayFormatRepeat))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "a[c]=1" {
			t.Errorf("expected 'a[c]=1', got %q", result)
		}
	})
}

// TestJSStringifyNonStringKeys tests stringifying with non-string keys in filter
// JS: t.test('stringifies non-string keys', ...)
// Note: In Go, map keys are always strings, but filter can contain non-string values
func TestJSStringifyNonStringKeys(t *testing.T) {
	// In JS: qs.stringify({ a: 'b', 'false': {}, 1e+22: 'c', d: 'e' }, {
	//     filter: ['a', false, null, 10000000000000000000000, S],
	//     allowDots: true,
	//     encodeDotInKeys: true
	// });
	// Result: 'a=b&1e%2B22=c&d=e'

	// In Go, we need to use string keys
	result, err := Stringify(map[string]any{
		"a":       "b",
		"false":   map[string]any{},
		"1e+22":   "c",
		"d":       "e",
	}, WithFilter([]string{"a", "1e+22", "d"}), WithStringifyAllowDots(true), WithEncodeDotInKeys(true))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include a=b, 1e%2B22=c (encoded +), and d=e
	// 'false' key maps to empty object so it's skipped
	if result != "a=b&1e%2B22=c&d=e" {
		t.Errorf("expected 'a=b&1e%%2B22=c&d=e', got %q", result)
	}
}

// TestSortArrayIndices tests that array indices are sorted as strings when SortArrayIndices is true
func TestSortArrayIndices(t *testing.T) {
	sortAsc := func(a, b string) bool { return a < b }

	// 12 element array - with string sort: 0, 1, 10, 11, 2, 3, 4, 5, 6, 7, 8, 9
	input := map[string]any{
		"arr": []any{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"},
	}

	// Without SortArrayIndices - numeric order
	resultNumeric, err := Stringify(input, WithEncode(false), WithSort(sortAsc))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}
	expectedNumeric := "arr[0]=a&arr[1]=b&arr[2]=c&arr[3]=d&arr[4]=e&arr[5]=f&arr[6]=g&arr[7]=h&arr[8]=i&arr[9]=j&arr[10]=k&arr[11]=l"
	if resultNumeric != expectedNumeric {
		t.Errorf("Without SortArrayIndices:\nGot:      %s\nExpected: %s", resultNumeric, expectedNumeric)
	}

	// With SortArrayIndices - string sort order (0, 1, 10, 11, 2, 3, ...)
	resultString, err := Stringify(input, WithEncode(false), WithSort(sortAsc), WithSortArrayIndices(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}
	expectedString := "arr[0]=a&arr[1]=b&arr[10]=k&arr[11]=l&arr[2]=c&arr[3]=d&arr[4]=e&arr[5]=f&arr[6]=g&arr[7]=h&arr[8]=i&arr[9]=j"
	if resultString != expectedString {
		t.Errorf("With SortArrayIndices:\nGot:      %s\nExpected: %s", resultString, expectedString)
	}
}
