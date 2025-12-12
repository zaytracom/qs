# QS Go Library - Full Compatibility Roadmap

This document describes all changes required for full compatibility between the Go implementation and the original Node.js [qs](https://github.com/ljharb/qs) library.

## Table of Contents
1. [Critical Changes](#critical-changes)
2. [Parse Options](#parse-options)
3. [Stringify Options](#stringify-options)
4. [General Improvements](#general-improvements)
5. [Testing](#testing)
6. [Implementation Phases](#implementation-phases)

---

## Critical Changes

### 1. Fix Options Merging

**Problem:** Current implementation uses `if custom.X { options.X = custom.X }`, which doesn't work for boolean `false` values.

**File:** `v1/qs.go:331-397` (Parse) and `v1/qs.go:991-1054` (Stringify)

**Solution:** Use pointers or a "was set" flag pattern:

```go
type ParseOptions struct {
    AllowDots     *bool  // nil = not set, use default
    // or
    allowDotsSet  bool   // private flag
    AllowDots     bool
}
```

**Alternative:** Use functional options pattern:

```go
func WithAllowDots(v bool) ParseOption {
    return func(o *ParseOptions) {
        o.AllowDots = v
        o.allowDotsSet = true
    }
}
```

### 2. Convert Map with Numeric Keys to Array

**Problem:**
```
Node.js: a[0]=b&a[1]=c → { a: ['b', 'c'] }
Go:      a[0]=b&a[1]=c → { a: { "0": "b", "1": "c" } }
```

**File:** `v1/qs.go` - function `parseObject`

**Solution:** After parsing, check if map can be converted to slice:

```go
func maybeConvertToArray(obj map[string]interface{}, options *ParseOptions) interface{} {
    if !options.ParseArrays {
        return obj
    }

    // Check that all keys are numbers from 0 to N-1
    // and max index <= arrayLimit
    allNumeric := true
    maxIndex := -1
    for k := range obj {
        idx, err := strconv.Atoi(k)
        if err != nil || idx < 0 {
            allNumeric = false
            break
        }
        if idx > maxIndex {
            maxIndex = idx
        }
    }

    if !allNumeric || maxIndex > options.ArrayLimit {
        return obj
    }

    // Convert to slice
    result := make([]interface{}, maxIndex+1)
    for k, v := range obj {
        idx, _ := strconv.Atoi(k)
        result[idx] = v
    }

    // If not allowSparse, compact the array
    if !options.AllowSparse {
        return compact(result)
    }
    return result
}
```

---

## Parse Options

### 3. Option `comma` - Parse Comma-Separated Values

**Priority:** High

**Description:** When `comma: true`, value `a=1,2,3` should parse as `{ a: ['1', '2', '3'] }`.

**File:** `v1/qs.go` - add to `Parse`

**Implementation:**

```go
func parseArrayValue(val string, options *ParseOptions) interface{} {
    if options.Comma && strings.Contains(val, ",") {
        parts := strings.Split(val, ",")
        result := make([]interface{}, len(parts))
        for i, p := range parts {
            result[i] = p
        }
        return result
    }
    return val
}
```

**Important edge cases:**
- URL-encoded comma `%2C` should NOT split: `a=b%2Cc` → `{ a: 'b,c' }`
- With brackets: `foo[]=1,2,3&foo[]=4,5,6` → `{ foo: [['1','2','3'], ['4','5','6']] }`

### 4. Option `duplicates` - Handle Duplicate Keys

**Priority:** Medium

**Description:** Three modes: `combine` (default), `first`, `last`

**File:** `v1/qs.go:134-139`

**Implementation:**

```go
func handleDuplicate(existing, newVal interface{}, duplicates string) interface{} {
    switch duplicates {
    case "first":
        return existing
    case "last":
        return newVal
    case "combine":
        fallthrough
    default:
        return merge(existing, newVal)
    }
}
```

### 5. Option `strictDepth` - Error on Depth Exceeded

**Priority:** Medium

**Description:** When `strictDepth: true` and depth is exceeded, throw error instead of truncating.

**File:** `v1/qs.go` - function `parseObject`

**Implementation:**

```go
if len(chain) > depth+1 {
    if options.StrictDepth {
        return fmt.Errorf("input depth exceeded depth option of %d and strictDepth is true", options.Depth)
    }
    // current truncation logic
}
```

### 6. Option `allowSparse` - Sparse Arrays

**Priority:** Low

**Description:** When `allowSparse: true`, preserve empty array elements:
```
a[4]=1&a[1]=2 → [nil, '2', nil, nil, '1']
```

**Current behavior (compact):**
```
a[4]=1&a[1]=2 → ['2', '1']
```

**Implementation:** Add `compact()` function and call it only when `allowSparse: false`

### 7. Option `charsetSentinel` and `charset: 'iso-8859-1'`

**Priority:** Low

**Description:** Support charset detection from query string:
- `utf8=%E2%9C%93` → UTF-8
- `utf8=%26%2310003%3B` → ISO-8859-1

**Implementation:**

```go
const (
    charsetSentinel = "utf8=%E2%9C%93"      // encodeURIComponent('✓')
    isoSentinel     = "utf8=%26%2310003%3B" // encodeURIComponent('&#10003;')
)

func detectCharset(str string, options *ParseOptions) string {
    if !options.CharsetSentinel {
        return options.Charset
    }
    if strings.Contains(str, charsetSentinel) {
        return "utf-8"
    }
    if strings.Contains(str, isoSentinel) {
        return "iso-8859-1"
    }
    return options.Charset
}
```

### 8. Option `interpretNumericEntities`

**Priority:** Low

**Description:** Convert HTML numeric entities to characters:
```
&#9786; → ☺
```

**Implementation:**

```go
var numericEntityRegex = regexp.MustCompile(`&#(\d+);`)

func interpretNumericEntities(str string) string {
    return numericEntityRegex.ReplaceAllStringFunc(str, func(match string) string {
        numStr := numericEntityRegex.FindStringSubmatch(match)[1]
        num, _ := strconv.Atoi(numStr)
        return string(rune(num))
    })
}
```

### 9. Option `allowPrototypes`

**Priority:** Low (Go-specific)

**Description:** In JS this prevents setting properties like `__proto__`, `constructor`, etc.
In Go this is not as critical, but for compatibility we can filter:

```go
var prototypeProperties = map[string]bool{
    "__proto__":      true,
    "constructor":    true,
    "hasOwnProperty": true,
    "toString":       true,
}

func isPrototypeProperty(key string) bool {
    return prototypeProperties[key]
}
```

### 10. Option `parseArrays: false`

**Priority:** Medium

**Description:** When `parseArrays: false`, `a[0]=b` → `{ a: { "0": "b" } }`, not `{ a: ['b'] }`

**Current state:** Option exists but not fully implemented

### 11. Decoder with Type (key/value)

**Priority:** Low

**Description:** Custom decoder should receive type: `key` or `value`

```go
type Decoder func(str string, defaultDecoder func(string) (string, error), charset string, keyType string) (string, error)
```

**Example usage:**
```go
decoder := func(str, _, charset, keyType string) (string, error) {
    if keyType == "key" {
        return strings.ToLower(str), nil
    }
    return strings.ToUpper(str), nil
}
```

### 12. RegExp Delimiter Support

**Priority:** Low

**Description:** In JS you can pass RegExp as delimiter:
```js
qs.parse('a=b; c=d', { delimiter: /[;,] */ })
```

**Go implementation:**

```go
type ParseOptions struct {
    Delimiter       string
    DelimiterRegexp *regexp.Regexp
}

func splitByDelimiter(str string, options *ParseOptions) []string {
    if options.DelimiterRegexp != nil {
        return options.DelimiterRegexp.Split(str, -1)
    }
    return strings.Split(str, options.Delimiter)
}
```

### 13. Parse Objects (not just strings)

**Priority:** Low

**Description:** In JS you can pass an object for additional parsing:
```js
qs.parse({ 'a[b]': 'c' }) // { a: { b: 'c' } }
```

**Go implementation:** Add overload or type check

---

## Stringify Options

### 14. Option `arrayFormat: 'comma'`

**Priority:** High

**Description:** Serialize arrays with comma separator:
```
{ a: ['b', 'c'] } → 'a=b,c'
```

**File:** `v1/qs.go` - function `stringify`

**Implementation:**

```go
case "comma":
    if len(arr) == 0 {
        return
    }
    values := make([]string, len(arr))
    for i, v := range arr {
        values[i] = fmt.Sprintf("%v", v)
    }
    *parts = append(*parts, prefix+"="+strings.Join(values, ","))
```

### 15. Option `commaRoundTrip`

**Priority:** Medium

**Description:** When `commaRoundTrip: true` and single element, add `[]`:
```
{ a: ['c'] } with comma → 'a[]=c' (so it parses back as array)
```

### 16. Option `skipNulls`

**Priority:** Medium

**Description:** Skip nil/null values:
```go
{ a: 'b', c: nil } with skipNulls → 'a=b'
```

**Current state:** Option declared but not used

### 17. Option `filter` (function or array)

**Priority:** Medium

**Description:**
1. As array: `filter: ['a', 'b']` - include only these keys
2. As function: `filter: func(prefix, value)` - custom filtering

**Implementation:**

```go
type FilterFunc func(prefix string, value interface{}) interface{}

type StringifyOptions struct {
    Filter     interface{} // []string or FilterFunc
}

func applyFilter(prefix string, value interface{}, filter interface{}) (interface{}, bool) {
    switch f := filter.(type) {
    case []string:
        for _, allowed := range f {
            if prefix == allowed {
                return value, true
            }
        }
        return nil, false
    case FilterFunc:
        result := f(prefix, value)
        return result, result != nil
    }
    return value, true
}
```

### 18. Option `sort`

**Priority:** Medium

**Description:** Sort keys:
```go
sort := func(a, b string) bool { return a < b }
qs.Stringify(data, &StringifyOptions{Sort: sort})
```

**Current state:** Option declared but not used

**Implementation:** Sort keys before iteration:

```go
keys := make([]string, 0, len(m))
for k := range m {
    keys = append(keys, k)
}
if options.Sort != nil {
    sort.Slice(keys, func(i, j int) bool {
        return options.Sort(keys[i], keys[j])
    })
}
for _, k := range keys {
    stringify(parts, m[k], options, newPrefix)
}
```

### 19. Option `encodeValuesOnly`

**Priority:** Medium

**Description:** Encode only values, not keys:
```
{ a: 'b c' } → 'a=b%20c'
{ 'a b': 'c' } → 'a b=c' (not 'a%20b=c')
```

**Implementation:**

```go
func stringify(...) {
    key := prefix
    if !options.EncodeValuesOnly {
        key = options.Encoder(prefix)
    }
    val := options.Encoder(fmt.Sprintf("%v", v))
    *parts = append(*parts, key+"="+val)
}
```

### 20. Option `encode: false`

**Priority:** Medium

**Description:** Disable URL encoding completely

**Current state:** Option declared but ignored

### 21. Option `allowEmptyArrays`

**Priority:** Low

**Description:** When `allowEmptyArrays: true`:
```
{ a: [], b: 'c' } → 'a[]&b=c'
```

By default empty arrays are skipped.

### 22. RFC1738 and RFC3986 Formats

**Priority:** Low

**Description:**
- RFC1738: spaces as `+`
- RFC3986: spaces as `%20` (default)

**Implementation:**

```go
type Format string

const (
    FormatRFC1738 Format = "RFC1738"
    FormatRFC3986 Format = "RFC3986"
)

func getFormatter(format Format) func(string) string {
    switch format {
    case FormatRFC1738:
        return func(s string) string {
            return strings.ReplaceAll(s, "%20", "+")
        }
    default:
        return func(s string) string { return s }
    }
}
```

### 23. Option `encodeDotInKeys`

**Priority:** Low

**Description:** Encode dots in keys:
```
{ 'name.obj': { first: 'John' } } with allowDots + encodeDotInKeys
→ 'name%2Eobj.first=John'
```

### 24. Option `charsetSentinel` for Stringify

**Priority:** Low

**Description:** Add sentinel at the beginning of query string:
```
{ a: 'b' } with charsetSentinel + utf-8 → 'utf8=%E2%9C%93&a=b'
```

### 25. Encoder with Type (key/value)

**Priority:** Low

**Description:** Similar to decoder - encoder receives type

```go
type Encoder func(str string, defaultEncoder func(string) string, charset string, keyType string) string
```

### 26. Cyclic Reference Handling

**Priority:** Low

**Description:** In JS, side-channel is used for cycle tracking.

**Go implementation:**

```go
func stringifyWithCycleDetection(obj interface{}, seen map[uintptr]bool) error {
    ptr := reflect.ValueOf(obj).Pointer()
    if seen[ptr] {
        return errors.New("cyclic object value")
    }
    seen[ptr] = true
    defer delete(seen, ptr)
    // ... stringify logic
}
```

### 27. Buffer ([]byte) Support

**Priority:** Low

**Description:** Handle `[]byte` as string

```go
case []byte:
    *parts = append(*parts, prefix+"="+options.Encoder(string(v)))
```

---

## General Improvements

### 28. Options Validation

**Priority:** High

**Description:** Validate option types and values:

```go
func validateParseOptions(opts *ParseOptions) error {
    if opts.Charset != "" && opts.Charset != "utf-8" && opts.Charset != "iso-8859-1" {
        return errors.New("charset must be utf-8, iso-8859-1, or empty")
    }
    if opts.Duplicates != "" && opts.Duplicates != "combine" &&
       opts.Duplicates != "first" && opts.Duplicates != "last" {
        return errors.New("duplicates must be combine, first, or last")
    }
    return nil
}
```

### 29. Don't Mutate Input Options

**Priority:** Medium

**Description:** Create a copy of options instead of modifying the passed ones

### 30. Export Utilities

**Priority:** Low

**Description:** Export `Encode`, `Decode` for compatibility:

```go
var (
    Encode = defaultEncoder
    Decode = defaultDecoder
)
```

### 31. Export Format Constants

**Priority:** Low

```go
var Formats = struct {
    RFC1738 Format
    RFC3986 Format
    Default Format
}{
    RFC1738: FormatRFC1738,
    RFC3986: FormatRFC3986,
    Default: FormatRFC3986,
}
```

---

## Testing

### 32. Port Tests from Node.js

**Priority:** High

**Description:** Port all tests from `.ref/test/parse.js` and `.ref/test/stringify.js`

**Number of tests to port:**
- `parse.js`: ~150 test cases
- `stringify.js`: ~120 test cases
- `empty-keys-cases.js`: ~20 test cases

### 33. Add Benchmark Comparison

**Priority:** Low

**Description:** Compare performance with Node.js implementation

---

## Implementation Phases

### Phase 1: Critical (2-3 days)
| # | Task | Priority | Effort |
|---|------|----------|--------|
| 1 | Fix options merging (boolean false) | Critical | 4h |
| 2 | Convert map → array for numeric keys | Critical | 6h |
| 3 | Option `comma` for parsing | High | 4h |
| 4 | Option `arrayFormat: 'comma'` for stringify | High | 3h |
| 5 | Port tests from Node.js | High | 8h |

### Phase 2: Important (3-4 days)
| # | Task | Priority | Effort |
|---|------|----------|--------|
| 6 | Option `duplicates` | Medium | 2h |
| 7 | Option `strictDepth` | Medium | 2h |
| 8 | Option `skipNulls` | Medium | 2h |
| 9 | Option `filter` | Medium | 4h |
| 10 | Option `sort` | Medium | 2h |
| 11 | Option `encodeValuesOnly` | Medium | 3h |
| 12 | Option `encode: false` | Medium | 2h |
| 13 | Option `parseArrays: false` | Medium | 2h |
| 14 | Option `commaRoundTrip` | Medium | 2h |

### Phase 3: Full Compatibility (5-7 days)
| # | Task | Priority | Effort |
|---|------|----------|--------|
| 15 | Option `allowSparse` | Low | 3h |
| 16 | Option `allowEmptyArrays` | Low | 2h |
| 17 | Charset support (iso-8859-1, charsetSentinel) | Low | 4h |
| 18 | Option `interpretNumericEntities` | Low | 2h |
| 19 | RFC1738/RFC3986 formats | Low | 2h |
| 20 | Option `encodeDotInKeys` | Low | 2h |
| 21 | Decoder/Encoder with key/value type | Low | 3h |
| 22 | Cyclic reference handling | Low | 3h |
| 23 | RegExp delimiter | Low | 2h |
| 24 | Parse objects (not just strings) | Low | 2h |
| 25 | Options validation | Low | 2h |
| 26 | Don't mutate input options | Low | 1h |
| 27 | Export utilities and constants | Low | 1h |
| 28 | Buffer ([]byte) support | Low | 1h |
| 29 | Benchmark comparison | Low | 2h |

---

## Summary

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1 | 5 tasks | 2-3 days |
| Phase 2 | 9 tasks | 3-4 days |
| Phase 3 | 15 tasks | 5-7 days |
| **Total** | **29 tasks** | **10-14 days** |

---

## Current Compatibility Status

### Fully Implemented ✅
- Basic key=value parsing
- Nested objects `a[b][c]=d`
- Arrays with `[]` notation `a[]=1&a[]=2`
- `allowDots` - dot notation `a.b.c=d`
- `strictNullHandling` - keys without values as `nil`
- `ignoreQueryPrefix` - ignore leading `?`
- Custom delimiters (`delimiter`)
- URL decoding
- Depth limiting (`depth`)
- `arrayFormat`: indices, brackets, repeat
- `addQueryPrefix`
- Struct tags `query:"fieldname"`
- Go-specific: `Marshal/Unmarshal` idiomatic functions

### Partially Implemented ⚠️
- `decodeDotInKeys` - option exists, partial implementation
- `allowEmptyArrays` - option exists, not fully working
- `parseArrays` - option exists, not fully implemented

### Not Implemented ❌
- `comma` (parse `a=1,2,3` as array)
- `arrayFormat: 'comma'` for stringify
- `charsetSentinel` / ISO-8859-1
- `interpretNumericEntities`
- `duplicates` (first/last/combine)
- `allowSparse`
- `strictDepth`
- `filter` (function/array)
- `sort`
- `encodeValuesOnly`
- `encode: false`
- `commaRoundTrip`
- `encodeDotInKeys` for stringify
- RFC1738 format
- Cyclic reference detection
- Decoder/Encoder with type parameter

---

## References

- [Original qs on npm](https://www.npmjs.com/package/qs)
- [GitHub repository](https://github.com/ljharb/qs)
- [API Documentation](https://github.com/ljharb/qs#readme)
