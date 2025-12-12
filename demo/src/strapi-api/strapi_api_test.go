package strapiapi

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

func parseQueryString(s string) map[string][]string {
	result := make(map[string][]string)
	if s == "" {
		return result
	}
	s = strings.TrimPrefix(s, "?")
	parts := strings.Split(s, "&")
	for _, part := range parts {
		idx := strings.Index(part, "=")
		var key, val string
		if idx == -1 {
			key = part
			val = ""
		} else {
			key = part[:idx]
			val = part[idx+1:]
		}
		result[key] = append(result[key], val)
	}
	for k := range result {
		sort.Strings(result[k])
	}
	return result
}

func queryStringsEqual(goQS, jsQS string) bool {
	return reflect.DeepEqual(parseQueryString(goQS), parseQueryString(jsQS))
}

func TestStrapiAPI_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	t.Run("filtering with operators", func(t *testing.T) {
		if !strings.Contains(readme, `$contains`) {
			t.Fatalf("README.md missing $contains operator example")
		}

		goQS, err := qs.Stringify(
			map[string]any{
				"filters": map[string]any{
					"title":     map[string]any{"$contains": "hello"},
					"createdAt": map[string]any{"$gte": "2023-01-01"},
				},
			},
			qs.WithEncodeValuesOnly(true),
		)
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}

		jsQS := demojs.Run(t, `
			console.log(qs.stringify(
				{
					filters: {
						title: { $contains: "hello" },
						createdAt: { $gte: "2023-01-01" },
					},
				},
				{ encodeValuesOnly: true }
			));
		`)

		if !queryStringsEqual(goQS, jsQS) {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "filters[title][$contains]=hello") || !strings.Contains(readme, "filters[createdAt][$gte]=2023-01-01") {
			t.Fatalf("README.md missing expected output for filtering")
		}
	})

	t.Run("sorting", func(t *testing.T) {
		if !strings.Contains(readme, `sort: ["title:asc", "createdAt:desc"]`) {
			t.Fatalf("README.md missing sorting example")
		}

		goQS, err := qs.Stringify(
			map[string]any{"sort": []any{"title:asc", "createdAt:desc"}},
			qs.WithEncodeValuesOnly(true),
		)
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}

		jsQS := demojs.Run(t, `
			console.log(qs.stringify({ sort: ["title:asc", "createdAt:desc"] }, { encodeValuesOnly: true }));
		`)

		if !queryStringsEqual(goQS, jsQS) {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "sort[0]=title:asc") || !strings.Contains(readme, "sort[1]=createdAt:desc") {
			t.Fatalf("README.md missing expected output for sorting")
		}
	})

	t.Run("pagination", func(t *testing.T) {
		if !strings.Contains(readme, `pagination: { page: 1, pageSize: 10 }`) {
			t.Fatalf("README.md missing pagination example")
		}

			goQS, err := qs.Stringify(
				map[string]any{"pagination": map[string]any{"page": 1, "pageSize": 10}},
				qs.WithEncodeValuesOnly(true),
				qs.WithSort(func(a, b string) bool { return a < b }),
			)
			if err != nil {
				t.Fatalf("go stringify: %v", err)
			}

		jsQS := demojs.Run(t, `
			console.log(qs.stringify(
				{ pagination: { page: 1, pageSize: 10 } },
				{ encodeValuesOnly: true }
			));
		`)

		if !queryStringsEqual(goQS, jsQS) {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "pagination[page]=1") || !strings.Contains(readme, "pagination[pageSize]=10") {
			t.Fatalf("README.md missing expected output for pagination")
		}
	})

	t.Run("population", func(t *testing.T) {
		if !strings.Contains(readme, `populate:`) {
			t.Fatalf("README.md missing populate example")
		}

		goQS, err := qs.Stringify(
			map[string]any{
				"populate": map[string]any{
					"author":     map[string]any{"fields": []any{"name", "email"}},
					"categories": map[string]any{"fields": []any{"name"}},
				},
			},
			qs.WithEncodeValuesOnly(true),
		)
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}

		jsQS := demojs.Run(t, `
			console.log(qs.stringify(
				{
					populate: {
						author: { fields: ["name", "email"] },
						categories: { fields: ["name"] },
					},
				},
				{ encodeValuesOnly: true }
			));
		`)

		if !queryStringsEqual(goQS, jsQS) {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "populate[author][fields][0]=name") ||
			!strings.Contains(readme, "populate[author][fields][1]=email") ||
			!strings.Contains(readme, "populate[categories][fields][0]=name") {
			t.Fatalf("README.md missing expected output for population")
		}
	})

	t.Run("complete query", func(t *testing.T) {
		goQS, err := qs.Stringify(
			map[string]any{
				"filters":    map[string]any{"status": map[string]any{"$eq": "published"}},
				"sort":       []any{"createdAt:desc"},
				"pagination": map[string]any{"page": 1, "pageSize": 25},
				"populate":   map[string]any{"author": map[string]any{"fields": []any{"name"}}},
			},
			qs.WithEncodeValuesOnly(true),
			qs.WithSort(func(a, b string) bool { return a < b }),
		)
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}

		jsQS := demojs.Run(t, `
			console.log(qs.stringify(
				{
					filters: { status: { $eq: "published" } },
					sort: ["createdAt:desc"],
					pagination: { page: 1, pageSize: 25 },
					populate: { author: { fields: ["name"] } },
				},
				{ encodeValuesOnly: true, sort: (a, b) => a.localeCompare(b) }
			));
		`)

		if !queryStringsEqual(goQS, jsQS) {
			t.Fatalf("mismatch:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		if !strings.Contains(readme, "filters[status][$eq]=published") ||
			!strings.Contains(readme, "pagination[page]=1") ||
			!strings.Contains(readme, "pagination[pageSize]=25") ||
			!strings.Contains(readme, "populate[author][fields][0]=name") ||
			!strings.Contains(readme, "sort[0]=createdAt:desc") {
			t.Fatalf("README.md missing expected output for complete query")
		}
	})
}
