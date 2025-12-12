package tests

import (
	"reflect"
	"testing"

	"github.com/zaytracom/qs/v1"
)

func TestParseSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  *qs.ParseOptions
		expected map[string]interface{}
	}{
		{
			name:     "parses simple string with number key",
			input:    "0=foo",
			expected: map[string]interface{}{"0": "foo"},
		},
		{
			name:     "parses string with plus signs",
			input:    "foo=c++",
			expected: map[string]interface{}{"foo": "c  "},
		},
		{
			name:     "parses string with special characters in key",
			input:    "a[>=]=23",
			expected: map[string]interface{}{"a": map[string]interface{}{">=": "23"}},
		},
		{
			name:     "parses string with multiple equals",
			input:    "a[<=>]==23",
			expected: map[string]interface{}{"a": map[string]interface{}{"<=>": "=23"}},
		},
		{
			name:     "parses string with double equals in key",
			input:    "a[==]=23",
			expected: map[string]interface{}{"a": map[string]interface{}{"==": "23"}},
		},
		{
			name:     "parses key without value with strict null handling",
			input:    "foo",
			options:  &qs.ParseOptions{StrictNullHandling: true},
			expected: map[string]interface{}{"foo": nil},
		},
		{
			name:     "parses key without value",
			input:    "foo",
			expected: map[string]interface{}{"foo": ""},
		},
		{
			name:     "parses key with empty value",
			input:    "foo=",
			expected: map[string]interface{}{"foo": ""},
		},
		{
			name:     "parses simple key-value pair",
			input:    "foo=bar",
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "parses string with spaces",
			input:    " foo = bar = baz ",
			expected: map[string]interface{}{" foo ": " bar = baz "},
		},
		{
			name:     "parses value with equals",
			input:    "foo=bar=baz",
			expected: map[string]interface{}{"foo": "bar=baz"},
		},
		{
			name:     "parses multiple key-value pairs",
			input:    "foo=bar&bar=baz",
			expected: map[string]interface{}{"foo": "bar", "bar": "baz"},
		},
		{
			name:     "parses with empty value",
			input:    "foo2=bar2&baz2=",
			expected: map[string]interface{}{"foo2": "bar2", "baz2": ""},
		},
		{
			name:     "parses with key without value and strict null handling",
			input:    "foo=bar&baz",
			options:  &qs.ParseOptions{StrictNullHandling: true},
			expected: map[string]interface{}{"foo": "bar", "baz": nil},
		},
		{
			name:     "parses with key without value",
			input:    "foo=bar&baz",
			expected: map[string]interface{}{"foo": "bar", "baz": ""},
		},
		{
			name:  "parses complex query string",
			input: "cht=p3&chd=t:60,40&chs=250x100&chl=Hello|World",
			expected: map[string]interface{}{
				"cht": "p3",
				"chd": "t:60,40",
				"chs": "250x100",
				"chl": "Hello|World",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, test.options)
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

func TestParseArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  *qs.ParseOptions
		expected map[string]interface{}
	}{
		{
			name:     "comma: false - array brackets",
			input:    "a[]=b&a[]=c",
			expected: map[string]interface{}{"a": []interface{}{"b", "c"}},
		},
		{
			name:     "comma: false - array indices",
			input:    "a[0]=b&a[1]=c",
			expected: map[string]interface{}{"a": map[string]interface{}{"0": "b", "1": "c"}},
		},
		{
			name:     "comma: false - comma in value",
			input:    "a=b,c",
			expected: map[string]interface{}{"a": "b,c"},
		},
		{
			name:     "comma: false - multiple values",
			input:    "a=b&a=c",
			expected: map[string]interface{}{"a": []interface{}{"b", "c"}},
		},
		// Note: comma: true tests would require implementation of comma support
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, test.options)
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

func TestParseAllowDots(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  *qs.ParseOptions
		expected map[string]interface{}
	}{
		{
			name:     "dot notation disabled",
			input:    "a.b=c",
			expected: map[string]interface{}{"a.b": "c"},
		},
		{
			name:    "dot notation enabled",
			input:   "a.b=c",
			options: &qs.ParseOptions{AllowDots: true},
			expected: map[string]interface{}{
				"a": map[string]interface{}{"b": "c"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, test.options)
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

func TestParseNestedObjects(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "single nested string",
			input: "a[b]=c",
			expected: map[string]interface{}{
				"a": map[string]interface{}{"b": "c"},
			},
		},
		{
			name:  "double nested string",
			input: "a[b][c]=d",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{"c": "d"},
				},
			},
		},
		{
			name:  "deeply nested with limit",
			input: "a[b][c][d][e][f][g][h]=i",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"d": map[string]interface{}{
								"e": map[string]interface{}{
									"f": map[string]interface{}{
										"[g][h]": "i", // Beyond depth limit
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "nested array",
			input: "a[b][]=c&a[b][]=d",
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": []interface{}{"c", "d"},
				},
			},
		},
		{
			name:  "complex nested structure",
			input: "user[name][first]=John&user[name][last]=Doe&user[age]=30",
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"name": map[string]interface{}{
						"first": "John",
						"last":  "Doe",
					},
					"age": "30",
				},
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

func TestParseEmptyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  *qs.ParseOptions
		expected map[string]interface{}
	}{
		{
			name:     "empty key with value",
			input:    "=value",
			expected: map[string]interface{}{"": "value"},
		},
		{
			name:     "empty key empty value",
			input:    "=",
			expected: map[string]interface{}{"": ""},
		},
		{
			name:     "key with empty brackets",
			input:    "key[]=",
			expected: map[string]interface{}{"key": []interface{}{""}},
		},
		{
			name:     "multiple empty values",
			input:    "a=&b=&c=value",
			expected: map[string]interface{}{"a": "", "b": "", "c": "value"},
		},
		{
			name:     "empty keys with strict null handling",
			input:    "key&empty=",
			options:  &qs.ParseOptions{StrictNullHandling: true},
			expected: map[string]interface{}{"key": nil, "empty": ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, test.options)
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

func TestParseURLEncoded(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:     "URL encoded spaces",
			input:    "name=John%20Doe&city=New%20York",
			expected: map[string]interface{}{"name": "John Doe", "city": "New York"},
		},
		{
			name:     "URL encoded special characters",
			input:    "symbols=%21%40%23%24%25%5E%26%2A%28%29",
			expected: map[string]interface{}{"symbols": "!@#$%^&*()"},
		},
		{
			name:     "URL encoded unicode",
			input:    "unicode=%D0%9F%D1%80%D0%B8%D0%B2%D0%B5%D1%82",
			expected: map[string]interface{}{"unicode": "Привет"},
		},
		{
			name:     "URL encoded plus signs",
			input:    "query=hello+world&lang=c%2B%2B",
			expected: map[string]interface{}{"query": "hello world", "lang": "c++"},
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

func TestParseCustomDelimiter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  *qs.ParseOptions
		expected map[string]interface{}
	}{
		{
			name:     "semicolon delimiter",
			input:    "a=1;b=2;c=3",
			options:  &qs.ParseOptions{Delimiter: ";"},
			expected: map[string]interface{}{"a": "1", "b": "2", "c": "3"},
		},
		{
			name:     "comma delimiter",
			input:    "x=foo,y=bar,z=baz",
			options:  &qs.ParseOptions{Delimiter: ","},
			expected: map[string]interface{}{"x": "foo", "y": "bar", "z": "baz"},
		},
		{
			name:     "pipe delimiter",
			input:    "name=John|age=30|city=NYC",
			options:  &qs.ParseOptions{Delimiter: "|"},
			expected: map[string]interface{}{"name": "John", "age": "30", "city": "NYC"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, test.options)
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

func TestParseQueryPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  *qs.ParseOptions
		expected map[string]interface{}
	}{
		{
			name:     "ignore query prefix enabled",
			input:    "?name=John&age=30",
			options:  &qs.ParseOptions{IgnoreQueryPrefix: true},
			expected: map[string]interface{}{"name": "John", "age": "30"},
		},
		{
			name:     "ignore query prefix disabled",
			input:    "?name=John&age=30",
			expected: map[string]interface{}{"?name": "John", "age": "30"},
		},
		{
			name:     "multiple question marks",
			input:    "??name=John&age=30",
			options:  &qs.ParseOptions{IgnoreQueryPrefix: true},
			expected: map[string]interface{}{"?name": "John", "age": "30"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := qs.Parse(test.input, test.options)
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

func TestParseComplexRealWorld(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]interface{}
	}{
		{
			name:  "e-commerce filter",
			input: "category=electronics&price[min]=100&price[max]=500&brands[]=Apple&brands[]=Samsung&sort=price_asc",
			expected: map[string]interface{}{
				"category": "electronics",
				"price": map[string]interface{}{
					"min": "100",
					"max": "500",
				},
				"brands": []interface{}{"Apple", "Samsung"},
				"sort":   "price_asc",
			},
		},
		{
			name:  "api query with includes",
			input: "include[]=user&include[]=comments&fields[posts]=title,content&fields[users]=name,email&page[number]=1&page[size]=10",
			expected: map[string]interface{}{
				"include": []interface{}{"user", "comments"},
				"fields": map[string]interface{}{
					"posts": "title,content",
					"users": "name,email",
				},
				"page": map[string]interface{}{
					"number": "1",
					"size":   "10",
				},
			},
		},
		{
			name:  "form with nested validation",
			input: "user[profile][name]=John&user[profile][email]=john@example.com&user[settings][theme]=dark&user[settings][notifications][email]=true&user[settings][notifications][push]=false",
			expected: map[string]interface{}{
				"user": map[string]interface{}{
					"profile": map[string]interface{}{
						"name":  "John",
						"email": "john@example.com",
					},
					"settings": map[string]interface{}{
						"theme": "dark",
						"notifications": map[string]interface{}{
							"email": "true",
							"push":  "false",
						},
					},
				},
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
