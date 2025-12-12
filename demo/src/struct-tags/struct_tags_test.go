package structtags

import (
	"os"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/democompare"
	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

type Filters struct {
	Status string   `query:"status" json:"status"`
	Tags   []string `query:"tags" json:"tags"`
}

type Query struct {
	Filters Filters `query:"filters" json:"filters"`
	Page    int     `query:"page" json:"page"`
}

func TestStructTags_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	const queryString = "filters.status=published&filters.tags[]=go&filters.tags[]=qs&page=2"
	const expectedJSON = `{"filters":{"status":"published","tags":["go","qs"]},"page":2}`

	t.Run("marshal parity with JS", func(t *testing.T) {
		if !strings.Contains(readme, "`query:\"filters\"") {
			t.Fatalf("README.md missing struct tags example")
		}
		in := Query{
			Filters: Filters{Status: "published", Tags: []string{"go", "qs"}},
			Page:    2,
		}

		goQS, err := qs.Marshal(
			in,
			qs.WithStringifyAllowDots(true),
			qs.WithArrayFormat(qs.ArrayFormatBrackets),
			qs.WithEncode(false),
			qs.WithSort(func(a, b string) bool { return a < b }),
		)
		if err != nil {
			t.Fatalf("go marshal: %v", err)
		}

		jsQS := demojs.Run(t, `
			console.log(qs.stringify(
				{ filters: { status: "published", tags: ["go", "qs"] }, page: 2 },
				{
					allowDots: true,
					arrayFormat: "brackets",
					encode: false,
					sort: (a, b) => a.localeCompare(b),
				}
			));
		`)

		if goQS != jsQS {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if goQS != queryString {
			t.Fatalf("unexpected output:\nGot:  %q\nWant: %q", goQS, queryString)
		}
		if !strings.Contains(readme, queryString) {
			t.Fatalf("README.md missing expected output %q", queryString)
		}
	})

	t.Run("unmarshal parity with JS parse", func(t *testing.T) {
		var out Query
		if err := qs.Unmarshal(queryString, &out, qs.WithAllowDots(true)); err != nil {
			t.Fatalf("go unmarshal: %v", err)
		}
		if out.Filters.Status != "published" || out.Page != 2 || len(out.Filters.Tags) != 2 || out.Filters.Tags[0] != "go" || out.Filters.Tags[1] != "qs" {
			t.Fatalf("unexpected struct: %#v", out)
		}

		goJSON := democompare.ToJSON(t, out)
		jsJSON := demojs.Run(t, `
			const parsed = qs.parse("filters.status=published&filters.tags[]=go&filters.tags[]=qs&page=2", { allowDots: true });
			parsed.page = Number(parsed.page);
			console.log(JSON.stringify(parsed));
		`)

		if !democompare.JSONEqual(t, goJSON, jsJSON) {
			t.Fatalf("mismatch:\nGo: %s\nJS: %s", goJSON, jsJSON)
		}
		if !strings.Contains(readme, expectedJSON) {
			t.Fatalf("README.md missing expected JSON %q", expectedJSON)
		}
	})
}
