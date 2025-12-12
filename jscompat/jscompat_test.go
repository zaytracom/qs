package jscompat

import (
	"encoding/json"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	qs "github.com/phl/qs/v2"
)

// runJS executes a JavaScript snippet and returns the output
func runJS(t *testing.T, code string) string {
	t.Helper()
	fullCode := `const qs = require('qs');` + code
	cmd := exec.Command("node", "-e", fullCode)
	cmd.Dir = "/Users/phl/Projects/qs/jscompat"
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("JS execution failed: %v\nOutput: %s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}

// toJSON converts any value to JSON string for comparison
func toJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	return string(b)
}

// parseJSON parses a JSON string into a Go value
func parseJSON(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	return v
}

// deepEqual compares two values semantically (ignoring key order)
func deepEqual(t *testing.T, go1, js string) bool {
	t.Helper()
	return reflect.DeepEqual(parseJSON(t, go1), parseJSON(t, js))
}

// normalizeForJSON converts ExplicitNullValue markers to nil for proper JSON comparison
func normalizeForJSON(v any) any {
	if v == nil {
		return nil
	}
	if qs.IsExplicitNull(v) {
		return nil
	}
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, v := range val {
			result[k] = normalizeForJSON(v)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, v := range val {
			result[i] = normalizeForJSON(v)
		}
		return result
	}
	return v
}

// compareQueryStrings compares two query strings semantically (ignoring param order)
func compareQueryStrings(t *testing.T, goQS, jsQS string) bool {
	t.Helper()
	goParams := parseQueryString(goQS)
	jsParams := parseQueryString(jsQS)
	return reflect.DeepEqual(goParams, jsParams)
}

// parseQueryString parses a query string into a map of key -> sorted values
func parseQueryString(s string) map[string][]string {
	return parseQueryStringWithDelimiter(s, "&")
}

// parseQueryStringWithDelimiter parses a query string with a custom delimiter
func parseQueryStringWithDelimiter(s, delimiter string) map[string][]string {
	result := make(map[string][]string)
	if s == "" {
		return result
	}
	// Strip query prefix if present
	if strings.HasPrefix(s, "?") {
		s = s[1:]
	}
	parts := strings.Split(s, delimiter)
	for _, part := range parts {
		idx := strings.Index(part, "=")
		var key, val string
		if idx == -1 {
			key = part
			val = ""
		} else {
			key = part[:idx]
			val = part[idx+1:]
		}
		result[key] = append(result[key], val)
	}
	return result
}

// compareQueryStringsWithDelimiter compares query strings with a custom delimiter
func compareQueryStringsWithDelimiter(t *testing.T, goQS, jsQS, delimiter string) bool {
	t.Helper()
	goParams := parseQueryStringWithDelimiter(goQS, delimiter)
	jsParams := parseQueryStringWithDelimiter(jsQS, delimiter)
	return reflect.DeepEqual(goParams, jsParams)
}

// Test 1: Deep nested objects with arrays
func TestDeepNestedComplex(t *testing.T) {
	input := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{
				"name":   "John Doe",
				"emails": []any{"john@example.com", "doe@work.com"},
				"settings": map[string]any{
					"theme": "dark",
					"notifications": map[string]any{
						"email": true,
						"sms":   false,
						"push":  []any{"morning", "evening"},
					},
				},
			},
			"tags": []any{"admin", "verified", "premium"},
		},
		"meta": map[string]any{
			"version":   "2.0",
			"timestamp": 1702400000,
		},
	}

	// Stringify
	goResult, err := qs.Stringify(input)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = {
			user: {
				profile: {
					name: "John Doe",
					emails: ["john@example.com", "doe@work.com"],
					settings: {
						theme: "dark",
						notifications: {
							email: true,
							sms: false,
							push: ["morning", "evening"]
						}
					}
				},
				tags: ["admin", "verified", "premium"]
			},
			meta: {
				version: "2.0",
				timestamp: 1702400000
			}
		};
		console.log(qs.stringify(input));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}

	// Parse
	goParsed, err := qs.Parse(jsResult)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse(`+"`"+jsResult+"`"+`)));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 2: Sparse arrays with nulls and strict null handling
// NOTE: In Go, nil in slices means "undefined/sparse slot", not "null".
// To represent explicit null, use qs.ExplicitNullValue.
// This test uses ExplicitNullValue for proper JS compatibility.
func TestSparseArraysNulls(t *testing.T) {
	// Use ExplicitNullValue to represent JS null (not Go nil which is undefined)
	input := map[string]any{
		"items": []any{qs.ExplicitNullValue, "first", qs.ExplicitNullValue, "third", qs.ExplicitNullValue},
		"config": map[string]any{
			"enabled": qs.ExplicitNullValue,
			"value":   "test",
		},
		"empty": qs.ExplicitNullValue,
	}

	goResult, err := qs.Stringify(input, qs.WithStringifyStrictNullHandling(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = {
			items: [null, "first", null, "third", null],
			config: { enabled: null, value: "test" },
			empty: null
		};
		console.log(qs.stringify(input, { strictNullHandling: true }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}

	// Parse with strictNullHandling and allowSparse
	// Note: Parse returns nil for null values after Compact converts ExplicitNullValue -> nil
	goParsed, err := qs.Parse(jsResult, qs.WithStrictNullHandling(true), qs.WithAllowSparse(true))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse(`+"`"+jsResult+"`"+`, { strictNullHandling: true, allowSparse: true })));`)

	// Normalize ExplicitNullValue -> nil for JSON comparison
	normalized := normalizeForJSON(goParsed)
	if !deepEqual(t, toJSON(t, normalized), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, normalized), jsParsedJSON)
	}
}

// Test 3: Array format - indices
func TestArrayFormatIndices(t *testing.T) {
	input := map[string]any{
		"colors":  []any{"red", "green", "blue"},
		"numbers": []any{1, 2, 3, 4, 5},
	}

	goResult, err := qs.Stringify(input, qs.WithArrayFormat(qs.ArrayFormatIndices))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ colors: ["red", "green", "blue"], numbers: [1, 2, 3, 4, 5] }, { arrayFormat: "indices" }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 4: Array format - brackets
func TestArrayFormatBrackets(t *testing.T) {
	input := map[string]any{
		"colors":  []any{"red", "green", "blue"},
		"numbers": []any{1, 2, 3, 4, 5},
	}

	goResult, err := qs.Stringify(input, qs.WithArrayFormat(qs.ArrayFormatBrackets))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ colors: ["red", "green", "blue"], numbers: [1, 2, 3, 4, 5] }, { arrayFormat: "brackets" }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 5: Array format - repeat
func TestArrayFormatRepeat(t *testing.T) {
	input := map[string]any{
		"colors":  []any{"red", "green", "blue"},
		"numbers": []any{1, 2, 3, 4, 5},
	}

	goResult, err := qs.Stringify(input, qs.WithArrayFormat(qs.ArrayFormatRepeat))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ colors: ["red", "green", "blue"], numbers: [1, 2, 3, 4, 5] }, { arrayFormat: "repeat" }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 6: Array format - comma
func TestArrayFormatComma(t *testing.T) {
	input := map[string]any{
		"colors":  []any{"red", "green", "blue"},
		"numbers": []any{1, 2, 3, 4, 5},
	}

	goResult, err := qs.Stringify(input, qs.WithArrayFormat(qs.ArrayFormatComma))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ colors: ["red", "green", "blue"], numbers: [1, 2, 3, 4, 5] }, { arrayFormat: "comma" }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 7: Dot notation with encoded dots in keys
func TestDotNotationEncoded(t *testing.T) {
	input := map[string]any{
		"user.name": "John",
		"config": map[string]any{
			"api.key": "secret123",
			"nested": map[string]any{
				"deep.value": 42,
			},
		},
	}

	goResult, err := qs.Stringify(input, qs.WithStringifyAllowDots(true), qs.WithEncodeDotInKeys(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = {
			"user.name": "John",
			config: {
				"api.key": "secret123",
				nested: { "deep.value": 42 }
			}
		};
		console.log(qs.stringify(input, { allowDots: true, encodeDotInKeys: true }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 8: Special characters and unicode
func TestSpecialCharsUnicode(t *testing.T) {
	input := map[string]any{
		"message":   "Hello, World! @#$%^&*()",
		"unicode":   "日本語テスト",
		"emoji":     "🎉🚀✨",
		"spaces":    "multiple   spaces   here",
		"quotes":    "\"quoted\" and 'single'",
		"ampersand": "tom&jerry",
		"equals":    "a=b=c",
	}

	goResult, err := qs.Stringify(input)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = {
			message: "Hello, World! @#$%^&*()",
			unicode: "日本語テスト",
			emoji: "🎉🚀✨",
			spaces: "multiple   spaces   here",
			quotes: '"quoted" and ' + "'single'",
			ampersand: "tom&jerry",
			equals: "a=b=c"
		};
		console.log(qs.stringify(input));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}

	// Parse
	goParsed, err := qs.Parse(jsResult)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse(`+"`"+jsResult+"`"+`)));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 9: Empty values and empty arrays
func TestEmptyValues(t *testing.T) {
	input := map[string]any{
		"emptyString": "",
		"emptyArray":  []any{},
		"nested": map[string]any{
			"also": map[string]any{
				"empty": "",
			},
		},
		"normalValue": "present",
	}

	goResult, err := qs.Stringify(input, qs.WithStringifyAllowEmptyArrays(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = {
			emptyString: "",
			emptyArray: [],
			nested: { also: { empty: "" } },
			normalValue: "present"
		};
		console.log(qs.stringify(input, { allowEmptyArrays: true }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 10: RFC1738 format
func TestRFCFormats(t *testing.T) {
	input := map[string]any{
		"space":   "hello world",
		"special": "a+b=c&d",
	}

	goResult, err := qs.Stringify(input, qs.WithFormat(qs.FormatRFC1738))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ space: "hello world", special: "a+b=c&d" }, { format: "RFC1738" }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 11: Filter and skip nulls with sort
func TestFilterSortSkipNulls(t *testing.T) {
	input := map[string]any{
		"zebra":  "last",
		"apple":  "first",
		"mango":  nil,
		"banana": "middle",
		"cherry": nil,
	}

	goResult, err := qs.Stringify(input,
		qs.WithSkipNulls(true),
		qs.WithSort(func(a, b string) bool { return a < b }))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = { zebra: "last", apple: "first", mango: null, banana: "middle", cherry: null };
		console.log(qs.stringify(input, { skipNulls: true, sort: (a, b) => a.localeCompare(b) }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 12: Depth limits
func TestDepthLimits(t *testing.T) {
	// Test stringify of deep object
	input := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": map[string]any{
							"f": map[string]any{
								"g": "deep",
							},
						},
					},
				},
			},
		},
	}

	goResult, err := qs.Stringify(input)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = { a: { b: { c: { d: { e: { f: { g: "deep" } } } } } } };
		console.log(qs.stringify(input));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}

	// Test parse with depth limit
	goParsed, err := qs.Parse(jsResult, qs.WithDepth(3))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse(`+"`"+jsResult+"`"+`, { depth: 3 })));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 13: Real-world API query
func TestRealWorldAPI(t *testing.T) {
	input := map[string]any{
		"filters": map[string]any{
			"status": []any{"active", "pending"},
			"created": map[string]any{
				"$gte": "2024-01-01",
				"$lte": "2024-12-31",
			},
			"tags": map[string]any{
				"$in": []any{"important", "urgent"},
			},
		},
		"pagination": map[string]any{
			"page":  1,
			"limit": 25,
			"sort":  "-createdAt",
		},
		"populate": []any{"user", "comments"},
		"select":   []any{"id", "title", "status", "createdAt"},
	}

	goResult, err := qs.Stringify(input)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = {
			filters: {
				status: ["active", "pending"],
				created: { "$gte": "2024-01-01", "$lte": "2024-12-31" },
				tags: { "$in": ["important", "urgent"] }
			},
			pagination: { page: 1, limit: 25, sort: "-createdAt" },
			populate: ["user", "comments"],
			select: ["id", "title", "status", "createdAt"]
		};
		console.log(qs.stringify(input));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}

	// Parse
	goParsed, err := qs.Parse(jsResult)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse(`+"`"+jsResult+"`"+`)));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 14: Query prefix
func TestQueryPrefix(t *testing.T) {
	input := map[string]any{"a": "b", "c": "d"}

	goResult, err := qs.Stringify(input, qs.WithStringifyAddQueryPrefix(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ a: "b", c: "d" }, { addQueryPrefix: true }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}

	// Parse with ignoreQueryPrefix
	goParsed, err := qs.Parse("?a=b&c=d", qs.WithIgnoreQueryPrefix(true))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("?a=b&c=d", { ignoreQueryPrefix: true })));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 15: Charset sentinel
func TestCharsetSentinel(t *testing.T) {
	input := map[string]any{"a": "b"}

	goResult, err := qs.Stringify(input, qs.WithStringifyCharsetSentinel(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ a: "b" }, { charsetSentinel: true }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 16: Comma round trip
func TestCommaRoundTrip(t *testing.T) {
	input := map[string]any{"a": []any{"b"}}

	goResult, err := qs.Stringify(input, qs.WithArrayFormat(qs.ArrayFormatComma), qs.WithCommaRoundTrip(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ a: ["b"] }, { arrayFormat: "comma", commaRoundTrip: true }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 17: Encode values only
func TestEncodeValuesOnly(t *testing.T) {
	input := map[string]any{"a[b]": "c d"}

	goResult, err := qs.Stringify(input, qs.WithEncodeValuesOnly(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ "a[b]": "c d" }, { encodeValuesOnly: true }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 18: Custom delimiter
func TestCustomDelimiter(t *testing.T) {
	input := map[string]any{"a": "1", "b": "2"}

	goResult, err := qs.Stringify(input, qs.WithStringifyDelimiter(";"))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ a: "1", b: "2" }, { delimiter: ";" }));`)

	if !compareQueryStringsWithDelimiter(t, goResult, jsResult, ";") {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}

	// Parse with custom delimiter
	goParsed, err := qs.Parse("a=1;b=2", qs.WithDelimiter(";"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a=1;b=2", { delimiter: ";" })));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 19: Array limit
func TestArrayLimit(t *testing.T) {
	goParsed, err := qs.Parse("a[100]=b", qs.WithArrayLimit(50))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a[100]=b", { arrayLimit: 50 })));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 20: Duplicates handling
func TestDuplicatesHandling(t *testing.T) {
	// Combine
	goParsedCombine, err := qs.Parse("a=1&a=2&a=3", qs.WithDuplicates(qs.DuplicateCombine))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsCombine := runJS(t, `console.log(JSON.stringify(qs.parse("a=1&a=2&a=3", { duplicates: "combine" })));`)
	if !deepEqual(t, toJSON(t, goParsedCombine), jsCombine) {
		t.Errorf("Combine mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsedCombine), jsCombine)
	}

	// First
	goParsedFirst, err := qs.Parse("a=1&a=2&a=3", qs.WithDuplicates(qs.DuplicateFirst))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsFirst := runJS(t, `console.log(JSON.stringify(qs.parse("a=1&a=2&a=3", { duplicates: "first" })));`)
	if !deepEqual(t, toJSON(t, goParsedFirst), jsFirst) {
		t.Errorf("First mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsedFirst), jsFirst)
	}

	// Last
	goParsedLast, err := qs.Parse("a=1&a=2&a=3", qs.WithDuplicates(qs.DuplicateLast))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsLast := runJS(t, `console.log(JSON.stringify(qs.parse("a=1&a=2&a=3", { duplicates: "last" })));`)
	if !deepEqual(t, toJSON(t, goParsedLast), jsLast) {
		t.Errorf("Last mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsedLast), jsLast)
	}
}

// Test 21: Nested arrays
func TestNestedArrays(t *testing.T) {
	input := map[string]any{
		"a": []any{
			[]any{"b", "c"},
			[]any{"d", "e"},
		},
	}

	goResult, err := qs.Stringify(input)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ a: [["b", "c"], ["d", "e"]] }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}

	// Parse
	goParsed, err := qs.Parse(jsResult)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse(`+"`"+jsResult+"`"+`)));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 22: Boolean and number handling
func TestBooleanAndNumbers(t *testing.T) {
	input := map[string]any{
		"bool":  true,
		"num":   42,
		"float": 3.14,
	}

	goResult, err := qs.Stringify(input)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ bool: true, num: 42, float: 3.14 }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 23: ISO-8859-1 charset
func TestISO88591Charset(t *testing.T) {
	input := map[string]any{"a": "ä"}

	goResult, err := qs.Stringify(input, qs.WithStringifyCharset(qs.CharsetISO88591))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ a: "ä" }, { charset: "iso-8859-1" }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 24: Mega complex test
func TestMegaComplex(t *testing.T) {
	input := map[string]any{
		"users": []any{
			map[string]any{
				"id":    1,
				"name":  "Alice",
				"roles": []any{"admin", "user"},
				"settings": map[string]any{
					"theme":              "dark",
					"notification.email": true,
				},
			},
			map[string]any{
				"id":    2,
				"name":  "Bob",
				"roles": []any{"user"},
				"settings": map[string]any{
					"theme":              "light",
					"notification.email": false,
				},
			},
		},
		"filters": map[string]any{
			"active": true,
			"created": map[string]any{
				"from": "2024-01-01",
				"to":   "2024-12-31",
			},
		},
		"search": "hello world",
		"tags":   []any{"important", "urgent", "review"},
		"pagination": map[string]any{
			"page": 1,
			"size": 20,
		},
		"special": "a=b&c=d",
		"unicode": "日本語",
		"empty":   "",
		"nullVal": nil,
	}

	goResult, err := qs.Stringify(input,
		qs.WithStringifyAllowDots(true),
		qs.WithEncodeDotInKeys(true),
		qs.WithArrayFormat(qs.ArrayFormatIndices))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = {
			users: [
				{ id: 1, name: "Alice", roles: ["admin", "user"], settings: { theme: "dark", "notification.email": true } },
				{ id: 2, name: "Bob", roles: ["user"], settings: { theme: "light", "notification.email": false } }
			],
			filters: { active: true, created: { from: "2024-01-01", to: "2024-12-31" } },
			search: "hello world",
			tags: ["important", "urgent", "review"],
			pagination: { page: 1, size: 20 },
			special: "a=b&c=d",
			unicode: "日本語",
			empty: "",
			nullVal: null
		};
		console.log(qs.stringify(input, { allowDots: true, encodeDotInKeys: true, arrayFormat: "indices" }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}
