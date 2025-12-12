package addqueryprefix

import (
	"os"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

func TestAddQueryPrefix_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	t.Run("default", func(t *testing.T) {
		if !strings.Contains(readme, `qs.stringify({ a: "b" }, { encode: false })`) {
			t.Fatalf("README.md missing JS example for default stringify")
		}
		goQS, err := qs.Stringify(map[string]any{"a": "b"}, qs.WithEncode(false))
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		jsQS := demojs.Run(t, `console.log(qs.stringify({ a: "b" }, { encode: false }));`)
		if goQS != jsQS {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "a=b") {
			t.Fatalf("README.md missing expected output %q", "a=b")
		}
	})

	t.Run("with addQueryPrefix", func(t *testing.T) {
		if !strings.Contains(readme, `addQueryPrefix: true`) {
			t.Fatalf("README.md missing JS example for addQueryPrefix")
		}
		goQS, err := qs.Stringify(
			map[string]any{"a": "b"},
			qs.WithStringifyAddQueryPrefix(true),
			qs.WithEncode(false),
		)
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		jsQS := demojs.Run(t, `console.log(qs.stringify({ a: "b" }, { addQueryPrefix: true, encode: false }));`)
		if goQS != jsQS {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "?a=b") {
			t.Fatalf("README.md missing expected output %q", "?a=b")
		}
	})
}
