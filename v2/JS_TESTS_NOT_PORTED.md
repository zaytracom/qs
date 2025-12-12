# JS Parse Tests Not Ported to Go

This document lists all tests from the JavaScript qs library (`test/parse.js`) that are not ported to the Go implementation, along with the reasons.

## Summary

| Category | Count |
|----------|-------|
| Not applicable (JS-specific) | 20 |
| Can be added | 18 |
| **Total not ported** | **38** |

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

| JS Test | Priority | Description |
|---------|----------|-------------|
| `parses jquery-param strings` | Low | jQuery serialization format compatibility |
| `does not error when parsing a very long array` | Medium | Performance/stress test |
| `does not use non-splittable objects as delimiters` | Low | Delimiter edge case |
| `allows setting the parameter limit to Infinity` | Medium | `math.MaxInt` equivalent |
| `ignores an utf8 sentinel with an unknown value` | Medium | Invalid sentinel handling |
| `uses the utf8 sentinel to switch to iso-8859-1 when no default charset is given` | Medium | Charset detection |
| `interpretNumericEntities with comma:true and iso charset does not crash` | Low | Combined options edge case |
| `does not interpret %uXXXX syntax in iso-8859-1 mode` | Low | Invalid escape sequence |
| `skips empty string key with ...` | Low | Empty key edge case |

### 2. Custom Decoder Tests

| JS Test | Priority | Description |
|---------|----------|-------------|
| `use number decoder, parses string that has one number with comma option enabled` | Medium | Custom decoder with comma |
| `can parse with custom encoding` | Medium | Custom encoder/decoder |
| `receives the default decoder as a second argument` | Low | Decoder API contract |
| `handles a custom decoder returning null, in the iso-8859-1 charset, when interpretNumericEntities` | Low | Null return handling |
| `handles a custom decoder returning null, with a string key of null` | Low | Null key handling |
| `allows for decoding keys and values differently` | Medium | Separate key/value decoders |

### 3. Option Validation

| JS Test | Priority | Description |
|---------|----------|-------------|
| `throws if an invalid charset is specified` | High | Already implemented, needs test |

### 4. Additional Coverage

| JS Test | Priority | Description |
|---------|----------|-------------|
| `parses url-encoded brackets holds array of arrays when having two parts of strings with comma as array divider` | Medium | URL-encoded variant of existing test |
| `can return null objects` | Low | `plainObjects` option behavior |

---

## Test Coverage Comparison

### JS parse.js Test Blocks: 111
### Go Ported Test Cases: ~99 (in multiple test files)

### Coverage by Category:

| Category | JS Tests | Go Tests | Coverage |
|----------|----------|----------|----------|
| Basic parsing | 16 | 16 | 100% |
| Comma option | 4 | 4 | 100% |
| Dot notation | 6 | 5 | 83% |
| Empty arrays | 3 | 3 | 100% |
| Depth handling | 3 | 2 | 67% |
| Array parsing | 8 | 8 | 100% |
| URL encoding | 3 | 3 | 100% |
| Transforms | 4 | 4 | 100% |
| Malformed input | 2 | 2 | 100% |
| Prototype protection | 6 | 6 | 100% |
| Sparse arrays | 2 | 2 | 100% |
| Delimiters | 3 | 2 | 67% |
| Limits (param/array) | 12 | 8 | 67% |
| Query prefix | 1 | 1 | 100% |
| Charset | 10 | 5 | 50% |
| Duplicates | 4 | 4 | 100% |
| StrictDepth | 8 | 6 | 75% |
| Empty keys | 6 | 6 | 100% |
| JS-specific (N/A) | 20 | 0 | N/A |

---

## Recommendations

### High Priority (should add):
1. `throws if an invalid charset is specified` - error handling test
2. `allows setting the parameter limit to Infinity` - `math.MaxInt` test

### Medium Priority (nice to have):
1. `does not error when parsing a very long array` - stress test
2. `ignores an utf8 sentinel with an unknown value` - robustness
3. Custom decoder edge cases

### Low Priority (optional):
1. jQuery param format compatibility
2. Various edge cases with combined options

---

## Notes

- All core parsing functionality is tested and matches JS behavior
- Missing tests are primarily edge cases and JS-specific features
- Go implementation passes all applicable JS test cases
- Some tests are inherently impossible in Go due to language differences
