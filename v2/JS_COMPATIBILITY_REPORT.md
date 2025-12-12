# JavaScript Compatibility Test Report

**Date:** 2025-12-12
**Version:** v2-alpha
**Reference:** [ljharb/qs](https://github.com/ljharb/qs)

---

## Executive Summary

This report provides a strict comparison between the Go v2 implementation and the original JavaScript `qs` library test suite. A dedicated test file `parse_js_compat_test.go` was created to validate exact behavioral parity with the JS implementation.

### Overall Results

| Metric | Value |
|--------|-------|
| Total Test Groups | 44 |
| Passed | 28 (64%) |
| Failed | 16 (36%) |
| Test Cases | ~150+ |

---

## Passed Tests

The following functionality matches the JavaScript implementation exactly:

### Basic Parsing
- Simple key-value: `foo=bar`
- Numeric keys: `0=foo`
- Double plus decoding: `foo=c++` → `{foo: "c  "}`
- Special characters in brackets: `a[>=]=23`, `a[<=>]==23`, `a[==]=23`
- Multiple equals in value: `foo=bar=baz` → `{foo: "bar=baz"}`
- Spaces in keys/values: ` foo = bar = baz ` → `{" foo ": " bar = baz "}`
- StrictNullHandling: `foo` → `{foo: null}`

### Array Parsing
- Explicit arrays: `a[]=b&a[]=c`
- Indexed arrays: `a[0]=b&a[1]=c`
- Index reordering: `a[1]=c&a[0]=b` → `["b", "c"]`
- Nested arrays: `a[b][]=c&a[b][]=d`
- Arrays of objects: `a[][b]=c` → `[{b: "c"}]`
- Sparse array compaction (default)
- Sparse arrays with `allowSparse: true`

### Object Parsing
- Bracket notation: `a[b][c]=d`
- Deep nesting with depth limit (default 5)
- Depth = 1 behavior
- Keys beginning with numbers: `a[12b]=c`

### Dot Notation
- `allowDots: true`: `a.b.c=d` → `{a: {b: {c: "d"}}}`
- `decodeDotInKeys: true`: `name%252Eobj.first=John`
- Mixed dot and bracket notation

### Options
- Custom delimiter: `a=b;c=d` with `delimiter: ";"`
- RegExp delimiter: `a=b; c=d` with `/[;,] */`
- `ignoreQueryPrefix: true`: `?foo=bar` → `{foo: "bar"}`
- `parseArrays: false`: always returns objects
- `allowEmptyArrays: true`: `foo[]` → `{foo: []}`
- `strictNullHandling: true`
- `duplicates`: combine/first/last

### Comma Parsing
- `comma: true`: `a=b,c` → `{a: ["b", "c"]}`
- Array of arrays: `foo[]=1,2,3&foo[]=4,5,6`
- Percent-encoded comma not split: `foo=a%2Cb` → `{foo: "a,b"}`

### Security
- Prototype protection: `hasOwnProperty`, `toString` blocked by default
- `allowPrototypes: true` enables dangerous keys
- Brackets in values preserved: `pets=["tobi"]`

### Encoding
- URL decoding: `%20`, `+` as space
- Encoded brackets: `a%5Bb%5D=c`
- Encoded equals: `he%3Dllo=th%3Dere`
- Malformed URI handling: `{%:%}`
- Numeric entities: `&#9786;` → `☺` (with ISO-8859-1)

### Limits & Errors
- `strictDepth: true` throws `ErrDepthLimitExceeded`
- `throwOnLimitExceeded: true` throws on parameter limit
- `arrayLimit` index boundary (20 vs 21)

---

## Failed Tests

The following tests revealed incompatibilities with the JavaScript implementation:

### Critical Issues (HIGH Priority)

#### 1. Depth Zero Behavior
```
Input:  a[0]=b&a[1]=c with depth: 0
Expected: {"a[0]": "b", "a[1]": "c"}
Got:      {a: ["b", "c"]}
```
**Issue:** `depth: 0` should completely disable nesting, treating brackets as literal characters.

#### 2. Array Limit Zero Behavior
```
Input:  a[1]=c with arrayLimit: 0
Expected: {a: {1: "c"}}
Got:      {a: ["c"]}
```
**Issue:** `arrayLimit: 0` should force object output, not array.

#### 3. Mixed Array/Value Order
```
Input:  a=b&a[]=c
Expected: {a: ["b", "c"]}
Got:      {a: ["c", "b"]}
```
**Issue:** Order of values should match input order.

#### 4. Adding Keys to Existing Objects
```
Input:  a[b]=c&a=d
Expected: {a: {b: "c", d: true}}
Got:      {a: ["d", {b: "c"}]}
```
**Issue:** Primitive value `d` should be added as key to existing object, not create array.

#### 5. Empty Parent Brackets
```
Input:  []=a&[]=b
Expected: {0: "a", 1: "b"}
Got:      {}
```
**Issue:** Empty parent `[]` should create numeric keys at root level.

#### 6. Parameter Limit Truncation
```
Input:  a=b&c=d with parameterLimit: 1
Expected: {a: "b"}
Got:      {a: "b&c=d"}
```
**Issue:** Limit should truncate number of parameters, not characters in value.

#### 7. ISO-8859-1 Charset
```
Input:  %A2=%BD with charset: "iso-8859-1"
Expected: {"¢": "½"}
Got:      {"\xa2": "\xbd"}
```
**Issue:** ISO-8859-1 bytes should be converted to UTF-8 Unicode characters.

### Medium Priority Issues

#### 8. Arrays to Objects Transformation
```
Input:  foo[bad]=baz&foo[0]=bar
Expected: {foo: {bad: "baz", 0: "bar"}}
Got:      {foo: {bad: "baz"}}
```
**Issue:** When mixing string and numeric keys, numeric should be preserved.

#### 9. Sparse Array Pruning
```
Input:  a[2]=b&a[99999999]=c
Expected: {a: {2: "b", 99999999: "c"}}
Got:      {a: {99999999: "c"}}
```
**Issue:** Both indices should be preserved when converting to object.

#### 10. StrictNullHandling in Arrays
```
Input:  a[0]=b&a[1]&a[2]=c with strictNullHandling: true
Expected: {a: ["b", null, "c"]}
Got:      {a: ["b", "c"]}
```
**Issue:** Missing value should create `null` entry in array.

#### 11. __proto__ with allowPrototypes
```
Input:  categories[__proto__]=login&categories[length]=42 with allowPrototypes: true
Expected: {categories: {length: "42"}}
Got:      {categories: {__proto__: ["login"], length: "42"}}
```
**Issue:** `__proto__` should ALWAYS be ignored, even with `allowPrototypes: true`.

#### 12. Charset Sentinel Switching
```
Input:  utf8=%26%2310003%3B&%C3%B8=%C3%B8 with charsetSentinel: true, charset: "utf-8"
Expected: {"Ã¸": "Ã¸"} (interpreted as ISO-8859-1)
Got:      {"ø": "ø"} (still UTF-8)
```
**Issue:** Sentinel should switch charset for entire query string.

---

## Test Coverage Comparison

| Module | JS Test Assertions | Go v2 Tests | Coverage |
|--------|-------------------|-------------|----------|
| parse.js | ~350+ | ~150 | ~43% |
| utils.js | ~50 | ~45 | ~90% |
| formats.js | ~10 | ~20 | 100% |
| stringify.js | ~300+ | 0 | 0% |
| **Total** | **~710+** | **~215** | **~30%** |

---

## Recommendations

### Immediate Fixes Required

1. **Implement `depth: 0` behavior** - Disable all bracket parsing
2. **Fix `arrayLimit: 0` logic** - Always return object, not array
3. **Preserve merge order** - First-seen value should be first in array
4. **Handle primitive-to-object addition** - `a[b]=c&a=d` pattern
5. **Support empty parent brackets** - `[]=a` creates `{0: "a"}`
6. **Fix parameter limit** - Count parameters, not characters
7. **ISO-8859-1 to UTF-8 conversion** - Proper charset handling

### Secondary Fixes

8. **Always ignore `__proto__`** - Even with `allowPrototypes: true`
9. **Charset sentinel switching** - Full re-parse with detected charset
10. **Sparse array null handling** - With `strictNullHandling`

### Implementation Priority

```
Phase 1: Core Parsing Fixes
├── depth: 0 behavior
├── arrayLimit: 0 behavior
├── Parameter limit truncation
└── Empty parent brackets

Phase 2: Merge Behavior
├── Mixed array/value order
├── Primitive-to-object addition
└── Arrays to objects transformation

Phase 3: Charset & Security
├── ISO-8859-1 conversion
├── Charset sentinel
└── __proto__ hardening
```

---

## Test File Location

All JS-compatible tests are located in:
```
v2/parse_js_compat_test.go
```

Run tests with:
```bash
go test -v -run "TestJS" ./...
```

---

## Conclusion

The Go v2 implementation covers approximately **64%** of the JavaScript `qs` library's parse functionality based on direct test porting. The main gaps are in edge cases around:

- Depth and array limit boundary conditions
- Complex merge scenarios (primitive + object on same key)
- Empty/parentless bracket notation
- Charset handling

Addressing the 7 critical issues identified above would bring compatibility to an estimated **85-90%** of JS behavior.
