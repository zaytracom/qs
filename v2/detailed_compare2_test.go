// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"reflect"
	"testing"
)

func TestDetailedJSCompat2(t *testing.T) {
	tests := []struct {
		input    string
		opts     []ParseOption
		expected map[string]any
		name     string
	}{
		// Mixed arrays and objects
		{"foo[0]=bar&foo[bad]=baz", nil, map[string]any{"foo": map[string]any{"0": "bar", "bad": "baz"}}, "mixed 1"},
		{"foo[bad]=baz&foo[0]=bar", nil, map[string]any{"foo": map[string]any{"bad": "baz", "0": "bar"}}, "mixed 2"},
		{"foo[bad]=baz&foo[]=bar", nil, map[string]any{"foo": map[string]any{"bad": "baz", "0": "bar"}}, "mixed 3"},
		{"foo[]=bar&foo[bad]=baz", nil, map[string]any{"foo": map[string]any{"0": "bar", "bad": "baz"}}, "mixed 4"},

		// Nested arrays
		{"a[0][0]=b&a[0][1]=c&a[1][0]=d", nil, map[string]any{"a": []any{[]any{"b", "c"}, []any{"d"}}}, "nested arr"},
		{"a[0][b]=c&a[1][d]=e", nil, map[string]any{"a": []any{map[string]any{"b": "c"}, map[string]any{"d": "e"}}}, "arr of obj"},

		// Empty strings in arrays
		{"a[]=b&a[]=&a[]=c", nil, map[string]any{"a": []any{"b", "", "c"}}, "empty in arr 1"},
		{"a[]=&a[]=b&a[]=c", nil, map[string]any{"a": []any{"", "b", "c"}}, "empty in arr 2"},

		// URL encoding
		{"a[b%20c]=d", nil, map[string]any{"a": map[string]any{"b c": "d"}}, "url enc key"},
		{"he%3Dllo=th%3Dere", nil, map[string]any{"he=llo": "th=ere"}, "url enc equals"},

		// Brackets in value
		{`pets=["tobi"]`, nil, map[string]any{"pets": `["tobi"]`}, "brackets val"},

		// Malformed
		{"{%:%}", []ParseOption{WithStrictNullHandling(true)}, map[string]any{"{%:%}": nil}, "malformed null"},
		{"foo=%:%}", nil, map[string]any{"foo": "%:%}"}, "malformed val"},

		// __proto__ always blocked
		{"categories[__proto__]=login&categories[length]=42", []ParseOption{WithAllowPrototypes(true)}, map[string]any{"categories": map[string]any{"length": "42"}}, "__proto__ blocked"},
		{"__proto__=bad", []ParseOption{WithAllowPrototypes(true)}, map[string]any{}, "__proto__ top"},

		// Encoded brackets
		{"a%5Bb%5D=c", nil, map[string]any{"a": map[string]any{"b": "c"}}, "enc brackets"},
		{"%5B0%5D=a&%5B1%5D=b", nil, map[string]any{"0": "a", "1": "b"}, "enc bracket idx"},

		// Charset sentinel
		{"utf8=%E2%9C%93&a=b", []ParseOption{WithCharsetSentinel(true)}, map[string]any{"a": "b"}, "charset sentinel"},

		// Numeric entities
		{"foo=%26%239786%3B", []ParseOption{WithCharset(CharsetISO88591), WithInterpretNumericEntities(true)}, map[string]any{"foo": "☺"}, "numeric entity"},

		// Parameter limit
		{"a=b&c=d&e=f", []ParseOption{WithParameterLimit(2)}, map[string]any{"a": "b", "c": "d"}, "param limit"},

		// StrictDepth ok
		{"a[b][c]=d", []ParseOption{WithDepth(2), WithStrictDepth(true)}, map[string]any{"a": map[string]any{"b": map[string]any{"c": "d"}}}, "strictDepth ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input, tt.opts...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Parse(%q)\n  got:  %#v\n  want: %#v", tt.input, result, tt.expected)
			}
		})
	}
}
