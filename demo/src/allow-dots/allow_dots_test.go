package allowdots

import (
	"os"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/democompare"
	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

func TestAllowDots_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	input := map[string]any{"a": map[string]any{"b": "c"}}

	t.Run("stringify default", func(t *testing.T) {
		if !strings.Contains(readme, `qs.stringify({ a: { b: "c" } }, { encode: false })`) {
			t.Fatalf("README.md missing JS example for default stringify")
		}
		goQS, err := qs.Stringify(input, qs.WithEncode(false))
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		jsQS := demojs.Run(t, `console.log(qs.stringify({ a: { b: "c" } }, { encode: false }));`)
		if goQS != jsQS {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "a[b]=c") {
			t.Fatalf("README.md missing expected output %q", "a[b]=c")
		}
	})

	t.Run("stringify allowDots", func(t *testing.T) {
		if !strings.Contains(readme, `qs.stringify({ a: { b: "c" } }, { allowDots: true, encode: false })`) {
			t.Fatalf("README.md missing JS example for allowDots stringify")
		}
		goQS, err := qs.Stringify(input, qs.WithStringifyAllowDots(true), qs.WithEncode(false))
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		jsQS := demojs.Run(t, `console.log(qs.stringify({ a: { b: "c" } }, { allowDots: true, encode: false }));`)
		if goQS != jsQS {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "a.b=c") {
			t.Fatalf("README.md missing expected output %q", "a.b=c")
		}
	})

	t.Run("parse default", func(t *testing.T) {
		if !strings.Contains(readme, `qs.parse("a.b=c")`) {
			t.Fatalf("README.md missing JS example for default parse")
		}
		goParsed, err := qs.Parse("a.b=c")
		if err != nil {
			t.Fatalf("go parse: %v", err)
		}
		goJSON := democompare.ToJSON(t, goParsed)
		jsJSON := demojs.Run(t, `console.log(JSON.stringify(qs.parse("a.b=c")));`)
		if !democompare.JSONEqual(t, goJSON, jsJSON) {
			t.Fatalf("mismatch:\nGo: %s\nJS: %s", goJSON, jsJSON)
		}
		if !strings.Contains(readme, `{"a.b":"c"}`) {
			t.Fatalf("README.md missing expected JSON %q", `{"a.b":"c"}`)
		}
	})

	t.Run("parse allowDots", func(t *testing.T) {
		if !strings.Contains(readme, `qs.parse("a.b=c", { allowDots: true })`) {
			t.Fatalf("README.md missing JS example for allowDots parse")
		}
		goParsed, err := qs.Parse("a.b=c", qs.WithAllowDots(true))
		if err != nil {
			t.Fatalf("go parse: %v", err)
		}
		goJSON := democompare.ToJSON(t, goParsed)
		jsJSON := demojs.Run(t, `console.log(JSON.stringify(qs.parse("a.b=c", { allowDots: true })));`)
		if !democompare.JSONEqual(t, goJSON, jsJSON) {
			t.Fatalf("mismatch:\nGo: %s\nJS: %s", goJSON, jsJSON)
		}
		if !strings.Contains(readme, `{"a":{"b":"c"}}`) {
			t.Fatalf("README.md missing expected JSON %q", `{"a":{"b":"c"}}`)
		}
	})
}
