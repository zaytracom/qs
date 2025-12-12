# QS v2 Implementation Plan - Full Node.js Port to Go

This document outlines the complete implementation plan for porting the JavaScript [qs library](https://github.com/ljharb/qs) to Go, while preserving Go-specific features from v1.

## Overview

**Goal**: Create a full-featured, 1:1 compatible port of the JavaScript `qs` library to Go.

**Source**: Node.js `qs` library (`.ref/` directory)
**Target**: Go package `github.com/zaytracom/qs/v2`

---

## Implementation Phases

### Phase 1: Core Infrastructure
- [x] **1.1 Project Setup** ✅ DONE
  - [x] Initialize Go module with proper versioning (`v2`)
  - [x] Set up directory structure
  - [ ] Configure CI/CD pipeline
  - **Tests**: Verify module initialization, imports work correctly

- [x] **1.2 Format Constants (formats.go)** ✅ DONE
  - [x] Define `Format` type
  - [x] Define `FormatRFC1738`, `FormatRFC3986` constants
  - [x] Set default format to RFC3986
  - [x] Implement `RFC1738` formatter function (spaces as `+`)
  - [x] Implement `RFC3986` formatter function (returns unchanged)
  - [x] Implement `Formatters` map with functions
  - [x] Implement `GetFormatter` helper function
  - **Tests**: ✅ All passing (formats_test.go)
    - Format constants exist
    - Formatter functions work correctly
    - RFC1738 converts `%20` to `+`
    - RFC3986 returns value unchanged
    - Unknown format fallback to RFC3986

- [x] **1.3 Utility Functions (utils.go)** ✅ DONE
  - [x] `hexTable` - Pre-computed hex encoding table (256 entries)
  - [x] `Encode(str, charset, format)` - RFC-compliant URL encoding
    - Handle UTF-8 encoding
    - Handle ISO-8859-1 encoding
    - Handle surrogate pairs
    - Support RFC1738/RFC3986 formats
  - [x] `Decode(str, charset)` - URL decoding
    - Handle `+` as space
    - Handle UTF-8 decoding
    - Handle ISO-8859-1 decoding
    - Graceful fallback on invalid sequences
  - [x] `Merge(target, source, allowPrototypes)` - Deep merge objects
  - [x] `Compact(value)` - Remove sparse array holes
  - [x] `ArrayToObject(source)` - Convert array to object
  - [x] `IsRegExp(obj)` - Check if value is regexp
  - [x] `Combine(a, b)` - Combine arrays
  - [x] `MaybeMap(val, fn)` - Map function over value or array
  - [x] `Assign(target, source)` - Copy map keys
  - **Tests**: ✅ All passing (utils_test.go)
    - Encode special characters correctly
    - Decode URL-encoded strings
    - Merge nested objects
    - Compact sparse arrays
    - Handle edge cases (empty strings, nil values)
    - Round-trip encode/decode

---

### Phase 2: Parse Implementation

- [ ] **2.1 Parse Options (parse.go)**
  ```go
  type ParseOptions struct {
      AllowDots                bool
      AllowEmptyArrays         bool
      AllowPrototypes          bool  // Go: less relevant, but for compatibility
      AllowSparse              bool
      ArrayLimit               int   // Default: 20
      Charset                  Charset // utf-8 | iso-8859-1
      CharsetSentinel          bool
      Comma                    bool
      DecodeDotInKeys          bool
      Decoder                  DecoderFunc
      Delimiter                string // Default: &, supports regex
      Depth                    int    // Default: 5
      Duplicates               DuplicateHandling // combine | first | last
      IgnoreQueryPrefix        bool
      InterpretNumericEntities bool
      ParameterLimit           int   // Default: 1000
      ParseArrays              bool  // Default: true
      PlainObjects             bool  // Go: less relevant
      StrictDepth              bool
      StrictNullHandling       bool
      ThrowOnLimitExceeded     bool
  }
  ```
  - **Tests**:
    - Default options are correct
    - Custom options override defaults
    - Invalid option values return errors

- [ ] **2.2 Core Parse Function**
  - [ ] `Parse(str string, opts ...ParseOptions) (map[string]any, error)`
  - [ ] Handle empty/nil/undefined input
  - [ ] Strip query prefix when `IgnoreQueryPrefix` is true
  - [ ] Split by delimiter (string or regex)
  - [ ] Respect `ParameterLimit`
  - [ ] Handle charset sentinel detection
  - **Tests**:
    - Parse simple key-value: `a=b` → `{a: "b"}`
    - Parse multiple params: `a=b&c=d`
    - Handle empty values: `a=` → `{a: ""}`
    - Handle missing values: `a` → `{a: ""}` or `{a: null}` with strictNullHandling

- [ ] **2.3 Nested Object Parsing**
  - [ ] `parseKeys(key, val, options)` - Parse bracket notation
  - [ ] Support bracket notation: `a[b][c]=d`
  - [ ] Support dot notation: `a.b.c=d` (when AllowDots=true)
  - [ ] Respect depth limit
  - [ ] Handle `StrictDepth` errors
  - **Tests**:
    - Nested object: `a[b][c]=d` → `{a: {b: {c: "d"}}}`
    - Depth limit: `a[b][c][d][e][f][g][h]=i` with depth=5
    - Dot notation: `a.b.c=d` with AllowDots
    - DecodeDotInKeys: `a%2Eb=c` → `{a.b: "c"}`

- [ ] **2.4 Array Parsing**
  - [ ] Parse bracket arrays: `a[]=b&a[]=c`
  - [ ] Parse indexed arrays: `a[0]=b&a[1]=c`
  - [ ] Respect `ArrayLimit`
  - [ ] Handle sparse arrays with `AllowSparse`
  - [ ] Compact sparse arrays when `AllowSparse=false`
  - [ ] Convert to object when exceeding ArrayLimit
  - **Tests**:
    - Bracket arrays: `a[]=b&a[]=c` → `{a: ["b", "c"]}`
    - Indexed arrays: `a[0]=b&a[1]=c` → `{a: ["b", "c"]}`
    - Mixed arrays: `a[]=b&a[0]=c`
    - ArrayLimit exceeded: convert to object
    - Sparse arrays: `a[1]=b&a[10]=c`
    - Arrays with objects: `a[0][b]=c`

- [ ] **2.5 Special Parsing Features**
  - [ ] Comma-separated values: `a=b,c,d` → `{a: ["b", "c", "d"]}`
  - [ ] Duplicate key handling: combine, first, last
  - [ ] Interpret numeric entities: `&#9786;` → `☺`
  - [ ] Charset detection via sentinel
  - [ ] URL-encoded bracket handling: `%5B`, `%5D`
  - **Tests**:
    - Comma parsing: `a=b,c` with `Comma=true`
    - Duplicates combine: `a=b&a=c` → `{a: ["b", "c"]}`
    - Duplicates first: `a=b&a=c` → `{a: "b"}`
    - Duplicates last: `a=b&a=c` → `{a: "c"}`
    - Numeric entities in ISO-8859-1

- [ ] **2.6 Prototype Protection**
  - [ ] Prevent `__proto__` injection
  - [ ] Handle `hasOwnProperty` keys
  - [ ] `AllowPrototypes` option
  - **Tests**:
    - `__proto__` is ignored
    - `hasOwnProperty` handling
    - `constructor[prototype]` protection

---

### Phase 3: Stringify Implementation

- [ ] **3.1 Stringify Options (stringify.go)**
  ```go
  type StringifyOptions struct {
      AddQueryPrefix     bool
      AllowDots          bool
      AllowEmptyArrays   bool
      ArrayFormat        ArrayFormat // indices | brackets | repeat | comma
      Charset            Charset
      CharsetSentinel    bool
      CommaRoundTrip     bool
      Delimiter          string // Default: &
      Encode             bool   // Default: true
      EncodeDotInKeys    bool
      Encoder            EncoderFunc
      EncodeValuesOnly   bool
      Filter             any    // func or []string
      Format             Format // RFC1738 | RFC3986
      Formatter          FormatterFunc
      SerializeDate      func(time.Time) string
      SkipNulls          bool
      Sort               func(a, b string) bool
      StrictNullHandling bool
  }
  ```
  - **Tests**: Same as parse options tests

- [ ] **3.2 Core Stringify Function**
  - [ ] `Stringify(obj any, opts ...StringifyOptions) (string, error)`
  - [ ] Handle nil/empty input → `""`
  - [ ] Handle primitive values
  - [ ] Add query prefix when requested
  - [ ] Join with delimiter
  - [ ] Add charset sentinel when requested
  - **Tests**:
    - Simple object: `{a: "b"}` → `a=b`
    - Multiple keys: `{a: "b", c: "d"}` → `a=b&c=d`
    - Query prefix: `?a=b`
    - Empty object: `{}` → `""`

- [ ] **3.3 Nested Object Stringify**
  - [ ] Bracket notation: `{a: {b: "c"}}` → `a[b]=c`
  - [ ] Dot notation: `{a: {b: "c"}}` → `a.b=c` (with AllowDots)
  - [ ] EncodeDotInKeys: `{a.b: "c"}` → `a%2Eb=c`
  - [ ] Deep nesting support
  - **Tests**:
    - Nested object: `{a: {b: "c"}}` → `a%5Bb%5D=c`
    - Dot notation: `{a: {b: "c"}}` → `a.b=c`
    - EncodeDotInKeys with nested keys
    - Very deep objects

- [ ] **3.4 Array Stringify**
  - [ ] Indices format: `a[0]=b&a[1]=c`
  - [ ] Brackets format: `a[]=b&a[]=c`
  - [ ] Repeat format: `a=b&a=c`
  - [ ] Comma format: `a=b,c`
  - [ ] `CommaRoundTrip` for single-element arrays
  - [ ] `AllowEmptyArrays` option
  - [ ] Sparse array handling
  - **Tests**:
    - All array formats
    - Empty arrays with/without AllowEmptyArrays
    - Single element with CommaRoundTrip
    - Arrays of objects
    - Nested arrays

- [ ] **3.5 Encoding Features**
  - [ ] URL encoding with RFC1738/RFC3986
  - [ ] `EncodeValuesOnly` option
  - [ ] Custom encoder function
  - [ ] ISO-8859-1 charset support
  - [ ] Numeric entity encoding for non-representable chars
  - **Tests**:
    - Space encoding: `%20` vs `+`
    - Special characters
    - EncodeValuesOnly: keys not encoded
    - Custom encoder function
    - ISO-8859-1 characters

- [ ] **3.6 Special Features**
  - [ ] Filter function/array
  - [ ] Sort function
  - [ ] SerializeDate function
  - [ ] SkipNulls option
  - [ ] StrictNullHandling
  - [ ] Cyclic reference detection
  - **Tests**:
    - Filter with array of keys
    - Filter with function
    - Sort keys
    - Date serialization
    - Null handling options
    - Cyclic reference error

---

### Phase 4: Go-Specific Features (from v1)

- [ ] **4.1 Struct Support**
  - [ ] `query` struct tag support
  - [ ] `ParseToStruct(str string, dest any, opts ...ParseOptions) error`
  - [ ] `StructToQueryString(obj any, opts ...StringifyOptions) (string, error)`
  - [ ] Nested struct support
  - [ ] Pointer field support
  - [ ] Slice/array field support
  - [ ] Map field support
  - **Tests**:
    - Simple struct parse/stringify
    - Nested struct parse/stringify
    - Pointer fields
    - Slice fields
    - Map fields
    - Field tag options

- [ ] **4.2 Marshal/Unmarshal API**
  - [ ] `Marshal(v any, opts ...StringifyOptions) (string, error)`
  - [ ] `Unmarshal(str string, v any, opts ...ParseOptions) error`
  - [ ] Automatic type detection
  - [ ] Support for all Go types
  - **Tests**:
    - Round-trip marshal/unmarshal
    - Various Go types
    - Error handling

- [ ] **4.3 Type Conversions**
  - [ ] String to int/float/bool conversion
  - [ ] Time parsing/formatting
  - [ ] Custom type support via interfaces
  - **Tests**:
    - All basic type conversions
    - Edge cases (empty strings, invalid numbers)
    - Custom types

---

### Phase 5: Advanced Features & Edge Cases

- [ ] **5.1 Error Handling**
  - [ ] `ParameterLimitExceeded` error
  - [ ] `ArrayLimitExceeded` error
  - [ ] `DepthLimitExceeded` error
  - [ ] `InvalidCharset` error
  - [ ] `CyclicReference` error
  - [ ] Graceful handling vs throwing (`ThrowOnLimitExceeded`)
  - **Tests**:
    - All error conditions
    - Error messages match JS behavior

- [ ] **5.2 Edge Cases**
  - [ ] Empty keys: `[]=b`
  - [ ] Keys starting with brackets: `[foo]=bar`
  - [ ] Malformed URI characters
  - [ ] Very long strings (performance)
  - [ ] Unicode characters (emoji, CJK)
  - [ ] Special characters in keys/values
  - **Tests**:
    - All edge cases from JS test suite
    - Unicode handling
    - Performance benchmarks

- [ ] **5.3 Charset Support**
  - [ ] UTF-8 (default)
  - [ ] ISO-8859-1
  - [ ] Charset sentinel auto-detection
  - [ ] Numeric entity interpretation
  - **Tests**:
    - Both charsets
    - Sentinel detection
    - Character conversion

---

### Phase 6: Documentation & Polish

- [ ] **6.1 API Documentation**
  - [ ] GoDoc comments for all exported types/functions
  - [ ] Usage examples in documentation
  - [ ] Migration guide from v1

- [ ] **6.2 README Updates**
  - [ ] Feature list
  - [ ] Quick start guide
  - [ ] API reference
  - [ ] Comparison with JS library

- [ ] **6.3 Examples**
  - [ ] Basic usage
  - [ ] Struct parsing
  - [ ] Framework integration (Gin, Echo)
  - [ ] Strapi-style queries

---

### Phase 7: Testing & Quality

- [ ] **7.1 Unit Tests**
  - [ ] Port all tests from JS test suite
  - [ ] Go-specific feature tests
  - [ ] Target >95% coverage

- [ ] **7.2 Integration Tests**
  - [ ] Round-trip parse/stringify compatibility
  - [ ] Cross-compatibility with JS library

- [ ] **7.3 Benchmarks**
  - [ ] Parse benchmarks (simple, complex, nested)
  - [ ] Stringify benchmarks
  - [ ] Memory allocation analysis
  - [ ] Comparison with v1

- [ ] **7.4 Fuzz Testing**
  - [ ] Parse fuzzing
  - [ ] Stringify fuzzing

---

## File Structure

```
v2/
├── formats.go       # Format constants and formatters
├── formats_test.go
├── utils.go         # Utility functions
├── utils_test.go
├── parse.go         # Parse implementation
├── parse_test.go
├── stringify.go     # Stringify implementation
├── stringify_test.go
├── struct.go        # Go struct support
├── struct_test.go
├── marshal.go       # Marshal/Unmarshal API
├── marshal_test.go
├── errors.go        # Error types
├── qs.go            # Main entry point, exports
├── doc.go           # Package documentation
├── benchmark_test.go
└── IMPLEMENTATION_PLAN.md
```

---

## Type Definitions

```go
// Charset represents supported character sets
type Charset string
const (
    CharsetUTF8     Charset = "utf-8"
    CharsetISO88591 Charset = "iso-8859-1"
)

// ArrayFormat represents array stringification formats
type ArrayFormat string
const (
    ArrayFormatIndices  ArrayFormat = "indices"
    ArrayFormatBrackets ArrayFormat = "brackets"
    ArrayFormatRepeat   ArrayFormat = "repeat"
    ArrayFormatComma    ArrayFormat = "comma"
)

// DuplicateHandling represents how duplicate keys are handled
type DuplicateHandling string
const (
    DuplicateCombine DuplicateHandling = "combine"
    DuplicateFirst   DuplicateHandling = "first"
    DuplicateLast    DuplicateHandling = "last"
)

// DecoderFunc is a custom decoder function signature
type DecoderFunc func(str string, defaultDecoder func(string) (string, error), charset Charset, kind string) (string, error)

// EncoderFunc is a custom encoder function signature
type EncoderFunc func(str string, defaultEncoder func(string) string, charset Charset, kind string, format Format) string

// FormatterFunc is a custom formatter function signature
type FormatterFunc func(str string) string
```

---

## Compatibility Notes

### Must Match JS Behavior
- All parse options and their effects
- All stringify options and their effects
- Array format outputs
- Encoding outputs (RFC1738 vs RFC3986)
- Error messages and conditions
- Edge case handling

### Go-Specific Extensions
- Struct tag support (`query:"fieldname"`)
- Marshal/Unmarshal functions
- Strong typing with generics where appropriate
- Context support for cancellation (optional)
- Concurrent-safe operations

---

## Timeline Estimate

| Phase | Description | Complexity |
|-------|-------------|------------|
| 1 | Core Infrastructure | Low |
| 2 | Parse Implementation | High |
| 3 | Stringify Implementation | High |
| 4 | Go-Specific Features | Medium |
| 5 | Advanced Features | Medium |
| 6 | Documentation | Low |
| 7 | Testing & Quality | Medium |

---

## Success Criteria

1. **Full JS Compatibility**: All JS library tests pass when ported
2. **Go Idioms**: Code follows Go conventions and best practices
3. **Performance**: Equal or better performance than v1
4. **Coverage**: >95% test coverage
5. **Documentation**: Complete GoDoc with examples
6. **Backwards Compatible**: v1 API patterns preserved where sensible
