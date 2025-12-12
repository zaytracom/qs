// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"reflect"
	"testing"
)

func TestDetailedJSCompat(t *testing.T) {
	tests := []struct {
		input    string
		opts     []ParseOption
		expected map[string]any
		name     string
	}{
		// parses a simple string
		{"0=foo", nil, map[string]any{"0": "foo"}, "0=foo"},
		{"foo=c++", nil, map[string]any{"foo": "c  "}, "foo=c++"},
		{"a[>=]=23", nil, map[string]any{"a": map[string]any{">=": "23"}}, "a[>=]=23"},
		{"a[<=>]==23", nil, map[string]any{"a": map[string]any{"<=>": "=23"}}, "a[<=>]==23"},
		{"a[==]=23", nil, map[string]any{"a": map[string]any{"==": "23"}}, "a[==]=23"},
		{"foo", []ParseOption{WithStrictNullHandling(true)}, map[string]any{"foo": nil}, "foo strictNull"},
		{"foo", nil, map[string]any{"foo": ""}, "foo empty"},
		{"foo=", nil, map[string]any{"foo": ""}, "foo="},
		{"foo=bar", nil, map[string]any{"foo": "bar"}, "foo=bar"},
		{" foo = bar = baz ", nil, map[string]any{" foo ": " bar = baz "}, "spaces"},
		{"foo=bar=baz", nil, map[string]any{"foo": "bar=baz"}, "foo=bar=baz"},
		{"foo=bar&bar=baz", nil, map[string]any{"foo": "bar", "bar": "baz"}, "foo=bar&bar=baz"},

		// comma tests
		{"a[]=b&a[]=c", nil, map[string]any{"a": []any{"b", "c"}}, "a[]=b&a[]=c"},
		{"a[0]=b&a[1]=c", nil, map[string]any{"a": []any{"b", "c"}}, "a[0]=b&a[1]=c"},
		{"a=b,c", nil, map[string]any{"a": "b,c"}, "a=b,c no comma"},
		{"a=b&a=c", nil, map[string]any{"a": []any{"b", "c"}}, "a=b&a=c"},
		{"a=b,c", []ParseOption{WithComma(true)}, map[string]any{"a": []any{"b", "c"}}, "a=b,c comma"},

		// dot notation
		{"a.b=c", nil, map[string]any{"a.b": "c"}, "a.b=c no dots"},
		{"a.b=c", []ParseOption{WithAllowDots(true)}, map[string]any{"a": map[string]any{"b": "c"}}, "a.b=c dots"},

		// nested
		{"a[b]=c", nil, map[string]any{"a": map[string]any{"b": "c"}}, "a[b]=c"},
		{"a[b][c]=d", nil, map[string]any{"a": map[string]any{"b": map[string]any{"c": "d"}}}, "a[b][c]=d"},
		{"a[b][c][d][e][f][g][h]=i", nil, map[string]any{"a": map[string]any{"b": map[string]any{"c": map[string]any{"d": map[string]any{"e": map[string]any{"f": map[string]any{"[g][h]": "i"}}}}}}}, "depth 5 default"},

		// depth
		{"a[b][c]=d", []ParseOption{WithDepth(1)}, map[string]any{"a": map[string]any{"b": map[string]any{"[c]": "d"}}}, "depth 1"},
		{"a[0]=b&a[1]=c", []ParseOption{WithDepth(0)}, map[string]any{"a[0]": "b", "a[1]": "c"}, "depth 0"},

		// arrays
		{"a[]=b", nil, map[string]any{"a": []any{"b"}}, "a[]=b"},
		{"a[1]=c&a[0]=b&a[2]=d", nil, map[string]any{"a": []any{"b", "c", "d"}}, "reorder indices"},
		{"a[1]=c", []ParseOption{WithArrayLimit(20)}, map[string]any{"a": []any{"c"}}, "sparse compact"},
		{"a[1]=c", []ParseOption{WithArrayLimit(0)}, map[string]any{"a": map[string]any{"1": "c"}}, "arrayLimit 0"},
		{"a[20]=a", []ParseOption{WithArrayLimit(20)}, map[string]any{"a": []any{"a"}}, "at limit"},
		{"a[21]=a", []ParseOption{WithArrayLimit(20)}, map[string]any{"a": map[string]any{"21": "a"}}, "over limit"},

		// sparse compact
		{"a[10]=1&a[2]=2", []ParseOption{WithArrayLimit(20)}, map[string]any{"a": []any{"2", "1"}}, "sparse compact"},

		// prototype
		{"a[hasOwnProperty]=b", []ParseOption{WithAllowPrototypes(false)}, map[string]any{}, "hasOwn blocked"},
		{"hasOwnProperty=b", []ParseOption{WithAllowPrototypes(false)}, map[string]any{}, "top hasOwn blocked"},
		{"a[hasOwnProperty]=b", []ParseOption{WithAllowPrototypes(true)}, map[string]any{"a": map[string]any{"hasOwnProperty": "b"}}, "hasOwn allowed"},

		// special
		{"a[b]=c&a=d", nil, map[string]any{"a": map[string]any{"b": "c", "d": true}}, "add key to obj"},
		{"[]=&a=b", nil, map[string]any{"0": "", "a": "b"}, "[] prefix"},
		{"[foo]=bar", nil, map[string]any{"foo": "bar"}, "[foo] key"},

		// duplicates
		{"foo=bar&foo=baz", nil, map[string]any{"foo": []any{"bar", "baz"}}, "dup combine"},
		{"foo=bar&foo=baz", []ParseOption{WithDuplicates(DuplicateFirst)}, map[string]any{"foo": "bar"}, "dup first"},
		{"foo=bar&foo=baz", []ParseOption{WithDuplicates(DuplicateLast)}, map[string]any{"foo": "baz"}, "dup last"},

		// empty arrays
		{"foo[]&bar=baz", []ParseOption{WithAllowEmptyArrays(true)}, map[string]any{"foo": []any{}, "bar": "baz"}, "empty array true"},
		{"foo[]&bar=baz", []ParseOption{WithAllowEmptyArrays(false)}, map[string]any{"foo": []any{""}, "bar": "baz"}, "empty array false"},

		// charset
		{"%A2=%BD", []ParseOption{WithCharset(CharsetISO88591)}, map[string]any{"¢": "½"}, "iso-8859-1"},
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
