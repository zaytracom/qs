# Stringify Fixes Plan for JS qs Library Compatibility

## Current Errors Analysis

After running tests, the following categories of issues were identified:

### 1. EncodeDotInKeys + AllowDots=false

**File:** `stringify.go`
**Problem:** When `encodeDotInKeys=true` and `allowDots=false`, Go implementation uses dot notation instead of bracket notation.

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

**Root cause in Go (stringify.go:243-246):**
```go
// If EncodeDotInKeys is true, AllowDots should also be true
if result.EncodeDotInKeys && !result.AllowDots {
    result.AllowDots = true  // INCORRECT!
}
```

**Fix:**
JS sets `allowDots=true` only if `allowDots` is **not explicitly set**. If `allowDots=false` is explicitly set, it should remain `false`.

```go
// WAS:
if result.EncodeDotInKeys && !result.AllowDots {
    result.AllowDots = true
}

// SHOULD BE:
// allowDots should NOT be automatically set to true
// if it was explicitly set to false
// Need to track whether AllowDots was explicitly set
```

**Solution:** Add `AllowDotsExplicit bool` field or use pointer `*bool` to track explicit setting.

---

### 2. Empty arrays in structures with `null` values

**File:** `stringify.go`
**Problem:** With `skipNulls=false`, empty strings in arrays are skipped instead of being output.

**Error example:**
```
input: {b: [""], c: "c"}
expected: "b[0]=&c=c"
got:      "c=c"
```

**JS behavior (lib/stringify.js:169-171):**
```javascript
if (skipNulls && value === null) {
    continue;
}
```
JS only checks for `null`, not empty strings or `undefined` in arrays.

**Root cause in Go (stringify.go:662-669):**
```go
// Skip nulls if requested, or skip nil in arrays (sparse array behavior like JS undefined)
if value == nil && isSlice(obj) {
    // In arrays, nil represents undefined (sparse slot) - always skip
    continue
}
if skipNulls && (value == nil || IsExplicitNull(value)) {
    continue
}
```

**Problem:** Go skips `nil` in arrays unconditionally. But empty string `""` should not be skipped.

**Fix:**
Ensure that `nil` is skipped only as sparse array slot, while empty values `""` are output.

---

### 3. strictNullHandling with arrays

**File:** `stringify.go`
**Problem:** With `strictNullHandling=true`, empty strings in arrays should be output as key without value.

**Error example:**
```
input: {b: [""], c: "c"}, strictNullHandling: true
expected: "b[0]&c=c"
got:      "c=c"
```

**Related to issue #2** - empty values in arrays are being skipped.

---

### 4. Filter with array of keys

**File:** `stringify.go`
**Problem:** Filter as array of keys is applied incorrectly - extra keys are added.

**Error example:**
```
input: {a: {b: [1, nil, 3]}}
filter: ["a", "b", "0", "2"]
expected: "a%5Bb%5D%5B0%5D=1&a%5Bb%5D%5B2%5D=3"
got:      "a%5Ba%5D=&a%5Bb%5D%5Ba%5D=1&a%5Bb%5D%5Bb%5D=1&a%5Bb%5D%5B0%5D=1&a%5Bb%5D%5B2%5D=3&a%5B0%5D=&a%5B2%5D=&b=&0=&2="
```

**JS behavior (lib/stringify.js:148-152):**
```javascript
if (isArray(filter)) {
    objKeys = filter;
} else {
    var keys = Object.keys(obj);
    objKeys = sort ? keys.sort(sort) : keys;
}
```
Filter as array is used only to determine which keys to output from current object, **not** for recursive application to all levels.

**Root cause in Go (stringify.go:591-596):**
```go
} else if filterSlice, ok := filter.([]string); ok {
    // Filter is array of keys
    objKeys = make([]any, len(filterSlice))
    for i, k := range filterSlice {
        objKeys[i] = k
    }
}
```
Go applies filter to all nesting levels, while JS applies only to the root.

**Fix:**
Filter as array should be applied **only at the top level** (in the main `Stringify` function), not in recursive `stringify`.

---

### 5. Filter function doesn't remove nil values

**File:** `stringify.go`
**Problem:** If filter function returns `nil`, the key should be skipped.

**Error example:**
```
input: {a: "b", c: nil, e: {f: time}}
filter: func returns nil for "c"
expected: "a=b&e%5Bf%5D=1257894000000"
got:      "a=b&c=&e%5Bf%5D=1257894000000"
```

**Fix:**
After applying filter function, check if `nil` was returned and skip such values.

---

### 6. Key order (map iteration order)

**File:** `stringify.go`
**Problem:** Go map doesn't guarantee iteration order. Tests expect a specific order.

**Error example:**
```
expected: "a=b&c[0]=d&c[1]=e&f[0][0]=g&f[1][0]=h"
got:      "f[0][0]=g&f[1][0]=h&a=b&c[0]=d&c[1]=e"
```

**Fix:**
Add `WithSort` to tests where specific order is required, or fix tests so they don't depend on order.

---

### 7. Empty key with nested array

**File:** `stringify.go`
**Problem:** When key is empty string `""` and contains an array, format should be `[][0]=val`.

**Error example:**
```
input: {"": [2, 3]}
expected: "[][0]=2&[][1]=3"
got:      "[0]=2&[1]=3"
```

**JS behavior (lib/stringify.js:174-176):**
```javascript
var keyPrefix = isArray(obj)
    ? typeof generateArrayPrefix === 'function' ? generateArrayPrefix(adjustedPrefix, encodedKey) : adjustedPrefix
    : adjustedPrefix + (allowDots ? '.' + encodedKey : '[' + encodedKey + ']');
```

With empty prefix and array, `generateArrayPrefix("")` should return `"[]"`.

**Root cause:**
In Go when prefix is empty and this is a root array, brackets are not added.

**Fix:**
For empty prefix with array, `[]` should be used as prefix.

---

### 8. repeat format with duplicated references

**File:** `stringify.go`
**Problem:** When using the same object in multiple places with repeat format.

**Related to handling of duplicated references.**

---

## Fix Priority

1. **Critical (affect core logic):**
   - #4 Filter with array of keys
   - #5 Filter function and nil
   - #2/#3 Empty values in arrays

2. **Medium priority:**
   - #1 EncodeDotInKeys + AllowDots
   - #7 Empty key with array

3. **Low priority (tests):**
   - #6 Key order - fix tests

---

## Detailed Fix Examples

### Fix #4: Filter array only at root level

**stringify.go - Stringify function (lines 765-771):**

```go
// WAS:
var filter any = normalizedOpts.Filter
var objKeys []string

// Handle filter
if filterFunc, ok := filter.(FilterFunc); ok {
    obj = filterFunc("", obj)
} else if fn, ok := filter.(func(string, any) any); ok {
    obj = fn("", obj)
} else if filterSlice, ok := filter.([]string); ok {
    objKeys = filterSlice
}
```

```go
// SHOULD BE:
var filter any = normalizedOpts.Filter
var objKeys []string
var filterForRecursion any = filter  // Filter function is passed recursively

// Handle filter
if filterFunc, ok := filter.(FilterFunc); ok {
    obj = filterFunc("", obj)
} else if fn, ok := filter.(func(string, any) any); ok {
    obj = fn("", obj)
} else if filterSlice, ok := filter.([]string); ok {
    objKeys = filterSlice
    filterForRecursion = nil  // Array filter is NOT passed to recursion
}
```

And in `stringify()` call pass `filterForRecursion` instead of `filter`.

---

### Fix #2/#3: Don't skip empty strings in arrays

**stringify.go - stringify function (lines 662-669):**

```go
// WAS:
// Skip nulls if requested, or skip nil in arrays (sparse array behavior like JS undefined)
if value == nil && isSlice(obj) {
    // In arrays, nil represents undefined (sparse slot) - always skip
    continue
}
if skipNulls && (value == nil || IsExplicitNull(value)) {
    continue
}
```

```go
// SHOULD BE:
// Skip sparse array slots (nil in arrays = undefined in JS)
if value == nil && isSlice(obj) {
    continue
}
// Skip nulls if requested (but not empty strings!)
if skipNulls && (value == nil || IsExplicitNull(value)) {
    continue
}
// Empty strings "" are NOT skipped - they should be output as key=
```

Issue: need to ensure that `""` (empty string) doesn't become `nil` somewhere earlier in the code.

---

### Fix #7: Empty key with array

**stringify.go - Stringify function (after getting keyValues):**

```go
// WAS (lines 815-848):
keyValues, err := stringify(
    value,
    key,  // <- key is used directly
    ...
)
```

```go
// SHOULD BE:
// For arrays at root level with empty key, need to use
// correct prefix format
var prefix string
if key == "" && isSlice(value) {
    // Empty key with array should generate [][0]=val
    prefix = ""  // generateArrayPrefix will add []
} else {
    prefix = key
}
keyValues, err := stringify(
    value,
    prefix,
    ...
)
```

Also need to check `arrayPrefixGenerators` - for empty prefix should return `[]`.

---

### Fix #1: EncodeDotInKeys should not automatically enable AllowDots

Need to change `StringifyOptions` structure or logic:

**Option 1: Use pointer**
```go
type StringifyOptions struct {
    AllowDots *bool  // nil = not set, true/false = explicitly set
    // ...
}
```

**Option 2: Separate field**
```go
type StringifyOptions struct {
    AllowDots         bool
    allowDotsExplicit bool  // internal field
    // ...
}
```

**Option 3: Logic in WithEncodeDotInKeys**
```go
// normalizeStringifyOptions:
// Do NOT automatically set AllowDots = true
// Instead check allowDots only where needed
```

Recommended **Option 3** - remove automatic setting and handle cases individually.

---

## JS vs Go Implementation Comparison

### JS lib/stringify.js key points:

1. **Line 142-147**: Comma format with arrays
```javascript
if (generateArrayPrefix === 'comma' && isArray(obj)) {
    if (encodeValuesOnly && encoder) {
        obj = utils.maybeMap(obj, encoder);
    }
    objKeys = [{ value: obj.length > 0 ? obj.join(',') || null : void undefined }];
}
```

2. **Line 169-171**: Skip only null, not undefined
```javascript
if (skipNulls && value === null) {
    continue;
}
```

3. **Line 174-176**: keyPrefix generation
```javascript
var keyPrefix = isArray(obj)
    ? typeof generateArrayPrefix === 'function' ? generateArrayPrefix(adjustedPrefix, encodedKey) : adjustedPrefix
    : adjustedPrefix + (allowDots ? '.' + encodedKey : '[' + encodedKey + ']');
```

4. **Line 301-302**: generateArrayPrefix for comma
```javascript
var generateArrayPrefix = arrayPrefixGenerators[options.arrayFormat];
var commaRoundTrip = generateArrayPrefix === 'comma' && options.commaRoundTrip;
```
`generateArrayPrefix` can be either string `'comma'` or function.

### Go differences:

1. Go uses `nil` for both sparse slots and explicit null - need to distinguish via `ExplicitNullValue`
2. Go automatically sets `AllowDots=true` when `EncodeDotInKeys=true` - incorrect
3. Go applies filter array recursively - incorrect
4. Go maps don't guarantee order - need sort for deterministic output
