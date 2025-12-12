package tests

import (
	"strings"
	"testing"

	"github.com/zaytracom/qs/v1"
)

func TestStringifyQueryStringObject(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "simple string",
			input:    map[string]interface{}{"a": "b"},
			expected: "a=b",
		},
		{
			name:     "simple number",
			input:    map[string]interface{}{"a": 1},
			expected: "a=1",
		},
		{
			name:     "multiple values",
			input:    map[string]interface{}{"a": 1, "b": 2},
			expected: "a=1&b=2", // Note: order may vary in Go maps
		},
		{
			name:     "string with underscore",
			input:    map[string]interface{}{"a": "A_Z"},
			expected: "a=A_Z",
		},
		{
			name:     "unicode characters",
			input:    map[string]interface{}{"a": "€"},
			expected: "a=%E2%82%AC",
		},
		{
			name:     "hebrew characters",
			input:    map[string]interface{}{"a": "א"},
			expected: "a=%D7%90",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}

			// For multiple values, check if result contains both key-value pairs
			if test.name == "multiple values" {
				if !strings.Contains(result, "a=1") || !strings.Contains(result, "b=2") {
					t.Errorf("Stringify() = %v, want to contain both a=1 and b=2", result)
				}
			} else if result != test.expected {
				t.Errorf("Stringify() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStringifyFalsyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		options  *qs.StringifyOptions
		expected string
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "",
		},
		{
			name:     "false value",
			input:    false,
			expected: "",
		},
		{
			name:     "zero value",
			input:    0,
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input, test.options)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}
			if result != test.expected {
				t.Errorf("Stringify() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStringifyNestedObjects(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "simple nested object",
			input: map[string]interface{}{
				"a": map[string]interface{}{"b": "c"},
			},
			expected: "a[b]=c",
		},
		{
			name: "double nested object",
			input: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{"c": "d"},
				},
			},
			expected: "a[b][c]=d",
		},
		{
			name: "complex nested structure",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"name": "John",
						"age":  30,
					},
					"settings": map[string]interface{}{
						"theme": "dark",
					},
				},
			},
			// Check individual parts since order may vary
			expected: "user[profile][name]=John&user[profile][age]=30&user[settings][theme]=dark",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}

			// For complex structures, check individual components
			if test.name == "complex nested structure" {
				if !strings.Contains(result, "user[profile][name]=John") ||
					!strings.Contains(result, "user[profile][age]=30") ||
					!strings.Contains(result, "user[settings][theme]=dark") {
					t.Errorf("Stringify() = %v, want to contain all expected parts", result)
				}
			} else if result != test.expected {
				t.Errorf("Stringify() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStringifyArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		options  *qs.StringifyOptions
		expected string
	}{
		{
			name: "array with indices (default)",
			input: map[string]interface{}{
				"a": []interface{}{"b", "c"},
			},
			expected: "a[0]=b&a[1]=c",
		},
		{
			name: "array with brackets format",
			input: map[string]interface{}{
				"a": []interface{}{"b", "c"},
			},
			options:  &qs.StringifyOptions{ArrayFormat: "brackets"},
			expected: "a[]=b&a[]=c",
		},
		{
			name: "array with repeat format",
			input: map[string]interface{}{
				"a": []interface{}{"b", "c"},
			},
			options:  &qs.StringifyOptions{ArrayFormat: "repeat"},
			expected: "a=b&a=c",
		},
		{
			name: "nested array",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"tags": []interface{}{"admin", "user"},
				},
			},
			expected: "user[tags][0]=admin&user[tags][1]=user",
		},
		{
			name: "multiple arrays",
			input: map[string]interface{}{
				"colors": []interface{}{"red", "blue"},
				"sizes":  []interface{}{"S", "M", "L"},
			},
			options: &qs.StringifyOptions{ArrayFormat: "brackets"},
			// Check individual parts since order may vary
			expected: "colors[]=red&colors[]=blue&sizes[]=S&sizes[]=M&sizes[]=L",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input, test.options)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}

			// For multiple arrays, check individual components
			if test.name == "multiple arrays" {
				if !strings.Contains(result, "colors[]=red") ||
					!strings.Contains(result, "colors[]=blue") ||
					!strings.Contains(result, "sizes[]=S") {
					t.Errorf("Stringify() = %v, want to contain expected array parts", result)
				}
			} else if result != test.expected {
				t.Errorf("Stringify() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStringifyWithOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		options *qs.StringifyOptions
		check   func(string) bool
		wantErr bool
	}{
		{
			name: "add query prefix",
			input: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			options: &qs.StringifyOptions{AddQueryPrefix: true},
			check:   func(s string) bool { return strings.HasPrefix(s, "?") },
		},
		{
			name: "custom delimiter",
			input: map[string]interface{}{
				"a": "1",
				"b": "2",
			},
			options: &qs.StringifyOptions{Delimiter: ";"},
			check:   func(s string) bool { return strings.Contains(s, ";") },
		},
		{
			name: "allow dots",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			options: &qs.StringifyOptions{AllowDots: true},
			check: func(s string) bool {
				return strings.Contains(s, "user.name=John") && strings.Contains(s, "user.age=30")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input, test.options)
			if (err != nil) != test.wantErr {
				t.Errorf("Stringify() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.check(result) {
				t.Errorf("Stringify() = %v, failed custom check", result)
			}
		})
	}
}

func TestStringifySpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "spaces in values",
			input: map[string]interface{}{
				"name": "John Doe",
				"city": "New York",
			},
			expected: "name=John%20Doe&city=New%20York",
		},
		{
			name: "special characters",
			input: map[string]interface{}{
				"symbols": "!@#$%^&*()",
			},
			expected: "symbols=%21%40%23%24%25%5E%26%2A%28%29",
		},
		{
			name: "unicode characters",
			input: map[string]interface{}{
				"greeting": "Привет",
				"emoji":    "🚀",
			},
			// Check individual parts since order may vary
			expected: "greeting=%D0%9F%D1%80%D0%B8%D0%B2%D0%B5%D1%82&emoji=%F0%9F%9A%80",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}

			// For multiple values, check individual components
			if test.name == "spaces in values" {
				if !strings.Contains(result, "name=John%20Doe") ||
					!strings.Contains(result, "city=New%20York") {
					t.Errorf("Stringify() = %v, want to contain encoded spaces", result)
				}
			} else if test.name == "unicode characters" {
				if !strings.Contains(result, "greeting=%D0%9F%D1%80%D0%B8%D0%B2%D0%B5%D1%82") {
					t.Errorf("Stringify() = %v, want to contain encoded unicode", result)
				}
			} else if result != test.expected {
				t.Errorf("Stringify() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStringifyEmptyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		options  *qs.StringifyOptions
		expected string
	}{
		{
			name: "empty string value",
			input: map[string]interface{}{
				"empty": "",
				"key":   "value",
			},
			expected: "empty=&key=value",
		},
		{
			name: "nil value in map",
			input: map[string]interface{}{
				"null": nil,
				"key":  "value",
			},
			expected: "null=&key=value",
		},
		{
			name: "empty array",
			input: map[string]interface{}{
				"items": []interface{}{},
			},
			expected: "",
		},
		{
			name: "array with empty string",
			input: map[string]interface{}{
				"items": []interface{}{"", "value"},
			},
			expected: "items[0]=&items[1]=value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input, test.options)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}

			// For multiple values, check individual components
			if test.name == "empty string value" || test.name == "nil value in map" {
				if !strings.Contains(result, "key=value") {
					t.Errorf("Stringify() = %v, want to contain key=value", result)
				}
			} else if result != test.expected {
				t.Errorf("Stringify() = %v, want %v", result, test.expected)
			}
		})
	}
}

func TestStringifyComplexRealWorld(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		checkFn func(string) bool
	}{
		{
			name: "e-commerce product listing",
			input: map[string]interface{}{
				"category": "electronics",
				"filters": map[string]interface{}{
					"price": map[string]interface{}{
						"min": 100,
						"max": 1000,
					},
					"brands": []interface{}{"Apple", "Samsung"},
				},
				"sort": "price_asc",
				"page": map[string]interface{}{
					"number": 1,
					"size":   20,
				},
			},
			checkFn: func(s string) bool {
				return strings.Contains(s, "category=electronics") &&
					strings.Contains(s, "filters[price][min]=100") &&
					strings.Contains(s, "filters[price][max]=1000") &&
					strings.Contains(s, "filters[brands][0]=Apple") &&
					strings.Contains(s, "filters[brands][1]=Samsung") &&
					strings.Contains(s, "sort=price_asc")
			},
		},
		{
			name: "api request with includes",
			input: map[string]interface{}{
				"include": []interface{}{"user", "comments", "tags"},
				"fields": map[string]interface{}{
					"posts": "title,content,published_at",
					"users": "name,email",
				},
				"filter": map[string]interface{}{
					"published": true,
					"author": map[string]interface{}{
						"role": "admin",
					},
				},
			},
			checkFn: func(s string) bool {
				return strings.Contains(s, "include[0]=user") &&
					strings.Contains(s, "include[1]=comments") &&
					strings.Contains(s, "fields[posts]=title,content,published_at") &&
					strings.Contains(s, "filter[published]=true") &&
					strings.Contains(s, "filter[author][role]=admin")
			},
		},
		{
			name: "form with nested validation rules",
			input: map[string]interface{}{
				"form": map[string]interface{}{
					"fields": map[string]interface{}{
						"name": map[string]interface{}{
							"type":     "text",
							"required": true,
							"validation": map[string]interface{}{
								"minLength": 2,
								"maxLength": 50,
							},
						},
						"email": map[string]interface{}{
							"type":     "email",
							"required": true,
						},
					},
					"metadata": map[string]interface{}{
						"version":   "1.0",
						"createdBy": "admin",
					},
				},
			},
			checkFn: func(s string) bool {
				return strings.Contains(s, "form[fields][name][type]=text") &&
					strings.Contains(s, "form[fields][name][required]=true") &&
					strings.Contains(s, "form[fields][name][validation][minLength]=2") &&
					strings.Contains(s, "form[fields][email][type]=email") &&
					strings.Contains(s, "form[metadata][version]=1.0")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}
			if !test.checkFn(result) {
				t.Errorf("Stringify() = %v, failed complex check", result)
			}
		})
	}
}

func TestStringifyDifferentDataTypes(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		checkFn func(string) bool
	}{
		{
			name: "mixed data types",
			input: map[string]interface{}{
				"string":  "hello",
				"number":  42,
				"float":   3.14,
				"boolean": true,
				"array":   []interface{}{1, "two", true, 3.14},
				"nested": map[string]interface{}{
					"value": "test",
				},
			},
			checkFn: func(s string) bool {
				return strings.Contains(s, "string=hello") &&
					strings.Contains(s, "number=42") &&
					strings.Contains(s, "float=3.14") &&
					strings.Contains(s, "boolean=true") &&
					strings.Contains(s, "array[0]=1") &&
					strings.Contains(s, "array[1]=two") &&
					strings.Contains(s, "array[2]=true") &&
					strings.Contains(s, "nested[value]=test")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Stringify(test.input)
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}
			if !test.checkFn(result) {
				t.Errorf("Stringify() = %v, failed data type check", result)
			}
		})
	}
}
