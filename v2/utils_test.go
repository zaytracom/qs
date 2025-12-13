// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"regexp"
	"testing"
)

// TestHexTable verifies the pre-computed hex table is correct.
func TestHexTable(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "%00"},
		{15, "%0F"},
		{16, "%10"},
		{32, "%20"}, // space
		{37, "%25"}, // %
		{38, "%26"}, // &
		{43, "%2B"}, // +
		{61, "%3D"}, // =
		{255, "%FF"},
	}

	for _, tt := range tests {
		if hexTable[tt.index] != tt.expected {
			t.Errorf("hexTable[%d] = %q, want %q", tt.index, hexTable[tt.index], tt.expected)
		}
	}
}

// TestEncode tests URL encoding with different charsets and formats.
func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		charset  Charset
		format   Format
		expected string
	}{
		// Empty string
		{"empty string", "", CharsetUTF8, FormatRFC3986, ""},

		// Unreserved characters (should not be encoded)
		{"unreserved lowercase", "abc", CharsetUTF8, FormatRFC3986, "abc"},
		{"unreserved uppercase", "ABC", CharsetUTF8, FormatRFC3986, "ABC"},
		{"unreserved digits", "123", CharsetUTF8, FormatRFC3986, "123"},
		{"unreserved special", "-._~", CharsetUTF8, FormatRFC3986, "-._~"},

		// Reserved characters (should be encoded)
		{"space", " ", CharsetUTF8, FormatRFC3986, "%20"},
		{"ampersand", "&", CharsetUTF8, FormatRFC3986, "%26"},
		{"equals", "=", CharsetUTF8, FormatRFC3986, "%3D"},
		{"plus", "+", CharsetUTF8, FormatRFC3986, "%2B"},
		{"percent", "%", CharsetUTF8, FormatRFC3986, "%25"},
		{"question", "?", CharsetUTF8, FormatRFC3986, "%3F"},
		{"hash", "#", CharsetUTF8, FormatRFC3986, "%23"},

		// RFC 1738 allows ( and ) unencoded
		{"parentheses RFC3986", "()", CharsetUTF8, FormatRFC3986, "%28%29"},
		{"parentheses RFC1738", "()", CharsetUTF8, FormatRFC1738, "()"},

		// UTF-8 multi-byte characters
		{"utf8 2-byte", "é", CharsetUTF8, FormatRFC3986, "%C3%A9"},
		{"utf8 3-byte", "☺", CharsetUTF8, FormatRFC3986, "%E2%98%BA"},
		{"utf8 4-byte emoji", "😀", CharsetUTF8, FormatRFC3986, "%F0%9F%98%80"},
		{"utf8 chinese", "中", CharsetUTF8, FormatRFC3986, "%E4%B8%AD"},

		// ISO-8859-1 encoding
		{"iso latin char", "é", CharsetISO88591, FormatRFC3986, "%E9"},
		{"iso space", " ", CharsetISO88591, FormatRFC3986, "%20"},
		// Characters outside Latin-1 are encoded as numeric entities
		{"iso non-latin", "☺", CharsetISO88591, FormatRFC3986, "%26%239786%3B"},
		{"iso emoji", "😀", CharsetISO88591, FormatRFC3986, "%26%23128512%3B"},

		// Mixed content
		{"mixed ascii and utf8", "hello world!", CharsetUTF8, FormatRFC3986, "hello%20world%21"},
		{"query string", "a=b&c=d", CharsetUTF8, FormatRFC3986, "a%3Db%26c%3Dd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Encode(tt.input, tt.charset, tt.format)
			if result != tt.expected {
				t.Errorf("Encode(%q, %q, %q) = %q, want %q",
					tt.input, tt.charset, tt.format, result, tt.expected)
			}
		})
	}
}

// TestDecode tests URL decoding with different charsets.
func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		charset  Charset
		expected string
	}{
		// Empty string
		{"empty string", "", CharsetUTF8, ""},

		// Plus as space
		{"plus to space", "hello+world", CharsetUTF8, "hello world"},
		{"multiple plus", "a+b+c", CharsetUTF8, "a b c"},

		// Percent encoding
		{"percent space", "hello%20world", CharsetUTF8, "hello world"},
		{"percent ampersand", "a%26b", CharsetUTF8, "a&b"},
		{"percent equals", "a%3Db", CharsetUTF8, "a=b"},

		// UTF-8 decoding
		{"utf8 2-byte", "%C3%A9", CharsetUTF8, "é"},
		{"utf8 3-byte", "%E2%98%BA", CharsetUTF8, "☺"},
		{"utf8 4-byte", "%F0%9F%98%80", CharsetUTF8, "😀"},

		// ISO-8859-1 decoding (converts Latin-1 to proper UTF-8)
		{"iso latin", "%E9", CharsetISO88591, "é"},
		{"iso space", "%20", CharsetISO88591, " "},
		{"iso plus", "a+b", CharsetISO88591, "a b"},

		// Invalid sequences (graceful fallback)
		{"invalid percent", "%GG", CharsetUTF8, "%GG"},
		{"incomplete percent", "%2", CharsetUTF8, "%2"},

		// Mixed content
		{"mixed", "hello%20world%21", CharsetUTF8, "hello world!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Decode(tt.input, tt.charset)
			if result != tt.expected {
				t.Errorf("Decode(%q, %q) = %q, want %q",
					tt.input, tt.charset, result, tt.expected)
			}
		})
	}
}

// TestMerge tests deep merging of values.
func TestMerge(t *testing.T) {
	t.Run("nil source returns target", func(t *testing.T) {
		target := map[string]any{"a": "b"}
		result := Merge(target, nil)
		if m, ok := result.(map[string]any); !ok || m["a"] != "b" {
			t.Errorf("expected target unchanged, got %v", result)
		}
	})

	t.Run("primitive source appends to slice target", func(t *testing.T) {
		target := []any{"a", "b"}
		result := Merge(target, "c")
		slice, ok := result.([]any)
		if !ok || len(slice) != 3 || slice[2] != "c" {
			t.Errorf("expected [a b c], got %v", result)
		}
	})

	t.Run("primitive source sets key on map target", func(t *testing.T) {
		target := map[string]any{"a": "b"}
		result := Merge(target, "c")
		m, ok := result.(map[string]any)
		if !ok || m["c"] != true {
			t.Errorf("expected {a:b, c:true}, got %v", result)
		}
	})

	t.Run("prototype key is normal in Go", func(t *testing.T) {
		target := map[string]any{}
		result := Merge(target, "__proto__")
		m := result.(map[string]any)
		if m["__proto__"] != true {
			t.Error("__proto__ should be a normal key")
		}
	})

	t.Run("merge two maps", func(t *testing.T) {
		target := map[string]any{"a": "1"}
		source := map[string]any{"b": "2"}
		result := Merge(target, source)
		m := result.(map[string]any)
		if m["a"] != "1" || m["b"] != "2" {
			t.Errorf("expected {a:1, b:2}, got %v", result)
		}
	})

	t.Run("deep merge nested maps", func(t *testing.T) {
		target := map[string]any{
			"a": map[string]any{"x": "1"},
		}
		source := map[string]any{
			"a": map[string]any{"y": "2"},
		}
		result := Merge(target, source)
		m := result.(map[string]any)
		nested := m["a"].(map[string]any)
		if nested["x"] != "1" || nested["y"] != "2" {
			t.Errorf("expected nested {x:1, y:2}, got %v", nested)
		}
	})

	t.Run("merge slices", func(t *testing.T) {
		target := []any{"a", "b"}
		source := []any{"c", "d", "e"}
		result := Merge(target, source)
		slice := result.([]any)
		// Existing indices are kept, new ones added
		if len(slice) < 3 {
			t.Errorf("expected at least 3 elements, got %v", slice)
		}
	})

	t.Run("primitive target with map source", func(t *testing.T) {
		result := Merge("a", map[string]any{"b": "c"})
		slice, ok := result.([]any)
		if !ok || len(slice) != 2 {
			t.Errorf("expected [a, {b:c}], got %v", result)
		}
	})
}

// TestArrayToObject tests slice to map conversion.
func TestArrayToObject(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected map[string]any
	}{
		{
			"empty slice",
			[]any{},
			map[string]any{},
		},
		{
			"simple slice",
			[]any{"a", "b", "c"},
			map[string]any{"0": "a", "1": "b", "2": "c"},
		},
		{
			"slice with nil",
			[]any{"a", nil, "c"},
			map[string]any{"0": "a", "2": "c"},
		},
		{
			"all nil",
			[]any{nil, nil, nil},
			map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrayToObject(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ArrayToObject(%v) length = %d, want %d", tt.input, len(result), len(tt.expected))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("ArrayToObject(%v)[%q] = %v, want %v", tt.input, k, result[k], v)
				}
			}
		})
	}
}

// TestCompact tests removal of nil holes from nested structures.
func TestCompact(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result := Compact(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("primitive value", func(t *testing.T) {
		result := Compact("hello")
		if result != "hello" {
			t.Errorf("expected hello, got %v", result)
		}
	})

	t.Run("slice with nil holes", func(t *testing.T) {
		input := []any{"a", nil, "b", nil, "c"}
		result := Compact(input)
		slice := result.([]any)
		if len(slice) != 3 {
			t.Errorf("expected 3 elements, got %d: %v", len(slice), slice)
		}
		if slice[0] != "a" || slice[1] != "b" || slice[2] != "c" {
			t.Errorf("expected [a b c], got %v", slice)
		}
	})

	t.Run("nested slice with nil holes", func(t *testing.T) {
		input := map[string]any{
			"arr": []any{"a", nil, "b"},
		}
		result := Compact(input)
		m := result.(map[string]any)
		arr := m["arr"].([]any)
		if len(arr) != 2 || arr[0] != "a" || arr[1] != "b" {
			t.Errorf("expected [a b], got %v", arr)
		}
	})

	t.Run("deeply nested", func(t *testing.T) {
		input := map[string]any{
			"level1": map[string]any{
				"level2": []any{"a", nil, "b"},
			},
		}
		result := Compact(input)
		m := result.(map[string]any)
		level1 := m["level1"].(map[string]any)
		level2 := level1["level2"].([]any)
		if len(level2) != 2 {
			t.Errorf("expected 2 elements in level2, got %v", level2)
		}
	})
}

// TestCombine tests combining values into a slice.
func TestCombine(t *testing.T) {
	tests := []struct {
		name     string
		a        any
		b        any
		expected []any
	}{
		{"two primitives", "a", "b", []any{"a", "b"}},
		{"slice and primitive", []any{"a", "b"}, "c", []any{"a", "b", "c"}},
		{"primitive and slice", "a", []any{"b", "c"}, []any{"a", "b", "c"}},
		{"two slices", []any{"a"}, []any{"b", "c"}, []any{"a", "b", "c"}},
		{"nil and value", nil, "a", []any{"a"}},
		{"value and nil", "a", nil, []any{"a"}},
		{"both nil", nil, nil, []any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Combine(tt.a, tt.b)
			if len(result) != len(tt.expected) {
				t.Errorf("Combine(%v, %v) length = %d, want %d", tt.a, tt.b, len(result), len(tt.expected))
				return
			}
			for i, v := range tt.expected {
				if result[i] != v {
					t.Errorf("Combine(%v, %v)[%d] = %v, want %v", tt.a, tt.b, i, result[i], v)
				}
			}
		})
	}
}

// TestMaybeMap tests applying a function to values.
func TestMaybeMap(t *testing.T) {
	double := func(v any) any {
		if s, ok := v.(string); ok {
			return s + s
		}
		return v
	}

	t.Run("single value", func(t *testing.T) {
		result := MaybeMap("a", double)
		if result != "aa" {
			t.Errorf("expected aa, got %v", result)
		}
	})

	t.Run("slice value", func(t *testing.T) {
		result := MaybeMap([]any{"a", "b"}, double)
		slice := result.([]any)
		if slice[0] != "aa" || slice[1] != "bb" {
			t.Errorf("expected [aa bb], got %v", slice)
		}
	})
}

// TestIsRegExp tests regular expression detection.
func TestIsRegExp(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"compiled regexp", regexp.MustCompile("test"), true},
		{"string", "test", false},
		{"nil", nil, false},
		{"int", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRegExp(tt.input)
			if result != tt.expected {
				t.Errorf("IsRegExp(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestAssign tests map assignment.
func TestAssign(t *testing.T) {
	target := map[string]any{"a": "1"}
	source := map[string]any{"b": "2", "c": "3"}

	result := Assign(target, source)

	if result["a"] != "1" || result["b"] != "2" || result["c"] != "3" {
		t.Errorf("Assign failed: got %v", result)
	}

	// Verify target was modified in place
	if target["b"] != "2" {
		t.Error("Assign should modify target in place")
	}
}

// TestEncodeDecodeRoundTrip tests that encode/decode are inverses.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []string{
		"hello world",
		"a=b&c=d",
		"special chars: !@#$%^&*()",
		"unicode: café ☺ 中文",
		"emoji: 😀🎉",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			encoded := Encode(input, CharsetUTF8, FormatRFC3986)
			decoded := Decode(encoded, CharsetUTF8)
			if decoded != input {
				t.Errorf("round-trip failed: %q -> %q -> %q", input, encoded, decoded)
			}
		})
	}
}
