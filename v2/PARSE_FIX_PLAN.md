# Parse Fix Plan - Comparison with JS Version (UPDATED)

## Overview

Current Go v2 implementation has **16 failed tests** out of 44 (36%). After detailed JS/Go code comparison, the real root causes are:

1. **CRITICAL**: Map iteration order (random in Go, insertion-order in JS)
2. `normalizeParseOptions` overwrites zero values with defaults
3. `splitByDelimiter` behavior differs from JS `String.split()`
4. `Merge` function has multiple behavioral differences
5. `__proto__` incorrectly allowed with `allowPrototypes: true`
6. ISO-8859-1 decoding produces invalid UTF-8

---

## Phase 0: CRITICAL — Map Iteration Order

### Problem

This is the **root cause** of ~50% of failed tests, completely missing from original plan.

**JS (parse.js:320-325):**
```js
var keys = Object.keys(tempObj);
for (var i = 0; i < keys.length; ++i) {
    var key = keys[i];
    var newObj = parseKeys(key, tempObj[key], options, typeof str === 'string');
    obj = utils.merge(obj, newObj, options);
}
```
`Object.keys()` returns keys in **insertion order** (ES2015+).

**Go (parse.go:872):**
```go
for key, val := range tempObj {
    newObj, err := parseKeys(key, val, &normalizedOpts, true)
    // ...
}
```
Go map iteration is **randomized** by design.

### Impact

For input `a[b]=c&a=d`:
- JS always processes `a[b]` first, then `a` → `{a: {b: "c", d: true}}`
- Go may process `a` first, then `a[b]` → `{a: ["d", {b: "c"}]}`

Affected tests:
- `TestJSMixedArrays` — order `["b", "c"]` vs `["c", "b"]`
- `TestJSAddKeysToObjects` — `{b: "c", d: true}` vs `["d", {b: "c"}]`
- `TestJSArraysToObjects` — key order in converted objects
- `TestJSArraysToObjectsDotNotation` — same issue

### Solution

**Option A: Track insertion order in parseValues**
```go
type orderedResult struct {
    keys   []string       // insertion order
    values map[string]any
}

func parseValues(str string, opts *ParseOptions) (orderedResult, error) {
    result := orderedResult{
        keys:   make([]string, 0),
        values: make(map[string]any),
    }
    // ...
    if _, exists := result.values[key]; !exists {
        result.keys = append(result.keys, key)
    }
    result.values[key] = val
    // ...
}
```

**Option B: Use ordered map library**
```go
import "github.com/elliotchance/orderedmap/v2"
```

**Option C: Sort keys deterministically**
Less ideal but simplest — sort keys alphabetically. Won't match JS exactly but will be deterministic.

### Files
- `parse.go:665-834` — parseValues return type
- `parse.go:870-886` — Parse iteration loop

### Tests affected
- `TestJSMixedArrays`
- `TestJSAddKeysToObjects`
- `TestJSArraysToObjects`
- `TestJSArraysToObjectsDotNotation`

---

## Phase 1: Core Parsing Fixes

### 1.1 Fix `depth: 0` and `arrayLimit: 0` behavior

**Problem:** Zero values are overwritten with defaults.

**JS (parse.js:287, 295):**
```js
arrayLimit: typeof opts.arrayLimit === 'number' ? opts.arrayLimit : defaults.arrayLimit,
depth: (typeof opts.depth === 'number' || opts.depth === false) ? +opts.depth : defaults.depth,
```
JS checks **type**, not value. `0` is a valid number.

**Go (parse.go:223-231):**
```go
if result.ArrayLimit == 0 {
    result.ArrayLimit = DefaultArrayLimit  // WRONG: overwrites explicit 0
}
if result.Depth == 0 {
    result.Depth = DefaultDepth  // WRONG: overwrites explicit 0
}
```

### Solution

**Option A: Use pointers**
```go
type ParseOptions struct {
    ArrayLimit *int  // nil = use default, 0 = explicitly zero
    Depth      *int
    // ...
}

func normalizeParseOptions(opts *ParseOptions) (ParseOptions, error) {
    // ...
    if result.ArrayLimit == nil {
        defaultVal := DefaultArrayLimit
        result.ArrayLimit = &defaultVal
    }
    if result.Depth == nil {
        defaultVal := DefaultDepth
        result.Depth = &defaultVal
    }
}
```

**Option B: Use sentinel value**
```go
const (
    ArrayLimitNotSet = -999
    DepthNotSet      = -999
)

// In DefaultParseOptions:
ArrayLimit: ArrayLimitNotSet,
Depth:      DepthNotSet,

// In normalizeParseOptions:
if result.ArrayLimit == ArrayLimitNotSet {
    result.ArrayLimit = DefaultArrayLimit
}
```

**Option C: Separate "Set" flags**
```go
type ParseOptions struct {
    ArrayLimit    int
    ArrayLimitSet bool
    Depth         int
    DepthSet      bool
}
```

### Files
- `parse.go:35-138` — ParseOptions struct
- `parse.go:197-244` — normalizeParseOptions
- `parse.go:278-282, 339-343` — WithArrayLimit, WithDepth

### Tests affected
- `TestJSDepthZero`
- `TestJSArrayIndices` (arrayLimit: 0)
- `TestJSArrayLimitOverride`
- `TestJSArrayLimitTests`

---

### 1.2 Fix ParameterLimit (split behavior)

**Problem:** JS and Go `split` behave differently with limit parameter.

**JS (parse.js:67-70):**
```js
var limit = options.parameterLimit === Infinity ? undefined : options.parameterLimit;
var parts = cleanStr.split(options.delimiter, limit);
```
JS `"a&b&c".split("&", 2)` returns `["a", "b"]` — exactly 2 parts, rest discarded.

**Go current behavior:**
```go
// For limit=1: loops 0 times, returns whole string as 1 part
// "a&b&c" with limit=1 → ["a&b&c"]  WRONG!
// Should be: ["a"]
```

**JS behavior table:**
| Input | Delimiter | Limit | Result |
|-------|-----------|-------|--------|
| `a&b&c` | `&` | 1 | `["a"]` |
| `a&b&c` | `&` | 2 | `["a", "b"]` |
| `a&b&c` | `&` | 5 | `["a", "b", "c"]` |

### Solution

```go
func splitByDelimiter(str string, delimiter string, delimiterRegexp *regexp.Regexp, limit int) []string {
    if limit <= 0 {
        // No limit
        if delimiterRegexp != nil {
            return delimiterRegexp.Split(str, -1)
        }
        return strings.Split(str, delimiter)
    }

    // Split with limit+1 to detect if there are more parts
    var parts []string
    if delimiterRegexp != nil {
        parts = delimiterRegexp.Split(str, limit+1)
    } else {
        parts = strings.SplitN(str, delimiter, limit+1)
    }

    // Truncate to limit (discard remainder like JS)
    if len(parts) > limit {
        parts = parts[:limit]
    }

    return parts
}
```

### Files
- `parse.go:453-482` — splitByDelimiter

### Tests affected
- `TestJSParameterLimit`
- `TestJSParameterLimitTests`

---

### 1.3 Support empty parent brackets `[]=a`

**Problem:** Keys like `[]` at root level should create indexed entries.

**JS behavior:**
```
Input:  []=&a=b
Expected: {0: "", a: "b"}

Input:  []=a&[]=b
Expected: {0: "a", 1: "b"}

Input:  [0]=a&[1]=b
Expected: {0: "a", 1: "b"}
```

**JS (parse.js:205-219):**
```js
var segment = options.depth > 0 && brackets.exec(key);
var parent = segment ? key.slice(0, segment.index) : key;

var keys = [];
if (parent) {
    keys.push(parent);
}

// For key="[]": segment matches at index 0, parent="", keys=[]
// Loop adds "[]" to keys
// parseObject(["[]"], val) creates array, which becomes {0: val}
```

**Go issue:**
The logic is similar but the final merge doesn't handle array→object conversion at root level properly.

### Root Cause Analysis

When `parseKeys("[]", "a", ...)` returns `["a"]` (an array), and we merge it into `{}` (empty map), Go's Merge doesn't convert the array to object with index keys.

**JS (utils.js:95-104):**
```js
return Object.keys(source).reduce(function (acc, key) {
    var value = source[key];
    // Object.keys(["a"]) returns ["0"]!
    acc[key] = value;
    return acc;
}, mergeTarget);
```

**Go (utils.go:322-330):**
```go
if sourceIsMap {
    for key, value := range sourceMap {
        // ...
    }
}
return mergeTarget  // If source is slice, it's IGNORED!
```

### Solution

Fix in Merge to handle `target=map, source=slice`:

```go
// After the map+map merge section, add:
if sourceIsSlice {
    // Convert slice indices to map keys (like JS Object.keys on array)
    for i, value := range sourceSlice {
        if value == nil {
            continue
        }
        key := strconv.Itoa(i)
        if existing, exists := mergeTarget[key]; exists {
            mergeTarget[key] = Merge(existing, value, allowPrototypes)
        } else {
            mergeTarget[key] = value
        }
    }
}
```

### Files
- `utils.go:232-333` — Merge function
- `parse.go:575-661` — parseKeys (minor adjustments)

### Tests affected
- `TestJSNoParentFound`
- `TestJSEmptyKeysWithBrackets`

---

## Phase 2: Merge Behavior Fixes

### 2.1 Fix slice+slice merge (push vs overwrite)

**Problem:** When merging arrays with same-index primitives, JS pushes to end, Go overwrites.

**JS (utils.js:79-92):**
```js
if (isArray(target) && isArray(source)) {
    source.forEach(function (item, i) {
        if (has.call(target, i)) {
            var targetItem = target[i];
            if (targetItem && typeof targetItem === 'object' && item && typeof item === 'object') {
                target[i] = merge(targetItem, item, options);
            } else {
                target.push(item);  // <- PUSH to end, not overwrite!
            }
        } else {
            target[i] = item;
        }
    });
    return target;
}
```

**Go (utils.go:300-307):**
```go
if targetItemIsMap && itemIsMap {
    targetSlice[i] = Merge(targetItem, item, allowPrototypes)
} else if targetItemIsSlice && itemIsSlice {
    targetSlice[i] = Merge(targetItem, item, allowPrototypes)
} else {
    targetSlice[i] = item  // <- WRONG: overwrites instead of push
}
```

### Solution

```go
if targetIsSlice && sourceIsSlice {
    for i, item := range sourceSlice {
        if item == nil {
            continue
        }

        if i < len(targetSlice) && targetSlice[i] != nil {
            targetItem := targetSlice[i]
            _, targetItemIsMap := targetItem.(map[string]any)
            _, itemIsMap := item.(map[string]any)
            _, targetItemIsSlice := targetItem.([]any)
            _, itemIsSlice := item.([]any)

            if (targetItemIsMap && itemIsMap) || (targetItemIsSlice && itemIsSlice) {
                targetSlice[i] = Merge(targetItem, item, allowPrototypes)
            } else {
                // Both exist but different types or primitives - PUSH to end
                targetSlice = append(targetSlice, item)
            }
        } else {
            // Extend if needed
            for len(targetSlice) <= i {
                targetSlice = append(targetSlice, nil)
            }
            targetSlice[i] = item
        }
    }
    return targetSlice
}
```

### Files
- `utils.go:278-311` — slice+slice merge block

### Tests affected
- `TestJSMixedArrays` (partially — also needs Phase 0)

---

### 2.2 Fix map+slice merge (missing case)

**Problem:** When target is map and source is slice, Go ignores the source entirely.

**JS (utils.js:95-104):**
```js
// Works for both map and array sources via Object.keys()
return Object.keys(source).reduce(function (acc, key) {
    var value = source[key];
    if (has.call(acc, key)) {
        acc[key] = merge(acc[key], value, options);
    } else {
        acc[key] = value;
    }
    return acc;
}, mergeTarget);
```

`Object.keys(["a", "b"])` returns `["0", "1"]`.

**Go (utils.go:322-330):**
```go
if sourceIsMap {
    for key, value := range sourceMap {
        // ...
    }
}
return mergeTarget  // sourceSlice is IGNORED!
```

### Solution

```go
// Source is a map, merge into target
if sourceIsMap {
    for key, value := range sourceMap {
        if existing, exists := mergeTarget[key]; exists {
            mergeTarget[key] = Merge(existing, value, allowPrototypes)
        } else {
            mergeTarget[key] = value
        }
    }
} else if sourceIsSlice {
    // Source is a slice - convert indices to string keys
    for i, value := range sourceSlice {
        if value == nil {
            continue
        }
        key := strconv.Itoa(i)
        if existing, exists := mergeTarget[key]; exists {
            mergeTarget[key] = Merge(existing, value, allowPrototypes)
        } else {
            mergeTarget[key] = value
        }
    }
}

return mergeTarget
```

### Files
- `utils.go:321-332` — add slice handling

### Tests affected
- `TestJSNoParentFound`
- `TestJSEmptyKeysWithBrackets`
- `TestJSArraysToObjects`

---

### 2.3 Fix ArrayToObject (preserve all indices)

**Problem:** ArrayToObject skips nil but should preserve index positions.

**JS (utils.js:36-45):**
```js
var arrayToObject = function arrayToObject(source, options) {
    var obj = options && options.plainObjects ? { __proto__: null } : {};
    for (var i = 0; i < source.length; ++i) {
        if (typeof source[i] !== 'undefined') {
            obj[i] = source[i];
        }
    }
    return obj;
};
```

JS checks for `undefined`, not null. In Go context, we should preserve non-nil values.

**Go (utils.go:343-351):**
```go
func ArrayToObject(source []any) map[string]any {
    result := make(map[string]any)
    for i, v := range source {
        if v != nil {
            result[strconv.Itoa(i)] = v
        }
    }
    return result
}
```

This is actually correct. The issue is elsewhere — when sparse arrays are created, some indices are lost before ArrayToObject is called.

### Real Issue

In `parseObject`, when creating sparse array with large index:
```go
if isValidIndex && opts.ParseArrays && index <= opts.ArrayLimit {
    arr := make([]any, index+1)
    arr[index] = leaf
    obj = arr
}
```

For input `a[2]=b&a[99999999]=c`:
- First: creates `arr[0..2]` with `arr[2]="b"`
- Second: creates `arr[0..99999999]` with `arr[99999999]="c"` — huge allocation!
- JS handles this by converting to object when index > arrayLimit

The arrayLimit check should trigger conversion to object for large indices.

### Files
- `utils.go:343-351` — ArrayToObject (OK as-is)
- `parse.go:549-553` — parseObject array creation

### Tests affected
- `TestJSPruneUndefined`

---

## Phase 3: Security & Charset

### 3.1 `__proto__` must ALWAYS be blocked

**Problem:** `__proto__` is allowed when `allowPrototypes: true`, but JS always blocks it.

**JS (parse.js:179-181):**
```js
} else if (decodedRoot !== '__proto__') {
    obj[decodedRoot] = leaf;
}
// NO else clause! __proto__ is always blocked.
```

**Go (parse.go:554-564):**
```go
} else if decodedRoot != "__proto__" {
    objMap[decodedRoot] = leaf
    obj = objMap
} else if opts.AllowPrototypes {  // <- SECURITY BUG!
    objMap[decodedRoot] = leaf
    obj = objMap
} else {
    obj = objMap
}
```

### Solution

Remove the `allowPrototypes` exception for `__proto__`:

```go
} else if decodedRoot != "__proto__" {
    objMap[decodedRoot] = leaf
    obj = objMap
} else {
    // __proto__ is ALWAYS blocked, regardless of allowPrototypes
    obj = objMap
}
```

### Files
- `parse.go:554-564` — parseObject

### Tests affected
- `TestJSDunderProto`

---

### 3.2 ISO-8859-1 to UTF-8 conversion

**Problem:** ISO-8859-1 bytes should become Unicode code points.

**JS (utils.js:116-118):**
```js
if (charset === 'iso-8859-1') {
    return strWithoutPlus.replace(/%[0-9a-f]{2}/gi, unescape);
}
```

JS `unescape('%A2')` returns character with code point U+00A2 (¢).

**Go (utils.go:196-198):**
```go
if hi >= 0 && lo >= 0 {
    result.WriteByte(byte(hi<<4 | lo))  // <- Writes raw byte, invalid UTF-8!
    i += 3
    continue
}
```

Byte `0xA2` alone is invalid UTF-8. Need to write Unicode code point.

### Solution

```go
if hi >= 0 && lo >= 0 {
    // Convert ISO-8859-1 byte to Unicode code point
    // ISO-8859-1 bytes 0x00-0xFF map directly to U+0000-U+00FF
    result.WriteRune(rune(hi<<4 | lo))
    i += 3
    continue
}
```

### Files
- `utils.go:186-207` — decodeISO88591

### Tests affected
- `TestJSCharset`

---

### 3.3 Charset Sentinel switching

**Problem:** Sentinel should switch charset interpretation for entire string.

**JS behavior:**
1. Scan for `utf8=` sentinel
2. If found, determine charset (UTF-8 or ISO-8859-1)
3. Decode ALL values using detected charset

**Go behavior:**
Charset detection works, but ISO-8859-1 decoding is broken (see 3.2).

After fixing 3.2, this should work. However, verify that:
1. Sentinel is detected before any value decoding
2. Detected charset is used for all subsequent decodes

### Files
- `parse.go:695-709` — charset detection (OK)
- `utils.go:164-183` — Decode function
- `utils.go:186-207` — decodeISO88591 (needs fix from 3.2)

### Tests affected
- `TestJSCharsetSentinel`

---

### 3.4 StrictNullHandling in arrays

**Problem:** Missing value should create null in array, not be skipped.

**Current behavior:**
```
Input:  a[0]=b&a[1]&a[2]=c with strictNullHandling: true
Expected: {a: ["b", null, "c"]}
Got:      {a: ["b", "c"]}
```

**Root cause:** `Compact` removes nil values.

**JS (utils.js:25-29):**
```js
for (var j = 0; j < obj.length; ++j) {
    if (typeof obj[j] !== 'undefined') {
        compacted.push(obj[j]);
    }
}
```

JS removes `undefined`, not `null`. In Go, we're removing all nil.

### Solution

When `strictNullHandling` is true, null values are meaningful and should not be compacted. Need to pass options to Compact or handle differently.

```go
func CompactWithOptions(value any, preserveNull bool) any {
    // ...
    if v == nil && !preserveNull {
        continue
    }
    // ...
}
```

Or: don't compact when strictNullHandling is true.

### Files
- `utils.go:355-407` — Compact functions
- `parse.go:888-894` — Compact call

### Tests affected
- `TestJSEmptyStringsInArrays`

---

## Implementation Order

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 0: CRITICAL - Iteration Order                         │
├─────────────────────────────────────────────────────────────┤
│ 0. parseValues returns ordered keys + Parse uses that order │
│    Impact: Fixes 4+ tests                                   │
│    Risk: HIGH (structural change)                           │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Core Parsing                                       │
├─────────────────────────────────────────────────────────────┤
│ 1.1 normalizeParseOptions - depth/arrayLimit zero values    │
│ 1.2 splitByDelimiter - parameter limit fix                  │
│ 1.3 parseKeys/Merge - empty parent brackets support         │
│    Impact: Fixes 6+ tests                                   │
│    Risk: MEDIUM                                             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Merge Behavior                                     │
├─────────────────────────────────────────────────────────────┤
│ 2.1 Merge slice+slice - push vs overwrite                   │
│ 2.2 Merge map+slice - handle slice source                   │
│ 2.3 ArrayToObject - verify index preservation               │
│    Impact: Fixes 3+ tests                                   │
│    Risk: MEDIUM-HIGH (affects many code paths)              │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Security & Charset                                 │
├─────────────────────────────────────────────────────────────┤
│ 3.1 __proto__ always blocked (SECURITY)                     │
│ 3.2 ISO-8859-1 decode - WriteRune                           │
│ 3.3 Charset sentinel (depends on 3.2)                       │
│ 3.4 StrictNullHandling - preserve null in arrays            │
│    Impact: Fixes 4+ tests                                   │
│    Risk: LOW-MEDIUM                                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Testing Strategy

After each phase:
```bash
# Run all JS compatibility tests
go test -v -run "TestJS" ./v2/...

# Run specific test
go test -v -run "TestJSDepthZero" ./v2/...

# With coverage
go test -v -run "TestJS" -coverprofile=cov.out ./v2/...
go tool cover -html=cov.out
```

### Phase 0 verification:
```bash
go test -v -run "TestJSMixedArrays|TestJSAddKeysToObjects|TestJSArraysToObjects" ./v2/...
```

### Phase 1 verification:
```bash
go test -v -run "TestJSDepthZero|TestJSArrayIndices|TestJSParameterLimit|TestJSNoParentFound" ./v2/...
```

### Phase 3 verification:
```bash
go test -v -run "TestJSDunderProto|TestJSCharset" ./v2/...
```

---

## Expected Outcome

| Phase | Tests Fixed | Cumulative Pass Rate |
|-------|-------------|---------------------|
| Before | 0 | 28/44 (64%) |
| Phase 0 | ~4 | 32/44 (73%) |
| Phase 1 | ~6 | 38/44 (86%) |
| Phase 2 | ~3 | 41/44 (93%) |
| Phase 3 | ~3 | 44/44 (100%) |

---

## Risk Assessment

| Fix | Risk Level | Reason |
|-----|------------|--------|
| Phase 0 (iteration order) | HIGH | Structural change to parseValues/Parse |
| Fix 1.1 (depth/arrayLimit) | MEDIUM | API change if using pointers |
| Fix 1.2 (split) | LOW | Isolated function |
| Fix 2.1-2.2 (Merge) | MEDIUM-HIGH | Core function, many callers |
| Fix 3.1 (__proto__) | LOW | Simple removal |
| Fix 3.2 (charset) | LOW | Isolated function |

**Recommendation:** Create comprehensive unit tests for Merge before modifying it.
