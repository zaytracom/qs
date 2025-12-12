# Parser Fixes Plan

## Summary

25 JS compatibility tests failing. Issues grouped by root cause.

---

# Part 1: Algorithm Documentation

## Algorithm A: Key Path Parsing (FSM)

**Purpose:** Parse key like `a[b][0][c]` into path segments.

**Input:** Key string, ParseOptions
**Output:** []pathSegment

```
STATE MACHINE:

    ┌──────────────────────────────────────────────────────────────┐
    │                                                              │
    │  States:                                                     │
    │    IDENT   - reading identifier characters                   │
    │    BRACKET - inside [...] brackets                           │
    │                                                              │
    │  Transitions:                                                │
    │                                                              │
    │    IDENT ──'['──► BRACKET  (emit current buffer as segment)  │
    │    IDENT ──'.'──► IDENT    (emit segment, if AllowDots)      │
    │    IDENT ──EOF──► END      (emit final segment)              │
    │                                                              │
    │    BRACKET ──']'──► IDENT  (emit bracket segment)            │
    │    BRACKET ──chr──► BRACKET (accumulate character)           │
    │                                                              │
    └──────────────────────────────────────────────────────────────┘

ALGORITHM parseKeyPath(key, opts):

    // SPECIAL CASE 1: depth=0 means no nesting
    IF opts.Depth == 0:
        RETURN [Segment{value: key, isIndex: false, isEmpty: false}]

    // SPECIAL CASE 2: key starts with '[' without proper bracket notation
    IF len(key) > 0 AND key[0] == '[':
        closeBracket = indexOf(key, ']')
        IF closeBracket == -1:
            // No closing bracket - entire key is literal (e.g., "[" or "[[")
            RETURN [Segment{value: key}]
        IF closeBracket == 1:
            // Empty brackets "[]..." at start - special root array case
            // Continue to normal parsing but remember no parent
            hasParent = false
        ELSE:
            // "[something]..." - normal bracket notation with no parent
            hasParent = false

    segments = []
    buffer = ""
    state = IDENT
    depth = 0

    FOR i = 0 TO len(key):
        c = key[i]

        // Check depth limit BEFORE processing
        IF depth >= opts.Depth:
            // Remainder becomes literal segment
            remainder = key[i:]
            segments.append(Segment{value: remainder})
            BREAK

        SWITCH state:
            CASE IDENT:
                IF c == '[':
                    IF buffer != "":
                        segments.append(makeSegment(buffer, opts))
                        buffer = ""
                    depth++
                    state = BRACKET

                ELSE IF c == '.' AND opts.AllowDots:
                    IF buffer != "":
                        segments.append(makeSegment(buffer, opts))
                        buffer = ""
                    depth++
                    // Stay in IDENT

                ELSE:
                    buffer += c

            CASE BRACKET:
                IF c == ']':
                    segments.append(makeBracketSegment(buffer, opts))
                    buffer = ""
                    state = IDENT
                ELSE:
                    buffer += c

    // Handle remaining buffer
    IF buffer != "":
        segments.append(makeSegment(buffer, opts))

    RETURN segments


FUNCTION makeBracketSegment(content, opts):
    // Empty brackets "[]"
    IF content == "":
        IF opts.ParseArrays:
            RETURN Segment{isEmpty: true}
        ELSE:
            // parseArrays=false: empty brackets become "0" key
            RETURN Segment{value: "0", isIndex: false}

    // Check if numeric index
    IF opts.ParseArrays:
        IF isNumeric(content) AND content == strconv.Itoa(parseInt(content)):
            index = parseInt(content)
            IF index >= 0 AND index <= opts.ArrayLimit:
                RETURN Segment{value: content, isIndex: true, index: index}
            // Index exceeds limit - treat as string key

    RETURN Segment{value: content, isIndex: false}
```

---

## Algorithm B: Tree Construction

**Purpose:** Build nested map/slice structure from path segments and value.

**Input:** []pathSegment, value, ParseOptions
**Output:** Modified tree (root map)

```
ALGORITHM insert(root, path, value, opts):

    // STEP 1: Security check - scan ALL segments for blocked keys
    FOR seg IN path:
        // __proto__ is ALWAYS blocked (security)
        IF seg.value == "__proto__":
            RETURN  // silently skip entire path

        // Other prototype keys blocked unless AllowPrototypes
        IF !opts.PlainObjects AND !opts.AllowPrototypes:
            IF isPrototypeKey(seg.value):  // constructor, hasOwnProperty, etc.
                RETURN  // silently skip

    // STEP 2: Handle single segment (direct key in root)
    IF len(path) == 1:
        insertAtRoot(root, path[0], value, opts)
        RETURN

    // STEP 3: Navigate/create path to parent of final segment
    current = root
    FOR i = 0 TO len(path) - 2:
        seg = path[i]
        nextSeg = path[i+1]

        current = ensureContainer(current, seg, nextSeg, path[:i], opts)
        IF current == nil:
            RETURN  // navigation failed

    // STEP 4: Insert value at final segment
    insertAtContainer(current, path[last], value, opts)


ALGORITHM ensureContainer(current, seg, nextSeg, parentPath, opts):

    SWITCH typeof(current):

        CASE map[string]any:
            key = seg.value
            IF seg.isEmpty:
                key = "0"  // empty bracket in map context

            existing = current[key]

            // CRITICAL: If existing is primitive but we need to descend,
            // convert to array with primitive as element [0]
            IF existing != nil AND isPrimitive(existing):
                IF nextSeg.isIndex OR nextSeg.isEmpty:
                    newArray = []any{existing}
                    current[key] = newArray
                    RETURN newArray
                ELSE:
                    // Next is string key - convert primitive to map?
                    // JS behavior: primitive becomes element, create array
                    newArray = []any{existing}
                    current[key] = newArray
                    RETURN newArray

            IF existing != nil:
                RETURN existing

            // Create new container based on next segment
            IF (nextSeg.isIndex OR nextSeg.isEmpty) AND opts.ParseArrays:
                newContainer = []any{}
            ELSE:
                newContainer = map[string]any{}

            current[key] = newContainer
            RETURN newContainer

        CASE []any:
            // String key on array - convert array to map
            IF !seg.isIndex AND !seg.isEmpty:
                mapVersion = arrayToMap(current)
                updateParent(parentPath, mapVersion)  // update reference
                RETURN ensureContainer(mapVersion, seg, nextSeg, parentPath, opts)

            idx = seg.index
            IF seg.isEmpty:
                idx = len(current)  // append position

            // Extend array if needed
            WHILE len(current) <= idx:
                current = append(current, nil)
            updateParent(parentPath, current)

            existing = current[idx]
            IF existing != nil AND isPrimitive(existing):
                // Same logic: convert primitive to container
                IF nextSeg.isIndex OR nextSeg.isEmpty:
                    newArray = []any{existing}
                    current[idx] = newArray
                    updateParent(parentPath, current)
                    RETURN newArray

            IF existing != nil:
                RETURN existing

            // Create new container
            IF (nextSeg.isIndex OR nextSeg.isEmpty) AND opts.ParseArrays:
                newContainer = []any{}
            ELSE:
                newContainer = map[string]any{}

            current[idx] = newContainer
            updateParent(parentPath, current)
            RETURN newContainer

    RETURN nil


ALGORITHM insertAtContainer(current, seg, value, opts):

    SWITCH typeof(current):

        CASE map[string]any:
            key = seg.value

            // Empty brackets with ParseArrays = array append
            IF seg.isEmpty AND opts.ParseArrays:
                IF opts.AllowEmptyArrays AND (value == "" OR isNull(value)):
                    IF current[key] == nil:
                        current[key] = []any{}
                    RETURN

                existing = current[key]
                IF existing == nil:
                    current[key] = []any{value}
                ELSE IF isSlice(existing):
                    current[key] = append(existing, value)
                ELSE:
                    current[key] = []any{existing, value}
                RETURN

            // Normal key insertion
            existing = current[key]
            IF existing == nil:
                current[key] = value
            ELSE:
                current[key] = handleDuplicate(existing, value, opts)

        CASE []any:
            // Similar logic for slice insertion...


ALGORITHM handleDuplicate(existing, newVal, opts):

    // Respect duplicate handling mode
    SWITCH opts.Duplicates:
        CASE "first": RETURN existing
        CASE "last":  RETURN newVal
        // "combine" falls through

    // CRITICAL JS BEHAVIOR: primitive added to map becomes key with true
    IF isMap(existing) AND isString(newVal):
        key = newVal.(string)
        IF key != "__proto__":
            IF opts.AllowPrototypes OR !isPrototypeKey(key):
                existing[key] = true
        RETURN existing

    // Primitive added to slice - append
    IF isSlice(existing):
        RETURN append(existing, newVal)

    // Two primitives - create array
    IF isPrimitive(existing) AND isPrimitive(newVal):
        RETURN []any{existing, newVal}

    // Fallback
    RETURN []any{existing, newVal}
```

---

## Algorithm C: Comma Value Splitting

**Purpose:** Split comma-separated values BEFORE URL decoding to preserve encoded commas.

**Input:** Raw (encoded) value string, charset, options
**Output:** Single value or []any of values

```
ALGORITHM processCommaValue(rawValue, charset, opts):

    // Only split if comma option enabled
    IF !opts.Comma:
        RETURN decode(rawValue, charset)

    // CRITICAL: Split by LITERAL comma before decoding
    // This preserves %2C (encoded comma) as single value

    parts = []string{}
    current = ""

    FOR i = 0 TO len(rawValue):
        c = rawValue[i]

        IF c == ',':
            // Literal comma - split point
            parts.append(current)
            current = ""
        ELSE:
            current += c

    parts.append(current)

    // Single part - return decoded string
    IF len(parts) == 1:
        RETURN decode(parts[0], charset)

    // Multiple parts - decode each, return array
    result = []any{}
    FOR part IN parts:
        result.append(decode(part, charset))

    RETURN result


EXAMPLE:
    Input:  "a%2Cb,c"  (comma=true)

    Split by literal ',':
        parts = ["a%2Cb", "c"]

    Decode each:
        result = ["a,b", "c"]

    NOT:
        decode first: "a,b,c"
        split: ["a", "b", "c"]  // WRONG - loses encoded comma
```

---

## Algorithm D: Charset Sentinel Detection

**Purpose:** Detect charset from utf8=✓ or utf8=&#10003; sentinel parameter.

**Input:** Query string parts array, default charset
**Output:** (detected charset, index to skip)

```
CONSTANTS:
    UTF8_SENTINEL = "utf8=%E2%9C%93"      // encodeURIComponent('✓')
    ISO_SENTINEL  = "utf8=%26%2310003%3B" // encodeURIComponent('&#10003;')

ALGORITHM detectCharset(parts, defaultCharset):

    FOR i, part IN parts:
        IF startsWith(part, "utf8="):
            rest = part[5:]  // after "utf8="

            IF rest == "%E2%9C%93":
                RETURN (UTF8, i)  // skip this index

            IF rest == "%26%2310003%3B":
                RETURN (ISO88591, i)  // skip this index

            // Unknown utf8= value - don't change charset, don't skip
            RETURN (defaultCharset, -1)

    // No sentinel found
    RETURN (defaultCharset, -1)


NOTE: Detection must happen on PARTS array (after split by delimiter),
      not on raw string. The skip index is array index, not string position.
```

---

## Algorithm E: ISO-8859-1 Decoding

**Purpose:** Decode percent-encoded string as ISO-8859-1 (Latin-1).

**Input:** Encoded string
**Output:** Decoded string with correct Unicode characters

```
ALGORITHM decodeISO88591(input):

    result = StringBuilder()

    i = 0
    WHILE i < len(input):
        c = input[i]

        IF c == '+':
            result.append(' ')
            i++

        ELSE IF c == '%' AND i+2 < len(input):
            hi = hexValue(input[i+1])
            lo = hexValue(input[i+2])

            IF hi >= 0 AND lo >= 0:
                byteVal = (hi << 4) | lo

                // KEY: ISO-8859-1 byte 0x00-0xFF maps directly to Unicode U+0000-U+00FF
                // Use WriteRune to output as UTF-8 encoded Unicode codepoint
                result.WriteRune(rune(byteVal))
                i += 3
                CONTINUE

            // Invalid hex - keep literal
            result.append(c)
            i++

        ELSE:
            result.append(c)
            i++

    RETURN result.String()


EXAMPLE:
    Input:  "%A2"  (ISO-8859-1)

    byteVal = 0xA2 = 162
    rune(162) = U+00A2 = '¢'

    Output: "¢"
```

---

## Algorithm F: Empty Bracket Root Handling

**Purpose:** Handle `[]=a&[]=b` as `{"0": "a", "1": "b"}`.

**Input:** Tokens with empty key and isEmpty segment
**Output:** Map with incrementing numeric string keys

```
ALGORITHM:

    When key path is just [] (empty brackets, no parent):

    path = [Segment{value: "", isEmpty: true}]

    In insertAtRoot or tree builder:

    IF path[0].isEmpty AND path[0].value == "":
        // Root-level empty bracket - use incrementing index
        // Track next available index for root empty brackets
        key = strconv.Itoa(rootEmptyBracketCounter)
        rootEmptyBracketCounter++
        root[key] = value
        RETURN

    This handles:
        Input:  "[]=a&[]=b&a=c"
        Output: {"0": "a", "1": "b", "a": "c"}

    Note: This is different from "a[]=x" which creates array at key "a".
    Root-level [] creates map with numeric string keys.
```

---

# Part 2: Issues and Fixes

## Issue 1: Depth 0 Handling

**Tests:** `TestJSDepthZero`, `TestDetailedJSCompat/depth_0`, `TestJSStrictDepthNoThrow`

**Problem:** When `depth=0`, keys should remain as literals without any nesting.

**Current behavior:**
```
Input:  "a[0]=b&a[1]=c" with depth=0
Got:    {a: {"[0]": "b", "[1]": "c"}}
Want:   {"a[0]": "b", "a[1]": "c"}
```

**Root cause:** `parseKeyPath()` in `parse_keypath.go` still parses brackets when depth=0.

**Fix algorithm:**
```
parseKeyPath(key, opts):
    IF opts.Depth == 0:
        // Return single segment with entire key as literal
        RETURN [Segment{value: key, isIndex: false, isEmpty: false}]

    // ... rest of parsing
```

**Also:** `StrictDepth` with `depth=0` should NOT throw - depth 0 means "don't parse nesting", not "error on any brackets".

---

## Issue 2: Mixed Primitive + Array Merge

**Tests:** `TestJSMixedArrays`, `TestJSEmptyKeys`, `TestDetailedJSCompat2/mixed_3`

**Problem:** When a key first has a primitive value, then array notation is used, the primitive should be converted to array element.

**Current behavior:**
```
Input:  "a=b&a[]=c"
Got:    {a: "b"}           // second value lost
Want:   {a: ["b", "c"]}
```

**Root cause:** `treeBuilder.insert()` doesn't handle converting existing primitive to array when path has array notation.

**Fix algorithm:**
```
insertPath(path, value):
    // Navigate to insertion point
    current = root
    for i = 0 to len(path)-2:
        seg = path[i]
        nextSeg = path[i+1]

        existing = getFromContainer(current, seg)

        // KEY FIX: If existing is primitive and we need to descend further,
        // convert primitive to array with primitive as first element
        IF existing != nil AND isPrimitive(existing) AND (nextSeg.isIndex OR nextSeg.isEmpty):
            newArray = []any{existing}
            setInContainer(current, seg, newArray)
            existing = newArray

        current = ensureContainer(current, seg, nextSeg)

    // Insert at final position
    insertAtContainer(current, path[last], value)
```

---

## Issue 3: Add Primitive Key to Object

**Tests:** `TestJSAddKeysToObjects`, `TestDetailedJSCompat/add_key_to_obj`

**Problem:** When object exists at key and primitive is added to same key, primitive becomes key with `true` value.

**Current behavior:**
```
Input:  "a[b]=c&a=d"
Got:    {a: [{b: "c"}, "d"]}   // creates array
Want:   {a: {b: "c", d: true}} // adds key to object
```

**Root cause:** `handleDuplicate()` creates array instead of checking if existing is map.

**Fix algorithm:**
```
handleDuplicate(existing, newVal, opts):
    SWITCH opts.Duplicates:
        CASE DuplicateFirst: RETURN existing
        CASE DuplicateLast: RETURN newVal

    // DuplicateCombine (default)

    // KEY FIX: If existing is map and new is primitive string,
    // add primitive as key with value true (JS behavior)
    IF isMap(existing) AND isString(newVal):
        key = newVal.(string)
        IF key != "__proto__" AND (opts.AllowPrototypes OR !isPrototypeProp(key)):
            existing[key] = true
        RETURN existing

    // If existing is slice, append
    IF isSlice(existing):
        RETURN append(existing, newVal)

    // Both primitives - create array
    RETURN []any{existing, newVal}
```

---

## Issue 4: Prototype Key Blocking

**Tests:** `TestJSPrototypeProtection`, `TestDetailedJSCompat/hasOwn_blocked`

**Problem:** `hasOwnProperty`, `constructor`, etc. should be blocked by default (not just `__proto__`).

**Current behavior:**
```
Input:  "a[hasOwnProperty]=b"
Got:    {a: {hasOwnProperty: "b"}}
Want:   {}  // entire path skipped
```

**Root cause:** `isPrototypeProp()` check exists but not applied consistently in `treeBuilder.insert()`.

**Fix algorithm:**
```
insert(path, value):
    // Check ALL segments for prototype keys (not just first)
    FOR seg IN path:
        IF seg.value == "__proto__":
            RETURN  // always block
        IF !opts.PlainObjects AND isPrototypeProp(seg.value) AND !opts.AllowPrototypes:
            RETURN  // block other prototype keys

    // Proceed with insertion...
```

---

## Issue 5: `__proto__` in Nested Paths

**Tests:** `TestJSDunderProto`, `TestDetailedJSCompat2/__proto___blocked`

**Problem:** `__proto__` should be blocked anywhere in path, but current code only checks first segment.

**Current behavior:**
```
Input:  "categories[__proto__][]=login"
Got:    {categories: {__proto__: ["login"]}}
Want:   {categories: {}}
```

**Root cause:** `__proto__` check only at path start, not during traversal.

**Fix:** Same as Issue 4 - check all segments before insertion.

---

## Issue 6: Keys Starting with `[`

**Tests:** `TestJSStartingWithOpeningBracket`, `TestDetailedJSCompat/[]_prefix`

**Problem:** Keys that start with `[` should preserve the `[` as part of key name.

**Current behavior:**
```
Input:  "[=toString"
Got:    {}
Want:   {"[": "toString"}

Input:  "[[=toString"
Got:    {"[": "toString"}
Want:   {"[[": "toString"}
```

**Root cause:** `parseKeyPath()` treats leading `[` as bracket notation start.

**Fix algorithm:**
```
parseKeyPath(key, opts):
    // Special case: key starts with [ but no matching ]
    // This is a literal key, not bracket notation
    IF key[0] == '[':
        eqPos = indexOf(key, '=')
        closeBracketPos = indexOf(key, ']')

        // If no ] before = (or no ] at all), treat whole thing as literal
        IF closeBracketPos == -1 OR (eqPos != -1 AND closeBracketPos > eqPos):
            RETURN [Segment{value: key}]

    // ... normal parsing
```

Wait, the `=` is already stripped by tokenizer. Let me reconsider:

```
parseKeyPath(key, opts):
    // Key is already without value, e.g. "[" or "[["
    IF key[0] == '[':
        // Find first ]
        closeBracket = indexOf(key, ']')
        IF closeBracket == -1:
            // No closing bracket - entire key is literal
            RETURN [Segment{value: key}]
        // Has closing bracket - check if it's at position 1 (empty brackets)
        // or further (content inside)
        // For "[hello[" - this has no ] so literal
        // For "[]" - empty brackets, special handling

    // ... normal parsing
```

Actually simpler: if key starts with `[` and there's no `]` at position 1 (meaning it's not `[]` or `[x]`), treat as literal.

---

## Issue 7: Empty Brackets Without Parent → Index 0

**Tests:** `TestJSNoParentFound`, `TestJSEmptyKeysWithBrackets`

**Problem:** `[]=a&[]=b` should become `{"0": "a", "1": "b"}`, not `{"": ["a", "b"]}`.

**Current behavior:**
```
Input:  "[]=a&[]=b"
Got:    {"": ["a", "b"]}
Want:   {"0": "a", "1": "b"}
```

**Root cause:** Empty bracket `[]` without parent key is treated as array append to empty key, but should use numeric indices as keys.

**Fix algorithm:**
```
// In parseKeyPath or treeBuilder:
// When key is just "[]" (empty brackets at root level),
// treat as incrementing numeric key

// Track counter per empty-bracket-at-root
insertEmptyBracketRoot(value):
    index = nextEmptyBracketIndex++
    root[strconv.Itoa(index)] = value
```

---

## Issue 8: Comma Split BEFORE Decode

**Tests:** `TestJSPercentEncodedComma`

**Problem:** `%2C` (encoded comma) should NOT be split - only literal `,` should split.

**Current behavior:**
```
Input:  "foo=a%2Cb" with comma=true
Got:    {foo: ["a", "b"]}   // split encoded comma
Want:   {foo: "a,b"}        // preserve encoded comma
```

**Root cause:** In `parse_fast.go:54-61`, comma split happens on `tok.value` which is already decoded.

**Fix algorithm:**
```
// In tokenizer, BEFORE decoding value:
processToken(part, charset, opts):
    keyPart, valPart = split by '='

    // Decode key
    key = decode(keyPart, charset)

    // Handle comma BEFORE decoding
    IF opts.Comma AND contains(valPart, ','):
        // Split by LITERAL comma only
        parts = splitByLiteralComma(valPart)  // not %2C
        values = []
        FOR part IN parts:
            values.append(decode(part, charset))
        RETURN token{key, values, hasValue: true}

    // Normal decode
    value = decode(valPart, charset)
    RETURN token{key, value, hasValue: true}
```

---

## Issue 9: ISO-8859-1 Charset Decoding

**Tests:** `TestJSCharset`, `TestJSCharsetSentinel`, `TestDetailedJSCompat/iso-8859-1`

**Problem:** ISO-8859-1 decoding not working correctly.

**Current behavior:**
```
Input:  "%A2=%BD" with charset=iso-8859-1
Got:    {"�": "�"}
Want:   {"¢": "½"}
```

**Root cause:** `fastDecode` or tokenizer not passing charset correctly to decoder.

**Fix:** Verify charset propagation through tokenize → parseToken → decoder.

---

## Issue 10: Charset Sentinel Detection Position

**Tests:** `TestJSCharsetSentinel`, `TestJSUTF8SentinelDetectsISO`

**Problem:** Sentinel detection returns string position, but skip uses array index.

**Current behavior:**
```
Input:  "a=%C3%B8&utf8=%26%2310003%3B" (sentinel at end)
Got:    {a: "��", utf8: "&#10003;"}  // sentinel not skipped, wrong charset
Want:   {a: "ø"}                      // sentinel skipped, ISO applied
```

**Root cause:** `detectCharsetFromSentinel` returns position in string, but `skipIndex` is used as parts array index.

**Fix algorithm:**
```
// Option A: Return the utf8=... parameter itself and skip by value match
detectCharsetFromSentinel(input, defaultCharset):
    // Find "utf8=" parameter
    // Return (charset, sentinelValue) instead of position

// Option B: Calculate array index from string position
// Find which part contains the sentinel by re-scanning parts
```

---

## Issue 11: ArrayLimit with ThrowOnLimitExceeded

**Tests:** `TestJSArrayLimitTests`

**Problem:** `ThrowOnLimitExceeded` should error when array index exceeds `ArrayLimit`.

**Current behavior:**
```
Input:  "a[999]=b" with arrayLimit=20, throwOnLimitExceeded=true
Got:    nil error
Want:   ErrArrayLimitExceeded
```

**Root cause:** Array limit check missing or not throwing.

**Fix:** In `makeBracketSegment` or `treeBuilder.insert`, check index against limit and return error if `ThrowOnLimitExceeded`.

---

## Issue 12: Empty Brackets with parseArrays=false

**Tests:** `TestJSDisableArrayParsing`

**Problem:** With `parseArrays=false`, `a[]=b` should become `{"a": {"0": "b"}}`.

**Current behavior:**
```
Input:  "a[]=b" with parseArrays=false
Got:    {a: {"": "b"}}
Want:   {a: {"0": "b"}}
```

**Root cause:** Empty bracket converted to empty string key instead of "0".

**Fix algorithm:**
```
makeBracketSegment(content, opts):
    IF content == "" AND !opts.ParseArrays:
        // Empty brackets with parseArrays=false → "0" key
        RETURN Segment{value: "0", isIndex: false}
    // ... rest
```

---

## Issue 13: Arrays to Objects Conversion

**Tests:** `TestJSArraysToObjects`

**Problem:** When string key is added to array, array should convert to object with numeric string keys.

**Current behavior:**
```
Input:  "foo[0]=bar&foo[bad]=baz"
Got:    {foo: {"": ["bar"], bad: "baz"}}
Want:   {foo: {"0": "bar", bad: "baz"}}
```

**Root cause:** `sliceToMap` conversion not triggered or done incorrectly.

**Fix:** When inserting string key into slice container, convert slice to map first.

---

## Implementation Order

1. **Issue 1: Depth 0** - Simple fix in `parseKeyPath`
2. **Issue 6: Keys starting with `[`** - Fix in `parseKeyPath`
3. **Issue 12: Empty brackets parseArrays=false** - Fix in `makeBracketSegment`
4. **Issue 8: Comma split before decode** - Fix in tokenizer
5. **Issue 9: ISO-8859-1** - Fix charset propagation
6. **Issue 10: Charset sentinel** - Fix detection/skip logic
7. **Issue 4+5: Prototype blocking** - Fix in `treeBuilder.insert`
8. **Issue 7: Empty brackets at root** - Fix root-level empty bracket handling
9. **Issue 2: Mixed primitive + array** - Fix merge logic
10. **Issue 3: Add key to object** - Fix `handleDuplicate`
11. **Issue 11: ArrayLimit throw** - Add error check
12. **Issue 13: Array to object** - Fix `sliceToMap` trigger

---

## Files to Modify

| File | Issues |
|------|--------|
| `parse_keypath.go` | 1, 6, 12 |
| `parse_tokenizer.go` | 8, 9, 10 |
| `parse_decode.go` | 9 |
| `parse_tree.go` | 2, 3, 4, 5, 7, 11, 13 |
| `parse_fast.go` | 8 (remove comma handling here) |
