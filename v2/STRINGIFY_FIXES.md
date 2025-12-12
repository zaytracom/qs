# Stringify Fixes Plan for JS qs Library Compatibility

## Current Errors Analysis

After running tests and comparing with JS source (`/Users/phl/Projects/qs/.ref/lib/stringify.js`), the following issues were identified:

---

### 1. EncodeDotInKeys + AllowDots=false ✅ CONFIRMED

**File:** `stringify.go`
**Status:** Test fails: `allowDots_false,_encodeDotInKeys_true`

**Problem:** When `encodeDotInKeys=true` and `allowDots=false` (explicitly set), Go implementation uses dot notation instead of bracket notation.

**Error example:**
```
expected: "name%252Eobj%5Bfirst%5D=John&name%252Eobj%5Blast%5D=Doe"
got:      "name%252Eobj.first=John&name%252Eobj.last=Doe"
```

**JS behavior (lib/stringify.js:255):**
```javascript
var allowDots = typeof opts.allowDots === 'undefined'
    ? opts.encodeDotInKeys === true ? true : defaults.allowDots
    : !!opts.allowDots;
```

**Key insight:** JS sets `allowDots=true` ONLY if:
1. `allowDots` is NOT explicitly set (`=== undefined`) AND
2. `encodeDotInKeys === true`

If `allowDots` is explicitly set to `false`, it MUST remain `false`.

**Root cause in Go:**

1. `stringify.go:243-246` (normalizeStringifyOptions):
```go
if result.EncodeDotInKeys && !result.AllowDots {
    result.AllowDots = true  // WRONG: ignores explicit false
}
```

2. `stringify.go:318-324` (WithEncodeDotInKeys):
```go
func WithEncodeDotInKeys(v bool) StringifyOption {
    return func(o *StringifyOptions) {
        o.EncodeDotInKeys = v
        if v {
            o.AllowDots = true  // WRONG: overwrites explicit setting
        }
    }
}
```

**Fix:** Track whether `AllowDots` was explicitly set.

**Option A (recommended): Add tracking field**
```go
type StringifyOptions struct {
    AllowDots         bool
    allowDotsSet      bool  // internal: tracks if AllowDots was explicitly set
    // ...
}

func WithStringifyAllowDots(v bool) StringifyOption {
    return func(o *StringifyOptions) {
        o.AllowDots = v
        o.allowDotsSet = true
    }
}

// In normalizeStringifyOptions:
if result.EncodeDotInKeys && !result.allowDotsSet {
    result.AllowDots = true  // Only set if not explicitly provided
}

// Remove auto-set from WithEncodeDotInKeys
```

**Option B: Remove auto-set entirely**
Simply remove lines 243-246 and 321-323. Users who want `encodeDotInKeys=true` with dots must also set `allowDots=true` explicitly.

---

### 2. Empty strings in arrays are skipped ✅ CONFIRMED

**File:** `stringify.go`
**Status:** Multiple test failures

**Problem:** Empty strings `""` in arrays are being skipped instead of output as `key=`.

**Error examples:**
```
input: {b: [""], c: "c"}
expected: "b[0]=&c=c"
got:      "c=c"

input: {b: [""], c: "c"}, strictNullHandling: true
expected: "b[0]&c=c"
got:      "c=c"
```

**JS behavior (lib/stringify.js:137-139, 169-171):**
```javascript
// Line 137-139: undefined returns empty array (no output)
if (typeof obj === 'undefined') {
    return values;
}

// Line 169-171: Only skip null with skipNulls, NOT undefined or empty string
if (skipNulls && value === null) {
    continue;
}
```

**Root cause in Go (stringify.go:662-665):**
```go
if value == nil && isSlice(obj) {
    // In arrays, nil represents undefined (sparse slot) - always skip
    continue
}
```

**Investigation needed:** The empty string `""` should NOT be `nil`. Need to trace where `""` becomes `nil` or why it's being skipped.

**Likely issue:** When building `objKeys` for arrays (line 612-616), the value retrieval might be returning `nil` for empty strings, OR the empty string is being converted somewhere.

**Fix approach:**
1. Trace the code path for `{b: [""]}`
2. Ensure `""` stays as `""` and is not converted to `nil`
3. Empty strings should reach the primitive handling (line 543) and output `key=`

---

### 3. strictNullHandling with empty strings ✅ CONFIRMED (related to #2)

Same root cause as #2. Once empty strings are properly handled, this should also work.

---

### 4. Filter array applied recursively ✅ CONFIRMED

**File:** `stringify.go`
**Status:** Test fails with garbage keys

**Problem:** Filter as array of keys is applied at ALL nesting levels instead of only at root.

**Error example:**
```
input: {a: {b: [1, nil, 3]}}
filter: ["a", "b", "0", "2"]
expected: "a%5Bb%5D%5B0%5D=1&a%5Bb%5D%5B2%5D=3"
got:      "a%5Ba%5D=&a%5Bb%5D%5Ba%5D=1&a%5Bb%5D%5Bb%5D=1&a%5Bb%5D%5B0%5D=1&a%5Bb%5D%5B2%5D=3&a%5B0%5D=&a%5B2%5D=&b=&0=&2="
```

**JS behavior (lib/stringify.js:148-153, 191, 287-293, 330):**

In the **exported function** (line 287-293):
```javascript
if (typeof options.filter === 'function') {
    filter = options.filter;
    obj = filter('', obj);
} else if (isArray(options.filter)) {
    filter = options.filter;
    objKeys = filter;  // Used for root keys only
}
// ...
// Line 330: filter is passed to stringify()
stringify(..., options.filter, ...)
```

In **recursive stringify** (line 148-153):
```javascript
if (generateArrayPrefix === 'comma' && isArray(obj)) {
    // ... comma handling
} else if (isArray(filter)) {
    objKeys = filter;  // Filter array used as keys for OBJECTS
} else {
    var keys = Object.keys(obj);
    objKeys = sort ? keys.sort(sort) : keys;
}
```

**Key insight:** The `else if (isArray(filter))` block is AFTER the array check. So for **arrays**, filter is NOT applied - indices are used. For **objects**, filter array specifies which keys to include.

**Root cause in Go (stringify.go:591-596):**
```go
} else if filterSlice, ok := filter.([]string); ok {
    objKeys = make([]any, len(filterSlice))
    for i, k := range filterSlice {
        objKeys[i] = k
    }
}
```

This runs BEFORE checking if `obj` is an array, and applies to all objects.

**Fix:**
```go
// In stringify(), reorder the conditions:
if generateArrayPrefix == nil && isSlice(obj) {
    // Comma format handling...
} else if isSlice(obj) {
    // Array: use indices, NOT filter
    slice := toSlice(obj)
    objKeys = make([]any, len(slice))
    for i := range slice {
        objKeys[i] = i
    }
} else if filterSlice, ok := filter.([]string); ok {
    // Object with filter array: use filter keys
    objKeys = make([]any, len(filterSlice))
    for i, k := range filterSlice {
        objKeys[i] = k
    }
} else {
    // Object without filter: use map keys
    // ...existing code...
}
```

---

### 5. Filter function returns nil → key not skipped ✅ CONFIRMED

**File:** `stringify.go`
**Status:** Test fails

**Problem:** When filter function returns `nil`, the key should be completely skipped.

**Error example:**
```
input: {a: "b", c: nil, e: {f: time}}
filter: func returns nil for "c"
expected: "a=b&e%5Bf%5D=1257894000000"
got:      "a=b&c=&e%5Bf%5D=1257894000000"
```

**JS behavior (lib/stringify.js:106-107):**
```javascript
if (typeof filter === 'function') {
    obj = filter(prefix, obj);
}
// Then later, if obj is null/undefined, it's handled appropriately
```

**Root cause in Go (stringify.go:509-514):**
```go
if filterFunc, ok := filter.(FilterFunc); ok {
    obj = filterFunc(prefix, obj)
} else if fn, ok := filter.(func(string, any) any); ok {
    obj = fn(prefix, obj)
}
// No check if obj became nil after filter!
```

**Fix:** After applying filter function, check if result is `nil` and return early:
```go
if filterFunc, ok := filter.(FilterFunc); ok {
    obj = filterFunc(prefix, obj)
    if obj == nil {
        return []string{}, nil  // Skip this key entirely
    }
} else if fn, ok := filter.(func(string, any) any); ok {
    obj = fn(prefix, obj)
    if obj == nil {
        return []string{}, nil
    }
}
```

---

### 6. Key order (map iteration) ✅ CONFIRMED

**File:** `stringify_jscompat_test.go`
**Status:** Test failures due to ordering

**Problem:** Go maps don't guarantee iteration order. Some tests expect specific order.

**Error example:**
```
expected: "a=b&c[]=d&c[]=e&f[][]=g&f[][]=h"
got:      "f[][]=g&f[][]=h&a=b&c[]=d&c[]=e"
```

**Fix:** This is a TEST issue, not implementation issue. Tests should either:
1. Add `WithSort(func(a, b string) bool { return a < b })` to get deterministic order
2. Compare results as unordered sets of key-value pairs

**Note:** The implementation correctly supports `WithSort` option. Tests need updating.

---

### 7. Empty key with nested array ✅ CONFIRMED

**File:** `stringify.go`
**Status:** Test fails

**Problem:** When key is empty string `""` containing nested empty key with array, format is wrong.

**Error example:**
```
input: {"": {"": [2, 3]}}
expected: "[][0]=2&[][1]=3"
got:      "[0]=2&[1]=3"
```

**JS behavior analysis:**
1. Root call: `key=""`, value=`{"": [2,3]}`
2. Recurse into object: `prefix=""`, `key=""` → `keyPrefix = "" + "[" + "" + "]"` = `"[]"`
3. Recurse into array: `prefix="[]"`, `generateArrayPrefix("[]", "0")` = `"[][0]"`

**JS (lib/stringify.js:174-176):**
```javascript
var keyPrefix = isArray(obj)
    ? typeof generateArrayPrefix === 'function' ? generateArrayPrefix(adjustedPrefix, encodedKey) : adjustedPrefix
    : adjustedPrefix + (allowDots ? '.' + encodedKey : '[' + encodedKey + ']');
```

For objects with empty key: `"" + "[" + "" + "]"` = `"[]"` ✓

**Root cause in Go:** When `adjustedPrefix=""` and `encodedKey=""`, the bracket notation should produce `"[]"`, but something is generating just `""`.

**Investigation:** Check Go's keyPrefix generation for empty keys in objects:
```go
// stringify.go:688-693
if allowDots {
    keyPrefix = adjustedPrefix + "." + encodedKey
} else {
    keyPrefix = adjustedPrefix + "[" + encodedKey + "]"
}
```

With `adjustedPrefix=""` and `encodedKey=""`: `"" + "[" + "" + "]"` = `"[]"` — this looks correct.

**Actual issue:** May be in how the initial prefix is built in `Stringify()` function, or in `generateArrayPrefix` handling.

**Direct JS test:**
```javascript
qs.stringify({'': [2,3]}, {encode: false})  // → "[0]=2&[1]=3" (NOT "[][0]=2")
qs.stringify({'': {'': [2,3]}}, {encode: false})  // → "[][0]=2&[][1]=3"
```

So `{"": [2,3]}` at root → `[0]=2` is CORRECT! The test case `{"": {"": [2,3]}}` expects `[][0]=2`.

**Fix:** Need to trace exact test case that's failing and verify expected value.

---

## Fix Priority (Updated)

### 1. Critical (core logic bugs):

1. **#4 Filter array** - causes garbage output, breaks filtering
2. **#5 Filter function nil** - filter doesn't work correctly
3. **#2/#3 Empty strings in arrays** - values silently dropped

### 2. Medium priority (option behavior):

4. **#1 EncodeDotInKeys + AllowDots** - option interaction bug
5. **#7 Empty key with array** - edge case with empty keys

### 3. Low priority (test fixes):

6. **#6 Key order** - fix tests to use sort or unordered comparison

---

## Implementation Order

```
1. Fix #4 (filter array) - reorder conditions in stringify()
2. Fix #5 (filter nil) - add nil check after filter function
3. Fix #2/#3 (empty strings) - trace and fix value handling
4. Fix #1 (allowDots) - add tracking field
5. Fix #7 (empty key) - investigate and fix prefix generation
6. Fix #6 (tests) - add sort to affected tests
```

---

## Code Locations Summary

| Issue | File | Lines | Function |
|-------|------|-------|----------|
| #1 | stringify.go | 243-246, 318-324 | normalizeStringifyOptions, WithEncodeDotInKeys |
| #2/#3 | stringify.go | 662-669 | stringify (skip logic) |
| #4 | stringify.go | 591-618 | stringify (objKeys building) |
| #5 | stringify.go | 509-514 | stringify (filter application) |
| #6 | stringify_jscompat_test.go | various | tests |
| #7 | stringify.go | 688-693 | stringify (keyPrefix generation) |

---

## JS Reference

Key file: `/Users/phl/Projects/qs/.ref/lib/stringify.js`

Key line numbers:
- 106-117: filter function application
- 137-139: undefined handling
- 142-153: objKeys building (comma, filter array, object keys)
- 155-161: prefix encoding and commaRoundTrip
- 163-176: main loop with skip logic and keyPrefix
- 255: allowDots calculation with encodeDotInKeys
- 287-293: root filter handling
