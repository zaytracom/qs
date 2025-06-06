# QS - Query String Library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/zaytra/qs.svg)](https://pkg.go.dev/github.com/zaytra/qs)
[![Go Report Card](https://goreportcard.com/badge/github.com/zaytra/qs)](https://goreportcard.com/report/github.com/zaytra/qs)

QS is a Go library for parsing and stringifying URL query strings with support for nested objects, arrays, and complex data structures. This is a port of the popular JavaScript [qs library](https://github.com/ljharb/qs).

## Features

- ✅ Parse query strings into nested Go data structures
- ✅ Stringify Go data structures into query strings
- ✅ Support for nested objects and arrays
- ✅ Multiple array formats (indices, brackets, repeat)
- ✅ URL encoding/decoding
- ✅ Customizable options and delimiters
- ✅ High performance with comprehensive benchmarks
- ✅ Extensive test coverage for complex scenarios

## Installation

```bash
go get github.com/zaytra/qs
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/zaytra/qs/qs"
)

func main() {
    // Parse a query string
    result, err := qs.Parse("name=John&age=30&skills[]=Go&skills[]=Python")
    if err != nil {
        panic(err)
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
        panic(err)
    }
    fmt.Println(queryString)
    // Output: user[name]=Jane&user[profile][age]=25&user[profile][skills][0]=JavaScript&user[profile][skills][1]=TypeScript
}
```

## Parsing Examples

### Simple Key-Value Pairs

```go
result, _ := qs.Parse("name=John&age=30&city=NYC")
// Result: map[string]interface{}{
//     "name": "John",
//     "age":  "30",
//     "city": "NYC",
// }
```

### Nested Objects

```go
result, _ := qs.Parse("user[profile][name]=John&user[profile][age]=30")
// Result: map[string]interface{}{
//     "user": map[string]interface{}{
//         "profile": map[string]interface{}{
//             "name": "John",
//             "age":  "30",
//         },
//     },
// }
```

### Arrays

```go
result, _ := qs.Parse("tags[]=golang&tags[]=programming&tags[]=web")
// Result: map[string]interface{}{
//     "tags": []interface{}{"golang", "programming", "web"},
// }
```

### Complex Nested Structures

```go
queryString := "products[0][id]=123&products[0][name]=Laptop&products[0][tags][]=electronics&products[0][tags][]=computers&filters[price][min]=100&filters[price][max]=1000"
result, _ := qs.Parse(queryString)
// Result: deeply nested map with products array and filters object
```

## Stringify Examples

### Basic Usage

```go
data := map[string]interface{}{
    "name": "John",
    "age":  30,
}
result, _ := qs.Stringify(data)
// Result: "age=30&name=John"
```

### Arrays with Different Formats

```go
data := map[string]interface{}{
    "items": []interface{}{"a", "b", "c"},
}

// Indices format (default)
result1, _ := qs.Stringify(data)
// Result: "items[0]=a&items[1]=b&items[2]=c"

// Brackets format
result2, _ := qs.Stringify(data, &qs.StringifyOptions{
    ArrayFormat: "brackets",
})
// Result: "items[]=a&items[]=b&items[]=c"

// Repeat format
result3, _ := qs.Stringify(data, &qs.StringifyOptions{
    ArrayFormat: "repeat",
})
// Result: "items=a&items=b&items=c"
```

## Options

### Parse Options

```go
options := &qs.ParseOptions{
    Delimiter:         "&",     // Parameter delimiter
    IgnoreQueryPrefix: true,    // Ignore leading '?'
    StrictNullHandling: false,  // Handle null values strictly
    // ... more options available
}

result, err := qs.Parse("?name=John&age=30", options)
```

### Stringify Options

```go
options := &qs.StringifyOptions{
    ArrayFormat:    "brackets", // Array format: "indices", "brackets", "repeat"
    AddQueryPrefix: true,       // Add '?' prefix
    Delimiter:      "&",        // Parameter delimiter
    // ... more options available
}

result, err := qs.Stringify(data, options)
```

## Advanced Usage

### Custom Delimiters

```go
// Parse with semicolon delimiter
result, _ := qs.Parse("name=John;age=30;city=NYC", &qs.ParseOptions{
    Delimiter: ";",
})

// Stringify with semicolon delimiter
data := map[string]interface{}{"a": "1", "b": "2"}
result, _ := qs.Stringify(data, &qs.StringifyOptions{
    Delimiter: ";",
})
// Result: "a=1;b=2"
```

### URL Encoding

```go
// Parse URL-encoded values
result, _ := qs.Parse("message=Hello%20World&emoji=%F0%9F%9A%80")
// Result: map[string]interface{}{
//     "message": "Hello World",
//     "emoji":   "🚀",
// }

// Stringify with URL encoding (enabled by default)
data := map[string]interface{}{
    "message": "Hello World!",
    "symbols": "@#$%^&*()",
}
result, _ := qs.Stringify(data)
// Result: URL-encoded query string
```

## Real-World Examples

See [`examples.md`](examples.md) for comprehensive real-world usage examples including:

- E-commerce product filtering
- API request building
- Form data processing
- Analytics and metrics
- Configuration management
- Search interfaces

## Performance

Benchmark results on Apple M1 Pro:

```
BenchmarkParseSimple-10       169478    6458 ns/op    11499 B/op    147 allocs/op
BenchmarkParseComplex-10       64339   18474 ns/op    27625 B/op    322 allocs/op
BenchmarkStringifySimple-10  2973129     400 ns/op      190 B/op     10 allocs/op
BenchmarkStringifyComplex-10  701146    1675 ns/op     1121 B/op     31 allocs/op
```

The library is optimized for performance while maintaining full compatibility with the JavaScript qs library.

## API Reference

### `Parse(str string, opts ...*ParseOptions) (map[string]interface{}, error)`

Parses a query string into a nested data structure.

**Parameters:**
- `str` - The query string to parse
- `opts` - Optional parsing options

**Returns:**
- `map[string]interface{}` - The parsed data structure
- `error` - Any parsing error

### `Stringify(obj interface{}, opts ...*StringifyOptions) (string, error)`

Converts a data structure into a query string.

**Parameters:**
- `obj` - The data structure to stringify
- `opts` - Optional stringify options

**Returns:**
- `string` - The generated query string
- `error` - Any stringification error

## Compatibility

This library aims to be compatible with the JavaScript [qs library](https://github.com/ljharb/qs) while following Go conventions and idioms.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

This library is inspired by and aims to be compatible with the JavaScript [qs library](https://github.com/ljharb/qs) by Jordan Harband.

# QS Go Library - Idiomatic Marshal/Unmarshal Functions

The QS library for Go now supports idiomatic `Marshal` and `Unmarshal` functions that automatically detect data types at runtime and choose the appropriate processing method.

## Key Features

### ✨ Automatic Type Detection
- **Runtime type detection** - functions automatically determine whether they work with struct, map, or slice
- **Unified API** - same functions for all data types
- **Full compatibility** - backward compatibility with existing API is preserved

### 🚀 Performance
```
BenchmarkMarshal-10                       759507              1491 ns/op
BenchmarkUnmarshal-10                     175050              7350 ns/op
BenchmarkMarshalComplex-10                406468              2964 ns/op
BenchmarkUnmarshalComplex-10               75964             15018 ns/op
```

### 📋 Supported Types
- **Structs** with query tags (`query:"field_name"`)
- **Maps** (`map[string]interface{}`, `map[string]string`, etc.)
- **Slices** and **Arrays**
- **Nested structures**
- **Pointers**
- **All basic Go types** (string, int, float, bool, etc.)

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/zaytra/qs/qs"
)

type User struct {
    Name     string  `query:"name"`
    Age      int     `query:"age"`
    Email    string  `query:"email"`
    IsActive bool    `query:"active"`
    Score    float64 `query:"score"`
}

func main() {
    // Marshal struct -> query string
    user := User{
        Name:     "John Doe",
        Age:      30,
        Email:    "john@example.com",
        IsActive: true,
        Score:    95.5,
    }

    queryString, err := qs.Marshal(user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Query: %s\n", queryString)

    // Unmarshal query string -> struct
    var newUser User
    err = qs.Unmarshal(queryString, &newUser)
    if err != nil {
        panic(err)
    }
    fmt.Printf("User: %+v\n", newUser)
}
```

## API Reference

### Marshal
```go
func Marshal(v interface{}, opts ...*StringifyOptions) (string, error)
```
Converts any data type to a query string. Automatically detects type at runtime.

### Unmarshal
```go
func Unmarshal(queryString string, v interface{}, opts ...*ParseOptions) error
```
Parses query string into the specified data type. Automatically detects target type at runtime.

## Usage Examples

### Runtime Type Detection
```go
testQuery := "name=Alice&age=28&q=search&tags[]=go&tags[]=api"

// Single API for different types
var userTarget User
var mapTarget map[string]interface{}
var filterTarget SearchFilter

qs.Unmarshal(testQuery, &userTarget)    // struct
qs.Unmarshal(testQuery, &mapTarget)     // map
qs.Unmarshal(testQuery, &filterTarget)  // another struct
```

### HTTP Handler
```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters into struct
    var params RequestParams
    if err := qs.Unmarshal(r.URL.RawQuery, &params); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Process...
    response := ResponseData{Status: "success", Data: params}

    // Create response query
    responseQuery, _ := qs.Marshal(response)
    w.Header().Set("X-Response-Query", responseQuery)
}
```

### Complex Nested Structures
```go
type ComplexData struct {
    User     User              `query:"user"`
    Filters  SearchFilter      `query:"filters"`
    Settings map[string]string `query:"settings"`
    Enabled  bool              `query:"enabled"`
}

// Marshal/Unmarshal work with any complexity
queryString, _ := qs.Marshal(complexData)
var parsed ComplexData
qs.Unmarshal(queryString, &parsed)
```

## Query Tags

The following tag options are supported:

```go
type Example struct {
    Name     string `query:"name"`        // Custom field name
    Age      int    `query:"age"`         // Automatic type conversion
    Email    string `query:"email"`       // URL encoding/decoding
    Hidden   string `query:"-"`           // Skip field
    Default  string                       // Use lowercase field name
}
```

## Compatibility

New functions are fully compatible with the existing API:
- `Parse()` - parsing to map
- `Stringify()` - creating query string from map
- `ParseToStruct()` - parsing to struct
- `StructToQueryString()` - creating query string from struct
- `MapToStruct()` - converting map to struct
- `StructToMap()` - converting struct to map
