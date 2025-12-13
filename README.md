# QS — Query String library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/zaytracom/qs/v2.svg)](https://pkg.go.dev/github.com/zaytracom/qs/v2)
[![CI](https://github.com/zaytracom/qs/actions/workflows/ci.yml/badge.svg)](https://github.com/zaytracom/qs/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zaytracom/qs/v2)](https://goreportcard.com/report/github.com/zaytracom/qs/v2)
[![codecov](https://codecov.io/gh/zaytracom/qs/branch/main/graph/badge.svg)](https://codecov.io/gh/zaytracom/qs)
[![GitHub release](https://img.shields.io/github/v/release/zaytracom/qs?include_prereleases)](https://github.com/zaytracom/qs/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Full-featured Go port of the popular JavaScript [`qs`](https://github.com/ljharb/qs) library — parse and stringify URL query strings with nested objects, arrays, and all the tricky edge cases.

## ✨ Features

- 🔍 **Parse** query strings into nested Go values (`map[string]any`, `[]any`) — see `qs.Parse` below.
- 📝 **Stringify** Go values into query strings (arrays, nested objects, filters/sort) — see `qs.Stringify` below.
- 🧩 **JS `qs` compatibility** — validated via the JS compatibility test suite.
- 🏷️ **Struct API** via `query` tags (`Marshal` / `Unmarshal`).
- 🎯 **Array formats**: indices, brackets, repeat, comma.
- ⚙️ **Limits + charset**: depth controls, UTF-8/ISO-8859-1, charset sentinel.
- 📋 **Encoding formats**: RFC 1738 / RFC 3986.

## Installation

```bash
go get github.com/zaytracom/qs/v2
```

## 🚀 Quick Start

### Parse query string into nested map

```go
import "github.com/zaytracom/qs/v2"

query := "user[name]=John&user[address][city]=NYC&tags[0]=go&tags[1]=qs"

result, _ := qs.Parse(query)
// map[string]any{
//   "user": map[string]any{
//     "name": "John",
//     "address": map[string]any{"city": "NYC"},
//   },
//   "tags": []any{"go", "qs"},
// }
```

### Stringify nested map into query string

```go
data := map[string]any{
    "filters": map[string]any{
        "status": "active",
        "role":   "admin",
    },
    "sort": []string{"name", "created_at"},
}

query, _ := qs.Stringify(data)
// "filters[status]=active&filters[role]=admin&sort[0]=name&sort[1]=created_at"
```

### Marshal/Unmarshal with structs

```go
type Request struct {
    Page   int      `query:"page"`
    Limit  int      `query:"limit"`
    Tags   []string `query:"tags"`
}

// Struct to query string
req := Request{Page: 1, Limit: 10, Tags: []string{"go", "qs"}}
query, _ := qs.Marshal(req)
// "page=1&limit=10&tags[0]=go&tags[1]=qs"

// Query string to struct
var parsed Request
qs.Unmarshal("page=2&limit=20&tags[0]=rust", &parsed)
// Request{Page: 2, Limit: 20, Tags: []string{"rust"}}
```

## Comparison with Other QS Libraries

| Feature | **qs (this repo)** | [gorilla/schema](https://github.com/gorilla/schema) | [google/go-querystring](https://github.com/google/go-querystring) | [go-playground/form](https://github.com/go-playground/form) | [ajg/form](https://github.com/ajg/form) |
|---|:---:|:---:|:---:|:---:|:---:|
| Encode struct | ✅ | ✅ | ✅ | ✅ | ✅ |
| Decode struct | ✅ | ✅ | ❌ | ✅ | ✅ |
| Struct tags | ✅ | ✅ | ✅ | ✅ | ✅ |
| Nested objects/arrays | ✅ | ❌ | ✅ | ✅ | ✅ |
| Nested `map[string]any` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Multiple array formats | ✅ | ❌ | ✅ | ✅ | ✅ |
| Depth/limits controls | ✅ | ❌ | ❌ | ❌ | ❌ |
| Charset sentinel + ISO-8859-1 | ✅ | ❌ | ❌ | ❌ | ❌ |
| Strict syntax validation | ✅ | ❌ | ❌ | ❌ | ❌ |

### Array formats (supported)

`qs` supports 4 array formats for both stringify and parse:

- **Indices**: `a[0]=x&a[1]=y` — explicit positions, good when order/indices matter (can represent sparse arrays).
- **Brackets**: `a[]=x&a[]=y` — no indices, order is the parameter order (common in HTML forms).
- **Repeat**: `a=x&a=y` — repeated key, simplest/most interoperable; semantics depend on how duplicates are handled.
- **Comma**: `a=x,y` — compact single value; requires comma-splitting on parse and can be ambiguous if elements contain commas.

## JS `qs` option compatibility

This library is a Go port of JS `qs`, so most options map 1:1. The table below highlights what exists on both sides and where Go differs.

### Parse options

| Option | JS `qs` | Go `qs` | Notes |
|:--|:--:|:--:|:--|
| `AllowDots` | ✅ | ✅ | Dot notation (`a.b=c`) [Read more](demo/src/allow-dots/README.md) |
| `AllowEmptyArrays` | ✅ | ✅ | `a[]=` creates `[]` vs `[""]` [Read more](demo/src/allow-empty-arrays/README.md) |
| `AllowSparse` | ✅ | ✅ | Preserve array gaps [Read more](demo/src/allow-sparse/README.md) |
| `ArrayLimit` | ✅ | ✅ | Max index for array notation [Read more](demo/src/array-limit/README.md) |
| `Charset` | ✅ | ✅ | `utf-8` / `iso-8859-1` [Read more](demo/src/charset/README.md) |
| `CharsetSentinel` | ✅ | ✅ | `utf8=✓` detection [Read more](demo/src/charset-sentinel/README.md) |
| `Comma` | ✅ | ✅ | `a=1,2` → array [Read more](demo/src/comma/README.md) |
| `DecodeDotInKeys` | ✅ | ✅ | `%2E` → `.` in keys [Read more](demo/src/decode-dot-in-keys/README.md) |
| `Decoder` | ✅ | ✅ | Custom decoder hook [Read more](demo/src/decoder/README.md) |
| `Delimiter` / regexp delimiter | ✅ | ✅ | Go supports `Delimiter` + `DelimiterRegexp` [Read more](demo/src/delimiter/README.md), [more](demo/src/delimiter-regexp/README.md) |
| `Depth` | ✅ | ✅ | Max nesting depth [Read more](demo/src/depth/README.md) |
| `Duplicates` | ✅ | ✅ | `combine` / `first` / `last` [Read more](demo/src/duplicates/README.md) |
| `IgnoreQueryPrefix` | ✅ | ✅ | Strip leading `?` [Read more](demo/src/ignore-query-prefix/README.md) |
| `InterpretNumericEntities` | ✅ | ✅ | ISO-8859-1 numeric entities [Read more](demo/src/interpret-numeric-entities/README.md) |
| `ParameterLimit` | ✅ | ✅ | Max number of params [Read more](demo/src/parameter-limit/README.md) |
| `ParseArrays` | ✅ | ✅ | Disable bracket parsing when false [Read more](demo/src/parse-arrays/README.md) |
| `StrictDepth` | ✅ | ✅ | Error when depth exceeded [Read more](demo/src/strict-depth/README.md) |
| `StrictNullHandling` | ✅ | ✅ | `a` → `null` vs `""` [Read more](demo/src/strict-null-handling/README.md) |
| `ThrowOnLimitExceeded` | ✅ | ✅ | Error on `ParameterLimit` / `ArrayLimit` [Read more](demo/src/throw-on-limit-exceeded/README.md) |
| `AllowPrototypes` / `PlainObjects` | ✅ | N/A | JS-only prototype pollution controls; in Go keys like `__proto__`, `constructor`, `prototype` are treated as normal map keys [Read more](demo/src/allow-prototypes-plain-objects/README.md) |
| `StrictMode` | ❌ | ✅ | Go-only: strict syntax validation (unmatched brackets, invalid percent-encoding, etc.) [Read more](demo/src/strict-mode/README.md) |

### Stringify options

| Option | JS `qs` | Go `qs` | Notes |
|:--|:--:|:--:|:--|
| `AddQueryPrefix` | ✅ | ✅ | Leading `?` [Read more](demo/src/add-query-prefix/README.md) |
| `AllowDots` | ✅ | ✅ | Dot output instead of brackets [Read more](demo/src/allow-dots/README.md) |
| `AllowEmptyArrays` | ✅ | ✅ | Output `key[]` for empty arrays [Read more](demo/src/allow-empty-arrays/README.md) |
| `ArrayFormat` | ✅ | ✅ | `indices` / `brackets` / `repeat` / `comma` [Read more](demo/src/array-format/README.md) |
| `Charset` | ✅ | ✅ | Output charset [Read more](demo/src/charset/README.md) |
| `CharsetSentinel` | ✅ | ✅ | Add `utf8=✓` [Read more](demo/src/charset-sentinel/README.md) |
| `CommaRoundTrip` | ✅ | ✅ | `comma` single-element round-trip [Read more](demo/src/comma-round-trip/README.md) |
| `Delimiter` | ✅ | ✅ | Join delimiter [Read more](demo/src/delimiter/README.md) |
| `Encode` | ✅ | ✅ | Percent-encode output [Read more](demo/src/encode/README.md) |
| `EncodeDotInKeys` | ✅ | ✅ | `.` → `%2E` in keys [Read more](demo/src/encode-dot-in-keys/README.md) |
| `Encoder` | ✅ | ✅ | Custom encoder hook [Read more](demo/src/encoder/README.md) |
| `EncodeValuesOnly` | ✅ | ✅ | Only encode values [Read more](demo/src/encode-values-only/README.md) |
| `Filter` | ✅ | ✅ | Function or allowlist [Read more](demo/src/filter/README.md) |
| `Format` | ✅ | ✅ | RFC1738 / RFC3986 [Read more](demo/src/format/README.md) |
| `SerializeDate` | ✅ | ✅ | Date formatting hook [Read more](demo/src/serialize-date/README.md) |
| `SkipNulls` | ✅ | ✅ | Drop null keys [Read more](demo/src/skip-nulls/README.md) |
| `Sort` | ✅ | ✅ | Custom key ordering [Read more](demo/src/sort/README.md) |
| `StrictNullHandling` | ✅ | ✅ | `null` → `a` vs `a=` [Read more](demo/src/strict-null-handling/README.md) |

### 🔥 Go-only extensions

| Feature | Go `qs` |
|:--|:--:|
| Struct API | ✅ (`Marshal` / `Unmarshal`, `query` tags) [Read more](demo/src/struct-tags/README.md) |
| `[]byte` decode API | ✅ (`UnmarshalBytes`) [Read more](demo/src/unmarshal-bytes/README.md) |
| `SortArrayIndices` | ✅ (matches JS key sorting behavior for array indices) [Read more](demo/src/sort-array-indices/README.md) |

## Parser architecture (Arena-backed, O(n))

Under the hood, `v2` uses a small lexer/parser in `v2/lang` that tokenizes the query string in a single pass and builds an arena-backed AST of `Span`s (offset/len views into the original input). This design keeps the hot path allocation-free in steady state when you reuse a `lang.Arena` (and is fully zero-copy via `ParseBytes`), while correctly handling the tricky `qs` key syntax (deeply nested brackets, percent-encoded `[`/`]` and dots, and `=` inside bracketed segments) and enabling strict error reporting for malformed keys (unmatched/unclosed brackets, invalid percent-encoding, etc.). It was added to make JS `qs`-compat parsing both fast and predictable for complex real-world keys; for the full grammar/AST details see `v2/LANGUAGE_SPECIFICATION.md`.

## Performance

Benchmarks on `darwin/arm64` (`go test -bench=. -benchmem` in `benchmarks/`). Lower is better.

### Stringify (encode)

Time / allocs (`μs/op`, `B/op`, `allocs/op`):

| Case | **qs (this repo)** | [go-playground/form](https://github.com/go-playground/form) | [gorilla/schema](https://github.com/gorilla/schema) | [google/go-querystring](https://github.com/google/go-querystring) | [ajg/form](https://github.com/ajg/form) |
|:--|--:|--:|--:|--:|--:|
| Simple struct | **0.10 / 208 / 2** | 0.34 / 485 / 10 | 0.67 / 256 / 14 | 1.00 / 656 / 20 | 1.48 / 1120 / 23 |
| Nested struct (`a[b]=x`) | **0.10 / 224 / 2** | 0.41 / 528 / 10 | — | 1.39 / 776 / 30 | 2.63 / 2072 / 41 |
| Array struct (`a[0]=x`) | **0.10 / 184 / 2** | 0.51 / 724 / 15 | — | 0.85 / 816 / 20 | 2.16 / 1472 / 32 |
| Giant dynamic map (`map[string]any`) | **16.70 / 18206 / 351** | — | — | — | — |

### Parse / Unmarshal (decode)

Time / allocs (`μs/op`, `B/op`, `allocs/op`). Benchmarks use raw query string input for all libs (includes `url.ParseQuery` overhead where applicable):

| Case | **qs (this repo)** | [go-playground/form](https://github.com/go-playground/form) | [gorilla/schema](https://github.com/gorilla/schema) | [google/go-querystring](https://github.com/google/go-querystring) | [ajg/form](https://github.com/ajg/form) |
|:--|--:|--:|--:|--:|--:|
| Simple struct | 1.22 / 1496 / 31 | **0.53 / 528 / 8** | 2.13 / 872 / 45 | — | 2.98 / 1024 / 37 |
| Nested struct (native format) | 1.92 / 1808 / 45 | **0.83 / 528 / 7** | — | — | 3.92 / 1736 / 40 |
| Array struct (native format) | 2.12 / 2176 / 49 | **0.88 / 848 / 14** | 1.43 / 1208 / 30 | — | 3.64 / 1507 / 38 |
| Dynamic map (`qs.Parse`) | 33.57 / 46119 / 678 | — | — | — | —

“Native format” = each library’s own nesting/array notation; `qs` uses JS `qs`-style brackets/indices, others may use dot or repeated keys.

## Documentation

- [Go Reference (v2)](https://pkg.go.dev/github.com/zaytracom/qs/v2)
- [GitHub Issues](https://github.com/zaytracom/qs/issues)

## Contributing

PRs and issue reports are welcome — see `CONTRIBUTING.md`.

## Code of Conduct

This project follows the Contributor Covenant — see `CODE_OF_CONDUCT.md`.

## License

Apache 2.0 — see [LICENSE](LICENSE)
