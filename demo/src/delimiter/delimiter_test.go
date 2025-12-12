package delimiter

import (
	"os"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/democompare"
	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

func TestDelimiter_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	t.Run("stringify delimiter ;", func(t *testing.T) {
		if !strings.Contains(readme, `delimiter: ";", encode: false`) {
			t.Fatalf("README.md missing JS example for delimiter stringify")
		}
		goQS, err := qs.Stringify(
			map[string]any{"a": "1", "b": "2"},
			qs.WithStringifyDelimiter(";"),
			qs.WithStringifyEncode(false),
			qs.WithStringifySort(func(a, b string) bool { return a < b }),
		)
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		jsQS := demojs.Run(t, `
			console.log(qs.stringify(
				{ a: "1", b: "2" },
				{ delimiter: ";", encode: false, sort: (a, b) => a.localeCompare(b) }
			));
		`)
		if goQS != jsQS {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "a=1;b=2") {
			t.Fatalf("README.md missing expected output %q", "a=1;b=2")
		}
	})

	t.Run("parse delimiter ;", func(t *testing.T) {
		if !strings.Contains(readme, `qs.parse("a=1;b=2", { delimiter: ";" })`) {
			t.Fatalf("README.md missing JS example for delimiter parse")
		}
		goParsed, err := qs.Parse("a=1;b=2", qs.WithParseDelimiter(";"))
		if err != nil {
			t.Fatalf("go parse: %v", err)
		}
		goJSON := democompare.ToJSON(t, goParsed)
		jsJSON := demojs.Run(t, `console.log(JSON.stringify(qs.parse("a=1;b=2", { delimiter: ";" })));`)
		if !democompare.JSONEqual(t, goJSON, jsJSON) {
			t.Fatalf("mismatch:\nGo: %s\nJS: %s", goJSON, jsJSON)
		}
		if !strings.Contains(readme, `{"a":"1","b":"2"}`) {
			t.Fatalf("README.md missing expected JSON %q", `{"a":"1","b":"2"}`)
		}
	})
}
