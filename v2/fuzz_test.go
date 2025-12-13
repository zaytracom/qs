package qs

import (
	"hash/crc32"
	"testing"
)

// FuzzParse is a safety fuzz test:
// - Parse must never panic or hang on arbitrary input.
// - If Parse succeeds, Stringify and Parse(Stringify(...)) should not panic or error.
func FuzzParse(f *testing.F) {
	// Seed corpus: common and pathological inputs.
	f.Add("")
	f.Add("a=b")
	f.Add("a=")
	f.Add("a")
	f.Add("a[b]=c")
	f.Add("a[b][0]=c&a[b][1]=d")
	f.Add("a[]=b&a[]=c")
	f.Add("a=b&a=c&a=d")
	f.Add("?a=b&c=d")
	f.Add("a.b=c")
	f.Add("a%5Bb%5D=c") // a[b]=c
	f.Add("a=%E2%98%BA") // UTF-8
	f.Add("a=%26%2310003%3B") // numeric entity
	f.Add("a=hello+world")
	f.Add("a=foo%26bar")
	f.Add("a[b=1") // unclosed bracket
	f.Add("a%")
	f.Add("a=%ZZ")
	f.Add("a=1;b=2") // delimiter variant

	f.Fuzz(func(t *testing.T, input string) {
		// Derive a deterministic set of options from the input.
		// This increases coverage without making the fuzzer non-reproducible.
		h := crc32.ChecksumIEEE([]byte(input))

		opts := []ParseOption{
			WithParseIgnoreQueryPrefix(h&(1<<0) != 0),
			WithParseAllowDots(h&(1<<1) != 0),
			WithParseAllowEmptyArrays(h&(1<<2) != 0),
			WithParseAllowSparse(h&(1<<3) != 0),
			WithParseStrictNullHandling(h&(1<<4) != 0),
		}
		if h&(1<<5) != 0 {
			opts = append(opts, WithParseComma(true))
		}
		if h&(1<<6) != 0 {
			opts = append(opts, WithParseDuplicates(DuplicateCombine))
		} else if h&(1<<7) != 0 {
			opts = append(opts, WithParseDuplicates(DuplicateFirst))
		} else {
			opts = append(opts, WithParseDuplicates(DuplicateLast))
		}

		parsed, err := Parse(input, opts...)
		if err != nil {
			return
		}

		// If parsing succeeded, stringification should be safe and reversible.
		qs1, err := Stringify(parsed, WithStringifySort(func(a, b string) bool { return a < b }))
		if err != nil {
			t.Fatalf("Stringify(Parse(%q)) failed: %v", input, err)
		}
		if _, err := Parse(qs1, opts...); err != nil {
			t.Fatalf("Parse(Stringify(Parse(%q))) failed: %v\nQS: %q", input, err, qs1)
		}
	})
}

// FuzzStringify is a safety fuzz test:
// - Stringify must never panic or hang on arbitrary (but valid) inputs.
// We keep the fuzz surface small and deterministic by constructing an input value
// from a fuzzed query string via Parse.
func FuzzStringify(f *testing.F) {
	f.Add("a=b")
	f.Add("a[b]=c&d[e]=f")
	f.Add("a=1&a=2&a=3")
	f.Add("arr[0]=x&arr[10]=y&arr[2]=z")

	f.Fuzz(func(t *testing.T, input string) {
		parsed, err := Parse(input,
			WithParseAllowDots(true),
			WithParseAllowSparse(true),
			WithParseAllowEmptyArrays(true),
			WithParseStrictNullHandling(true),
			WithParseIgnoreQueryPrefix(true),
		)
		if err != nil {
			return
		}

		_, err = Stringify(parsed,
			WithStringifySort(func(a, b string) bool { return a < b }),
			WithStringifySortArrayIndices(true),
		)
		if err != nil {
			t.Fatalf("Stringify failed after Parse(%q): %v", input, err)
		}
	})
}
