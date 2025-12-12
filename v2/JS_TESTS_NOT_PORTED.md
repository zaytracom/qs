# JS Parse Tests Not Ported to Go

This document lists all tests from the JavaScript qs library (`test/parse.js`) that are not ported to the Go implementation, along with the reasons.

## Summary

| Category | Count |
|----------|-------|
| Not applicable (JS-specific) | 20 |
| Can be added | 3 |
| **Total not ported** | **23** |

> **Note**: 15 tests from the "Can be added" category were ported on 2025-12-12.

---

## Not Applicable - JS Runtime/Type System Specific

These tests cannot be ported because they test JavaScript-specific behavior that doesn't exist in Go.

### 1. Type Validation Tests (Go has static typing)

| JS Test | Reason |
|---------|--------|
| `throws when decodeDotInKeys is not of type boolean` | Go: compile-time type checking, `bool` type enforced |
| `throws when allowEmptyArrays is not of type boolean` | Go: compile-time type checking, `bool` type enforced |
| `throws error with wrong decoder` | Go: function signature enforced at compile time |
| `throws error when throwOnLimitExceeded is present but not boolean` | Go: compile-time type checking |
| `throws error when throwOnLimitExceeded is present but not boolean for array limit` | Go: compile-time type checking |

### 2. JS Object Input Tests (Go Parse accepts string only)

| JS Test | Reason |
|---------|--------|
| `parses semi-parsed strings` | JS: `qs.parse({ a: 'b' })` - object input |
| `parses an object` | JS: `qs.parse({ a: { b: 'c' } })` - object input |
| `parses values with comma as array divider` | JS: object input with comma parsing |
| `parses an object in dot notation` | JS: object input with dot notation |
| `parses an object and not child values` | JS: object input handling |
| `does not crash when parsing circular references` | JS: circular object references |
| `does not crash when parsing deep objects` | JS: deeply nested object input |

### 3. Node.js Specific

| JS Test | Reason |
|---------|--------|
| `parses buffers correctly` | Node.js `Buffer` type doesn't exist in Go |
| `does not blow up when Buffer global is missing` | Node.js runtime check |

### 4. JS Prototype/Runtime Specific

| JS Test | Reason |
|---------|--------|
| `parses null objects correctly` | JS: `Object.create(null)` / `{ __proto__: null }` |
| `parses dates correctly` | JS: `Date` object input |
| `parses regular expressions correctly` | JS: `RegExp` object input |
| `does not crash when the global Object prototype is frozen` | JS: `Object.freeze(Object.prototype)` |
| `does not throw when a native prototype has an enumerable property` | JS: prototype pollution test |

### 5. Other JS-Specific

| JS Test | Reason |
|---------|--------|
| `uses original key when depth = false` | JS: `depth: false` is falsy, Go uses `int` |
| `does not mutate the options argument` | Go: structs are copied by value |

---

## Can Be Added - Applicable to Go

These tests could be ported to Go but are currently missing.

### 1. Edge Cases Not Yet Tested

| JS Test | Priority | Status | Description |
|---------|----------|--------|-------------|
| `parses jquery-param strings` | Low | ✅ Ported | jQuery serialization format compatibility |
| `does not error when parsing a very long array` | Medium | ✅ Ported | Performance/stress test |
| `does not use non-splittable objects as delimiters` | Low | N/A | Not applicable in Go (no JS objects) |
| `allows setting the parameter limit to Infinity` | Medium | ✅ Ported | `math.MaxInt` equivalent |
| `ignores an utf8 sentinel with an unknown value` | Medium | ✅ Ported | Invalid sentinel handling |
| `uses the utf8 sentinel to switch to iso-8859-1 when no default charset is given` | Medium | ✅ Ported | Charset detection |
| `interpretNumericEntities with comma:true and iso charset does not crash` | Low | ✅ Ported | Combined options edge case |
| `does not interpret %uXXXX syntax in iso-8859-1 mode` | Low | ✅ Ported | Invalid escape sequence |
| `skips empty string key with ...` | Low | ✅ Ported | Empty key edge case |

### 2. Custom Decoder Tests

| JS Test | Priority | Status | Description |
|---------|----------|--------|-------------|
| `use number decoder, parses string that has one number with comma option enabled` | Medium | ✅ Ported | Custom decoder with comma |
| `can parse with custom encoding` | Medium | ✅ Ported | Custom encoder/decoder |
| `receives the default decoder as a second argument` | Low | N/A | Go uses different API (charset passed as arg) |
| `handles a custom decoder returning null, in the iso-8859-1 charset, when interpretNumericEntities` | Low | ✅ Ported | Error handling test |
| `handles a custom decoder returning null, with a string key of null` | Low | N/A | Go has no null keys |
| `allows for decoding keys and values differently` | Medium | ✅ Ported | Separate key/value decoders |

### 3. Option Validation

| JS Test | Priority | Status | Description |
|---------|----------|--------|-------------|
| `throws if an invalid charset is specified` | High | ✅ Existing | Already tested in parse_test.go |

### 4. Additional Coverage

| JS Test | Priority | Status | Description |
|---------|----------|--------|-------------|
| `parses url-encoded brackets holds array of arrays when having two parts of strings with comma as array divider` | Medium | ✅ Ported | URL-encoded variant of existing test |
| `can return null objects` | Low | ✅ Ported | `plainObjects` option behavior |

---

## Test Coverage Comparison

### JS parse.js Test Blocks: 111
### Go Ported Test Cases: ~114 (in multiple test files)

### Coverage by Category:

| Category | JS Tests | Go Tests | Coverage |
|----------|----------|----------|----------|
| Basic parsing | 16 | 16 | 100% |
| Comma option | 4 | 4 | 100% |
| Dot notation | 6 | 5 | 83% |
| Empty arrays | 3 | 3 | 100% |
| Depth handling | 3 | 2 | 67% |
| Array parsing | 8 | 9 | 100% |
| URL encoding | 3 | 4 | 100% |
| Transforms | 4 | 4 | 100% |
| Malformed input | 2 | 2 | 100% |
| Prototype protection | 6 | 6 | 100% |
| Sparse arrays | 2 | 2 | 100% |
| Delimiters | 3 | 3 | 100% |
| Limits (param/array) | 12 | 10 | 83% |
| Query prefix | 1 | 1 | 100% |
| Charset | 10 | 8 | 80% |
| Duplicates | 4 | 4 | 100% |
| StrictDepth | 8 | 6 | 75% |
| Empty keys | 6 | 8 | 100% |
| Custom Decoder | 6 | 5 | 83% |
| JS-specific (N/A) | 20 | 0 | N/A |

---

## Recommendations

All high and medium priority tests have been ported. The remaining tests are either:
- Not applicable to Go (JS-specific features)
- Low priority edge cases with N/A status

---

## Notes

- All core parsing functionality is tested and matches JS behavior
- Missing tests are primarily JS-specific features that don't exist in Go
- Go implementation passes all applicable JS test cases
- Some tests are inherently impossible in Go due to language differences
- Tests ported on 2025-12-12 are in `parse_js_compat_test.go`
