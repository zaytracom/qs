package sortarrayindices

import (
	"os"
	"strings"
	"testing"

	"github.com/zaytracom/qs/demo/internal/demojs"
	qs "github.com/zaytracom/qs/v2"
)

func TestSortArrayIndices_ReadmeExamples(t *testing.T) {
	readmeBytes, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(readmeBytes)

	// 12-element array: indices 0-11, string sort gives 0,1,10,11,2,3,4,5,6,7,8,9
	input := map[string]any{"arr": []any{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}}
	sortAsc := func(a, b string) bool { return a < b }

	t.Run("JS with sort", func(t *testing.T) {
		jsCode := `console.log(qs.stringify({ arr: ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"] }, { encode: false, sort: (a, b) => a.localeCompare(b) }));`
		if !strings.Contains(readme, `qs.stringify({ arr: ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"] }, { encode: false, sort: (a, b) => a.localeCompare(b) })`) {
			t.Fatalf("README.md missing JS example for sort")
		}
		jsQS := demojs.Run(t, jsCode)
		expectedJS := "arr[0]=a&arr[1]=b&arr[10]=k&arr[11]=l&arr[2]=c&arr[3]=d&arr[4]=e&arr[5]=f&arr[6]=g&arr[7]=h&arr[8]=i&arr[9]=j"
		if jsQS != expectedJS {
			t.Fatalf("JS output mismatch:\nGot:      %q\nExpected: %q", jsQS, expectedJS)
		}
		if !strings.Contains(readme, expectedJS) {
			t.Fatalf("README.md missing expected JS output %q", expectedJS)
		}
	})

	t.Run("Go with Sort (default - numeric order)", func(t *testing.T) {
		goQS, err := qs.Stringify(input, qs.WithEncode(false), qs.WithSort(sortAsc))
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		expectedGo := "arr[0]=a&arr[1]=b&arr[2]=c&arr[3]=d&arr[4]=e&arr[5]=f&arr[6]=g&arr[7]=h&arr[8]=i&arr[9]=j&arr[10]=k&arr[11]=l"
		if goQS != expectedGo {
			t.Fatalf("Go output mismatch:\nGot:      %q\nExpected: %q", goQS, expectedGo)
		}
		if !strings.Contains(readme, expectedGo) {
			t.Fatalf("README.md missing expected Go output %q", expectedGo)
		}
	})

	t.Run("Go with Sort and SortArrayIndices (matches JS)", func(t *testing.T) {
		goQS, err := qs.Stringify(input, qs.WithEncode(false), qs.WithSort(sortAsc), qs.WithSortArrayIndices(true))
		if err != nil {
			t.Fatalf("go stringify: %v", err)
		}
		// Should match JS output exactly
		jsCode := `console.log(qs.stringify({ arr: ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"] }, { encode: false, sort: (a, b) => a.localeCompare(b) }));`
		jsQS := demojs.Run(t, jsCode)
		if goQS != jsQS {
			t.Fatalf("Go with SortArrayIndices should match JS:\nGo: %q\nJS: %q", goQS, jsQS)
		}
		expectedOutput := "arr[0]=a&arr[1]=b&arr[10]=k&arr[11]=l&arr[2]=c&arr[3]=d&arr[4]=e&arr[5]=f&arr[6]=g&arr[7]=h&arr[8]=i&arr[9]=j"
		if !strings.Contains(readme, expectedOutput) {
			t.Fatalf("README.md missing expected output for SortArrayIndices")
		}
	})
}
