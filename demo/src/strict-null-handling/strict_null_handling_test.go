package strictnullhandling

import (
	"os"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/democompare"
	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

func TestStrictNullHandling_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	t.Run("stringify default", func(t *testing.T) {
		if !strings.Contains(readme, `qs.stringify({ a: null }, { encode: false })`) {
			t.Fatalf("README.md missing JS example for default stringify")
		}
		goQS, err := qs.Stringify(map[string]any{"a": nil}, qs.WithStringifyEncode(false))
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		jsQS := demojs.Run(t, `console.log(qs.stringify({ a: null }, { encode: false }));`)
		if goQS != jsQS {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if goQS != "a=" {
			t.Fatalf("unexpected output: %q", goQS)
		}
		if !strings.Contains(readme, "a=") {
			t.Fatalf("README.md missing expected output %q", "a=")
		}
	})

	t.Run("stringify strictNullHandling", func(t *testing.T) {
		if !strings.Contains(readme, `strictNullHandling: true`) {
			t.Fatalf("README.md missing JS example for strictNullHandling stringify")
		}
		goQS, err := qs.Stringify(
			map[string]any{"a": nil},
			qs.WithStringifyStrictNullHandling(true),
			qs.WithStringifyEncode(false),
		)
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		jsQS := demojs.Run(t, `console.log(qs.stringify({ a: null }, { strictNullHandling: true, encode: false }));`)
		if goQS != jsQS {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if goQS != "a" {
			t.Fatalf("unexpected output: %q", goQS)
		}
		if !strings.Contains(readme, "\n// a\n") && !strings.Contains(readme, "\n// a\r\n") {
			t.Fatalf("README.md missing expected output %q", "a")
		}
	})

	t.Run("parse default", func(t *testing.T) {
		if !strings.Contains(readme, `qs.parse("a&b=")`) {
			t.Fatalf("README.md missing JS example for default parse")
		}
		goParsed, err := qs.Parse("a&b=")
		if err != nil {
			t.Fatalf("go parse: %v", err)
		}
		goJSON := democompare.ToJSON(t, goParsed)
		jsJSON := demojs.Run(t, `console.log(JSON.stringify(qs.parse("a&b=")));`)
		if !democompare.JSONEqual(t, goJSON, jsJSON) {
			t.Fatalf("mismatch:\nGo: %s\nJS: %s", goJSON, jsJSON)
		}
		if !strings.Contains(readme, `{"a":"","b":""}`) {
			t.Fatalf("README.md missing expected JSON %q", `{"a":"","b":""}`)
		}
	})

	t.Run("parse strictNullHandling", func(t *testing.T) {
		if !strings.Contains(readme, `qs.parse("a&b=", { strictNullHandling: true })`) {
			t.Fatalf("README.md missing JS example for strictNullHandling parse")
		}
		goParsed, err := qs.Parse("a&b=", qs.WithParseStrictNullHandling(true))
		if err != nil {
			t.Fatalf("go parse: %v", err)
		}
		goJSON := democompare.ToJSON(t, goParsed)
		jsJSON := demojs.Run(t, `console.log(JSON.stringify(qs.parse("a&b=", { strictNullHandling: true })));`)
		if !democompare.JSONEqual(t, goJSON, jsJSON) {
			t.Fatalf("mismatch:\nGo: %s\nJS: %s", goJSON, jsJSON)
		}
		if !strings.Contains(readme, `{"a":null,"b":""}`) {
			t.Fatalf("README.md missing expected JSON %q", `{"a":null,"b":""}`)
		}
	})
}
