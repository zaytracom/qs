// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"testing"
	"time"
)

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
				return nil // Remove 'b' - but in JS this actually results in key=
			}
			return value
		})
		result, err := Stringify(input, WithFilter(filter), WithSort(func(a, b string) bool { return a < b }))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// In JS, returning undefined from filter just means the value becomes undefined
		// which still gets stringified as key= (or key with strictNullHandling)
		if result != "a=1&b=" {
			t.Errorf("expected 'a=1&b=', got %q", result)
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
