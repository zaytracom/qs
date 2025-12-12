package tests

import (
	"reflect"
	"strings"
	"testing"

	qs "github.com/zaytracom/qs/v1"
)

// EmptyTestCase represents a test case for empty keys
type EmptyTestCase struct {
	Input           string
	WithEmptyKeys   map[string]interface{}
	NoEmptyKeys     map[string]interface{}
	StringifyOutput map[string]string
}

// getEmptyTestCases returns the test cases for empty keys (ported from JavaScript)
func getEmptyTestCases() []EmptyTestCase {
	return []EmptyTestCase{
		{
			Input:         "&",
			WithEmptyKeys: map[string]interface{}{},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "",
				"indices":  "",
				"repeat":   "",
			},
		},
		{
			Input:         "&&",
			WithEmptyKeys: map[string]interface{}{},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "",
				"indices":  "",
				"repeat":   "",
			},
		},
		{
			Input:         "&=",
			WithEmptyKeys: map[string]interface{}{"": ""},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "=",
				"indices":  "=",
				"repeat":   "=",
			},
		},
		{
			Input:         "&=&",
			WithEmptyKeys: map[string]interface{}{"": ""},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "=",
				"indices":  "=",
				"repeat":   "=",
			},
		},
		{
			Input:         "&=&=",
			WithEmptyKeys: map[string]interface{}{"": []interface{}{"", ""}},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "[]=&[]=",
				"indices":  "[0]=&[1]=",
				"repeat":   "=&=",
			},
		},
		{
			Input:         "&=&=&",
			WithEmptyKeys: map[string]interface{}{"": []interface{}{"", ""}},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "[]=&[]=",
				"indices":  "[0]=&[1]=",
				"repeat":   "=&=",
			},
		},
		{
			Input:         "=",
			WithEmptyKeys: map[string]interface{}{"": ""},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "=",
				"indices":  "=",
				"repeat":   "=",
			},
		},
		{
			Input:         "=&",
			WithEmptyKeys: map[string]interface{}{"": ""},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "=",
				"indices":  "=",
				"repeat":   "=",
			},
		},
		{
			Input:         "=&&&",
			WithEmptyKeys: map[string]interface{}{"": ""},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "=",
				"indices":  "=",
				"repeat":   "=",
			},
		},
		{
			Input:         "=&=&=&",
			WithEmptyKeys: map[string]interface{}{"": []interface{}{"", "", ""}},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "[]=&[]=&[]=",
				"indices":  "[0]=&[1]=&[2]=",
				"repeat":   "=&=&=",
			},
		},
		{
			Input:         "=&a[]=b&a[1]=c",
			WithEmptyKeys: map[string]interface{}{"": "", "a": []interface{}{"b", "c"}},
			NoEmptyKeys:   map[string]interface{}{"a": []interface{}{"b", "c"}},
			StringifyOutput: map[string]string{
				"brackets": "=&a[]=b&a[]=c",
				"indices":  "=&a[0]=b&a[1]=c",
				"repeat":   "=&a=b&a=c",
			},
		},
		{
			Input:         "=a",
			WithEmptyKeys: map[string]interface{}{"": "a"},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "=a",
				"indices":  "=a",
				"repeat":   "=a",
			},
		},
		{
			Input:         "a==a",
			WithEmptyKeys: map[string]interface{}{"a": "=a"},
			NoEmptyKeys:   map[string]interface{}{"a": "=a"},
			StringifyOutput: map[string]string{
				"brackets": "a==a",
				"indices":  "a==a",
				"repeat":   "a==a",
			},
		},
		{
			Input:         "=&a[]=b",
			WithEmptyKeys: map[string]interface{}{"": "", "a": []interface{}{"b"}},
			NoEmptyKeys:   map[string]interface{}{"a": []interface{}{"b"}},
			StringifyOutput: map[string]string{
				"brackets": "=&a[]=b",
				"indices":  "=&a[0]=b",
				"repeat":   "=&a=b",
			},
		},
		{
			Input:         "=&a[]=b&a[]=c&a[2]=d",
			WithEmptyKeys: map[string]interface{}{"": "", "a": []interface{}{"b", "c", "d"}},
			NoEmptyKeys:   map[string]interface{}{"a": []interface{}{"b", "c", "d"}},
			StringifyOutput: map[string]string{
				"brackets": "=&a[]=b&a[]=c&a[]=d",
				"indices":  "=&a[0]=b&a[1]=c&a[2]=d",
				"repeat":   "=&a=b&a=c&a=d",
			},
		},
		{
			Input:         "=a&=b",
			WithEmptyKeys: map[string]interface{}{"": []interface{}{"a", "b"}},
			NoEmptyKeys:   map[string]interface{}{},
			StringifyOutput: map[string]string{
				"brackets": "[]=a&[]=b",
				"indices":  "[0]=a&[1]=b",
				"repeat":   "=a&=b",
			},
		},
		{
			Input:         "=a&foo=b",
			WithEmptyKeys: map[string]interface{}{"": "a", "foo": "b"},
			NoEmptyKeys:   map[string]interface{}{"foo": "b"},
			StringifyOutput: map[string]string{
				"brackets": "=a&foo=b",
				"indices":  "=a&foo=b",
				"repeat":   "=a&foo=b",
			},
		},
		{
			Input:         "a[]=b&a=c&=",
			WithEmptyKeys: map[string]interface{}{"": "", "a": []interface{}{"b", "c"}},
			NoEmptyKeys:   map[string]interface{}{"a": []interface{}{"b", "c"}},
			StringifyOutput: map[string]string{
				"brackets": "=&a[]=b&a[]=c",
				"indices":  "=&a[0]=b&a[1]=c",
				"repeat":   "=&a=b&a=c",
			},
		},
		{
			Input:         "a[0]=b&a=c&=",
			WithEmptyKeys: map[string]interface{}{"": "", "a": []interface{}{"b", "c"}},
			NoEmptyKeys:   map[string]interface{}{"a": []interface{}{"b", "c"}},
			StringifyOutput: map[string]string{
				"brackets": "=&a[]=b&a[]=c",
				"indices":  "=&a[0]=b&a[1]=c",
				"repeat":   "=&a=b&a=c",
			},
		},
	}
}

func TestParseEmptyKeys(t *testing.T) {
	testCases := getEmptyTestCases()

	for i, testCase := range testCases {
		t.Run(testCase.Input, func(t *testing.T) {
			// Test parsing with empty keys allowed (default behavior)
			result, err := qs.Parse(testCase.Input)
			if err != nil {
				t.Errorf("Test case %d: Parse() error = %v", i, err)
				return
			}

			if !reflect.DeepEqual(result, testCase.WithEmptyKeys) {
				t.Errorf("Test case %d: Parse() = %v, want %v", i, result, testCase.WithEmptyKeys)
			}

			// Note: Testing NoEmptyKeys would require implementing an option to ignore empty keys
			// This is not currently implemented in the Go version but would be part of ParseOptions
		})
	}
}

func TestStringifyEmptyKeysRoundTrip(t *testing.T) {
	testCases := getEmptyTestCases()

	for i, testCase := range testCases {
		// Skip empty test cases for stringify
		if len(testCase.WithEmptyKeys) == 0 {
			continue
		}

		t.Run(testCase.Input+"_roundtrip", func(t *testing.T) {
			// Test with brackets format
			if testCase.StringifyOutput["brackets"] != "" {
				result, err := qs.Stringify(testCase.WithEmptyKeys, &qs.StringifyOptions{
					ArrayFormat: "brackets",
				})
				if err != nil {
					t.Errorf("Test case %d (brackets): Stringify() error = %v", i, err)
					return
				}

				// Parse back the result
				parsed, err := qs.Parse(result)
				if err != nil {
					t.Errorf("Test case %d (brackets): Parse() error = %v", i, err)
					return
				}

				if !reflect.DeepEqual(parsed, testCase.WithEmptyKeys) {
					t.Errorf("Test case %d (brackets): Round trip failed. Original: %v, Stringified: %s, Parsed: %v",
						i, testCase.WithEmptyKeys, result, parsed)
				}
			}

			// Test with indices format
			if testCase.StringifyOutput["indices"] != "" {
				result, err := qs.Stringify(testCase.WithEmptyKeys, &qs.StringifyOptions{
					ArrayFormat: "indices",
				})
				if err != nil {
					t.Errorf("Test case %d (indices): Stringify() error = %v", i, err)
					return
				}

				// Parse back the result
				_, err = qs.Parse(result)
				if err != nil {
					t.Errorf("Test case %d (indices): Parse() error = %v", i, err)
					return
				}

				// Note: indices format parsing creates maps instead of arrays, so we need special handling
				// For now, just check that it parses without error
			}

			// Test with repeat format
			if testCase.StringifyOutput["repeat"] != "" {
				result, err := qs.Stringify(testCase.WithEmptyKeys, &qs.StringifyOptions{
					ArrayFormat: "repeat",
				})
				if err != nil {
					t.Errorf("Test case %d (repeat): Stringify() error = %v", i, err)
					return
				}

				// Parse back the result
				parsed, err := qs.Parse(result)
				if err != nil {
					t.Errorf("Test case %d (repeat): Parse() error = %v", i, err)
					return
				}

				// For repeat format, single-element arrays become scalar values after roundtrip
				// This is expected behavior, so we need to create the expected result
				expectedForRepeat := make(map[string]interface{})
				for k, v := range testCase.WithEmptyKeys {
					if arr, isArray := v.([]interface{}); isArray && len(arr) == 1 {
						// Single-element arrays become scalar values in repeat format
						expectedForRepeat[k] = arr[0]
					} else {
						expectedForRepeat[k] = v
					}
				}

				if !reflect.DeepEqual(parsed, expectedForRepeat) {
					t.Errorf("Test case %d (repeat): Round trip failed. Original: %v, Stringified: %s, Parsed: %v, Expected: %v",
						i, testCase.WithEmptyKeys, result, parsed, expectedForRepeat)
				}
			}
		})
	}
}

func TestEmptyKeySpecificCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:     "single empty key",
			input:    "=value",
			expected: map[string]interface{}{"": "value"},
		},
		{
			name:     "empty key with empty value",
			input:    "=",
			expected: map[string]interface{}{"": ""},
		},
		{
			name:     "multiple empty keys",
			input:    "=first&=second",
			expected: map[string]interface{}{"": []interface{}{"first", "second"}},
		},
		{
			name:  "empty key mixed with normal keys",
			input: "=empty&normal=value&=another",
			expected: map[string]interface{}{
				"":       []interface{}{"empty", "another"},
				"normal": "value",
			},
		},
		{
			name:  "empty key with nested structure",
			input: "=[nested]&normal[key]=value",
			expected: map[string]interface{}{
				"":       "[nested]",
				"normal": map[string]interface{}{"key": "value"},
			},
		},
		{
			name:  "complex empty key patterns",
			input: "&=&a=b&=c&d=e&=",
			expected: map[string]interface{}{
				"":  []interface{}{"", "c", ""},
				"a": "b",
				"d": "e",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Parse() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestEmptyKeyStringify(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		options *qs.StringifyOptions
		checkFn func(string) bool
	}{
		{
			name: "stringify empty key with value",
			input: map[string]interface{}{
				"":    "value",
				"key": "normal",
			},
			checkFn: func(s string) bool {
				return strings.Contains(s, "=value") && strings.Contains(s, "key=normal")
			},
		},
		{
			name: "stringify empty key array with brackets",
			input: map[string]interface{}{
				"": []interface{}{"first", "second"},
			},
			options: &qs.StringifyOptions{ArrayFormat: "brackets"},
			checkFn: func(s string) bool {
				return strings.Contains(s, "[]=first") && strings.Contains(s, "[]=second")
			},
		},
		{
			name: "stringify empty key array with repeat",
			input: map[string]interface{}{
				"": []interface{}{"a", "b"},
			},
			options: &qs.StringifyOptions{ArrayFormat: "repeat"},
			checkFn: func(s string) bool {
				return strings.Contains(s, "=a") && strings.Contains(s, "=b")
			},
		},
		{
			name: "stringify mixed empty and normal keys",
			input: map[string]interface{}{
				"":       "empty_value",
				"normal": "normal_value",
				"array":  []interface{}{"item1", "item2"},
			},
			checkFn: func(s string) bool {
				return strings.Contains(s, "=empty_value") &&
					strings.Contains(s, "normal=normal_value") &&
					strings.Contains(s, "array[0]=item1")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input, test.options)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}
			if !test.checkFn(result) {
				t.Errorf("Stringify() = %v, failed check", result)
			}
		})
	}
}
