package arrayformat

import (
	"os"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

func TestArrayFormat_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	input := map[string]any{"a": []any{"b", "c"}}

	cases := []struct {
		name     string
		goOpts   []qs.StringifyOption
		jsOpts   string
		expected string
	}{
		{
			name:     "default (indices)",
			goOpts:   []qs.StringifyOption{qs.WithStringifyEncode(false)},
			jsOpts:   `{ encode: false }`,
			expected: "a[0]=b&a[1]=c",
		},
		{
			name:     "brackets",
			goOpts:   []qs.StringifyOption{qs.WithStringifyArrayFormat(qs.ArrayFormatBrackets), qs.WithStringifyEncode(false)},
			jsOpts:   `{ arrayFormat: "brackets", encode: false }`,
			expected: "a[]=b&a[]=c",
		},
		{
			name:     "repeat",
			goOpts:   []qs.StringifyOption{qs.WithStringifyArrayFormat(qs.ArrayFormatRepeat), qs.WithStringifyEncode(false)},
			jsOpts:   `{ arrayFormat: "repeat", encode: false }`,
			expected: "a=b&a=c",
		},
		{
			name:     "comma",
			goOpts:   []qs.StringifyOption{qs.WithStringifyArrayFormat(qs.ArrayFormatComma), qs.WithStringifyEncode(false)},
			jsOpts:   `{ arrayFormat: "comma", encode: false }`,
			expected: "a=b,c",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(readme, tc.jsOpts) {
				t.Fatalf("README.md missing JS options snippet %q", tc.jsOpts)
			}
			goQS, err := qs.Stringify(input, tc.goOpts...)
			if err != nil {
				t.Fatalf("go stringify: %v", err)
			}
			jsQS := demojs.Run(t, `console.log(qs.stringify({ a: ["b", "c"] }, `+tc.jsOpts+`));`)
			if goQS != jsQS {
				t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
			}
			if goQS != tc.expected {
				t.Fatalf("unexpected output:\nGot:  %q\nWant: %q", goQS, tc.expected)
			}
			if !strings.Contains(readme, tc.expected) {
				t.Fatalf("README.md missing expected output %q", tc.expected)
			}
		})
	}
}
