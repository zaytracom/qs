# QS — Query String library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/zaytracom/qs/v2.svg)](https://pkg.go.dev/github.com/zaytracom/qs/v2)
[![CI](https://github.com/zaytracom/qs/actions/workflows/ci.yml/badge.svg)](https://github.com/zaytracom/qs/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zaytracom/qs/v2)](https://goreportcard.com/report/github.com/zaytracom/qs/v2)
[![codecov](https://codecov.io/gh/zaytracom/qs/branch/main/graph/badge.svg)](https://codecov.io/gh/zaytracom/qs)
[![GitHub release](https://img.shields.io/github/v/release/zaytracom/qs?include_prereleases)](https://github.com/zaytracom/qs/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Go port of the popular JavaScript [`qs`](https://github.com/ljharb/qs) library — parse and stringify URL query strings with nested objects, arrays, and all the tricky edge cases.

## ✨ Features

- 🔍 Parse query strings into nested Go data structures (`map[string]any`, `[]any`)
- 📝 Stringify Go values into query strings (arrays, nested objects, filters/sort)
- 🧩 Full JS `qs` compatibility (validated via JS compatibility test suite)
- 🏷️ Struct support via `query` tags (`Marshal` / `Unmarshal`)
- 🎯 Multiple array formats: indices, brackets, repeat, comma
- ⚙️ Depth control and charset support (UTF-8/ISO-8859-1)
- 📋 RFC 1738/3986 encoding formats

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

## Comparison with Other Libraries

| Feature | zaytra | [gorilla](https://github.com/gorilla/schema) | [go-playground](https://github.com/go-playground/form) | [ajg](https://github.com/ajg/form) | [google](https://github.com/google/go-querystring) |
|---------|:------:|:------:|:------:|:------:|:------:|
| Encode | ✅ | ✅ | ✅ | ✅ | ✅ |
| Decode | ✅ | ✅ | ✅ | ✅ | ❌ |
| Struct Tags | ✅ | ✅ | ✅ | ✅ | ✅ |
| Nested Objects | ✅ | ❌ | ✅ | ✅ | ✅ |
| Nested Arrays | ✅ | ❌ | ✅ | ✅ | ❌ |
| Dynamic map | ✅ | ❌ | ❌ | ❌ | ❌ |
| Array Formats | 4️⃣ | 1️⃣ | 2️⃣ | 1️⃣ | 5️⃣ |
| Depth Control | ✅ | ❌ | ❌ | ❌ | ❌ |

**Array Formats:** `indices` (`a[0]=x`), `brackets` (`a[]=x`), `repeat` (`a=x&a=y`), `comma` (`a=x,y`)

## Performance

### Stringify (struct → query string) — lower is better

| Benchmark | zaytra | gorilla | google | go-playground | ajg |
|:----------|-------:|--------:|-------:|--------------:|----:|
| Simple struct | **96 ns** | 667 ns | 952 ns | 314 ns | 1430 ns |
| Nested struct | **98 ns** | — | 1340 ns | 397 ns | 2592 ns |
| Array struct | **95 ns** | — | 820 ns | 501 ns | 2107 ns |
| Dynamic map | 16 μs | — | — | — | — |

zaytra is **3-26x faster** than alternatives for encoding.

### Parse (query string → struct) — lower is better

| Benchmark | zaytra | gorilla | google | go-playground | ajg |
|:----------|-------:|--------:|-------:|--------------:|----:|
| Simple struct | 12 μs | 2 μs | — | 515 ns | 2.9 μs |
| Nested struct | 14 μs | — | — | 800 ns | — |
| Array struct | 17 μs | — | — | 834 ns | — |
| Dynamic map | 90 μs | — | — | — | — |

## Documentation

- [Go Reference (v2)](https://pkg.go.dev/github.com/zaytracom/qs/v2)
- [GitHub Issues](https://github.com/zaytracom/qs/issues)

## License

Apache 2.0 — see [LICENSE](LICENSE)
