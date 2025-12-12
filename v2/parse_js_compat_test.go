// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

// This file contains tests ported from the JavaScript qs library test suite.
// Each test case is designed to match the exact behavior of the JS implementation.
// Reference: https://github.com/ljharb/qs/blob/main/test/parse.js

package qs

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Helper function to compare results with expected values
func assertEqual(t *testing.T, got, want any, msg string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: got %#v, want %#v", msg, got, want)
	}
}

// ===========================================
// JS Test: "parses a simple string"
// ===========================================
func TestJSParseSimpleString(t *testing.T) {
	tests := []struct {
		input    string
		opts     []ParseOption
		expected map[string]any
	}{
		// st.deepEqual(qs.parse('0=foo'), { 0: 'foo' });
		{"0=foo", nil, map[string]any{"0": "foo"}},

		// st.deepEqual(qs.parse('foo=c++'), { foo: 'c  ' });
		{"foo=c++", nil, map[string]any{"foo": "c  "}},

		// st.deepEqual(qs.parse('a[>=]=23'), { a: { '>=': '23' } });
		{"a[>=]=23", nil, map[string]any{"a": map[string]any{">=": "23"}}},

		// st.deepEqual(qs.parse('a[<=>]==23'), { a: { '<=>': '=23' } });
		{"a[<=>]==23", nil, map[string]any{"a": map[string]any{"<=>": "=23"}}},

		// st.deepEqual(qs.parse('a[==]=23'), { a: { '==': '23' } });
		{"a[==]=23", nil, map[string]any{"a": map[string]any{"==": "23"}}},

		// st.deepEqual(qs.parse('foo', { strictNullHandling: true }), { foo: null });
		{"foo", []ParseOption{WithStrictNullHandling(true)}, map[string]any{"foo": nil}},

		// st.deepEqual(qs.parse('foo'), { foo: '' });
		{"foo", nil, map[string]any{"foo": ""}},

		// st.deepEqual(qs.parse('foo='), { foo: '' });
		{"foo=", nil, map[string]any{"foo": ""}},

		// st.deepEqual(qs.parse('foo=bar'), { foo: 'bar' });
		{"foo=bar", nil, map[string]any{"foo": "bar"}},

		// st.deepEqual(qs.parse(' foo = bar = baz '), { ' foo ': ' bar = baz ' });
		{" foo = bar = baz ", nil, map[string]any{" foo ": " bar = baz "}},

		// st.deepEqual(qs.parse('foo=bar=baz'), { foo: 'bar=baz' });
		{"foo=bar=baz", nil, map[string]any{"foo": "bar=baz"}},

		// st.deepEqual(qs.parse('foo=bar&bar=baz'), { foo: 'bar', bar: 'baz' });
		{"foo=bar&bar=baz", nil, map[string]any{"foo": "bar", "bar": "baz"}},

		// st.deepEqual(qs.parse('foo2=bar2&baz2='), { foo2: 'bar2', baz2: '' });
		{"foo2=bar2&baz2=", nil, map[string]any{"foo2": "bar2", "baz2": ""}},

		// st.deepEqual(qs.parse('foo=bar&baz', { strictNullHandling: true }), { foo: 'bar', baz: null });
		{"foo=bar&baz", []ParseOption{WithStrictNullHandling(true)}, map[string]any{"foo": "bar", "baz": nil}},

		// st.deepEqual(qs.parse('foo=bar&baz'), { foo: 'bar', baz: '' });
		{"foo=bar&baz", nil, map[string]any{"foo": "bar", "baz": ""}},

		// st.deepEqual(qs.parse('cht=p3&chd=t:60,40&chs=250x100&chl=Hello|World'), {...}
		{"cht=p3&chd=t:60,40&chs=250x100&chl=Hello|World", nil, map[string]any{
			"cht": "p3",
			"chd": "t:60,40",
			"chs": "250x100",
			"chl": "Hello|World",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := Parse(tt.input, tt.opts...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertEqual(t, result, tt.expected, tt.input)
		})
	}
}

// ===========================================
// JS Test: "comma: false"
// ===========================================
func TestJSCommaFalse(t *testing.T) {
	// st.deepEqual(qs.parse('a[]=b&a[]=c'), { a: ['b', 'c'] });
	result, _ := Parse("a[]=b&a[]=c")
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "a[]=b&a[]=c")

	// st.deepEqual(qs.parse('a[0]=b&a[1]=c'), { a: ['b', 'c'] });
	result, _ = Parse("a[0]=b&a[1]=c")
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "a[0]=b&a[1]=c")

	// st.deepEqual(qs.parse('a=b,c'), { a: 'b,c' });
	result, _ = Parse("a=b,c")
	assertEqual(t, result, map[string]any{"a": "b,c"}, "a=b,c")

	// st.deepEqual(qs.parse('a=b&a=c'), { a: ['b', 'c'] });
	result, _ = Parse("a=b&a=c")
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "a=b&a=c")
}

// ===========================================
// JS Test: "comma: true"
// ===========================================
func TestJSCommaTrue(t *testing.T) {
	// st.deepEqual(qs.parse('a[]=b&a[]=c', { comma: true }), { a: ['b', 'c'] });
	result, _ := Parse("a[]=b&a[]=c", WithComma(true))
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "a[]=b&a[]=c with comma")

	// st.deepEqual(qs.parse('a[0]=b&a[1]=c', { comma: true }), { a: ['b', 'c'] });
	result, _ = Parse("a[0]=b&a[1]=c", WithComma(true))
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "a[0]=b&a[1]=c with comma")

	// st.deepEqual(qs.parse('a=b,c', { comma: true }), { a: ['b', 'c'] });
	result, _ = Parse("a=b,c", WithComma(true))
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "a=b,c with comma")

	// st.deepEqual(qs.parse('a=b&a=c', { comma: true }), { a: ['b', 'c'] });
	result, _ = Parse("a=b&a=c", WithComma(true))
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "a=b&a=c with comma")
}

// ===========================================
// JS Test: "allows enabling dot notation"
// ===========================================
func TestJSAllowsDotNotation(t *testing.T) {
	// st.deepEqual(qs.parse('a.b=c'), { 'a.b': 'c' });
	result, _ := Parse("a.b=c")
	assertEqual(t, result, map[string]any{"a.b": "c"}, "a.b=c without allowDots")

	// st.deepEqual(qs.parse('a.b=c', { allowDots: true }), { a: { b: 'c' } });
	result, _ = Parse("a.b=c", WithAllowDots(true))
	assertEqual(t, result, map[string]any{"a": map[string]any{"b": "c"}}, "a.b=c with allowDots")
}

// ===========================================
// JS Test: "decode dot keys correctly"
// ===========================================
func TestJSDecodeDotKeys(t *testing.T) {
	// with allowDots false and decodeDotInKeys false
	result, _ := Parse("name%252Eobj.first=John&name%252Eobj.last=Doe")
	assertEqual(t, result, map[string]any{
		"name%2Eobj.first": "John",
		"name%2Eobj.last":  "Doe",
	}, "allowDots false, decodeDotInKeys false")

	// with allowDots true and decodeDotInKeys false
	result, _ = Parse("name.obj.first=John&name.obj.last=Doe", WithAllowDots(true))
	assertEqual(t, result, map[string]any{
		"name": map[string]any{
			"obj": map[string]any{
				"first": "John",
				"last":  "Doe",
			},
		},
	}, "allowDots true, decodeDotInKeys false")

	// with allowDots true and decodeDotInKeys false (encoded dot)
	result, _ = Parse("name%252Eobj.first=John&name%252Eobj.last=Doe", WithAllowDots(true))
	assertEqual(t, result, map[string]any{
		"name%2Eobj": map[string]any{
			"first": "John",
			"last":  "Doe",
		},
	}, "allowDots true, encoded dot, decodeDotInKeys false")

	// with allowDots true and decodeDotInKeys true
	result, _ = Parse("name%252Eobj.first=John&name%252Eobj.last=Doe", WithAllowDots(true), WithDecodeDotInKeys(true))
	assertEqual(t, result, map[string]any{
		"name.obj": map[string]any{
			"first": "John",
			"last":  "Doe",
		},
	}, "allowDots true, decodeDotInKeys true")
}

// ===========================================
// JS Test: "allows empty arrays in obj values"
// ===========================================
func TestJSAllowEmptyArrays(t *testing.T) {
	// st.deepEqual(qs.parse('foo[]&bar=baz', { allowEmptyArrays: true }), { foo: [], bar: 'baz' });
	result, _ := Parse("foo[]&bar=baz", WithAllowEmptyArrays(true))
	assertEqual(t, result, map[string]any{"foo": []any{}, "bar": "baz"}, "allowEmptyArrays true")

	// st.deepEqual(qs.parse('foo[]&bar=baz', { allowEmptyArrays: false }), { foo: [''], bar: 'baz' });
	result, _ = Parse("foo[]&bar=baz", WithAllowEmptyArrays(false))
	assertEqual(t, result, map[string]any{"foo": []any{""}, "bar": "baz"}, "allowEmptyArrays false")
}

// ===========================================
// JS Test: "allowEmptyArrays + strictNullHandling"
// ===========================================
func TestJSAllowEmptyArraysStrictNull(t *testing.T) {
	result, _ := Parse("testEmptyArray[]", WithStrictNullHandling(true), WithAllowEmptyArrays(true))
	assertEqual(t, result, map[string]any{"testEmptyArray": []any{}}, "allowEmptyArrays + strictNullHandling")
}

// ===========================================
// JS Test: nested strings parsing
// ===========================================
func TestJSNestedStrings(t *testing.T) {
	// t.deepEqual(qs.parse('a[b]=c'), { a: { b: 'c' } }, 'parses a single nested string');
	result, _ := Parse("a[b]=c")
	assertEqual(t, result, map[string]any{"a": map[string]any{"b": "c"}}, "single nested")

	// t.deepEqual(qs.parse('a[b][c]=d'), { a: { b: { c: 'd' } } }, 'parses a double nested string');
	result, _ = Parse("a[b][c]=d")
	assertEqual(t, result, map[string]any{"a": map[string]any{"b": map[string]any{"c": "d"}}}, "double nested")

	// t.deepEqual(qs.parse('a[b][c][d][e][f][g][h]=i'), { a: { b: { c: { d: { e: { f: { '[g][h]': 'i' } } } } } } }, 'defaults to a depth of 5');
	result, _ = Parse("a[b][c][d][e][f][g][h]=i")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": map[string]any{
							"f": map[string]any{
								"[g][h]": "i",
							},
						},
					},
				},
			},
		},
	}, "defaults to depth 5")
}

// ===========================================
// JS Test: "only parses one level when depth = 1"
// ===========================================
func TestJSDepthOne(t *testing.T) {
	// st.deepEqual(qs.parse('a[b][c]=d', { depth: 1 }), { a: { b: { '[c]': 'd' } } });
	result, _ := Parse("a[b][c]=d", WithDepth(1))
	assertEqual(t, result, map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"[c]": "d",
			},
		},
	}, "depth 1 - a[b][c]=d")

	// st.deepEqual(qs.parse('a[b][c][d]=e', { depth: 1 }), { a: { b: { '[c][d]': 'e' } } });
	result, _ = Parse("a[b][c][d]=e", WithDepth(1))
	assertEqual(t, result, map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"[c][d]": "e",
			},
		},
	}, "depth 1 - a[b][c][d]=e")
}

// ===========================================
// JS Test: "uses original key when depth = 0"
// ===========================================
func TestJSDepthZero(t *testing.T) {
	// st.deepEqual(qs.parse('a[0]=b&a[1]=c', { depth: 0 }), { 'a[0]': 'b', 'a[1]': 'c' });
	result, _ := Parse("a[0]=b&a[1]=c", WithDepth(0))
	assertEqual(t, result, map[string]any{"a[0]": "b", "a[1]": "c"}, "depth 0 - arrays")

	// st.deepEqual(qs.parse('a[0][0]=b&a[0][1]=c&a[1]=d&e=2', { depth: 0 }), { 'a[0][0]': 'b', 'a[0][1]': 'c', 'a[1]': 'd', e: '2' });
	result, _ = Parse("a[0][0]=b&a[0][1]=c&a[1]=d&e=2", WithDepth(0))
	assertEqual(t, result, map[string]any{
		"a[0][0]": "b",
		"a[0][1]": "c",
		"a[1]":    "d",
		"e":       "2",
	}, "depth 0 - nested")
}

// ===========================================
// JS Test: "parses a simple array"
// ===========================================
func TestJSSimpleArray(t *testing.T) {
	// t.deepEqual(qs.parse('a=b&a=c'), { a: ['b', 'c'] }, 'parses a simple array');
	result, _ := Parse("a=b&a=c")
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "simple array")
}

// ===========================================
// JS Test: "parses an explicit array"
// ===========================================
func TestJSExplicitArray(t *testing.T) {
	// st.deepEqual(qs.parse('a[]=b'), { a: ['b'] });
	result, _ := Parse("a[]=b")
	assertEqual(t, result, map[string]any{"a": []any{"b"}}, "a[]=b")

	// st.deepEqual(qs.parse('a[]=b&a[]=c'), { a: ['b', 'c'] });
	result, _ = Parse("a[]=b&a[]=c")
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "a[]=b&a[]=c")

	// st.deepEqual(qs.parse('a[]=b&a[]=c&a[]=d'), { a: ['b', 'c', 'd'] });
	result, _ = Parse("a[]=b&a[]=c&a[]=d")
	assertEqual(t, result, map[string]any{"a": []any{"b", "c", "d"}}, "a[]=b&a[]=c&a[]=d")
}

// ===========================================
// JS Test: "parses a mix of simple and explicit arrays"
// ===========================================
func TestJSMixedArrays(t *testing.T) {
	tests := []struct {
		input    string
		opts     []ParseOption
		expected map[string]any
	}{
		{"a=b&a[]=c", nil, map[string]any{"a": []any{"b", "c"}}},
		{"a[]=b&a=c", nil, map[string]any{"a": []any{"b", "c"}}},
		{"a[0]=b&a=c", nil, map[string]any{"a": []any{"b", "c"}}},
		{"a=b&a[0]=c", nil, map[string]any{"a": []any{"b", "c"}}},
		{"a[1]=b&a=c", []ParseOption{WithArrayLimit(20)}, map[string]any{"a": []any{"b", "c"}}},
		{"a[]=b&a=c", []ParseOption{WithArrayLimit(0)}, map[string]any{"a": []any{"b", "c"}}},
		{"a[]=b&a=c", nil, map[string]any{"a": []any{"b", "c"}}},
		{"a=b&a[1]=c", []ParseOption{WithArrayLimit(20)}, map[string]any{"a": []any{"b", "c"}}},
		{"a=b&a[]=c", []ParseOption{WithArrayLimit(0)}, map[string]any{"a": []any{"b", "c"}}},
		{"a=b&a[]=c", nil, map[string]any{"a": []any{"b", "c"}}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := Parse(tt.input, tt.opts...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertEqual(t, result, tt.expected, tt.input)
		})
	}
}

// ===========================================
// JS Test: "parses a nested array"
// ===========================================
func TestJSNestedArray(t *testing.T) {
	// st.deepEqual(qs.parse('a[b][]=c&a[b][]=d'), { a: { b: ['c', 'd'] } });
	result, _ := Parse("a[b][]=c&a[b][]=d")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{
			"b": []any{"c", "d"},
		},
	}, "nested array")

	// st.deepEqual(qs.parse('a[>=]=25'), { a: { '>=': '25' } });
	result, _ = Parse("a[>=]=25")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{">=": "25"},
	}, "special chars in key")
}

// ===========================================
// JS Test: "allows to specify array indices"
// ===========================================
func TestJSArrayIndices(t *testing.T) {
	// st.deepEqual(qs.parse('a[1]=c&a[0]=b&a[2]=d'), { a: ['b', 'c', 'd'] });
	result, _ := Parse("a[1]=c&a[0]=b&a[2]=d")
	assertEqual(t, result, map[string]any{"a": []any{"b", "c", "d"}}, "reorder indices")

	// st.deepEqual(qs.parse('a[1]=c&a[0]=b'), { a: ['b', 'c'] });
	result, _ = Parse("a[1]=c&a[0]=b")
	assertEqual(t, result, map[string]any{"a": []any{"b", "c"}}, "two indices")

	// st.deepEqual(qs.parse('a[1]=c', { arrayLimit: 20 }), { a: ['c'] });
	result, _ = Parse("a[1]=c", WithArrayLimit(20))
	assertEqual(t, result, map[string]any{"a": []any{"c"}}, "sparse compacted")

	// st.deepEqual(qs.parse('a[1]=c', { arrayLimit: 0 }), { a: { 1: 'c' } });
	result, _ = Parse("a[1]=c", WithArrayLimit(0))
	assertEqual(t, result, map[string]any{"a": map[string]any{"1": "c"}}, "arrayLimit 0 = object")

	// st.deepEqual(qs.parse('a[1]=c'), { a: ['c'] });
	result, _ = Parse("a[1]=c")
	assertEqual(t, result, map[string]any{"a": []any{"c"}}, "default compacts")
}

// ===========================================
// JS Test: "limits specific array indices to arrayLimit"
// ===========================================
func TestJSArrayLimitIndices(t *testing.T) {
	// st.deepEqual(qs.parse('a[20]=a', { arrayLimit: 20 }), { a: ['a'] });
	result, _ := Parse("a[20]=a", WithArrayLimit(20))
	assertEqual(t, result, map[string]any{"a": []any{"a"}}, "at limit = array")

	// st.deepEqual(qs.parse('a[21]=a', { arrayLimit: 20 }), { a: { 21: 'a' } });
	result, _ = Parse("a[21]=a", WithArrayLimit(20))
	assertEqual(t, result, map[string]any{"a": map[string]any{"21": "a"}}, "over limit = object")

	// st.deepEqual(qs.parse('a[20]=a'), { a: ['a'] });
	result, _ = Parse("a[20]=a")
	assertEqual(t, result, map[string]any{"a": []any{"a"}}, "default 20 at limit")

	// st.deepEqual(qs.parse('a[21]=a'), { a: { 21: 'a' } });
	result, _ = Parse("a[21]=a")
	assertEqual(t, result, map[string]any{"a": map[string]any{"21": "a"}}, "default 20 over limit")
}

// ===========================================
// JS Test: "supports keys that begin with a number"
// ===========================================
func TestJSKeysBeginWithNumber(t *testing.T) {
	// t.deepEqual(qs.parse('a[12b]=c'), { a: { '12b': 'c' } }, 'supports keys that begin with a number');
	result, _ := Parse("a[12b]=c")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"12b": "c"},
	}, "key begins with number")
}

// ===========================================
// JS Test: "supports encoded = signs"
// ===========================================
func TestJSEncodedEquals(t *testing.T) {
	// st.deepEqual(qs.parse('he%3Dllo=th%3Dere'), { 'he=llo': 'th=ere' });
	result, _ := Parse("he%3Dllo=th%3Dere")
	assertEqual(t, result, map[string]any{"he=llo": "th=ere"}, "encoded equals")
}

// ===========================================
// JS Test: "is ok with url encoded strings"
// ===========================================
func TestJSUrlEncodedStrings(t *testing.T) {
	// st.deepEqual(qs.parse('a[b%20c]=d'), { a: { 'b c': 'd' } });
	result, _ := Parse("a[b%20c]=d")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"b c": "d"},
	}, "encoded space in key")

	// st.deepEqual(qs.parse('a[b]=c%20d'), { a: { b: 'c d' } });
	result, _ = Parse("a[b]=c%20d")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"b": "c d"},
	}, "encoded space in value")
}

// ===========================================
// JS Test: "allows brackets in the value"
// ===========================================
func TestJSBracketsInValue(t *testing.T) {
	// st.deepEqual(qs.parse('pets=["tobi"]'), { pets: '["tobi"]' });
	result, _ := Parse(`pets=["tobi"]`)
	assertEqual(t, result, map[string]any{"pets": `["tobi"]`}, "brackets in value")

	// st.deepEqual(qs.parse('operators=[">=", "<="]'), { operators: '[">=", "<="]' });
	result, _ = Parse(`operators=[">=", "<="]`)
	assertEqual(t, result, map[string]any{"operators": `[">=", "<="]`}, "operators in value")
}

// ===========================================
// JS Test: "allows empty values"
// ===========================================
func TestJSEmptyValues(t *testing.T) {
	// st.deepEqual(qs.parse(''), {});
	result, _ := Parse("")
	assertEqual(t, result, map[string]any{}, "empty string")

	// Note: JS also tests null and undefined which don't exist in Go
}

// ===========================================
// JS Test: "transforms arrays to objects"
// ===========================================
func TestJSArraysToObjects(t *testing.T) {
	// st.deepEqual(qs.parse('foo[0]=bar&foo[bad]=baz'), { foo: { 0: 'bar', bad: 'baz' } });
	result, _ := Parse("foo[0]=bar&foo[bad]=baz")
	assertEqual(t, result, map[string]any{
		"foo": map[string]any{"0": "bar", "bad": "baz"},
	}, "mixed numeric and string keys")

	// st.deepEqual(qs.parse('foo[bad]=baz&foo[0]=bar'), { foo: { bad: 'baz', 0: 'bar' } });
	result, _ = Parse("foo[bad]=baz&foo[0]=bar")
	assertEqual(t, result, map[string]any{
		"foo": map[string]any{"bad": "baz", "0": "bar"},
	}, "string then numeric")

	// st.deepEqual(qs.parse('foo[bad]=baz&foo[]=bar'), { foo: { bad: 'baz', 0: 'bar' } });
	result, _ = Parse("foo[bad]=baz&foo[]=bar")
	assertEqual(t, result, map[string]any{
		"foo": map[string]any{"bad": "baz", "0": "bar"},
	}, "string then empty bracket")

	// st.deepEqual(qs.parse('foo[]=bar&foo[bad]=baz'), { foo: { 0: 'bar', bad: 'baz' } });
	result, _ = Parse("foo[]=bar&foo[bad]=baz")
	assertEqual(t, result, map[string]any{
		"foo": map[string]any{"0": "bar", "bad": "baz"},
	}, "empty bracket then string")

	// st.deepEqual(qs.parse('foo[bad]=baz&foo[]=bar&foo[]=foo'), { foo: { bad: 'baz', 0: 'bar', 1: 'foo' } });
	result, _ = Parse("foo[bad]=baz&foo[]=bar&foo[]=foo")
	assertEqual(t, result, map[string]any{
		"foo": map[string]any{"bad": "baz", "0": "bar", "1": "foo"},
	}, "string then multiple empty brackets")

	// st.deepEqual(qs.parse('foo[0][a]=a&foo[0][b]=b&foo[1][a]=aa&foo[1][b]=bb'), { foo: [{ a: 'a', b: 'b' }, { a: 'aa', b: 'bb' }] });
	result, _ = Parse("foo[0][a]=a&foo[0][b]=b&foo[1][a]=aa&foo[1][b]=bb")
	assertEqual(t, result, map[string]any{
		"foo": []any{
			map[string]any{"a": "a", "b": "b"},
			map[string]any{"a": "aa", "b": "bb"},
		},
	}, "array of objects")
}

// ===========================================
// JS Test: "transforms arrays to objects (dot notation)"
// ===========================================
func TestJSArraysToObjectsDotNotation(t *testing.T) {
	// st.deepEqual(qs.parse('foo[0].baz=bar&fool.bad=baz', { allowDots: true }), { foo: [{ baz: 'bar' }], fool: { bad: 'baz' } });
	result, _ := Parse("foo[0].baz=bar&fool.bad=baz", WithAllowDots(true))
	assertEqual(t, result, map[string]any{
		"foo":  []any{map[string]any{"baz": "bar"}},
		"fool": map[string]any{"bad": "baz"},
	}, "dot notation with array")

	// st.deepEqual(qs.parse('foo[0].baz=bar&fool.bad.boo=baz', { allowDots: true }), { foo: [{ baz: 'bar' }], fool: { bad: { boo: 'baz' } } });
	result, _ = Parse("foo[0].baz=bar&fool.bad.boo=baz", WithAllowDots(true))
	assertEqual(t, result, map[string]any{
		"foo":  []any{map[string]any{"baz": "bar"}},
		"fool": map[string]any{"bad": map[string]any{"boo": "baz"}},
	}, "dot notation nested")

	// st.deepEqual(qs.parse('foo[0][0].baz=bar&fool.bad=baz', { allowDots: true }), { foo: [[{ baz: 'bar' }]], fool: { bad: 'baz' } });
	result, _ = Parse("foo[0][0].baz=bar&fool.bad=baz", WithAllowDots(true))
	assertEqual(t, result, map[string]any{
		"foo":  []any{[]any{map[string]any{"baz": "bar"}}},
		"fool": map[string]any{"bad": "baz"},
	}, "nested array with dot")

	// st.deepEqual(qs.parse('foo.bad=baz&foo[0]=bar', { allowDots: true }), { foo: { bad: 'baz', 0: 'bar' } });
	result, _ = Parse("foo.bad=baz&foo[0]=bar", WithAllowDots(true))
	assertEqual(t, result, map[string]any{
		"foo": map[string]any{"bad": "baz", "0": "bar"},
	}, "dot then bracket")
}

// ===========================================
// JS Test: "correctly prunes undefined values when converting an array to an object"
// ===========================================
func TestJSPruneUndefined(t *testing.T) {
	// st.deepEqual(qs.parse('a[2]=b&a[99999999]=c'), { a: { 2: 'b', 99999999: 'c' } });
	result, _ := Parse("a[2]=b&a[99999999]=c")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"2": "b", "99999999": "c"},
	}, "prune undefined in sparse")
}

// ===========================================
// JS Test: "supports malformed uri characters"
// ===========================================
func TestJSMalformedUri(t *testing.T) {
	// st.deepEqual(qs.parse('{%:%}', { strictNullHandling: true }), { '{%:%}': null });
	result, _ := Parse("{%:%}", WithStrictNullHandling(true))
	assertEqual(t, result, map[string]any{"{%:%}": nil}, "malformed as key null")

	// st.deepEqual(qs.parse('{%:%}='), { '{%:%}': '' });
	result, _ = Parse("{%:%}=")
	assertEqual(t, result, map[string]any{"{%:%}": ""}, "malformed as key empty")

	// st.deepEqual(qs.parse('foo=%:%}'), { foo: '%:%}' });
	result, _ = Parse("foo=%:%}")
	assertEqual(t, result, map[string]any{"foo": "%:%}"}, "malformed in value")
}

// ===========================================
// JS Test: "doesn't produce empty keys"
// ===========================================
func TestJSNoEmptyKeys(t *testing.T) {
	// st.deepEqual(qs.parse('_r=1&'), { _r: '1' });
	result, _ := Parse("_r=1&")
	assertEqual(t, result, map[string]any{"_r": "1"}, "trailing &")
}

// ===========================================
// JS Test: "parses arrays of objects"
// ===========================================
func TestJSArraysOfObjects(t *testing.T) {
	// st.deepEqual(qs.parse('a[][b]=c'), { a: [{ b: 'c' }] });
	result, _ := Parse("a[][b]=c")
	assertEqual(t, result, map[string]any{
		"a": []any{map[string]any{"b": "c"}},
	}, "a[][b]=c")

	// st.deepEqual(qs.parse('a[0][b]=c'), { a: [{ b: 'c' }] });
	result, _ = Parse("a[0][b]=c")
	assertEqual(t, result, map[string]any{
		"a": []any{map[string]any{"b": "c"}},
	}, "a[0][b]=c")
}

// ===========================================
// JS Test: "allows for empty strings in arrays"
// ===========================================
func TestJSEmptyStringsInArrays(t *testing.T) {
	// st.deepEqual(qs.parse('a[]=b&a[]=&a[]=c'), { a: ['b', '', 'c'] });
	result, _ := Parse("a[]=b&a[]=&a[]=c")
	assertEqual(t, result, map[string]any{"a": []any{"b", "", "c"}}, "empty string in array")

	// st.deepEqual(qs.parse('a[0]=b&a[1]&a[2]=c&a[19]=', { strictNullHandling: true, arrayLimit: 20 }), { a: ['b', null, 'c', ''] });
	result, _ = Parse("a[0]=b&a[1]&a[2]=c&a[19]=", WithStrictNullHandling(true), WithArrayLimit(20))
	assertEqual(t, result, map[string]any{"a": []any{"b", nil, "c", ""}}, "null then empty")

	// st.deepEqual(qs.parse('a[]=b&a[]&a[]=c&a[]=', { strictNullHandling: true, arrayLimit: 0 }), { a: ['b', null, 'c', ''] });
	result, _ = Parse("a[]=b&a[]&a[]=c&a[]=", WithStrictNullHandling(true), WithArrayLimit(0))
	assertEqual(t, result, map[string]any{"a": []any{"b", nil, "c", ""}}, "brackets null then empty")

	// st.deepEqual(qs.parse('a[0]=b&a[1]=&a[2]=c&a[19]', { strictNullHandling: true, arrayLimit: 20 }), { a: ['b', '', 'c', null] });
	result, _ = Parse("a[0]=b&a[1]=&a[2]=c&a[19]", WithStrictNullHandling(true), WithArrayLimit(20))
	assertEqual(t, result, map[string]any{"a": []any{"b", "", "c", nil}}, "empty then null")

	// st.deepEqual(qs.parse('a[]=&a[]=b&a[]=c'), { a: ['', 'b', 'c'] });
	result, _ = Parse("a[]=&a[]=b&a[]=c")
	assertEqual(t, result, map[string]any{"a": []any{"", "b", "c"}}, "leading empty")
}

// ===========================================
// JS Test: "compacts sparse arrays"
// ===========================================
func TestJSCompactsSparseArrays(t *testing.T) {
	// st.deepEqual(qs.parse('a[10]=1&a[2]=2', { arrayLimit: 20 }), { a: ['2', '1'] });
	result, _ := Parse("a[10]=1&a[2]=2", WithArrayLimit(20))
	assertEqual(t, result, map[string]any{"a": []any{"2", "1"}}, "compacted sparse")

	// st.deepEqual(qs.parse('a[1][b][2][c]=1', { arrayLimit: 20 }), { a: [{ b: [{ c: '1' }] }] });
	result, _ = Parse("a[1][b][2][c]=1", WithArrayLimit(20))
	assertEqual(t, result, map[string]any{
		"a": []any{map[string]any{"b": []any{map[string]any{"c": "1"}}}},
	}, "nested compacted")

	// st.deepEqual(qs.parse('a[1][2][3][c]=1', { arrayLimit: 20 }), { a: [[[{ c: '1' }]]] });
	result, _ = Parse("a[1][2][3][c]=1", WithArrayLimit(20))
	assertEqual(t, result, map[string]any{
		"a": []any{[]any{[]any{map[string]any{"c": "1"}}}},
	}, "deeply nested compacted")

	// st.deepEqual(qs.parse('a[1][2][3][c][1]=1', { arrayLimit: 20 }), { a: [[[{ c: ['1'] }]]] });
	result, _ = Parse("a[1][2][3][c][1]=1", WithArrayLimit(20))
	assertEqual(t, result, map[string]any{
		"a": []any{[]any{[]any{map[string]any{"c": []any{"1"}}}}},
	}, "deeply nested with array value")
}

// ===========================================
// JS Test: "parses sparse arrays"
// ===========================================
func TestJSParsesSparseArrays(t *testing.T) {
	// st.deepEqual(qs.parse('a[4]=1&a[1]=2', { allowSparse: true }), { a: [, '2', , , '1'] });
	result, _ := Parse("a[4]=1&a[1]=2", WithAllowSparse(true))
	expected := []any{nil, "2", nil, nil, "1"}
	assertEqual(t, result, map[string]any{"a": expected}, "sparse array")

	// st.deepEqual(qs.parse('a[1][b][2][c]=1', { allowSparse: true }), { a: [, { b: [, , { c: '1' }] }] });
	result, _ = Parse("a[1][b][2][c]=1", WithAllowSparse(true))
	assertEqual(t, result, map[string]any{
		"a": []any{nil, map[string]any{"b": []any{nil, nil, map[string]any{"c": "1"}}}},
	}, "nested sparse")
}

// ===========================================
// JS Test: "parses a string with an alternative string delimiter"
// ===========================================
func TestJSAlternativeDelimiter(t *testing.T) {
	// st.deepEqual(qs.parse('a=b;c=d', { delimiter: ';' }), { a: 'b', c: 'd' });
	result, _ := Parse("a=b;c=d", WithDelimiter(";"))
	assertEqual(t, result, map[string]any{"a": "b", "c": "d"}, "semicolon delimiter")
}

// ===========================================
// JS Test: "parses a string with an alternative RegExp delimiter"
// ===========================================
func TestJSRegExpDelimiter(t *testing.T) {
	// st.deepEqual(qs.parse('a=b; c=d', { delimiter: /[;,] */ }), { a: 'b', c: 'd' });
	re := regexp.MustCompile(`[;,] *`)
	result, _ := Parse("a=b; c=d", WithDelimiterRegexp(re))
	assertEqual(t, result, map[string]any{"a": "b", "c": "d"}, "regexp delimiter")
}

// ===========================================
// JS Test: "allows overriding parameter limit"
// ===========================================
func TestJSParameterLimit(t *testing.T) {
	// st.deepEqual(qs.parse('a=b&c=d', { parameterLimit: 1 }), { a: 'b' });
	result, _ := Parse("a=b&c=d", WithParameterLimit(1))
	assertEqual(t, result, map[string]any{"a": "b"}, "parameter limit 1")
}

// ===========================================
// JS Test: "allows overriding array limit"
// ===========================================
func TestJSArrayLimitOverride(t *testing.T) {
	// st.deepEqual(qs.parse('a[0]=b', { arrayLimit: -1 }), { a: { 0: 'b' } });
	result, _ := Parse("a[0]=b", WithArrayLimit(-1))
	assertEqual(t, result, map[string]any{"a": map[string]any{"0": "b"}}, "arrayLimit -1")

	// st.deepEqual(qs.parse('a[0]=b', { arrayLimit: 0 }), { a: ['b'] });
	result, _ = Parse("a[0]=b", WithArrayLimit(0))
	assertEqual(t, result, map[string]any{"a": []any{"b"}}, "arrayLimit 0")

	// st.deepEqual(qs.parse('a[-1]=b', { arrayLimit: -1 }), { a: { '-1': 'b' } });
	result, _ = Parse("a[-1]=b", WithArrayLimit(-1))
	assertEqual(t, result, map[string]any{"a": map[string]any{"-1": "b"}}, "negative index, limit -1")

	// st.deepEqual(qs.parse('a[-1]=b', { arrayLimit: 0 }), { a: { '-1': 'b' } });
	result, _ = Parse("a[-1]=b", WithArrayLimit(0))
	assertEqual(t, result, map[string]any{"a": map[string]any{"-1": "b"}}, "negative index, limit 0")

	// st.deepEqual(qs.parse('a[0]=b&a[1]=c', { arrayLimit: -1 }), { a: { 0: 'b', 1: 'c' } });
	result, _ = Parse("a[0]=b&a[1]=c", WithArrayLimit(-1))
	assertEqual(t, result, map[string]any{"a": map[string]any{"0": "b", "1": "c"}}, "multiple with limit -1")

	// st.deepEqual(qs.parse('a[0]=b&a[1]=c', { arrayLimit: 0 }), { a: { 0: 'b', 1: 'c' } });
	result, _ = Parse("a[0]=b&a[1]=c", WithArrayLimit(0))
	assertEqual(t, result, map[string]any{"a": map[string]any{"0": "b", "1": "c"}}, "multiple with limit 0")
}

// ===========================================
// JS Test: "allows disabling array parsing"
// ===========================================
func TestJSDisableArrayParsing(t *testing.T) {
	// var indices = qs.parse('a[0]=b&a[1]=c', { parseArrays: false });
	// st.deepEqual(indices, { a: { 0: 'b', 1: 'c' } });
	result, _ := Parse("a[0]=b&a[1]=c", WithParseArrays(false))
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"0": "b", "1": "c"},
	}, "parseArrays false indices")

	// var emptyBrackets = qs.parse('a[]=b', { parseArrays: false });
	// st.deepEqual(emptyBrackets, { a: { 0: 'b' } });
	result, _ = Parse("a[]=b", WithParseArrays(false))
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"0": "b"},
	}, "parseArrays false brackets")
}

// ===========================================
// JS Test: "allows for query string prefix"
// ===========================================
func TestJSQueryPrefix(t *testing.T) {
	// st.deepEqual(qs.parse('?foo=bar', { ignoreQueryPrefix: true }), { foo: 'bar' });
	result, _ := Parse("?foo=bar", WithIgnoreQueryPrefix(true))
	assertEqual(t, result, map[string]any{"foo": "bar"}, "strip prefix")

	// st.deepEqual(qs.parse('foo=bar', { ignoreQueryPrefix: true }), { foo: 'bar' });
	result, _ = Parse("foo=bar", WithIgnoreQueryPrefix(true))
	assertEqual(t, result, map[string]any{"foo": "bar"}, "no prefix")

	// st.deepEqual(qs.parse('?foo=bar', { ignoreQueryPrefix: false }), { '?foo': 'bar' });
	result, _ = Parse("?foo=bar", WithIgnoreQueryPrefix(false))
	assertEqual(t, result, map[string]any{"?foo": "bar"}, "keep prefix")
}

// ===========================================
// JS Test: "parses string with comma as array divider"
// ===========================================
func TestJSCommaAsArrayDivider(t *testing.T) {
	// st.deepEqual(qs.parse('foo=bar,tee', { comma: true }), { foo: ['bar', 'tee'] });
	result, _ := Parse("foo=bar,tee", WithComma(true))
	assertEqual(t, result, map[string]any{"foo": []any{"bar", "tee"}}, "simple comma")

	// st.deepEqual(qs.parse('foo[bar]=coffee,tee', { comma: true }), { foo: { bar: ['coffee', 'tee'] } });
	result, _ = Parse("foo[bar]=coffee,tee", WithComma(true))
	assertEqual(t, result, map[string]any{
		"foo": map[string]any{"bar": []any{"coffee", "tee"}},
	}, "nested comma")

	// st.deepEqual(qs.parse('foo=', { comma: true }), { foo: '' });
	result, _ = Parse("foo=", WithComma(true))
	assertEqual(t, result, map[string]any{"foo": ""}, "empty with comma")

	// st.deepEqual(qs.parse('foo', { comma: true }), { foo: '' });
	result, _ = Parse("foo", WithComma(true))
	assertEqual(t, result, map[string]any{"foo": ""}, "no value with comma")

	// st.deepEqual(qs.parse('foo', { comma: true, strictNullHandling: true }), { foo: null });
	result, _ = Parse("foo", WithComma(true), WithStrictNullHandling(true))
	assertEqual(t, result, map[string]any{"foo": nil}, "null with comma")
}

// ===========================================
// JS Test: "parses brackets holds array of arrays when having two parts of strings with comma"
// ===========================================
func TestJSBracketsArrayOfArrays(t *testing.T) {
	// st.deepEqual(qs.parse('foo[]=1,2,3&foo[]=4,5,6', { comma: true }), { foo: [['1', '2', '3'], ['4', '5', '6']] });
	result, _ := Parse("foo[]=1,2,3&foo[]=4,5,6", WithComma(true))
	assertEqual(t, result, map[string]any{
		"foo": []any{[]any{"1", "2", "3"}, []any{"4", "5", "6"}},
	}, "array of arrays")

	// st.deepEqual(qs.parse('foo[]=1,2,3&foo[]=', { comma: true }), { foo: [['1', '2', '3'], ''] });
	result, _ = Parse("foo[]=1,2,3&foo[]=", WithComma(true))
	assertEqual(t, result, map[string]any{
		"foo": []any{[]any{"1", "2", "3"}, ""},
	}, "array then empty")

	// st.deepEqual(qs.parse('foo[]=1,2,3&foo[]=,', { comma: true }), { foo: [['1', '2', '3'], ['', '']] });
	result, _ = Parse("foo[]=1,2,3&foo[]=,", WithComma(true))
	assertEqual(t, result, map[string]any{
		"foo": []any{[]any{"1", "2", "3"}, []any{"", ""}},
	}, "array then comma only")

	// st.deepEqual(qs.parse('foo[]=1,2,3&foo[]=a', { comma: true }), { foo: [['1', '2', '3'], 'a'] });
	result, _ = Parse("foo[]=1,2,3&foo[]=a", WithComma(true))
	assertEqual(t, result, map[string]any{
		"foo": []any{[]any{"1", "2", "3"}, "a"},
	}, "array then single")
}

// ===========================================
// JS Test: "parses comma delimited array while having percent-encoded comma treated as normal text"
// ===========================================
func TestJSPercentEncodedComma(t *testing.T) {
	// st.deepEqual(qs.parse('foo=a%2Cb', { comma: true }), { foo: 'a,b' });
	result, _ := Parse("foo=a%2Cb", WithComma(true))
	assertEqual(t, result, map[string]any{"foo": "a,b"}, "encoded comma not split")

	// st.deepEqual(qs.parse('foo=a%2C%20b,d', { comma: true }), { foo: ['a, b', 'd'] });
	result, _ = Parse("foo=a%2C%20b,d", WithComma(true))
	assertEqual(t, result, map[string]any{"foo": []any{"a, b", "d"}}, "mixed encoded and literal")

	// st.deepEqual(qs.parse('foo=a%2C%20b,c%2C%20d', { comma: true }), { foo: ['a, b', 'c, d'] });
	result, _ = Parse("foo=a%2C%20b,c%2C%20d", WithComma(true))
	assertEqual(t, result, map[string]any{"foo": []any{"a, b", "c, d"}}, "both encoded")
}

// ===========================================
// JS Test: "does not allow overwriting prototype properties"
// ===========================================
func TestJSPrototypeProtection(t *testing.T) {
	// st.deepEqual(qs.parse('a[hasOwnProperty]=b', { allowPrototypes: false }), {});
	result, _ := Parse("a[hasOwnProperty]=b", WithAllowPrototypes(false))
	assertEqual(t, result, map[string]any{}, "hasOwnProperty blocked")

	// st.deepEqual(qs.parse('hasOwnProperty=b', { allowPrototypes: false }), {});
	result, _ = Parse("hasOwnProperty=b", WithAllowPrototypes(false))
	assertEqual(t, result, map[string]any{}, "top-level hasOwnProperty blocked")

	// st.deepEqual(qs.parse('toString', { allowPrototypes: false }), {});
	result, _ = Parse("toString", WithAllowPrototypes(false))
	assertEqual(t, result, map[string]any{}, "toString blocked")
}

// ===========================================
// JS Test: "can allow overwriting prototype properties"
// ===========================================
func TestJSAllowPrototypes(t *testing.T) {
	// st.deepEqual(qs.parse('a[hasOwnProperty]=b', { allowPrototypes: true }), { a: { hasOwnProperty: 'b' } });
	result, _ := Parse("a[hasOwnProperty]=b", WithAllowPrototypes(true))
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"hasOwnProperty": "b"},
	}, "hasOwnProperty allowed")

	// st.deepEqual(qs.parse('hasOwnProperty=b', { allowPrototypes: true }), { hasOwnProperty: 'b' });
	result, _ = Parse("hasOwnProperty=b", WithAllowPrototypes(true))
	assertEqual(t, result, map[string]any{"hasOwnProperty": "b"}, "top-level hasOwnProperty allowed")

	// st.deepEqual(qs.parse('toString', { allowPrototypes: true }), { toString: '' });
	result, _ = Parse("toString", WithAllowPrototypes(true))
	assertEqual(t, result, map[string]any{"toString": ""}, "toString allowed")
}

// ===========================================
// JS Test: "params starting with a closing bracket"
// ===========================================
func TestJSStartingWithClosingBracket(t *testing.T) {
	// st.deepEqual(qs.parse(']=toString'), { ']': 'toString' });
	result, _ := Parse("]=toString")
	assertEqual(t, result, map[string]any{"]": "toString"}, "] key")

	// st.deepEqual(qs.parse(']]=toString'), { ']]': 'toString' });
	result, _ = Parse("]]=toString")
	assertEqual(t, result, map[string]any{"]]": "toString"}, "]] key")

	// st.deepEqual(qs.parse(']hello]=toString'), { ']hello]': 'toString' });
	result, _ = Parse("]hello]=toString")
	assertEqual(t, result, map[string]any{"]hello]": "toString"}, "]hello] key")
}

// ===========================================
// JS Test: "params starting with a starting bracket"
// ===========================================
func TestJSStartingWithOpeningBracket(t *testing.T) {
	// st.deepEqual(qs.parse('[=toString'), { '[': 'toString' });
	result, _ := Parse("[=toString")
	assertEqual(t, result, map[string]any{"[": "toString"}, "[ key")

	// st.deepEqual(qs.parse('[[=toString'), { '[[': 'toString' });
	result, _ = Parse("[[=toString")
	assertEqual(t, result, map[string]any{"[[": "toString"}, "[[ key")

	// st.deepEqual(qs.parse('[hello[=toString'), { '[hello[': 'toString' });
	result, _ = Parse("[hello[=toString")
	assertEqual(t, result, map[string]any{"[hello[": "toString"}, "[hello[ key")
}

// ===========================================
// JS Test: "add keys to objects"
// ===========================================
func TestJSAddKeysToObjects(t *testing.T) {
	// st.deepEqual(qs.parse('a[b]=c&a=d'), { a: { b: 'c', d: true } });
	result, _ := Parse("a[b]=c&a=d")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"b": "c", "d": true},
	}, "add primitive to object")

	// st.deepEqual(qs.parse('a[b]=c&a=toString'), { a: { b: 'c' } });
	result, _ = Parse("a[b]=c&a=toString")
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"b": "c"},
	}, "toString blocked")

	// st.deepEqual(qs.parse('a[b]=c&a=toString', { allowPrototypes: true }), { a: { b: 'c', toString: true } });
	result, _ = Parse("a[b]=c&a=toString", WithAllowPrototypes(true))
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"b": "c", "toString": true},
	}, "toString allowed")
}

// ===========================================
// JS Test: "dunder proto is ignored"
// ===========================================
func TestJSDunderProto(t *testing.T) {
	// var payload = 'categories[__proto__]=login&categories[__proto__]&categories[length]=42';
	// var result = qs.parse(payload, { allowPrototypes: true });
	// st.deepEqual(result, { categories: { length: '42' } });
	result, _ := Parse("categories[__proto__]=login&categories[__proto__]&categories[length]=42", WithAllowPrototypes(true))
	assertEqual(t, result, map[string]any{
		"categories": map[string]any{"length": "42"},
	}, "__proto__ ignored")

	// var query = qs.parse('categories[__proto__]=cats&categories[__proto__]=dogs&categories[some][json]=toInject', { allowPrototypes: true });
	// st.deepEqual(query.categories, { some: { json: 'toInject' } });
	result, _ = Parse("categories[__proto__]=cats&categories[__proto__]=dogs&categories[some][json]=toInject", WithAllowPrototypes(true))
	assertEqual(t, result, map[string]any{
		"categories": map[string]any{
			"some": map[string]any{"json": "toInject"},
		},
	}, "__proto__ values ignored")

	// st.deepEqual(qs.parse('foo[__proto__][hidden]=value&foo[bar]=stuffs', { allowPrototypes: true }), { foo: { bar: 'stuffs' } });
	result, _ = Parse("foo[__proto__][hidden]=value&foo[bar]=stuffs", WithAllowPrototypes(true))
	assertEqual(t, result, map[string]any{
		"foo": map[string]any{"bar": "stuffs"},
	}, "__proto__ nested ignored")
}

// ===========================================
// JS Test: charset tests
// ===========================================
func TestJSCharset(t *testing.T) {
	// st.deepEqual(qs.parse('%A2=%BD', { charset: 'iso-8859-1' }), { '¢': '½' });
	result, _ := Parse("%A2=%BD", WithCharset(CharsetISO88591))
	assertEqual(t, result, map[string]any{"¢": "½"}, "iso-8859-1 charset")
}

// ===========================================
// JS Test: charset sentinel tests
// ===========================================
func TestJSCharsetSentinel(t *testing.T) {
	urlEncodedCheckmarkInUtf8 := "%E2%9C%93"
	urlEncodedOSlashInUtf8 := "%C3%B8"
	urlEncodedNumCheckmark := "%26%2310003%3B"

	// prefers utf-8 charset specified by sentinel to default iso-8859-1
	result, _ := Parse("utf8="+urlEncodedCheckmarkInUtf8+"&"+urlEncodedOSlashInUtf8+"="+urlEncodedOSlashInUtf8,
		WithCharsetSentinel(true), WithCharset(CharsetISO88591))
	assertEqual(t, result, map[string]any{"ø": "ø"}, "utf8 sentinel overrides iso")

	// prefers iso-8859-1 charset specified by sentinel to default utf-8
	result, _ = Parse("utf8="+urlEncodedNumCheckmark+"&"+urlEncodedOSlashInUtf8+"="+urlEncodedOSlashInUtf8,
		WithCharsetSentinel(true), WithCharset(CharsetUTF8))
	assertEqual(t, result, map[string]any{"Ã¸": "Ã¸"}, "iso sentinel overrides utf8")

	// does not require utf8 sentinel to be defined before parameters
	result, _ = Parse("a="+urlEncodedOSlashInUtf8+"&utf8="+urlEncodedNumCheckmark,
		WithCharsetSentinel(true), WithCharset(CharsetUTF8))
	assertEqual(t, result, map[string]any{"a": "Ã¸"}, "sentinel position independent")
}

// ===========================================
// JS Test: interprets numeric entities
// ===========================================
func TestJSNumericEntities(t *testing.T) {
	urlEncodedNumSmiley := "%26%239786%3B"

	// st.deepEqual(qs.parse('foo=' + urlEncodedNumSmiley, { charset: 'iso-8859-1', interpretNumericEntities: true }), { foo: '☺' });
	result, _ := Parse("foo="+urlEncodedNumSmiley,
		WithCharset(CharsetISO88591), WithInterpretNumericEntities(true))
	assertEqual(t, result, map[string]any{"foo": "☺"}, "numeric entity interpreted")

	// st.deepEqual(qs.parse('foo=' + urlEncodedNumSmiley, { charset: 'iso-8859-1' }), { foo: '&#9786;' });
	result, _ = Parse("foo="+urlEncodedNumSmiley, WithCharset(CharsetISO88591))
	assertEqual(t, result, map[string]any{"foo": "&#9786;"}, "numeric entity not interpreted")

	// st.deepEqual(qs.parse('foo=' + urlEncodedNumSmiley, { charset: 'utf-8', interpretNumericEntities: true }), { foo: '&#9786;' });
	result, _ = Parse("foo="+urlEncodedNumSmiley,
		WithCharset(CharsetUTF8), WithInterpretNumericEntities(true))
	assertEqual(t, result, map[string]any{"foo": "&#9786;"}, "numeric entity not interpreted in utf8")
}

// ===========================================
// JS Test: "continues parsing when no parent is found"
// ===========================================
func TestJSNoParentFound(t *testing.T) {
	// st.deepEqual(qs.parse('[]=&a=b'), { 0: '', a: 'b' });
	result, _ := Parse("[]=&a=b")
	assertEqual(t, result, map[string]any{"0": "", "a": "b"}, "empty bracket as index")

	// st.deepEqual(qs.parse('[]&a=b', { strictNullHandling: true }), { 0: null, a: 'b' });
	result, _ = Parse("[]&a=b", WithStrictNullHandling(true))
	assertEqual(t, result, map[string]any{"0": nil, "a": "b"}, "empty bracket null")

	// st.deepEqual(qs.parse('[foo]=bar'), { foo: 'bar' });
	result, _ = Parse("[foo]=bar")
	assertEqual(t, result, map[string]any{"foo": "bar"}, "[foo] as key")
}

// ===========================================
// JS Test: "duplicates option"
// ===========================================
func TestJSDuplicatesOption(t *testing.T) {
	// t.deepEqual(qs.parse('foo=bar&foo=baz'), { foo: ['bar', 'baz'] }, 'duplicates: default, combine');
	result, _ := Parse("foo=bar&foo=baz")
	assertEqual(t, result, map[string]any{"foo": []any{"bar", "baz"}}, "default combine")

	// t.deepEqual(qs.parse('foo=bar&foo=baz', { duplicates: 'combine' }), { foo: ['bar', 'baz'] });
	result, _ = Parse("foo=bar&foo=baz", WithDuplicates(DuplicateCombine))
	assertEqual(t, result, map[string]any{"foo": []any{"bar", "baz"}}, "explicit combine")

	// t.deepEqual(qs.parse('foo=bar&foo=baz', { duplicates: 'first' }), { foo: 'bar' });
	result, _ = Parse("foo=bar&foo=baz", WithDuplicates(DuplicateFirst))
	assertEqual(t, result, map[string]any{"foo": "bar"}, "first")

	// t.deepEqual(qs.parse('foo=bar&foo=baz', { duplicates: 'last' }), { foo: 'baz' });
	result, _ = Parse("foo=bar&foo=baz", WithDuplicates(DuplicateLast))
	assertEqual(t, result, map[string]any{"foo": "baz"}, "last")
}

// ===========================================
// JS Test: "strictDepth option - throw cases"
// ===========================================
func TestJSStrictDepthThrow(t *testing.T) {
	// throws when depth exceeds limit with strictDepth: true
	_, err := Parse("a[b][c][d][e][f][g][h][i]=j", WithDepth(1), WithStrictDepth(true))
	if err != ErrDepthLimitExceeded {
		t.Errorf("expected ErrDepthLimitExceeded, got %v", err)
	}

	// throws for multiple nested arrays
	_, err = Parse("a[0][1][2][3][4]=b", WithDepth(3), WithStrictDepth(true))
	if err != ErrDepthLimitExceeded {
		t.Errorf("expected ErrDepthLimitExceeded for arrays, got %v", err)
	}

	// throws for nested objects and arrays
	_, err = Parse("a[b][c][0][d][e]=f", WithDepth(3), WithStrictDepth(true))
	if err != ErrDepthLimitExceeded {
		t.Errorf("expected ErrDepthLimitExceeded for mixed, got %v", err)
	}
}

// ===========================================
// JS Test: "strictDepth option - non-throw cases"
// ===========================================
func TestJSStrictDepthNoThrow(t *testing.T) {
	// when depth is 0 and strictDepth true, do not throw
	_, err := Parse("a[b][c][d][e]=true&a[b][c][d][f]=42", WithDepth(0), WithStrictDepth(true))
	if err != nil {
		t.Errorf("depth 0 should not throw, got %v", err)
	}

	// parses successfully when within limit
	result, err := Parse("a[b]=c", WithDepth(1), WithStrictDepth(true))
	if err != nil {
		t.Errorf("should not throw when within limit, got %v", err)
	}
	assertEqual(t, result, map[string]any{"a": map[string]any{"b": "c"}}, "within depth limit")

	// does not throw when exactly at limit
	result, err = Parse("a[b][c]=d", WithDepth(2), WithStrictDepth(true))
	if err != nil {
		t.Errorf("should not throw at exact limit, got %v", err)
	}
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"b": map[string]any{"c": "d"}},
	}, "at exact depth limit")
}

// ===========================================
// JS Test: parameter limit tests
// ===========================================
func TestJSParameterLimitTests(t *testing.T) {
	// does not throw when within limit
	result, err := Parse("a=1&b=2&c=3", WithParameterLimit(5), WithThrowOnLimitExceeded(true))
	if err != nil {
		t.Errorf("should not throw within limit, got %v", err)
	}
	assertEqual(t, result, map[string]any{"a": "1", "b": "2", "c": "3"}, "within parameter limit")

	// throws when exceeded
	_, err = Parse("a=1&b=2&c=3&d=4&e=5&f=6", WithParameterLimit(3), WithThrowOnLimitExceeded(true))
	if err != ErrParameterLimitExceeded {
		t.Errorf("expected ErrParameterLimitExceeded, got %v", err)
	}

	// silently truncates without throwOnLimitExceeded
	result, _ = Parse("a=1&b=2&c=3&d=4&e=5", WithParameterLimit(3))
	assertEqual(t, result, map[string]any{"a": "1", "b": "2", "c": "3"}, "truncated silently")

	// silently truncates with throwOnLimitExceeded: false
	result, _ = Parse("a=1&b=2&c=3&d=4&e=5", WithParameterLimit(3), WithThrowOnLimitExceeded(false))
	assertEqual(t, result, map[string]any{"a": "1", "b": "2", "c": "3"}, "truncated with false")
}

// ===========================================
// JS Test: array limit tests
// ===========================================
func TestJSArrayLimitTests(t *testing.T) {
	// does not throw when within limit
	result, err := Parse("a[]=1&a[]=2&a[]=3", WithArrayLimit(5), WithThrowOnLimitExceeded(true))
	if err != nil {
		t.Errorf("should not throw within array limit, got %v", err)
	}
	assertEqual(t, result, map[string]any{"a": []any{"1", "2", "3"}}, "within array limit")

	// throws when exceeded
	_, err = Parse("a[]=1&a[]=2&a[]=3&a[]=4", WithArrayLimit(3), WithThrowOnLimitExceeded(true))
	if err != ErrArrayLimitExceeded {
		t.Errorf("expected ErrArrayLimitExceeded, got %v", err)
	}

	// converts to object when index greater than limit
	result, _ = Parse("a[1]=1&a[2]=2&a[3]=3&a[4]=4&a[5]=5&a[6]=6", WithArrayLimit(5))
	assertEqual(t, result, map[string]any{
		"a": map[string]any{"1": "1", "2": "2", "3": "3", "4": "4", "5": "5", "6": "6"},
	}, "converted to object")
}

// ===========================================
// JS Test: empty keys cases
// ===========================================
func TestJSEmptyKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]any
	}{
		{"&", map[string]any{}},
		{"&&", map[string]any{}},
		{"&=", map[string]any{}},
		{"&=&", map[string]any{}},
		{"&=&=", map[string]any{}},
		{"&=&=&", map[string]any{}},
		{"=", map[string]any{}},
		{"=&", map[string]any{}},
		{"=&&&", map[string]any{}},
		{"=&=&=&", map[string]any{}},
		{"=&a[]=b&a[1]=c", map[string]any{"a": []any{"b", "c"}}},
		{"=a", map[string]any{}},
		{"a==a", map[string]any{"a": "=a"}},
		{"=&a[]=b", map[string]any{"a": []any{"b"}}},
		{"=&a[]=b&a[]=c&a[2]=d", map[string]any{"a": []any{"b", "c", "d"}}},
		{"=a&=b", map[string]any{}},
		{"=a&foo=b", map[string]any{"foo": "b"}},
		{"a[]=b&a=c&=", map[string]any{"a": []any{"b", "c"}}},
		{"a[0]=b&a=c&=", map[string]any{"a": []any{"b", "c"}}},
		{"a=b&a[]=c&=", map[string]any{"a": []any{"b", "c"}}},
		{"a=b&a[0]=c&=", map[string]any{"a": []any{"b", "c"}}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertEqual(t, result, tt.expected, tt.input)
		})
	}
}

// ===========================================
// JS Test: empty keys with brackets
// ===========================================
func TestJSEmptyKeysWithBrackets(t *testing.T) {
	// st.deepEqual(qs.parse('[]=a&[]=b& []=1'), { 0: 'a', 1: 'b', ' ': ['1'] }); - noEmptyKeys behavior
	result, _ := Parse("[]=a&[]=b& []=1")
	assertEqual(t, result, map[string]any{
		"0": "a", "1": "b", " ": []any{"1"},
	}, "empty bracket indices")

	// st.deepEqual(qs.parse('[0]=a&[1]=b&a[0]=1&a[1]=2'), { 0: 'a', 1: 'b', a: ['1', '2'] });
	result, _ = Parse("[0]=a&[1]=b&a[0]=1&a[1]=2")
	assertEqual(t, result, map[string]any{
		"0": "a", "1": "b", "a": []any{"1", "2"},
	}, "numeric bracket keys")

	// st.deepEqual(qs.parse('[deep]=a&[deep]=2'), { deep: ['a', '2'] });
	result, _ = Parse("[deep]=a&[deep]=2")
	assertEqual(t, result, map[string]any{
		"deep": []any{"a", "2"},
	}, "deep key")

	// st.deepEqual(qs.parse('%5B0%5D=a&%5B1%5D=b'), { 0: 'a', 1: 'b' });
	result, _ = Parse("%5B0%5D=a&%5B1%5D=b")
	assertEqual(t, result, map[string]any{
		"0": "a", "1": "b",
	}, "encoded brackets")
}

// ===========================================
// JS Test: "allows setting the parameter limit to Infinity"
// Go equivalent: math.MaxInt
// ===========================================
func TestJSParameterLimitInfinity(t *testing.T) {
	// Generate a query string with many parameters
	var parts []string
	for i := 0; i < 2000; i++ {
		parts = append(parts, "a"+strconv.Itoa(i)+"="+strconv.Itoa(i))
	}
	input := strings.Join(parts, "&")

	// With math.MaxInt, all parameters should be parsed
	result, err := Parse(input, WithParameterLimit(int(^uint(0)>>1))) // math.MaxInt
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2000 {
		t.Errorf("expected 2000 parameters, got %d", len(result))
	}
}

// ===========================================
// JS Test: "does not error when parsing a very long array"
// ===========================================
func TestJSVeryLongArray(t *testing.T) {
	// Generate a very long array
	var parts []string
	for i := 0; i < 5000; i++ {
		parts = append(parts, "a[]="+strconv.Itoa(i))
	}
	input := strings.Join(parts, "&")

	// Should parse without error
	result, err := Parse(input, WithParameterLimit(10000), WithArrayLimit(10000))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := result["a"].([]any)
	if !ok {
		t.Fatalf("result['a'] should be array, got %T", result["a"])
	}
	if len(arr) != 5000 {
		t.Errorf("expected 5000 elements, got %d", len(arr))
	}
}

// ===========================================
// JS Test: "ignores an utf8 sentinel with an unknown value"
// When utf8 parameter has unknown value (not ✓ or &#10003;), it should be
// treated as a regular parameter and included in results.
// ===========================================
func TestJSIgnoreUnknownUTF8Sentinel(t *testing.T) {
	// When charset sentinel is enabled but utf8 has unknown value:
	// - charset should remain default (UTF-8)
	// - utf8 parameter should be included in result as regular param
	result, err := Parse("utf8=unknown&a=%C3%B8", WithCharsetSentinel(true), WithCharset(CharsetUTF8))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// utf8=unknown is NOT a valid sentinel, so it should be included as regular parameter
	if result["utf8"] != "unknown" {
		t.Errorf("utf8 should be 'unknown' (not a valid sentinel), got %v", result["utf8"])
	}

	// ø should be decoded as UTF-8 (unknown sentinel keeps default charset)
	if result["a"] != "ø" {
		t.Errorf("a should be 'ø', got %v", result["a"])
	}
}

// ===========================================
// JS Test: "uses the utf8 sentinel to switch to iso-8859-1 when no default charset is given"
// ===========================================
func TestJSUTF8SentinelDetectsISO(t *testing.T) {
	urlEncodedNumCheckmark := "%26%2310003%3B"
	urlEncodedOSlashInUtf8 := "%C3%B8"

	// When charset sentinel is present for ISO-8859-1, should switch charset
	// Note: default charset is UTF-8, but sentinel should switch to ISO-8859-1
	result, err := Parse("utf8="+urlEncodedNumCheckmark+"&a="+urlEncodedOSlashInUtf8,
		WithCharsetSentinel(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The ø encoded as UTF-8 bytes interpreted as ISO-8859-1 becomes "Ã¸"
	if result["a"] != "Ã¸" {
		t.Errorf("a should be 'Ã¸' (UTF-8 bytes as ISO-8859-1), got %q", result["a"])
	}
}

// ===========================================
// JS Test: "interpretNumericEntities with comma:true and iso charset does not crash"
// ===========================================
func TestJSNumericEntitiesWithComma(t *testing.T) {
	urlEncodedNumSmiley := "%26%239786%3B"

	result, err := Parse("foo="+urlEncodedNumSmiley+",bar",
		WithCharset(CharsetISO88591),
		WithInterpretNumericEntities(true),
		WithComma(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := result["foo"].([]any)
	if !ok {
		t.Fatalf("result['foo'] should be array, got %T: %v", result["foo"], result["foo"])
	}

	if len(arr) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr))
	}

	// First element should be the smiley
	if arr[0] != "☺" {
		t.Errorf("arr[0] should be '☺', got %q", arr[0])
	}

	// Second element should be "bar"
	if arr[1] != "bar" {
		t.Errorf("arr[1] should be 'bar', got %q", arr[1])
	}
}

// ===========================================
// JS Test: "use number decoder"
// ===========================================
func TestJSCustomDecoderWithComma(t *testing.T) {
	// Custom decoder that parses numbers
	numberDecoder := func(str string, charset Charset, kind string) (string, error) {
		// Just return the string, we'll check if comma parsing works with custom decoder
		return str, nil
	}

	result, err := Parse("a=1,2,3", WithDecoder(numberDecoder), WithComma(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := result["a"].([]any)
	if !ok {
		t.Fatalf("result['a'] should be array, got %T", result["a"])
	}

	if len(arr) != 3 {
		t.Errorf("expected 3 elements, got %d", len(arr))
	}
}

// ===========================================
// JS Test: "can parse with custom encoding"
// ===========================================
func TestJSCustomEncoding(t *testing.T) {
	// Custom decoder that transforms values
	customDecoder := func(str string, charset Charset, kind string) (string, error) {
		// Simple transformation: uppercase all values
		if kind == "value" {
			return strings.ToUpper(str), nil
		}
		return str, nil
	}

	result, err := Parse("a=hello&b=world", WithDecoder(customDecoder))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["a"] != "HELLO" {
		t.Errorf("result['a'] = %v, want 'HELLO'", result["a"])
	}
	if result["b"] != "WORLD" {
		t.Errorf("result['b'] = %v, want 'WORLD'", result["b"])
	}
}

// ===========================================
// JS Test: "allows for decoding keys and values differently"
// ===========================================
func TestJSDecodeKeyValueDifferently(t *testing.T) {
	customDecoder := func(str string, charset Charset, kind string) (string, error) {
		if kind == "key" {
			return "key_" + str, nil
		}
		return "val_" + str, nil
	}

	result, err := Parse("a=b", WithDecoder(customDecoder))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["key_a"] != "val_b" {
		t.Errorf("result = %v, expected {key_a: val_b}", result)
	}
}

// ===========================================
// JS Test: custom decoder returning error
// Errors from decoder should be propagated for both keys and values.
// ===========================================
func TestJSCustomDecoderError(t *testing.T) {
	// Test error from key decoder (should be propagated)
	keyErrorDecoder := func(str string, charset Charset, kind string) (string, error) {
		if kind == "key" && str == "badkey" {
			return "", errors.New("key decode error")
		}
		return str, nil
	}

	_, err := Parse("badkey=value", WithDecoder(keyErrorDecoder))
	if err == nil {
		t.Error("expected error from key decoder")
	}

	// Test error from value decoder (should also be propagated)
	valueErrorDecoder := func(str string, charset Charset, kind string) (string, error) {
		if kind == "value" && str == "badvalue" {
			return "", errors.New("value decode error")
		}
		return str, nil
	}

	_, err = Parse("a=badvalue", WithDecoder(valueErrorDecoder))
	if err == nil {
		t.Error("expected error from value decoder")
	}
}

// ===========================================
// JS Test: "parses jquery-param strings"
// jQuery serialization format compatibility
// ===========================================
func TestJSJQueryParamStrings(t *testing.T) {
	// jQuery.param serializes spaces as +
	result, err := Parse("a=hello+world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["a"] != "hello world" {
		t.Errorf("result['a'] = %q, want 'hello world'", result["a"])
	}

	// Multiple parameters with +
	result, err = Parse("name=John+Doe&city=New+York")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "John Doe" {
		t.Errorf("result['name'] = %q, want 'John Doe'", result["name"])
	}
	if result["city"] != "New York" {
		t.Errorf("result['city'] = %q, want 'New York'", result["city"])
	}

	// Full jQuery format test from JS parse.js:
	// readable = 'filter[0][]=int1&filter[0][]==&filter[0][]=77&filter[]=and&filter[2][]=int2&filter[2][]==&filter[2][]=8'
	// encoded = 'filter%5B0%5D%5B%5D=int1&filter%5B0%5D%5B%5D=%3D&filter%5B0%5D%5B%5D=77&filter%5B%5D=and&filter%5B2%5D%5B%5D=int2&filter%5B2%5D%5B%5D=%3D&filter%5B2%5D%5B%5D=8'
	// expected = { filter: [['int1', '=', '77'], 'and', ['int2', '=', '8']] }
	result, err = Parse("filter%5B0%5D%5B%5D=int1&filter%5B0%5D%5B%5D=%3D&filter%5B0%5D%5B%5D=77&filter%5B%5D=and&filter%5B2%5D%5B%5D=int2&filter%5B2%5D%5B%5D=%3D&filter%5B2%5D%5B%5D=8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := result["filter"].([]any)
	if !ok {
		t.Fatalf("filter should be array, got %T: %v", result["filter"], result["filter"])
	}

	if len(filter) != 3 {
		t.Fatalf("expected 3 elements, got %d: %v", len(filter), filter)
	}

	// First element: ['int1', '=', '77']
	first, ok := filter[0].([]any)
	if !ok {
		t.Errorf("filter[0] should be array, got %T: %v", filter[0], filter[0])
	} else {
		assertEqual(t, first, []any{"int1", "=", "77"}, "filter[0]")
	}

	// Second element: 'and'
	if filter[1] != "and" {
		t.Errorf("filter[1] = %v, expected 'and'", filter[1])
	}

	// Third element: ['int2', '=', '8']
	third, ok := filter[2].([]any)
	if !ok {
		t.Errorf("filter[2] should be array, got %T: %v", filter[2], filter[2])
	} else {
		assertEqual(t, third, []any{"int2", "=", "8"}, "filter[2]")
	}
}

// ===========================================
// JS Test: "does not interpret %uXXXX syntax in iso-8859-1 mode"
// ===========================================
func TestJSNoPercentUInISO(t *testing.T) {
	// %uXXXX is not a valid escape sequence and should be preserved
	result, err := Parse("a=%u0041", WithCharset(CharsetISO88591))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT be converted to "A" - %uXXXX is not valid URL encoding
	if result["a"] != "%u0041" {
		t.Errorf("result['a'] = %q, want '%%u0041'", result["a"])
	}
}

// ===========================================
// JS Test: "parses url-encoded brackets holds array of arrays"
// ===========================================
func TestJSURLEncodedBracketsArrayOfArrays(t *testing.T) {
	// st.deepEqual(qs.parse('foo%5B%5D=1,2,3&foo%5B%5D=4,5,6', { comma: true }), { foo: [['1', '2', '3'], ['4', '5', '6']] });
	result, err := Parse("foo%5B%5D=1,2,3&foo%5B%5D=4,5,6", WithComma(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, result, map[string]any{
		"foo": []any{[]any{"1", "2", "3"}, []any{"4", "5", "6"}},
	}, "url-encoded brackets array of arrays")
}

// ===========================================
// JS Test: "can return null objects" (plainObjects option)
// ===========================================
func TestJSPlainObjects(t *testing.T) {
	// In Go, PlainObjects doesn't have the same effect as JS (no prototype chain)
	// but we test that the option is accepted and parsing works
	result, err := Parse("a[b]=c", WithPlainObjects(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a, ok := result["a"].(map[string]any)
	if !ok {
		t.Fatalf("result['a'] should be map, got %T", result["a"])
	}
	if a["b"] != "c" {
		t.Errorf("result['a']['b'] = %v, want 'c'", a["b"])
	}

	// With plainObjects, prototype keys like "toString" should be allowed
	result, err = Parse("toString=foo", WithPlainObjects(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["toString"] != "foo" {
		t.Errorf("result['toString'] = %v, want 'foo'", result["toString"])
	}
}

// ===========================================
// Additional edge cases
// ===========================================
func TestJSSkipsEmptyStringKey(t *testing.T) {
	// Empty key parts should be handled correctly
	result, err := Parse("=value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty key should be skipped
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestJSMultipleEmptyKeys(t *testing.T) {
	// Multiple empty keys
	result, err := Parse("=a&=b&=c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All empty keys should be skipped
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestJSDelimiterEdgeCases(t *testing.T) {
	// Multi-character delimiter
	result, err := Parse("a=1||b=2||c=3", WithDelimiter("||"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["a"] != "1" || result["b"] != "2" || result["c"] != "3" {
		t.Errorf("multi-char delimiter failed: %v", result)
	}
}
