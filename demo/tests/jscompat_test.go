package jscompat

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	qs "github.com/zaytracom/qs/v2"
)

// runJS executes a JavaScript snippet and returns the output
func runJS(t *testing.T, code string) string {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("node is required for jscompat tests: %v", err)
	}

	demoDir, err := demoRootDir()
	if err != nil {
		t.Fatalf("failed to locate demo directory: %v", err)
	}

	fullCode := `const qs = require('qs');` + code

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "node", "-e", fullCode)
	cmd.Dir = demoDir
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("JS execution timed out: %v\nOutput: %s", ctx.Err(), string(out))
	}
	if err != nil {
		t.Fatalf("JS execution failed: %v\nOutput: %s", err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func demoRootDir() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller failed")
	}
	// This file lives in <repo>/demo/tests/jscompat_test.go
	return filepath.Dir(filepath.Dir(thisFile)), nil
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

// compareQueryStrings compares two query strings semantically (ignoring param order)
func compareQueryStrings(t *testing.T, goQS, jsQS string) bool {
	t.Helper()
	goParams := parseQueryString(goQS)
	jsParams := parseQueryString(jsQS)
	return reflect.DeepEqual(goParams, jsParams)
}

// parseQueryString parses a query string into a map of key -> values.
// Values are sorted to make comparison independent of param order.
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
	s = strings.TrimPrefix(s, "?")
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
	for k := range result {
		sort.Strings(result[k])
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

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
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

// =============================================================================
// MISSING PARSE OPTIONS TESTS
// =============================================================================

// Test 25: AllowPrototypes - allows __proto__ and similar keys
func TestAllowPrototypes(t *testing.T) {
	// By default, __proto__ is blocked
	goParsedBlocked, err := qs.Parse("a[__proto__][b]=c")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsBlockedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a[__proto__][b]=c")));`)
	if !deepEqual(t, toJSON(t, goParsedBlocked), jsBlockedJSON) {
		t.Errorf("Blocked mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsedBlocked), jsBlockedJSON)
	}

	// With allowPrototypes: true
	goParsedAllowed, err := qs.Parse("a[__proto__][b]=c", qs.WithAllowPrototypes(true))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsAllowedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a[__proto__][b]=c", { allowPrototypes: true })));`)
	if !deepEqual(t, toJSON(t, goParsedAllowed), jsAllowedJSON) {
		t.Errorf("Allowed mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsedAllowed), jsAllowedJSON)
	}
}

// Test 26: DelimiterRegexp - regex delimiter
func TestDelimiterRegexp(t *testing.T) {
	re := regexp.MustCompile(`[;,]`)
	goParsed, err := qs.Parse("a=1;b=2,c=3", qs.WithDelimiterRegexp(re))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a=1;b=2,c=3", { delimiter: /[;,]/ })));`)
	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 27: InterpretNumericEntities - HTML entities in ISO-8859-1
func TestInterpretNumericEntities(t *testing.T) {
	goParsed, err := qs.Parse("a=%26%23945%3B", qs.WithCharset(qs.CharsetISO88591), qs.WithInterpretNumericEntities(true))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a=%26%23945%3B", { charset: "iso-8859-1", interpretNumericEntities: true })));`)
	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 28: ParameterLimit
func TestParameterLimit(t *testing.T) {
	goParsed, err := qs.Parse("a=1&b=2&c=3&d=4&e=5", qs.WithParameterLimit(3))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a=1&b=2&c=3&d=4&e=5", { parameterLimit: 3 })));`)
	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 29: ParseArrays false - disable array parsing
func TestParseArraysFalse(t *testing.T) {
	goParsed, err := qs.Parse("a[0]=b&a[1]=c", qs.WithParseArrays(false))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a[0]=b&a[1]=c", { parseArrays: false })));`)
	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 30: PlainObjects
func TestPlainObjects(t *testing.T) {
	goParsed, err := qs.Parse("a[hasOwnProperty]=b", qs.WithPlainObjects(true))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a[hasOwnProperty]=b", { plainObjects: true })));`)
	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 31: StrictDepth - error when depth exceeded
func TestStrictDepth(t *testing.T) {
	// Should NOT error when within depth
	goParsedOK, err := qs.Parse("a[b][c]=d", qs.WithDepth(2), qs.WithStrictDepth(true))
	if err != nil {
		t.Fatalf("Parse should not fail within depth: %v", err)
	}
	jsOKJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a[b][c]=d", { depth: 2, strictDepth: true })));`)
	if !deepEqual(t, toJSON(t, goParsedOK), jsOKJSON) {
		t.Errorf("OK parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsedOK), jsOKJSON)
	}

	// Should error when exceeding depth
	_, err = qs.Parse("a[b][c][d]=e", qs.WithDepth(2), qs.WithStrictDepth(true))
	if err == nil {
		t.Errorf("Expected error when strictDepth exceeded, got nil")
	}

	// JS throws an error too
	jsError := runJS(t, `
		try {
			qs.parse("a[b][c][d]=e", { depth: 2, strictDepth: true });
			console.log("no_error");
		} catch(e) {
			console.log("error");
		}
	`)
	if jsError != "error" {
		t.Errorf("JS should throw error for strictDepth exceeded")
	}
}

// Test 32: ThrowOnLimitExceeded - error when parameter limit exceeded
func TestThrowOnLimitExceeded(t *testing.T) {
	// Should error when exceeding parameter limit
	_, err := qs.Parse("a=1&b=2&c=3", qs.WithParameterLimit(2), qs.WithThrowOnLimitExceeded(true))
	if err == nil {
		t.Errorf("Expected error when parameterLimit exceeded with throwOnLimitExceeded")
	}

	// JS throws too
	jsError := runJS(t, `
		try {
			qs.parse("a=1&b=2&c=3", { parameterLimit: 2, throwOnLimitExceeded: true });
			console.log("no_error");
		} catch(e) {
			console.log("error");
		}
	`)
	if jsError != "error" {
		t.Errorf("JS should throw error for parameterLimit exceeded")
	}
}

// Test 33: Custom Decoder
func TestCustomDecoder(t *testing.T) {
	// Custom decoder that adds prefix to values
	decoder := func(str string, charset qs.Charset, kind string) (string, error) {
		decoded := qs.Decode(str, charset)
		if kind == "value" {
			return "PREFIX_" + decoded, nil
		}
		return decoded, nil
	}

	goParsed, err := qs.Parse("foo=bar", qs.WithDecoder(decoder))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// JS decoder takes 4 args: (str, defaultDecoder, charset, kind)
	jsParsedJSON := runJS(t, `
		const decoder = (str, defaultDecoder, charset, kind) => {
			const decoded = defaultDecoder(str, defaultDecoder, charset);
			if (kind === 'value') return 'PREFIX_' + decoded;
			return decoded;
		};
		console.log(JSON.stringify(qs.parse("foo=bar", { decoder })));
	`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Custom decoder mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// =============================================================================
// MISSING STRINGIFY OPTIONS TESTS
// =============================================================================

// Test 34: Encode false - disable URL encoding
func TestEncodeDisabled(t *testing.T) {
	input := map[string]any{"a": "hello world", "b": "foo&bar"}

	goResult, err := qs.Stringify(input,
		qs.WithEncode(false),
		qs.WithSort(func(a, b string) bool { return a < b }),
	)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ a: "hello world", b: "foo&bar" }, {
			encode: false,
			sort: (a, b) => a.localeCompare(b)
		}));
	`)

	if goResult != jsResult {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 35: Custom Encoder
func TestCustomEncoder(t *testing.T) {
	// Custom encoder that adds a suffix to values
	encoder := func(str string, charset qs.Charset, kind string, format qs.Format) string {
		encoded := qs.Encode(str, charset, format)
		if kind == "value" {
			return encoded + "_SUFFIX"
		}
		return encoded
	}

	input := map[string]any{"a": "test"}
	goResult, err := qs.Stringify(input, qs.WithEncoder(encoder))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	// JS encoder takes 4 args: (str, defaultEncoder, charset, kind)
	jsResult := runJS(t, `
		const encoder = (str, defaultEncoder, charset, kind) => {
			const encoded = defaultEncoder(str, defaultEncoder, charset);
			if (kind === 'value') return encoded + '_SUFFIX';
			return encoded;
		};
		console.log(qs.stringify({ a: "test" }, { encoder }));
	`)

	if goResult != jsResult {
		t.Errorf("Custom encoder mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 36: Filter as function
func TestFilterFunction(t *testing.T) {
	input := map[string]any{
		"a": "1",
		"b": "2",
		"c": "3",
	}

	// Filter function that only includes keys starting with 'a' or 'b'
	filter := func(prefix string, value any) any {
		if strings.HasPrefix(prefix, "a") || strings.HasPrefix(prefix, "b") {
			return value
		}
		return nil
	}

	goResult, err := qs.Stringify(input, qs.WithFilter(filter))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const filter = (prefix, value) => {
			if (prefix.startsWith('a') || prefix.startsWith('b')) return value;
			return;
		};
		console.log(qs.stringify({ a: "1", b: "2", c: "3" }, { filter }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Filter function mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 37: Filter as array of keys
func TestFilterArray(t *testing.T) {
	input := map[string]any{
		"a": "1",
		"b": "2",
		"c": "3",
	}

	goResult, err := qs.Stringify(input, qs.WithFilter([]string{"a", "c"}))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ a: "1", b: "2", c: "3" }, { filter: ["a", "c"] }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Filter array mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 38: SerializeDate
func TestSerializeDate(t *testing.T) {
	// Use a fixed timestamp for testing
	// Note: We can't directly compare time.Time with JS Date, so we test the serialize function behavior
	input := map[string]any{
		"date": "2024-01-15T10:30:00Z",
	}

	goResult, err := qs.Stringify(input)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `console.log(qs.stringify({ date: "2024-01-15T10:30:00Z" }));`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Date stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// =============================================================================
// IMPORTANT OPTION COMBINATIONS
// =============================================================================

// Test 39: allowDots + depth + strictDepth
func TestAllowDotsWithDepthStrict(t *testing.T) {
	goParsed, err := qs.Parse("a.b.c.d=e", qs.WithAllowDots(true), qs.WithDepth(2), qs.WithStrictDepth(false))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a.b.c.d=e", { allowDots: true, depth: 2, strictDepth: false })));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 40: comma + arrayLimit
func TestCommaWithArrayLimit(t *testing.T) {
	goParsed, err := qs.Parse("a=1,2,3,4,5", qs.WithComma(true), qs.WithArrayLimit(2))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse("a=1,2,3,4,5", { comma: true, arrayLimit: 2 })));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 41: strictNullHandling + allowEmptyArrays
func TestStrictNullWithEmptyArrays(t *testing.T) {
	input := map[string]any{
		"a":     nil,
		"b":     []any{},
		"c":     "value",
		"empty": []any{},
	}

	goResult, err := qs.Stringify(input,
		qs.WithStringifyStrictNullHandling(true),
		qs.WithStringifyAllowEmptyArrays(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ a: null, b: [], c: "value", empty: [] },
			{ strictNullHandling: true, allowEmptyArrays: true }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 42: arrayFormat brackets + encodeValuesOnly
func TestBracketsWithEncodeValuesOnly(t *testing.T) {
	input := map[string]any{
		"items": []any{"hello world", "foo&bar"},
	}

	goResult, err := qs.Stringify(input,
		qs.WithArrayFormat(qs.ArrayFormatBrackets),
		qs.WithEncodeValuesOnly(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ items: ["hello world", "foo&bar"] },
			{ arrayFormat: "brackets", encodeValuesOnly: true }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 43: charset ISO-8859-1 + charsetSentinel + interpretNumericEntities
func TestISO8859WithSentinelAndEntities(t *testing.T) {
	// Stringify with ISO charset and sentinel
	input := map[string]any{"a": "ä", "b": "test"}

	goResult, err := qs.Stringify(input,
		qs.WithStringifyCharset(qs.CharsetISO88591),
		qs.WithStringifyCharsetSentinel(true))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ a: "ä", b: "test" },
			{ charset: "iso-8859-1", charsetSentinel: true }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 44: allowDots + encodeDotInKeys + sort
func TestDotsWithEncodingAndSort(t *testing.T) {
	input := map[string]any{
		"z.key": "last",
		"a.key": "first",
		"m.key": "middle",
		"nested": map[string]any{
			"x.val": 1,
			"a.val": 2,
		},
	}

	goResult, err := qs.Stringify(input,
		qs.WithStringifyAllowDots(true),
		qs.WithEncodeDotInKeys(true),
		qs.WithSort(func(a, b string) bool { return a < b }))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const input = {
			"z.key": "last",
			"a.key": "first",
			"m.key": "middle",
			nested: { "x.val": 1, "a.val": 2 }
		};
		console.log(qs.stringify(input,
			{ allowDots: true, encodeDotInKeys: true, sort: (a, b) => a.localeCompare(b) }));
	`)

	// For sorted output, we can compare directly
	if goResult != jsResult {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 45: duplicates + parameterLimit
func TestDuplicatesWithParameterLimit(t *testing.T) {
	// Test that parameter limit is applied before duplicate handling
	goParsed, err := qs.Parse("a=1&a=2&a=3&b=4&b=5", qs.WithDuplicates(qs.DuplicateCombine), qs.WithParameterLimit(3))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `
		console.log(JSON.stringify(qs.parse("a=1&a=2&a=3&b=4&b=5",
			{ duplicates: "combine", parameterLimit: 3 })));
	`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 46: skipNulls + filter + arrayFormat
func TestSkipNullsWithFilterAndArrayFormat(t *testing.T) {
	input := map[string]any{
		"a": nil,
		"b": []any{"x", "y"},
		"c": "keep",
		"d": nil,
	}

	goResult, err := qs.Stringify(input,
		qs.WithSkipNulls(true),
		qs.WithFilter([]string{"a", "b", "c"}),
		qs.WithArrayFormat(qs.ArrayFormatBrackets))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ a: null, b: ["x", "y"], c: "keep", d: null },
			{ skipNulls: true, filter: ["a", "b", "c"], arrayFormat: "brackets" }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 47: ignoreQueryPrefix + delimiter + depth
func TestIgnorePrefixWithDelimiterAndDepth(t *testing.T) {
	goParsed, err := qs.Parse("?a[b][c]=1;d[e]=2",
		qs.WithIgnoreQueryPrefix(true),
		qs.WithDelimiter(";"),
		qs.WithDepth(1))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `
		console.log(JSON.stringify(qs.parse("?a[b][c]=1;d[e]=2",
			{ ignoreQueryPrefix: true, delimiter: ";", depth: 1 })));
	`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 48: addQueryPrefix + format RFC1738 + encode
func TestQueryPrefixWithFormatRFC1738(t *testing.T) {
	input := map[string]any{
		"search": "hello world",
		"filter": "a+b",
	}

	goResult, err := qs.Stringify(input,
		qs.WithStringifyAddQueryPrefix(true),
		qs.WithFormat(qs.FormatRFC1738))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({ search: "hello world", filter: "a+b" },
			{ addQueryPrefix: true, format: "RFC1738" }));
	`)

	if !compareQueryStrings(t, goResult, jsResult) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 49: Parse and Stringify round-trip with complex options
func TestRoundTripComplex(t *testing.T) {
	// Start with a complex input
	input := map[string]any{
		"users": []any{
			map[string]any{"name": "Alice", "active": true},
			map[string]any{"name": "Bob", "active": false},
		},
		"filters": map[string]any{
			"status": []any{"pending", "active"},
		},
		"page": 1,
	}

	// Stringify with specific options
	goStringified, err := qs.Stringify(input, qs.WithArrayFormat(qs.ArrayFormatIndices))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsStringified := runJS(t, `
		const input = {
			users: [{ name: "Alice", active: true }, { name: "Bob", active: false }],
			filters: { status: ["pending", "active"] },
			page: 1
		};
		console.log(qs.stringify(input, { arrayFormat: "indices" }));
	`)

	if !compareQueryStrings(t, goStringified, jsStringified) {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goStringified, jsStringified)
	}

	// Parse back
	goParsed, err := qs.Parse(goStringified)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse(`+"`"+jsStringified+"`"+`)));`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Round-trip parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 50: All parse options together (reasonable combination)
func TestAllParseOptionsCombined(t *testing.T) {
	goParsed, err := qs.Parse("?utf8=%E2%9C%93&a.b=1&c[0]=x&c[1]=y&d=hello%20world",
		qs.WithAllowDots(true),
		qs.WithCharset(qs.CharsetUTF8),
		qs.WithCharsetSentinel(true),
		qs.WithDepth(5),
		qs.WithIgnoreQueryPrefix(true),
		qs.WithParameterLimit(100))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	jsParsedJSON := runJS(t, `
		console.log(JSON.stringify(qs.parse("?utf8=%E2%9C%93&a.b=1&c[0]=x&c[1]=y&d=hello%20world", {
			allowDots: true,
			charset: "utf-8",
			charsetSentinel: true,
			depth: 5,
			ignoreQueryPrefix: true,
			parameterLimit: 100
		})));
	`)

	if !deepEqual(t, toJSON(t, goParsed), jsParsedJSON) {
		t.Errorf("Parse mismatch:\nGo: %s\nJS: %s", toJSON(t, goParsed), jsParsedJSON)
	}
}

// Test 51: All stringify options together (reasonable combination)
func TestAllStringifyOptionsCombined(t *testing.T) {
	input := map[string]any{
		"search":  "hello world",
		"tags":    []any{"a", "b"},
		"filters": map[string]any{"active": true},
		"empty":   nil,
	}

	goResult, err := qs.Stringify(input,
		qs.WithStringifyAddQueryPrefix(true),
		qs.WithStringifyAllowDots(true),
		qs.WithArrayFormat(qs.ArrayFormatBrackets),
		qs.WithStringifyCharset(qs.CharsetUTF8),
		qs.WithFormat(qs.FormatRFC3986),
		qs.WithSkipNulls(true),
		qs.WithSort(func(a, b string) bool { return a < b }))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		console.log(qs.stringify({
			search: "hello world",
			tags: ["a", "b"],
			filters: { active: true },
			empty: null
		}, {
			addQueryPrefix: true,
			allowDots: true,
			arrayFormat: "brackets",
			charset: "utf-8",
			format: "RFC3986",
			skipNulls: true,
			sort: (a, b) => a.localeCompare(b)
		}));
	`)

	// With sort, order should match
	if goResult != jsResult {
		t.Errorf("Stringify mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 52: Formatter via Format option
// Note: In Go, Formatter is an internal implementation detail tied to Format.
// This is already tested via TestRFCFormats. The Format option (RFC1738/RFC3986)
// determines which formatter is used internally.
// RFC1738: spaces become +
// RFC3986: spaces become %20

// Test 52: SerializeDate with time.Time
func TestSerializeDateWithTime(t *testing.T) {
	// Use ISO format for date serialization
	serializeDate := func(t time.Time) string {
		return t.Format("2006-01-02")
	}

	// Fixed date: 2024-06-15
	date := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	input := map[string]any{"date": date}

	goResult, err := qs.Stringify(input, qs.WithSerializeDate(serializeDate))
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	jsResult := runJS(t, `
		const serializeDate = (d) => d.toISOString().split('T')[0];
		console.log(qs.stringify({ date: new Date('2024-06-15T10:30:00Z') }, { serializeDate }));
	`)

	if goResult != jsResult {
		t.Errorf("SerializeDate mismatch:\nGo: %s\nJS: %s", goResult, jsResult)
	}
}

// Test 54: SerializeDate with default (RFC3339)
func TestSerializeDateDefault(t *testing.T) {
	// Test that default serialization uses RFC3339
	date := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	input := map[string]any{"date": date}

	goResult, err := qs.Stringify(input)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}

	// JS default is toISOString which is similar to RFC3339
	jsResult := runJS(t, `
		console.log(qs.stringify({ date: new Date('2024-06-15T10:30:00Z') }));
	`)

	// Compare parsed values since format might differ slightly
	goParsed, err := qs.Parse(goResult)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	jsParsedJSON := runJS(t, `console.log(JSON.stringify(qs.parse(`+"`"+jsResult+"`"+`)));`)

	goDateStr, ok := goParsed["date"].(string)
	if !ok {
		t.Fatalf("expected Go parsed date to be string, got %T", goParsed["date"])
	}
	var jsParsed map[string]any
	if err := json.Unmarshal([]byte(jsParsedJSON), &jsParsed); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	jsDateStr, ok := jsParsed["date"].(string)
	if !ok {
		t.Fatalf("expected JS parsed date to be string, got %T", jsParsed["date"])
	}

	// Both should represent the same date
	goTime, err := time.Parse(time.RFC3339Nano, goDateStr)
	if err != nil {
		t.Fatalf("failed to parse Go time %q: %v", goDateStr, err)
	}
	jsTime, err := time.Parse(time.RFC3339Nano, jsDateStr)
	if err != nil {
		t.Fatalf("failed to parse JS time %q: %v", jsDateStr, err)
	}

	if !goTime.Equal(jsTime) {
		t.Errorf("Date values differ:\nGo: %s (%v)\nJS: %s (%v)", goDateStr, goTime, jsDateStr, jsTime)
	}
}
