# Parse Fix Plan - Comparison with JS Version

## Overview

Current Go v2 implementation has **16 failed tests** out of 44 (36%). Main discrepancies with JS `qs` are related to:
1. Handling edge cases for depth/arrayLimit
2. Merge logic for values
3. Empty bracket handling
4. Charset encoding

---

## Phase 1: Critical Core Parsing Fixes

### 1.1 Fix `depth: 0` behavior

**Problem:** `depth: 0` should completely disable nesting
```
Input:  a[0]=b&a[1]=c with depth: 0
Expected: {"a[0]": "b", "a[1]": "c"}
Got:      {a: ["b", "c"]}
```

**JS code (parse.js:205-206):**
```js
var segment = options.depth > 0 && brackets.exec(key);
var parent = segment ? key.slice(0, segment.index) : key;
```

**Go code (parse.go:596-600):**
```go
if opts.Depth > 0 && segment != nil {
    parent = key[:segment[0]]
} else {
    parent = key
}
```

**Root Cause:**
Problem in `normalizeParseOptions` (line 226-228):
```go
if result.Depth == 0 {
    result.Depth = DefaultDepth  // <- This overwrites depth: 0 to 5!
}
```

**Solution:**
1. Add field `DepthSet bool` or use pointer `*int`
2. Or use special value (e.g., -1 = use default)

**Files:**
- `parse.go:197-244` - normalizeParseOptions
- `parse.go:339-343` - WithDepth

---

### 1.2 Fix `arrayLimit: 0` behavior

**Problem:** `arrayLimit: 0` should always return object, not array
```
Input:  a[1]=c with arrayLimit: 0
Expected: {a: {1: "c"}}
Got:      {a: ["c"]}
```

**JS code (parse.js:170-176):**
```js
if (
    !isNaN(index)
    && root !== decodedRoot
    && String(index) === decodedRoot
    && index >= 0
    && (options.parseArrays && index <= options.arrayLimit)  // <- Key condition
) {
    obj = [];
    obj[index] = leaf;
}
```

**Go code (parse.go:549):**
```go
} else if isValidIndex && opts.ParseArrays && index <= opts.ArrayLimit {
```

**Root Cause:**
Same issue - `normalizeParseOptions` overwrites `ArrayLimit: 0` to 20:
```go
if result.ArrayLimit == 0 {
    result.ArrayLimit = DefaultArrayLimit  // <- Overwrites!
}
```

**Solution:**
Same as depth - use explicit flag or pointer.

**Files:**
- `parse.go:223-225` - normalizeParseOptions
- `parse.go:549` - parseObject

---

### 1.3 Fix ParameterLimit (count parameters vs characters)

**Problem:** Limit should count number of parameters, not characters
```
Input:  a=b&c=d with parameterLimit: 1
Expected: {a: "b"}
Got:      {a: "b&c=d"}
```

**JS code (parse.js:66-74):**
```js
var limit = options.parameterLimit === Infinity ? undefined : options.parameterLimit;
var parts = cleanStr.split(options.delimiter,
    options.throwOnLimitExceeded ? limit + 1 : limit);

if (options.throwOnLimitExceeded && parts.length > limit) {
    throw new RangeError('Parameter limit exceeded...');
}
```

**Go code (parse.go:680-691):**
```go
limit := opts.ParameterLimit
if opts.ThrowOnLimitExceeded {
    limit = opts.ParameterLimit + 1
}
parts := splitByDelimiter(cleanStr, opts.Delimiter, opts.DelimiterRegexp, limit)
```

**Root Cause:**
Current `splitByDelimiter` implementation is incorrect. Verify:
1. JS `split(delimiter, limit)` splits into `limit` parts, not characters
2. Go should use `strings.SplitN` correctly

**Files:**
- `parse.go:453-482` - splitByDelimiter
- `parse.go:680-691` - parseValues

---

### 1.4 Support empty parent brackets `[]=a`

**Problem:** `[]=a` at root level should create `{0: "a"}`
```
Input:  []=a&[]=b
Expected: {0: "a", 1: "b"}
Got:      {}
```

**JS code (parse.js:205-219):**
```js
var segment = options.depth > 0 && brackets.exec(key);
var parent = segment ? key.slice(0, segment.index) : key;

var keys = [];
if (parent) {
    // ... push parent
    keys.push(parent);
}
```
When `key = "[]"`, `parent = ""` (empty string), and key is skipped.

**But** in JS this works through different mechanism - parseValues creates key `[]`,
which is then processed by parseObject.

**Go problem:**
In `parseKeys` (line 576-577):
```go
if givenKey == "" {
    return nil, nil
}
```
Key `[]` is not empty, but `parent = ""` handling skips it.

**Solution:**
Need to change logic in `parseKeys` to handle case when parent is empty,
but bracket segments exist.

**Files:**
- `parse.go:575-661` - parseKeys

---

## Phase 2: Merge Behavior Fixes

### 2.1 Mixed array/value order

**Problem:** Value order should match input order
```
Input:  a=b&a[]=c
Expected: {a: ["b", "c"]}
Got:      {a: ["c", "b"]}
```

**JS code (utils.js:79-93):**
```js
if (isArray(target) && isArray(source)) {
    source.forEach(function (item, i) {
        if (has.call(target, i)) {
            var targetItem = target[i];
            if (targetItem && typeof targetItem === 'object' && item && typeof item === 'object') {
                target[i] = merge(targetItem, item, options);
            } else {
                target.push(item);  // <- push preserves order
            }
        } else {
            target[i] = item;
        }
    });
    return target;
}
```

**Go code (utils.go:277-311):**
Problem is Go overwrites elements by index instead of push.

**Solution:**
1. Change array merge logic in `Merge`
2. When merging primitive with array - append to end

**Files:**
- `utils.go:232-333` - Merge

---

### 2.2 Adding keys to objects `a[b]=c&a=d`

**Problem:** Primitive should be added as key to existing object
```
Input:  a[b]=c&a=d
Expected: {a: {b: "c", d: true}}
Got:      {a: ["d", {b: "c"}]}
```

**JS code (utils.js:53-67):**
```js
if (typeof source !== 'object' && typeof source !== 'function') {
    if (isArray(target)) {
        target.push(source);
    } else if (target && typeof target === 'object') {
        if (
            (options && (options.plainObjects || options.allowPrototypes))
            || !has.call(Object.prototype, source)
        ) {
            target[source] = true;  // <- Adds as key!
        }
    } else {
        return [target, source];
    }
    return target;
}
```

**Go code (utils.go:242-256):**
```go
if sourceIsPrimitive {
    if targetSlice, ok := target.([]any); ok {
        return append(targetSlice, source)
    }
    if targetMap, ok := target.(map[string]any); ok {
        key, isString := source.(string)
        if isString && (allowPrototypes || !isPrototypeKey(key)) {
            targetMap[key] = true  // <- This works!
        }
        return targetMap
    }
    return []any{target, source}
}
```

**Root Cause:**
Problem is elsewhere - when merging `{b: "c"}` with `"d"` first target is map,
but somewhere the operation order is wrong. Need to check key iteration order
in `Parse`.

**Files:**
- `parse.go:871-886` - Parse main loop
- `utils.go:232-333` - Merge

---

### 2.3 Arrays to Objects transformation

**Problem:** When mixing string and numeric keys, all should be preserved
```
Input:  foo[bad]=baz&foo[0]=bar
Expected: {foo: {bad: "baz", 0: "bar"}}
Got:      {foo: {bad: "baz"}}
```

**JS code (utils.js:74-77):**
```js
var mergeTarget = target;
if (isArray(target) && !isArray(source)) {
    mergeTarget = arrayToObject(target, options);
}
```

**Go code (utils.go:314-319):**
```go
var mergeTarget map[string]any
if targetIsSlice {
    mergeTarget = ArrayToObject(targetSlice)
} else {
    mergeTarget = targetMap
}
```

**Root Cause:**
Problem is `ArrayToObject` skips nil values, but in JS
arrayToObject preserves all indices where `typeof source[i] !== 'undefined'`.

Also need to check logic in `parseObject` - possibly array is created
incorrectly initially.

**Files:**
- `utils.go:343-351` - ArrayToObject
- `parse.go:503-571` - parseObject

---

## Phase 3: Charset & Security

### 3.1 ISO-8859-1 to UTF-8 conversion

**Problem:** ISO-8859-1 bytes should be converted to UTF-8 Unicode
```
Input:  %A2=%BD with charset: "iso-8859-1"
Expected: {"¢": "½"}
Got:      {"\xa2": "\xbd"}
```

**JS code (utils.js:114-126):**
```js
var decode = function (str, defaultDecoder, charset) {
    var strWithoutPlus = str.replace(/\+/g, ' ');
    if (charset === 'iso-8859-1') {
        return strWithoutPlus.replace(/%[0-9a-f]{2}/gi, unescape);
    }
    // ... utf-8
};
```

**Go code (utils.go:186-207):**
```go
func decodeISO88591(str string) string {
    // ... decodes bytes as-is
}
```

**Solution:**
JS `unescape` interprets each %XX as Latin-1 character and returns
its Unicode equivalent. Go should do the same:
```go
// Each byte 0x00-0xFF should map to Unicode code point U+0000-U+00FF
result.WriteRune(rune(byte(hi<<4 | lo)))  // Instead of WriteByte
```

**Files:**
- `utils.go:186-207` - decodeISO88591

---

### 3.2 `__proto__` should ALWAYS be blocked

**Problem:** `__proto__` should be ignored even with `allowPrototypes: true`
```
Input:  categories[__proto__]=login with allowPrototypes: true
Expected: {}  // __proto__ ignored
Got:      {categories: {__proto__: ["login"]}}
```

**JS code (parse.js:179-181):**
```js
} else if (decodedRoot !== '__proto__') {
    obj[decodedRoot] = leaf;
}
```
Note - check for `__proto__` happens **OUTSIDE** the `allowPrototypes` block.

**Go code (parse.go:554-564):**
```go
} else if decodedRoot != "__proto__" {
    objMap[decodedRoot] = leaf
    obj = objMap
} else if opts.AllowPrototypes {  // <- This is wrong!
    objMap[decodedRoot] = leaf
    obj = objMap
}
```

**Solution:**
Remove `opts.AllowPrototypes` block for `__proto__`.
`__proto__` should **ALWAYS** be ignored regardless of options.

**Files:**
- `parse.go:554-564` - parseObject

---

### 3.3 Charset Sentinel switching

**Problem:** Sentinel should switch charset for entire string
```
Input:  utf8=%26%2310003%3B&%C3%B8=%C3%B8 with charsetSentinel: true, charset: "utf-8"
Expected: {"Ã¸": "Ã¸"} (interpreted as ISO-8859-1)
Got:      {"ø": "ø"} (still UTF-8)
```

**JS code:** Sentinel is determined before parsing values and applied to all.

**Go code (parse.go:695-709):**
```go
if opts.CharsetSentinel {
    for i, part := range parts {
        if strings.HasPrefix(part, "utf8=") {
            if part == charsetSentinel {
                charset = CharsetUTF8
            } else if part == isoSentinel {
                charset = CharsetISO88591
            }
            // ...
        }
    }
}
```

**Root Cause:**
Problem is that `%C3%B8` is first decoded by URL decoder, which
transforms it to "ø" (UTF-8). But if sentinel indicates ISO-8859-1,
need to interpret bytes 0xC3 0xB8 as two separate Latin-1 characters.

**Files:**
- `parse.go:695-709` - charset detection
- `utils.go:164-183` - Decode

---

## Phase 4: Additional Fixes

### 4.1 StrictNullHandling in arrays

**Problem:** Missing value should create null in array
```
Input:  a[0]=b&a[1]&a[2]=c with strictNullHandling: true
Expected: {a: ["b", null, "c"]}
Got:      {a: ["b", "c"]}
```

**Files:**
- `parse.go` - parseValues
- `utils.go` - Compact (should not remove null)

---

### 4.2 Sparse array pruning

**Problem:** When converting to object, both indices should be preserved
```
Input:  a[2]=b&a[99999999]=c
Expected: {a: {2: "b", 99999999: "c"}}
Got:      {a: {99999999: "c"}}
```

**Files:**
- `utils.go:343-351` - ArrayToObject

---

## Implementation Order

```
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Critical Fixes                                      │
├─────────────────────────────────────────────────────────────┤
│ 1. normalizeParseOptions - depth/arrayLimit zero values     │
│ 2. splitByDelimiter - parameter limit fix                   │
│ 3. parseKeys - empty parent brackets support                │
│ 4. parseObject - __proto__ always blocked                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Merge Behavior                                      │
├─────────────────────────────────────────────────────────────┤
│ 5. Merge - order preservation when merging arrays           │
│ 6. Merge - primitive-to-object addition                     │
│ 7. ArrayToObject - preserve all indices                     │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Charset & Edge Cases                                │
├─────────────────────────────────────────────────────────────┤
│ 8. decodeISO88591 - UTF-8 conversion                        │
│ 9. Charset sentinel - full string re-interpretation         │
│ 10. StrictNullHandling - null preservation in arrays        │
└─────────────────────────────────────────────────────────────┘
```

---

## Testing

After each fix:
```bash
# Run specific test
go test -v -run "TestJSDepthZero" ./v2/...

# Run all JS-compatible tests
go test -v -run "TestJS" ./v2/...

# With coverage
go test -v -run "TestJS" -coverprofile=cov.out ./v2/...
go tool cover -html=cov.out
```

---

## Expected Outcome

After all fixes:
- **44/44 tests** should pass (100%)
- JS qs compatibility ~95%+ for parse functionality
