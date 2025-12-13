# QS — Query String library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/zaytracom/qs/v2.svg)](https://pkg.go.dev/github.com/zaytracom/qs/v2)
[![CI](https://github.com/zaytracom/qs/actions/workflows/ci.yml/badge.svg)](https://github.com/zaytracom/qs/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zaytracom/qs/v2)](https://goreportcard.com/report/github.com/zaytracom/qs/v2)
[![codecov](https://codecov.io/gh/zaytracom/qs/branch/main/graph/badge.svg)](https://codecov.io/gh/zaytracom/qs)
[![GitHub release](https://img.shields.io/github/v/release/zaytracom/qs?include_prereleases)](https://github.com/zaytracom/qs/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Full-featured Go port of the popular JavaScript [`qs`](https://github.com/ljharb/qs) library — parse and stringify URL query strings with nested objects, arrays, and all the tricky edge cases.

## Table of Contents

- [Features](#-features)
- [Installation](#installation)
- [Quick Start](#-quick-start)
- [Real-World Example: Strapi API](#-real-world-example-strapi-api)
- [Comparison with Other Libraries](#comparison-with-other-qs-libraries)
- [Array Formats](#array-formats-supported)
- [JS `qs` Option Compatibility](#js-qs-option-compatibility)
  - [Parse Options](#parse-options)
  - [Stringify Options](#stringify-options)
  - [Go-only Extensions](#-go-only-extensions)
- [Parser Architecture](#parser-architecture-arena-backed-on)
- [Performance](#performance)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## ✨ Features

- 🔍 **Parse** query strings into nested Go values (`map[string]any`, `[]any`) — see `qs.Parse` below.
- 📝 **Stringify** Go values into query strings (arrays, nested objects, filters/sort) — see `qs.Stringify` below.
- 🌳 **Any nesting depth** — structs, dynamic maps (`map[string]any`), slices, and other JSON-like Go values of arbitrary complexity.
- 🧩 **JS `qs` compatibility** — validated via the JS compatibility test suite.
- 🏷️ **Struct API** via `query` tags (`Marshal` / `Unmarshal`).
- 🎯 **Array formats**: indices, brackets, repeat, comma.
- ⚙️ **Limits + charset**: depth controls, UTF-8/ISO-8859-1, charset sentinel.
- 📋 **Encoding formats**: RFC 1738 / RFC 3986.
- ⚡ **Fast** — arena-backed parser, minimal allocations, single-pass processing.
- 🧪 **Well-tested** — high test coverage, JS compatibility test suite, [benchmarks](#performance).

### When should you use `qs`?

Use `qs` if you need:

* Designing APIs that accept **large, dynamic, deeply nested query parameters**
* Supporting **nested filters and deep object syntax** in URLs (e.g. `filter[price][gte]=100`)
* Parsing query strings into **`map[string]any`**, not only predefined structs
* Full **JavaScript `qs` compatibility** (Strapi, Keystone, JSON:API-style or custom APIs)
* **Strict, predictable parsing and validation** of complex query strings

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

## 🌐 Real-World Example: Strapi-style APIs (server + client)

Many APIs (Strapi, Keystone, and similar) use JavaScript `qs`-style query strings for filtering, sorting, pagination, and population.
`qs` helps in both directions:

- When you're building a **client**, you can generate complex GET query strings reliably.
- When you're designing an **API**, you can parse the same query syntax into a nested `map[string]any` and handle it consistently.

```go
import (
    "net/http"

    "github.com/zaytracom/qs/v2"
)

// Client side: build a Strapi-style query.
query, _ := qs.Stringify(map[string]any{
    "filters": map[string]any{
        "status":      map[string]any{"$eq": "published"},
        "publishedAt": map[string]any{"$gte": "2024-01-01"},
        "author": map[string]any{
            "verified": map[string]any{"$eq": true},
        },
    },
    "sort": []any{"publishedAt:desc", "title:asc"},
    "pagination": map[string]any{
        "page":     1,
        "pageSize": 10,
    },
    "populate": map[string]any{
        "author":     map[string]any{"fields": []any{"name", "avatar"}},
        "categories": map[string]any{"fields": []any{"name", "slug"}},
        "cover":      map[string]any{"fields": []any{"url"}},
    },
    "fields": []any{"title", "slug", "excerpt", "publishedAt"},
},
    // Keep brackets in keys unescaped (matches Strapi expectations and is more readable).
    qs.WithStringifyEncodeValuesOnly(true),
)

resp, _ := http.Get("https://api.example.com/api/articles?" + query)

// Server side: parse the same syntax when your API receives it.
func handler(w http.ResponseWriter, r *http.Request) {
    params, err := qs.Parse(r.URL.RawQuery)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // params is a nested map/slice structure, e.g. params["filters"].(map[string]any)["status"]...
    _ = params
    w.WriteHeader(http.StatusOK)
}
```

See more Strapi examples in `demo/src/strapi-api/README.md`.

## Comparison with Other QS Libraries

| Feature                       | **qs (this repo)** | [gorilla/schema](https://github.com/gorilla/schema) | [google/go-querystring](https://github.com/google/go-querystring) | [go-playground/form](https://github.com/go-playground/form) | [ajg/form](https://github.com/ajg/form) |
| ----------------------------- | :----------------: | :-------------------------------------------------: | :---------------------------------------------------------------: | :---------------------------------------------------------: | :-------------------------------------: |
| Encode struct                 |         ✅         |                         ✅                          |                                ✅                                 |                             ✅                              |                   ✅                    |
| Decode struct                 |         ✅         |                         ✅                          |                                ❌                                 |                             ✅                              |                   ✅                    |
| Struct tags                   |         ✅         |                         ✅                          |                                ✅                                 |                             ✅                              |                   ✅                    |
| Nested objects/arrays         |         ✅         |                         ❌                          |                                ✅                                 |                             ✅                              |                   ✅                    |
| Nested `map[string]any`       |         ✅         |                         ❌                          |                                ❌                                 |                             ❌                              |                   ❌                    |
| Multiple array formats        |         ✅         |                         ❌                          |                                ✅                                 |                             ✅                              |                   ✅                    |
| Depth/limits controls         |         ✅         |                         ❌                          |                                ❌                                 |                             ❌                              |                   ❌                    |
| Charset sentinel + ISO-8859-1 |         ✅         |                         ❌                          |                                ❌                                 |                             ❌                              |                   ❌                    |
| Strict syntax validation      |         ✅         |                         ❌                          |                                ❌                                 |                             ❌                              |                   ❌                    |

### Array formats (supported)

`qs` supports 4 array formats for both stringify and parse:

- **Indices**: `a[0]=x&a[1]=y` — explicit positions, good when order/indices matter (can represent sparse arrays).
- **Brackets**: `a[]=x&a[]=y` — no indices, order is the parameter order (common in HTML forms).
- **Repeat**: `a=x&a=y` — repeated key, simplest/most interoperable; semantics depend on how duplicates are handled.
- **Comma**: `a=x,y` — compact single value; requires comma-splitting on parse and can be ambiguous if elements contain commas.

## JS `qs` option compatibility

This library is a Go port of JS `qs`, so most options map 1:1. The table below highlights what exists on both sides and where Go differs.

### Parse options

Most used:

- `AllowDots` — dot notation (`a.b=c`)
- `StrictMode` — strict syntax validation (unmatched brackets, invalid percent-encoding, etc.)
- `Depth` / `StrictDepth` — nesting depth controls
- `ParameterLimit` / `ThrowOnLimitExceeded` — control worst-case input size
- `Comma` / `Duplicates` — common real-world semantics knobs

<details>
<summary>Full parse options table</summary>

| Option                             | JS `qs` | Go `qs` | Notes                                                                                                                                                                               |
| :--------------------------------- | :-----: | :-----: | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `AllowDots`                        |   ✅    |   ✅    | Dot notation (`a.b=c`) [→](demo/src/allow-dots/README.md)                                                                                                                           |
| `AllowEmptyArrays`                 |   ✅    |   ✅    | `a[]=` creates `[]` vs `[""]` [→](demo/src/allow-empty-arrays/README.md)                                                                                                            |
| `AllowSparse`                      |   ✅    |   ✅    | Preserve array gaps [→](demo/src/allow-sparse/README.md)                                                                                                                            |
| `ArrayLimit`                       |   ✅    |   ✅    | Max index for array notation [→](demo/src/array-limit/README.md)                                                                                                                    |
| `Charset`                          |   ✅    |   ✅    | `utf-8` / `iso-8859-1` [→](demo/src/charset/README.md)                                                                                                                              |
| `CharsetSentinel`                  |   ✅    |   ✅    | `utf8=✓` detection [→](demo/src/charset-sentinel/README.md)                                                                                                                         |
| `Comma`                            |   ✅    |   ✅    | `a=1,2` → array [→](demo/src/comma/README.md)                                                                                                                                       |
| `DecodeDotInKeys`                  |   ✅    |   ✅    | `%2E` → `.` in keys [→](demo/src/decode-dot-in-keys/README.md)                                                                                                                      |
| `Decoder`                          |   ✅    |   ✅    | Custom decoder hook [→](demo/src/decoder/README.md)                                                                                                                                 |
| `Delimiter` / regexp delimiter     |   ✅    |   ✅    | Go supports `Delimiter` + `DelimiterRegexp` [→](demo/src/delimiter/README.md), [more](demo/src/delimiter-regexp/README.md)                                                          |
| `Depth`                            |   ✅    |   ✅    | Max nesting depth [→](demo/src/depth/README.md)                                                                                                                                     |
| `Duplicates`                       |   ✅    |   ✅    | `combine` / `first` / `last` [→](demo/src/duplicates/README.md)                                                                                                                     |
| `IgnoreQueryPrefix`                |   ✅    |   ✅    | Strip leading `?` [→](demo/src/ignore-query-prefix/README.md)                                                                                                                       |
| `InterpretNumericEntities`         |   ✅    |   ✅    | ISO-8859-1 numeric entities [→](demo/src/interpret-numeric-entities/README.md)                                                                                                      |
| `ParameterLimit`                   |   ✅    |   ✅    | Max number of params [→](demo/src/parameter-limit/README.md)                                                                                                                        |
| `ParseArrays`                      |   ✅    |   ✅    | Disable bracket parsing when false [→](demo/src/parse-arrays/README.md)                                                                                                             |
| `StrictDepth`                      |   ✅    |   ✅    | Error when depth exceeded [→](demo/src/strict-depth/README.md)                                                                                                                      |
| `StrictNullHandling`               |   ✅    |   ✅    | `a` → `null` vs `""` [→](demo/src/strict-null-handling/README.md)                                                                                                                   |
| `ThrowOnLimitExceeded`             |   ✅    |   ✅    | Error on `ParameterLimit` / `ArrayLimit` [→](demo/src/throw-on-limit-exceeded/README.md)                                                                                            |
| `AllowPrototypes` / `PlainObjects` |   ✅    |   N/A   | JS-only prototype pollution controls; in Go keys like `__proto__`, `constructor`, `prototype` are treated as normal map keys [→](demo/src/allow-prototypes-plain-objects/README.md) |
| `StrictMode`                       |   ❌    |   ✅    | Go-only: strict syntax validation (unmatched brackets, invalid percent-encoding, etc.) [→](demo/src/strict-mode/README.md)                                                          |

</details>

### Stringify options

Most used:

- `ArrayFormat` — indices / brackets / repeat / comma
- `EncodeValuesOnly` — keeps brackets in keys (common for Strapi-style APIs)
- `AllowDots` — dot output instead of brackets
- `Sort` — stable key ordering
- `Format` — RFC1738 / RFC3986

<details>
<summary>Full stringify options table</summary>

| Option               | JS `qs` | Go `qs` | Notes                                                                            |
| :------------------- | :-----: | :-----: | :------------------------------------------------------------------------------- |
| `AddQueryPrefix`     |   ✅    |   ✅    | Leading `?` [→](demo/src/add-query-prefix/README.md)                             |
| `AllowDots`          |   ✅    |   ✅    | Dot output instead of brackets [→](demo/src/allow-dots/README.md)                |
| `AllowEmptyArrays`   |   ✅    |   ✅    | Output `key[]` for empty arrays [→](demo/src/allow-empty-arrays/README.md)       |
| `ArrayFormat`        |   ✅    |   ✅    | `indices` / `brackets` / `repeat` / `comma` [→](demo/src/array-format/README.md) |
| `Charset`            |   ✅    |   ✅    | Output charset [→](demo/src/charset/README.md)                                   |
| `CharsetSentinel`    |   ✅    |   ✅    | Add `utf8=✓` [→](demo/src/charset-sentinel/README.md)                            |
| `CommaRoundTrip`     |   ✅    |   ✅    | `comma` single-element round-trip [→](demo/src/comma-round-trip/README.md)       |
| `Delimiter`          |   ✅    |   ✅    | Join delimiter [→](demo/src/delimiter/README.md)                                 |
| `Encode`             |   ✅    |   ✅    | Percent-encode output [→](demo/src/encode/README.md)                             |
| `EncodeDotInKeys`    |   ✅    |   ✅    | `.` → `%2E` in keys [→](demo/src/encode-dot-in-keys/README.md)                   |
| `Encoder`            |   ✅    |   ✅    | Custom encoder hook [→](demo/src/encoder/README.md)                              |
| `EncodeValuesOnly`   |   ✅    |   ✅    | Only encode values [→](demo/src/encode-values-only/README.md)                    |
| `Filter`             |   ✅    |   ✅    | Function or allowlist [→](demo/src/filter/README.md)                             |
| `Format`             |   ✅    |   ✅    | RFC1738 / RFC3986 [→](demo/src/format/README.md)                                 |
| `SerializeDate`      |   ✅    |   ✅    | Date formatting hook [→](demo/src/serialize-date/README.md)                      |
| `SkipNulls`          |   ✅    |   ✅    | Drop null keys [→](demo/src/skip-nulls/README.md)                                |
| `Sort`               |   ✅    |   ✅    | Custom key ordering [→](demo/src/sort/README.md)                                 |
| `StrictNullHandling` |   ✅    |   ✅    | `null` → `a` vs `a=` [→](demo/src/strict-null-handling/README.md)                |

</details>

### 🔥 Go-only extensions

| Feature             |                                              Go `qs`                                              |
| :------------------ | :-----------------------------------------------------------------------------------------------: |
| Struct API          |          ✅ (`Marshal` / `Unmarshal`, `query` tags) [→](demo/src/struct-tags/README.md)           |
| `[]byte` decode API |                   ✅ (`UnmarshalBytes`) [→](demo/src/unmarshal-bytes/README.md)                   |
| `SortArrayIndices`  | ✅ (matches JS key sorting behavior for array indices) [→](demo/src/sort-array-indices/README.md) |

## Parser architecture (Arena-backed, O(n))

Under the hood, `v2` uses a small lexer/parser in `v2/lang` that tokenizes the query string in a single pass and builds an arena-backed AST of `Span`s (offset/len views into the original input). This design keeps the hot path allocation-free in steady state when you reuse a `lang.Arena` (and is fully zero-copy via `ParseBytes`), while correctly handling the tricky `qs` key syntax (deeply nested brackets, percent-encoded `[`/`]` and dots, and `=` inside bracketed segments) and enabling strict error reporting for malformed keys (unmatched/unclosed brackets, invalid percent-encoding, etc.). It was added to make JS `qs`-compat parsing both fast and predictable for complex real-world keys; for the full grammar/AST details see `v2/LANGUAGE_SPECIFICATION.md`.

## Performance

Benchmarks on `darwin/arm64` (`go test -bench=. -benchmem` in `benchmarks/`). Lower is better.

### Stringify (encode)

Time / allocs (`μs/op`, `B/op`, `allocs/op`):

<details>
<summary>Benchmark table</summary>

| Case                                 |      **qs (this repo)** | [go-playground/form](https://github.com/go-playground/form) | [gorilla/schema](https://github.com/gorilla/schema) | [google/go-querystring](https://github.com/google/go-querystring) | [ajg/form](https://github.com/ajg/form) |
| :----------------------------------- | ----------------------: | ----------------------------------------------------------: | --------------------------------------------------: | ----------------------------------------------------------------: | --------------------------------------: |
| Simple struct                        |      **0.10 / 208 / 2** |                                             0.34 / 485 / 10 |                                     0.67 / 256 / 14 |                                                   1.00 / 656 / 20 |                        1.48 / 1120 / 23 |
| Nested struct (`a[b]=x`)             |      **0.10 / 224 / 2** |                                             0.41 / 528 / 10 |                                                   — |                                                   1.39 / 776 / 30 |                        2.63 / 2072 / 41 |
| Array struct (`a[0]=x`)              |      **0.10 / 184 / 2** |                                             0.51 / 724 / 15 |                                                   — |                                                   0.85 / 816 / 20 |                        2.16 / 1472 / 32 |
| Giant dynamic map (`map[string]any`) | **16.70 / 18206 / 351** |                                                           — |                                                   — |                                                                 — |                                       — |

</details>

### Parse / Unmarshal (decode)

Time / allocs (`μs/op`, `B/op`, `allocs/op`). Benchmarks use raw query string input for all libs (includes `url.ParseQuery` overhead where applicable):

<details>
<summary>Benchmark table</summary>

| Case                          |  **qs (this repo)** | [go-playground/form](https://github.com/go-playground/form) | [gorilla/schema](https://github.com/gorilla/schema) | [google/go-querystring](https://github.com/google/go-querystring) | [ajg/form](https://github.com/ajg/form) |
| :---------------------------- | ------------------: | ----------------------------------------------------------: | --------------------------------------------------: | ----------------------------------------------------------------: | --------------------------------------: |
| Simple struct                 |    1.22 / 1496 / 31 |                                          **0.53 / 528 / 8** |                                     2.13 / 872 / 45 |                                                                 — |                        2.98 / 1024 / 37 |
| Nested struct (native format) |    1.92 / 1808 / 45 |                                          **0.83 / 528 / 7** |                                                   — |                                                                 — |                        3.92 / 1736 / 40 |
| Array struct (native format)  |    2.12 / 2176 / 49 |                                         **0.88 / 848 / 14** |                                    1.43 / 1208 / 30 |                                                                 — |                        3.64 / 1507 / 38 |
| Dynamic map (`qs.Parse`)      | 33.57 / 46119 / 678 |                                                           — |                                                   — |                                                                 — |                                       — |

</details>

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
