package ignorequeryprefix

import (
	"os"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/democompare"
	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

func TestIgnoreQueryPrefix_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	t.Run("default", func(t *testing.T) {
		if !strings.Contains(readme, `qs.parse("?a=b&c=d")`) {
			t.Fatalf("README.md missing JS example for default parse")
		}
		goParsed, err := qs.Parse("?a=b&c=d")
		if err != nil {
			t.Fatalf("go parse: %v", err)
		}
		goJSON := democompare.ToJSON(t, goParsed)
		jsJSON := demojs.Run(t, `console.log(JSON.stringify(qs.parse("?a=b&c=d")));`)
		if !democompare.JSONEqual(t, goJSON, jsJSON) {
			t.Fatalf("mismatch:\nGo: %s\nJS: %s", goJSON, jsJSON)
		}
		if !strings.Contains(readme, `{"?a":"b","c":"d"}`) {
			t.Fatalf("README.md missing expected JSON %q", `{"?a":"b","c":"d"}`)
		}
	})

	t.Run("with ignoreQueryPrefix", func(t *testing.T) {
		if !strings.Contains(readme, `ignoreQueryPrefix: true`) {
			t.Fatalf("README.md missing JS example for ignoreQueryPrefix")
		}
		goParsed, err := qs.Parse("?a=b&c=d", qs.WithParseIgnoreQueryPrefix(true))
		if err != nil {
			t.Fatalf("go parse: %v", err)
		}
		goJSON := democompare.ToJSON(t, goParsed)
		jsJSON := demojs.Run(t, `console.log(JSON.stringify(qs.parse("?a=b&c=d", { ignoreQueryPrefix: true })));`)
		if !democompare.JSONEqual(t, goJSON, jsJSON) {
			t.Fatalf("mismatch:\nGo: %s\nJS: %s", goJSON, jsJSON)
		}
		if !strings.Contains(readme, `{"a":"b","c":"d"}`) {
			t.Fatalf("README.md missing expected JSON %q", `{"a":"b","c":"d"}`)
		}
	})
}
