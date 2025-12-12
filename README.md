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

## 📦 Installation

```bash
go get github.com/zaytracom/qs/v2
```

## 📋 Comparison with Other Libraries

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

## 📊 Performance

### Stringify — lower is better

| Benchmark | zaytra | gorilla | go-playground | ajg | google |
|:----------|-------:|--------:|--------------:|----:|-------:|
| Simple struct | 97 ns | 667 ns | 317 ns | 1438 ns | 959 ns |
| Nested struct | 101 ns | — | 396 ns | 2563 ns | 1349 ns |
| Array struct | 97 ns | — | 509 ns | 2181 ns | 848 ns |
| Dynamic map | 17 μs | — | — | — | — |

### Parse — lower is better

| Benchmark | zaytra | gorilla | go-playground | ajg | google |
|:----------|-------:|--------:|--------------:|----:|-------:|
| Simple struct | 12.4 μs | 1.7 μs | 121 ns | 2.5 μs | — |
| Nested struct | 14.7 μs | — | 343 ns | — | — |
| Array struct | 16.6 μs | — | 402 ns | — | — |
| Dynamic map | 91 μs | — | — | — | — |

## 📚 Documentation

- [Go Reference (v2)](https://pkg.go.dev/github.com/zaytracom/qs/v2)
- [GitHub Issues](https://github.com/zaytracom/qs/issues)

## 📄 License

Apache 2.0 — see [LICENSE](LICENSE)
