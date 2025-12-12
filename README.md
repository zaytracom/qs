# QS — Query String library for Go (v2-beta)

[![Go Reference (v2)](https://pkg.go.dev/badge/github.com/zaytracom/qs/v2.svg)](https://pkg.go.dev/github.com/zaytracom/qs/v2)
[![CI](https://github.com/zaytracom/qs/actions/workflows/ci.yml/badge.svg)](https://github.com/zaytracom/qs/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zaytracom/qs/v2)](https://goreportcard.com/report/github.com/zaytracom/qs/v2)
[![codecov](https://codecov.io/gh/zaytracom/qs/branch/main/graph/badge.svg)](https://codecov.io/gh/zaytracom/qs)
[![GitHub release](https://img.shields.io/github/v/release/zaytracom/qs?include_prereleases)](https://github.com/zaytracom/qs/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://github.com/zaytracom/qs)

**QS v2-beta** is a production-ready Go port of the popular JavaScript [`qs`](https://github.com/ljharb/qs) library for parsing and stringifying URL query strings (nested objects, arrays, and all the tricky edge cases).

**Status: beta.** The v2 API is considered stable enough for testing, but may still change before the final v2 release.

**v2 is a full port from JS** and is validated against the real npm `qs` via the JS compatibility suite in `demo/`.

## ✨ Features

- 🔍 **Parse query strings** into nested Go data structures (`map[string]any`, `[]any`, …)
- 📝 **Stringify Go values** into query strings (arrays, nested objects, filters/sort, RFC formats)
- 🧩 **JS `qs` compatibility** (port + compatibility tests)
- 🏷️ **Struct support via `query` tags** (`Marshal` / `Unmarshal`, `ParseToStruct`, `StructToQueryString`)
- 🎯 **Strapi-style API support** for CMS and API applications
- 🧪 **Extensive tests** + Go/JS parity demos in `demo/src/*`

## 🆕 v2-beta vs v1

- **v2-beta (recommended for new users)**: `github.com/zaytracom/qs/v2` (Go 1.21+)
- **v1 (legacy)**: `github.com/zaytracom/qs/v1` (kept for compatibility)

Key v2 API changes:

- Options moved from `*Options` structs to functional options (`...ParseOption`, `...StringifyOption`)
- Strongly-typed enums for key options (`ArrayFormat`, `Format`, `Charset`, …)

See `CHANGELOG.md` for details.

## 🧪 Demo (JS compatibility)

The repository includes a demo suite that runs the real npm `qs` and compares outputs with Go:

- `demo/src/*` — feature-focused docs with JS ↔ Go examples (tested)
- `demo/tests/jscompat_test.go` — broader JS compatibility tests

Run locally:

```bash
cd demo
npm ci
go test ./...
```

## 📦 Installation

### v2-beta (recommended)

```bash
go get github.com/zaytracom/qs/v2@v2.0.0-beta.2
```

### v1 (legacy)

```bash
go get github.com/zaytracom/qs/v1
```

## 🚀 Get started (v2-beta)

### Basic Usage

```go
package main

import (
	"fmt"
	"log"

	qs "github.com/zaytracom/qs/v2"
)

func main() {
	// Parse a query string (nested objects + arrays)
	result, err := qs.Parse("name=John&age=30&skills[]=Go&skills[]=Python")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", result)

	// Stringify a nested structure
	data := map[string]any{
		"user": map[string]any{
			"name": "Jane",
			"profile": map[string]any{
				"age":    25,
				"skills": []any{"JavaScript", "TypeScript"},
			},
		},
	}

	queryString, err := qs.Stringify(
		data,
		qs.WithStringifyEncodeValuesOnly(true), // keep keys readable (brackets), still encode values
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(queryString)
	// user[name]=Jane&user[profile][age]=25&user[profile][skills][0]=JavaScript&user[profile][skills][1]=TypeScript
}
```

### Struct Parsing with Query Tags

```go
type User struct {
	Name     string  `query:"name"`
	Age      int     `query:"age"`
	Email    string  `query:"email"`
	IsActive bool    `query:"active"`
	Score    float64 `query:"score"`
}

func main() {
	// Parse to struct
	queryString := "name=John&age=30&email=john@example.com&active=true&score=95.5"
	var user User
	if err := qs.ParseToStruct(queryString, &user); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User: %+v\n", user)

	// Convert struct to query string
	newUser := User{
		Name:     "Alice",
		Age:      25,
		Email:    "alice@example.com",
		IsActive: true,
		Score:    88.5,
	}

	queryString, err := qs.StructToQueryString(newUser, qs.WithStringifyEncodeValuesOnly(true))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Query: %s\n", queryString)
}
```

### Idiomatic Marshal/Unmarshal

```go
func main() {
	// Marshal with automatic type detection
	user := User{Name: "John", Age: 30}
	queryString, err := qs.Marshal(user, qs.WithStringifyEncodeValuesOnly(true))
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal with automatic type detection
	var newUser User
	if err := qs.Unmarshal(queryString, &newUser); err != nil {
		log.Fatal(err)
	}
}
```

## 📚 Documentation

### Core Functions

| Function | Description |
|----------|-------------|
| `Parse(str, opts ...ParseOption)` | Parse query string to `map[string]any` |
| `Stringify(obj, opts ...StringifyOption)` | Convert a value to a query string |
| `Marshal(v, opts ...StringifyOption)` | Convert any Go value to a query string |
| `Unmarshal(str, v, opts ...ParseOption)` | Parse query string into any Go value |
| `ParseToStruct(str, dest, opts ...ParseOption)` | Parse query string into a struct |
| `StructToQueryString(obj, opts ...StringifyOption)` | Convert a struct to a query string |

### Options (quick peek)

Array formats:

```go
data := map[string]any{
	"items": []any{"a", "b", "c"},
}

qs.Stringify(data, qs.WithStringifyArrayFormat(qs.ArrayFormatIndices), qs.WithStringifyEncode(false))
// items[0]=a&items[1]=b&items[2]=c

qs.Stringify(data, qs.WithStringifyArrayFormat(qs.ArrayFormatBrackets), qs.WithStringifyEncode(false))
// items[]=a&items[]=b&items[]=c

qs.Stringify(data, qs.WithStringifyArrayFormat(qs.ArrayFormatRepeat), qs.WithStringifyEncode(false))
// items=a&items=b&items=c

qs.Stringify(data, qs.WithStringifyArrayFormat(qs.ArrayFormatComma), qs.WithStringifyEncode(false))
// items=a,b,c
```

### Nested Objects

```go
// Parse nested structures
result, _ := qs.Parse("user[profile][name]=John&user[profile][age]=30")
// Result: map[user:map[profile:map[age:30 name:John]]]

// Create nested structures
data := map[string]any{
	"user": map[string]any{
		"profile": map[string]any{
			"name": "John",
			"age":  30,
		},
	},
}
queryString, _ := qs.Stringify(data, qs.WithStringifyEncodeValuesOnly(true))
// user[profile][age]=30&user[profile][name]=John
```

More JS ↔ Go examples (tested): start with `demo/src/array-format/README.md`, `demo/src/struct-tags/README.md`, and `demo/src/strapi-api/README.md`.

## 🏗️ Advanced Usage

### Parse options example

```go
result, err := qs.Parse(
	"?name=John&age=30",
	qs.WithParseIgnoreQueryPrefix(true),
	qs.WithParseDepth(5),
	qs.WithParseParameterLimit(1000),
)
_ = result
_ = err
```

### Stringify options example

```go
result, err := qs.Stringify(
	map[string]any{"items": []any{"a", "b"}},
	qs.WithStringifyAddQueryPrefix(true),
	qs.WithStringifyArrayFormat(qs.ArrayFormatBrackets),
	qs.WithStringifyEncodeValuesOnly(true),
)
_ = result
_ = err
```

### Strapi-style APIs

Perfect for building CMS and API applications:

```go
type StrapiQuery struct {
	Filters    map[string]any `query:"filters"`
	Sort       []string       `query:"sort"`
	Fields     []string       `query:"fields"`
	Populate   map[string]any `query:"populate"`
	Pagination map[string]any `query:"pagination"`
	Locale     string         `query:"locale"`
}

// Parse complex Strapi query
strapiQuery := "filters[title][$contains]=golang&sort[]=publishedAt:desc&populate[author][fields][]=name"
var query StrapiQuery
err := qs.Unmarshal(strapiQuery, &query)
```

### Framework Integration

Works seamlessly with popular Go web frameworks:

#### Gin Framework

```go
func getArticles(c *gin.Context) {
    var query ArticleQuery
    if err := qs.Unmarshal(c.Request.URL.RawQuery, &query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    articles := getArticlesFromDB(query)
    c.JSON(200, gin.H{"data": articles})
}
```

#### Echo Framework

```go
func getProducts(c echo.Context) error {
    var query ProductQuery
    if err := qs.Unmarshal(c.Request().URL.RawQuery, &query); err != nil {
        return echo.NewHTTPError(400, err.Error())
    }

    products := getProductsFromDB(query)
    return c.JSON(200, products)
}
```

## 📊 Performance

QS is optimized for high performance. Run local benchmarks to see numbers on your machine:

```bash
make bench-v2
```

Example output:

```
BenchmarkParseSimple-10       169478    6458 ns/op    11499 B/op    147 allocs/op
BenchmarkParseComplex-10       64339   18474 ns/op    27625 B/op    322 allocs/op
BenchmarkStringifySimple-10  2973129     400 ns/op      190 B/op     10 allocs/op
BenchmarkStringifyComplex-10  701146    1675 ns/op     1121 B/op     31 allocs/op
BenchmarkMarshal-10           759507    1491 ns/op      520 B/op     15 allocs/op
BenchmarkUnmarshal-10         175050    7350 ns/op     3825 B/op     95 allocs/op
```

## 🔧 Real-World Examples

### E-commerce Product Filtering

```go
type ProductQuery struct {
    Filters struct {
        Category   string  `query:"category"`
        PriceMin   float64 `query:"price_min"`
        PriceMax   float64 `query:"price_max"`
        InStock    bool    `query:"in_stock"`
        Brand      []string `query:"brand"`
    } `query:"filters"`
    Sort       []string `query:"sort"`
    Pagination struct {
        Page     int `query:"page"`
        PageSize int `query:"page_size"`
    } `query:"pagination"`
}

// Parse: /api/products?filters[category]=electronics&filters[price_min]=100&filters[price_max]=1000&sort[]=price:asc
```

### Search Interface

```go
type SearchQuery struct {
    Query    string   `query:"q"`
    Filters  map[string]interface{} `query:"filters"`
    Facets   []string `query:"facets"`
    Sort     []string `query:"sort"`
    Page     int      `query:"page"`
    PageSize int      `query:"page_size"`
}
```

### Analytics Dashboard

```go
type AnalyticsQuery struct {
    Metrics   []string `query:"metrics"`
    DateRange struct {
        Start string `query:"start"`
        End   string `query:"end"`
    } `query:"date_range"`
    GroupBy   []string `query:"group_by"`
    Filters   map[string]interface{} `query:"filters"`
}
```

## 🧪 Testing

Run the test suite:

```bash
# Run all Go tests (v1 + v2 + demo)
make test

# Demo tests require Node.js + npm deps
cd demo && npm ci && go test ./...
```

## 🤝 Contributing

Please review `CODE_OF_CONDUCT.md` before participating.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by the JavaScript [qs library](https://github.com/ljharb/qs) by Jordan Harband
- Thanks to all contributors who have helped improve this library

## 📞 Support

- 📖 [Documentation (v2)](https://pkg.go.dev/github.com/zaytracom/qs/v2)
- 📖 [Documentation (v1)](https://pkg.go.dev/github.com/zaytracom/qs/v1)
- 🐛 [Issue Tracker](https://github.com/zaytracom/qs/issues)
- 💬 [Discussions](https://github.com/zaytracom/qs/discussions)
