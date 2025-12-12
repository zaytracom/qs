# Test Status Table

## Summary

| Status | Count | Percentage |
|--------|-------|------------|
| PASS | 28 | 64% |
| FAIL | 16 | 36% |
| **Total** | **44** | **100%** |

---

## Detailed Test Results

| # | Test Name | Status | Description |
|---|-----------|--------|-------------|
| 1 | `TestJSParseSimpleString` | PASS | Basic parsing: `foo=bar`, `0=foo`, `foo=c++`, special chars |
| 2 | `TestJSCommaFalse` | PASS | Default comma behavior: `a=b,c` stays as string |
| 3 | `TestJSCommaTrue` | PASS | Comma splitting: `a=b,c` → `["b","c"]` |
| 4 | `TestJSAllowsDotNotation` | PASS | Dot notation: `a.b=c` with `allowDots` |
| 5 | `TestJSDecodeDotKeys` | PASS | Decode `%2E` in keys with `decodeDotInKeys` |
| 6 | `TestJSAllowEmptyArrays` | PASS | Empty arrays: `foo[]` → `[]` |
| 7 | `TestJSAllowEmptyArraysStrictNull` | PASS | Empty arrays with `strictNullHandling` |
| 8 | `TestJSNestedStrings` | PASS | Nested objects: `a[b][c]=d` |
| 9 | `TestJSDepthOne` | PASS | Depth limit = 1 |
| 10 | `TestJSDepthZero` | **FAIL** | `depth: 0` should disable all nesting |
| 11 | `TestJSSimpleArray` | PASS | Duplicate keys: `a=b&a=c` → `["b","c"]` |
| 12 | `TestJSExplicitArray` | PASS | Bracket arrays: `a[]=b&a[]=c` |
| 13 | `TestJSMixedArrays` | **FAIL** | Order issue: `a=b&a[]=c` wrong order |
| 14 | `TestJSNestedArray` | PASS | Nested arrays: `a[b][]=c` |
| 15 | `TestJSArrayIndices` | **FAIL** | `arrayLimit: 0` should return object |
| 16 | `TestJSArrayLimitIndices` | PASS | Array limit boundary (20 vs 21) |
| 17 | `TestJSKeysBeginWithNumber` | PASS | Keys like `a[12b]=c` |
| 18 | `TestJSEncodedEquals` | PASS | Encoded `=`: `he%3Dllo=th%3Dere` |
| 19 | `TestJSUrlEncodedStrings` | PASS | URL encoding in keys/values |
| 20 | `TestJSBracketsInValue` | PASS | Brackets in value: `pets=["tobi"]` |
| 21 | `TestJSEmptyValues` | PASS | Empty string input |
| 22 | `TestJSArraysToObjects` | **FAIL** | Mixed string/numeric keys not preserved |
| 23 | `TestJSArraysToObjectsDotNotation` | **FAIL** | Same issue with dot notation |
| 24 | `TestJSPruneUndefined` | **FAIL** | Sparse indices lost in conversion |
| 25 | `TestJSMalformedUri` | PASS | Malformed URI: `{%:%}` |
| 26 | `TestJSNoEmptyKeys` | PASS | Trailing `&` ignored |
| 27 | `TestJSArraysOfObjects` | PASS | `a[][b]=c` → `[{b:"c"}]` |
| 28 | `TestJSEmptyStringsInArrays` | **FAIL** | `strictNullHandling` null in arrays |
| 29 | `TestJSCompactsSparseArrays` | PASS | Sparse array compaction |
| 30 | `TestJSParsesSparseArrays` | PASS | `allowSparse: true` |
| 31 | `TestJSAlternativeDelimiter` | PASS | Custom delimiter `;` |
| 32 | `TestJSRegExpDelimiter` | PASS | RegExp delimiter |
| 33 | `TestJSParameterLimit` | **FAIL** | Limit counts wrong (chars vs params) |
| 34 | `TestJSArrayLimitOverride` | **FAIL** | `arrayLimit: 0` multiple elements |
| 35 | `TestJSDisableArrayParsing` | PASS | `parseArrays: false` |
| 36 | `TestJSQueryPrefix` | PASS | `ignoreQueryPrefix: true` |
| 37 | `TestJSCommaAsArrayDivider` | PASS | Comma in nested values |
| 38 | `TestJSBracketsArrayOfArrays` | PASS | `foo[]=1,2,3&foo[]=4,5,6` |
| 39 | `TestJSPercentEncodedComma` | PASS | `%2C` not split |
| 40 | `TestJSPrototypeProtection` | PASS | `hasOwnProperty` blocked |
| 41 | `TestJSAllowPrototypes` | PASS | `allowPrototypes: true` |
| 42 | `TestJSStartingWithClosingBracket` | PASS | Keys like `]=x` |
| 43 | `TestJSStartingWithOpeningBracket` | PASS | Keys like `[=x` |
| 44 | `TestJSAddKeysToObjects` | **FAIL** | `a[b]=c&a=d` should add to object |
| 45 | `TestJSDunderProto` | **FAIL** | `__proto__` leaks with `allowPrototypes` |
| 46 | `TestJSCharset` | **FAIL** | ISO-8859-1 not converted to UTF-8 |
| 47 | `TestJSCharsetSentinel` | **FAIL** | Sentinel charset switch fails |
| 48 | `TestJSNumericEntities` | PASS | `&#9786;` → `☺` |
| 49 | `TestJSNoParentFound` | **FAIL** | `[]=a` should create `{0:"a"}` |
| 50 | `TestJSDuplicatesOption` | PASS | `duplicates: combine/first/last` |
| 51 | `TestJSStrictDepthThrow` | PASS | `strictDepth` throws error |
| 52 | `TestJSStrictDepthNoThrow` | PASS | Within depth limit OK |
| 53 | `TestJSParameterLimitTests` | **FAIL** | Truncation behavior wrong |
| 54 | `TestJSArrayLimitTests` | **FAIL** | Object conversion incomplete |
| 55 | `TestJSEmptyKeys` | **FAIL** | Some edge cases with empty keys |
| 56 | `TestJSEmptyKeysWithBrackets` | **FAIL** | `[0]=a` should create `{0:"a"}` |

---

## Failed Tests by Category

### Depth & Limits (4 tests)

| Test | Issue | Expected | Got |
|------|-------|----------|-----|
| `TestJSDepthZero` | `depth:0` ignores brackets | `{"a[0]":"b"}` | `{a:["b"]}` |
| `TestJSArrayIndices` | `arrayLimit:0` = object | `{a:{1:"c"}}` | `{a:["c"]}` |
| `TestJSParameterLimit` | Count params not chars | `{a:"b"}` | `{a:"b&c=d"}` |
| `TestJSArrayLimitOverride` | `arrayLimit:0` multi | `{a:{0:"b",1:"c"}}` | `{a:["b","c"]}` |

### Merge Behavior (5 tests)

| Test | Issue | Expected | Got |
|------|-------|----------|-----|
| `TestJSMixedArrays` | Order preserved | `["b","c"]` | `["c","b"]` |
| `TestJSArraysToObjects` | String+numeric keys | `{bad:"baz",0:"bar"}` | `{bad:"baz"}` |
| `TestJSArraysToObjectsDot` | Same with dots | `{bad:"baz",0:"bar"}` | `{bad:"baz"}` |
| `TestJSPruneUndefined` | Keep all indices | `{2:"b",99999999:"c"}` | `{99999999:"c"}` |
| `TestJSAddKeysToObjects` | Primitive to object | `{b:"c",d:true}` | `["d",{b:"c"}]` |

### Empty/Parent Brackets (3 tests)

| Test | Issue | Expected | Got |
|------|-------|----------|-----|
| `TestJSNoParentFound` | `[]=a` creates index | `{0:"a"}` | `{}` |
| `TestJSEmptyKeys` | Edge cases | Various | Partial |
| `TestJSEmptyKeysWithBrackets` | `[0]=a` at root | `{0:"a"}` | `{}` |

### Charset (2 tests)

| Test | Issue | Expected | Got |
|------|-------|----------|-----|
| `TestJSCharset` | ISO→UTF-8 conversion | `{"¢":"½"}` | `{"\xa2":"\xbd"}` |
| `TestJSCharsetSentinel` | Sentinel switches charset | `{"Ã¸":"Ã¸"}` | `{"ø":"ø"}` |

### Other (2 tests)

| Test | Issue | Expected | Got |
|------|-------|----------|-----|
| `TestJSDunderProto` | `__proto__` always blocked | `{length:"42"}` | `{__proto__:[...]}` |
| `TestJSEmptyStringsInArrays` | null in array | `["b",null,"c"]` | `["b","c"]` |

---

## Priority Matrix

| Priority | Tests | Impact |
|----------|-------|--------|
| **P0 Critical** | `DepthZero`, `ParameterLimit`, `NoParentFound` | Core parsing broken |
| **P1 High** | `ArrayLimit*`, `MixedArrays`, `AddKeysToObjects` | Common use cases |
| **P2 Medium** | `Charset*`, `DunderProto`, `EmptyStrings` | Edge cases |
| **P3 Low** | `PruneUndefined`, `ArraysToObjects` | Rare scenarios |

---

## Run Command

```bash
# Run all JS compatibility tests
go test -v -run "TestJS" ./...

# Run specific test
go test -v -run "TestJSDepthZero" ./...

# Run with coverage
go test -v -run "TestJS" -coverprofile=coverage.out ./...
```
