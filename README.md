# QS - Query String Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/zaytracom/qs.svg)](https://pkg.go.dev/github.com/zaytracom/qs)
[![Go Report Card](https://goreportcard.com/badge/github.com/zaytracom/qs)](https://goreportcard.com/report/github.com/zaytracom/qs)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub release](https://img.shields.io/github/release/zaytracom/qs.svg)](https://github.com/zaytracom/qs/releases)
[![Build Status](https://github.com/zaytracom/qs/workflows/CI/badge.svg)](https://github.com/zaytracom/qs/actions)

**QS** is a powerful, high-performance Go library for parsing and stringifying URL query strings with support for nested objects, arrays, and complex data structures. This library is inspired by and compatible with the popular JavaScript [qs library](https://github.com/ljharb/qs), while providing Go-specific features like struct parsing with query tags and idiomatic Marshal/Unmarshal functions.

## ✨ Features

- 🔍 **Parse query strings** into nested Go data structures
- 📝 **Stringify Go data structures** into query strings
- 🏗️ **Support for nested objects and arrays** with multiple formatting options
- 🏷️ **Struct parsing with query tags** for type-safe operations
- 🔄 **Idiomatic Marshal/Unmarshal functions** with automatic type detection
- 🎯 **Strapi-style API support** for CMS and API applications
- ⚡ **High performance** with comprehensive benchmarks
- 🛡️ **Extensive customization** through options
- 📚 **Comprehensive documentation** and examples
- 🧪 **Extensive test coverage** (>95%)
- 🌐 **Framework agnostic** - works with any Go web framework

## 📦 Installation

```bash
go get github.com/zaytracom/qs/v1
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/zaytracom/qs/v1"
)

func main() {
    // Parse a query string
    result, err := qs.Parse("name=John&age=30&skills[]=Go&skills[]=Python")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%+v\n", result)
    // Output: map[age:30 name:John skills:[Go Python]]

    // Stringify a data structure
    data := map[string]interface{}{
        "user": map[string]interface{}{
            "name": "Jane",
            "profile": map[string]interface{}{
                "age": 25,
                "skills": []interface{}{"JavaScript", "TypeScript"},
            },
        },
    }

    queryString, err := qs.Stringify(data)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(queryString)
    // Output: user[name]=Jane&user[profile][age]=25&user[profile][skills][0]=JavaScript&user[profile][skills][1]=TypeScript
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
    err := qs.ParseToStruct(queryString, &user)
    if err != nil {
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

    queryString, err = qs.StructToQueryString(&newUser)
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
    queryString, err := qs.Marshal(user)
    if err != nil {
        log.Fatal(err)
    }

    // Unmarshal with automatic type detection
    var newUser User
    err = qs.Unmarshal(queryString, &newUser)
    if err != nil {
        log.Fatal(err)
    }
}
```

## 📚 Documentation

### Core Functions

| Function | Description |
|----------|-------------|
| `Parse(str, opts)` | Parse query string to map |
| `Stringify(obj, opts)` | Convert data structure to query string |
| `Marshal(v, opts)` | Convert any type to query string (idiomatic) |
| `Unmarshal(str, v, opts)` | Parse query string to any type (idiomatic) |
| `ParseToStruct(str, dest, opts)` | Parse query string to struct |
| `StructToQueryString(obj, opts)` | Convert struct to query string |

### Array Formats

The library supports multiple array formats:

```go
data := map[string]interface{}{
    "items": []interface{}{"a", "b", "c"},
}

// Indices format (default)
qs.Stringify(data)
// Result: "items[0]=a&items[1]=b&items[2]=c"

// Brackets format
qs.Stringify(data, &qs.StringifyOptions{ArrayFormat: "brackets"})
// Result: "items[]=a&items[]=b&items[]=c"

// Repeat format
qs.Stringify(data, &qs.StringifyOptions{ArrayFormat: "repeat"})
// Result: "items=a&items=b&items=c"
```

### Nested Objects

```go
// Parse nested structures
result, _ := qs.Parse("user[profile][name]=John&user[profile][age]=30")
// Result: map[user:map[profile:map[age:30 name:John]]]

// Create nested structures
data := map[string]interface{}{
    "user": map[string]interface{}{
        "profile": map[string]interface{}{
            "name": "John",
            "age":  30,
        },
    },
}
queryString, _ := qs.Stringify(data)
// Result: "user[profile][age]=30&user[profile][name]=John"
```

### Custom Options

#### Parse Options

```go
options := &qs.ParseOptions{
    Delimiter:         "&",     // Parameter delimiter
    IgnoreQueryPrefix: true,    // Ignore leading '?'
    ArrayLimit:        20,      // Maximum array elements
    Depth:            5,        // Maximum nesting depth
    ParameterLimit:   1000,     // Maximum parameters
}

result, err := qs.Parse("?name=John&age=30", options)
```

#### Stringify Options

```go
options := &qs.StringifyOptions{
    ArrayFormat:    "brackets", // Array format
    AddQueryPrefix: true,       // Add '?' prefix
    Delimiter:      "&",        // Parameter delimiter
    Encode:         true,       // URL encoding
}

result, err := qs.Stringify(data, options)
```

## 🏗️ Advanced Usage

### Strapi-style APIs

Perfect for building CMS and API applications:

```go
type StrapiQuery struct {
    Filters    map[string]interface{} `query:"filters"`
    Sort       []string               `query:"sort"`
    Fields     []string               `query:"fields"`
    Populate   map[string]interface{} `query:"populate"`
    Pagination map[string]interface{} `query:"pagination"`
    Locale     string                 `query:"locale"`
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

QS is optimized for high performance:

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
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

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

- 📖 [Documentation](https://pkg.go.dev/github.com/zaytracom/qs)
- 🐛 [Issue Tracker](https://github.com/zaytracom/qs/issues)
- 💬 [Discussions](https://github.com/zaytracom/qs/discussions)
