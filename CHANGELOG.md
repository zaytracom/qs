# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-12-06

### 🎉 Initial Release

This is the first stable release of the QS library for Go - a powerful, high-performance library for parsing and stringifying URL query strings with support for nested objects, arrays, and complex data structures.

### ✨ Features

#### Core Functionality
- **Parse query strings** into nested Go data structures with `Parse()` function
- **Stringify Go data structures** into query strings with `Stringify()` function
- **Support for nested objects and arrays** with multiple formatting options
- **Multiple array formats**: indices (`a[0]=1&a[1]=2`), brackets (`a[]=1&a[]=2`), repeat (`a=1&a=2`)

#### Idiomatic Go API
- **Marshal/Unmarshal functions** with automatic type detection at runtime
- **Struct parsing with query tags** for type-safe operations (`query:"field_name"`)
- **ParseToStruct()** and **StructToQueryString()** for direct struct operations
- **MapToStruct()** and **StructToMap()** for data conversion

#### Advanced Features
- **Strapi-style API support** for CMS and API applications
- **Extensive customization** through ParseOptions and StringifyOptions
- **URL encoding/decoding** with proper character handling
- **Custom delimiters** and formatting options
- **Deep nesting support** with configurable depth limits
- **Error handling** with detailed error messages

#### Performance & Quality
- **High performance** with comprehensive benchmarks
- **Memory efficient** with optimized allocations
- **Extensive test coverage** (>95%) with edge case testing
- **Framework agnostic** - works with any Go web framework

### 📚 Supported Data Types

- **Primitive types**: string, int (all variants), uint (all variants), float32/64, bool
- **Complex types**: maps, slices, arrays, structs
- **Nested structures**: unlimited nesting with configurable depth
- **Pointers**: automatic allocation and dereferencing
- **Interfaces**: interface{} support
- **Time values**: time.Time with custom serialization

### 🔧 Framework Integration

Works seamlessly with popular Go web frameworks:
- **Gin** - the most popular Go web framework
- **Echo** - high-performance minimalist framework
- **Chi** - lightweight router
- **net/http** - standard library
- **Fiber** - Express-inspired framework
- **Any other framework** - framework agnostic design

### 📊 Performance

Optimized for high-performance applications:
- `BenchmarkParseSimple-10`: 169,478 ops/s
- `BenchmarkParseComplex-10`: 64,339 ops/s
- `BenchmarkStringifySimple-10`: 2,973,129 ops/s
- `BenchmarkStringifyComplex-10`: 701,146 ops/s
- `BenchmarkMarshal-10`: 759,507 ops/s
- `BenchmarkUnmarshal-10`: 175,050 ops/s

### 🌐 Compatibility

- **Go version**: Go 1.19+ (tested with Go 1.21-1.24)
- **JavaScript qs compatibility**: Compatible with popular JavaScript qs library patterns
- **Backward compatibility**: Stable API following semantic versioning

### 📦 Package Structure

```
github.com/zaytracom/qs/v1
├── qs.go           # Main package with all functionality
├── qs_test.go      # Comprehensive test suite
└── examples/       # Usage examples and demos
    ├── strapi_demo.go       # Strapi-style API examples
    ├── example_struct.go    # Struct parsing examples
    └── EXAMPLES.md          # Comprehensive usage guide
```

### 🎯 Use Cases

Perfect for:
- **REST API query parsing** - complex filtering and pagination
- **E-commerce applications** - product filtering and search
- **CMS systems** - Strapi-style content queries
- **Analytics dashboards** - metrics and reporting queries
- **Search interfaces** - faceted search and filtering
- **Configuration management** - URL-based configuration
- **Form processing** - complex form data handling

### 📄 License

Licensed under the Apache License 2.0 - see [LICENSE](LICENSE) for details.

### 🙏 Acknowledgments

- Inspired by the JavaScript [qs library](https://github.com/ljharb/qs) by Jordan Harband
- Built with ❤️ for the Go community

---

*This changelog follows the [Keep a Changelog](https://keepachangelog.com/) format.*
