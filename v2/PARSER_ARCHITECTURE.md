# Query String Parser Architecture

## Overview

High-performance query string parser compatible with Node.js `qs` library.
Single-pass architecture with O(n) complexity for most operations.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INPUT STRING                                    │
│                    "a[b][c]=value&d.e=test&f[]=1&f[]=2"                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         PHASE 1: PREPROCESSING                               │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  1. Strip query prefix (?) if IgnoreQueryPrefix=true                │    │
│  │  2. Detect charset from sentinel (utf8=✓ or utf8=&#10003;)          │    │
│  │  3. Decode %5B/%5D to [ ] for bracket detection                     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      PHASE 2: TOKENIZATION (Single Pass)                     │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Input: "a[b]=1&c=2&d"                                              │    │
│  │                                                                      │    │
│  │  Algorithm: Linear scan with state tracking                         │    │
│  │  - State: SCANNING_KEY | SCANNING_VALUE                             │    │
│  │  - Track positions: keyStart, keyEnd, valueStart, valueEnd          │    │
│  │  - On '&' or EOF: emit token, reset state                           │    │
│  │  - On '=' (first): transition to SCANNING_VALUE                     │    │
│  │                                                                      │    │
│  │  Output: [(key="a[b]", val="1"), (key="c", val="2"), (key="d", ∅)]  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Complexity: O(n) single pass                                                │
│  Allocations: 1 slice for tokens, pre-allocated based on input length       │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                    PHASE 3: KEY PATH PARSING (Per Token)                     │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Input: "a[b][0][c]" or "a.b.0.c" (with AllowDots)                  │    │
│  │                                                                      │    │
│  │  Finite State Machine:                                               │    │
│  │  ┌──────────┐    '['     ┌──────────┐    ']'    ┌──────────┐        │    │
│  │  │  IDENT   │ ─────────► │ BRACKET  │ ────────► │  IDENT   │        │    │
│  │  └──────────┘            └──────────┘           └──────────┘        │    │
│  │       │                       │                      │               │    │
│  │       │ '.' (if AllowDots)    │ any char             │               │    │
│  │       ▼                       ▼                      │               │    │
│  │  emit segment            accumulate              loop back           │    │
│  │                                                                      │    │
│  │  Output: [Segment{value="a"}, Segment{value="b"},                   │    │
│  │           Segment{value="0", isIndex=true, index=0},                │    │
│  │           Segment{value="c"}]                                       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Special Cases:                                                              │
│  - Depth=0: Return entire key as single literal segment                     │
│  - Depth=N: After N segments, remainder becomes literal                     │
│  - Key starts with '[': Handle as bracket without parent                    │
│  - Empty brackets '[]': Mark as isEmpty=true for array append               │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     PHASE 4: VALUE PROCESSING                                │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  1. URL Decode (charset-aware)                                      │    │
│  │     - UTF-8: Standard percent decoding                              │    │
│  │     - ISO-8859-1: Byte-to-Unicode mapping (0x00-0xFF → U+0000-00FF) │    │
│  │                                                                      │    │
│  │  2. Comma Split (if Comma=true)                                     │    │
│  │     IMPORTANT: Split BEFORE decoding to preserve %2C                │    │
│  │     "a%2Cb,c" → split by literal ',' → ["a%2Cb", "c"]              │    │
│  │              → decode each → ["a,b", "c"]                           │    │
│  │                                                                      │    │
│  │  3. Numeric Entities (if InterpretNumericEntities + ISO-8859-1)     │    │
│  │     "&#9786;" → "☺" (U+263A)                                        │    │
│  │                                                                      │    │
│  │  4. Null Handling                                                   │    │
│  │     - No '=' sign + StrictNullHandling: value = null                │    │
│  │     - No '=' sign + !StrictNullHandling: value = ""                 │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      PHASE 5: TREE CONSTRUCTION                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Algorithm: Direct path insertion (no recursive merge)              │    │
│  │                                                                      │    │
│  │  For each (path, value) pair:                                       │    │
│  │  1. Security check: Skip if any segment is "__proto__"              │    │
│  │  2. Prototype check: Skip if segment in prototypeKeys && !AllowProto│    │
│  │  3. Navigate/create containers along path                           │    │
│  │  4. Insert value at leaf with duplicate handling                    │    │
│  │                                                                      │    │
│  │  Container Creation Rules:                                          │    │
│  │  ┌─────────────────────────────────────────────────────────────┐    │    │
│  │  │ Next Segment    │ ParseArrays │ Container Type              │    │    │
│  │  ├─────────────────┼─────────────┼─────────────────────────────┤    │    │
│  │  │ isIndex=true    │ true        │ []any (slice)               │    │    │
│  │  │ isEmpty=true    │ true        │ []any (slice)               │    │    │
│  │  │ isIndex=true    │ false       │ map[string]any              │    │    │
│  │  │ string key      │ any         │ map[string]any              │    │    │
│  │  └─────────────────┴─────────────┴─────────────────────────────┘    │    │
│  │                                                                      │    │
│  │  Duplicate Handling:                                                │    │
│  │  ┌─────────────────────────────────────────────────────────────┐    │    │
│  │  │ Mode     │ Behavior                                         │    │    │
│  │  ├──────────┼─────────────────────────────────────────────────┤    │    │
│  │  │ combine  │ [existing...] + [new...] → merged array         │    │    │
│  │  │ first    │ keep existing, ignore new                       │    │    │
│  │  │ last     │ replace with new                                │    │    │
│  │  └──────────┴─────────────────────────────────────────────────┘    │    │
│  │                                                                      │    │
│  │  Mixed Type Merge (JS qs behavior):                                 │    │
│  │  ┌─────────────────────────────────────────────────────────────┐    │    │
│  │  │ Existing │ New              │ Result                        │    │    │
│  │  ├──────────┼──────────────────┼───────────────────────────────┤    │    │
│  │  │ string   │ array notation   │ [string, ...newValues]        │    │    │
│  │  │ map      │ primitive        │ map[primitive] = true         │    │    │
│  │  │ slice    │ primitive        │ append(slice, primitive)      │    │    │
│  │  │ slice    │ string key       │ convert to map, add key       │    │    │
│  │  └──────────┴──────────────────┴───────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                       PHASE 6: POST-PROCESSING                               │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  1. Sparse Array Compaction (if AllowSparse=false)                  │    │
│  │     [a, nil, b, nil, c] → [a, b, c]                                 │    │
│  │                                                                      │    │
│  │  2. ExplicitNull Conversion                                         │    │
│  │     ExplicitNullValue markers → nil                                 │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              OUTPUT                                          │
│                         map[string]any                                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Data Structures

### Token
```go
type token struct {
    // Raw positions in input (for zero-copy when possible)
    keyStart, keyEnd   int
    valStart, valEnd   int

    // Decoded values (populated lazily or during processing)
    key      string
    value    string
    hasValue bool  // distinguishes "key" from "key="
}
```

### Path Segment
```go
type pathSegment struct {
    value   string  // segment text (decoded)
    isIndex bool    // true if valid numeric index within ArrayLimit
    index   int     // numeric value (valid only if isIndex)
    isEmpty bool    // true for [] (empty brackets)
}
```

### Tree Builder
```go
type treeBuilder struct {
    root map[string]any
    opts *ParseOptions
}
```

---

## Algorithm Details

### Algorithm 1: Charset Sentinel Detection

**Purpose**: Determine charset before parsing values

**Input**: Raw query string, default charset
**Output**: Detected charset, sentinel position (or -1)

```
FUNCTION detectCharset(input, defaultCharset):
    // Scan for "utf8=" pattern
    pos = findSubstring(input, "utf8=")
    IF pos == -1:
        RETURN (defaultCharset, -1)

    rest = input[pos+5:]

    // Check UTF-8 sentinel: %E2%9C%93 (encoded ✓)
    IF startsWith(rest, "%E2%9C%93"):
        RETURN (UTF8, pos)

    // Check ISO-8859-1 sentinel: %26%2310003%3B (encoded &#10003;)
    IF startsWith(rest, "%26%2310003%3B"):
        RETURN (ISO88591, pos)

    // Unknown sentinel value - don't change charset, don't skip
    RETURN (defaultCharset, -1)
```

**Complexity**: O(n) worst case, typically O(1) as sentinel is usually at start

---

### Algorithm 2: Single-Pass Tokenizer

**Purpose**: Extract key-value pairs in one pass

**Input**: Preprocessed query string, options
**Output**: Array of tokens

```
FUNCTION tokenize(input, opts):
    tokens = []
    charset = opts.Charset
    skipIndex = -1

    // Phase 1: Detect charset if sentinel enabled
    IF opts.CharsetSentinel:
        (charset, skipIndex) = detectCharset(input, charset)

    // Phase 2: Single-pass tokenization
    state = SCANNING_KEY
    keyStart = 0
    keyEnd = -1
    valueStart = -1
    paramCount = 0

    FOR i = 0 TO len(input):
        c = input[i] (or DELIMITER at EOF)

        IF c == DELIMITER OR i == len(input):
            IF paramCount >= opts.ParameterLimit:
                IF opts.ThrowOnLimitExceeded:
                    RETURN ERROR
                BREAK

            // Emit token if we have content
            IF keyEnd > keyStart OR (keyEnd == -1 AND i > keyStart):
                token = createToken(input, keyStart, keyEnd, valueStart, i, charset, opts)
                IF token.key != "" AND tokenIndex != skipIndex:
                    tokens.append(token)
                paramCount++

            // Reset for next token
            keyStart = i + 1
            keyEnd = -1
            valueStart = -1
            state = SCANNING_KEY

        ELSE IF c == '=' AND state == SCANNING_KEY:
            keyEnd = i
            valueStart = i + 1
            state = SCANNING_VALUE

    RETURN tokens
```

**Complexity**: O(n) - single pass through input

---

### Algorithm 3: Key Path Parser (Finite State Machine)

**Purpose**: Parse key like "a[b][c]" into path segments

**Input**: Key string, options
**Output**: Array of path segments

```
STATE_DIAGRAM:
    ┌─────────────────────────────────────────────────────┐
    │                                                     │
    │    ┌──────┐                    ┌──────────┐        │
    │    │      │◄───── ']' ─────────│          │        │
    │    │IDENT │                    │ BRACKET  │        │
    │    │      │────── '[' ────────►│          │        │
    │    └──┬───┘                    └────┬─────┘        │
    │       │                             │              │
    │       │ '.' (AllowDots)             │ any char     │
    │       │                             │              │
    │       ▼                             ▼              │
    │   emit segment              accumulate char        │
    │                                                     │
    └─────────────────────────────────────────────────────┘

FUNCTION parseKeyPath(key, opts):
    // Special case: depth=0 means no nesting
    IF opts.Depth == 0:
        RETURN [Segment{value: key, isIndex: false}]

    segments = []
    buffer = ""
    state = IDENT
    depth = 0

    FOR i = 0 TO len(key):
        c = key[i]

        // Check depth limit
        IF depth >= opts.Depth:
            // Remainder becomes literal segment
            segments.append(Segment{value: key[i:]})
            BREAK

        SWITCH state:
            CASE IDENT:
                IF c == '[':
                    IF buffer != "":
                        segments.append(createSegment(buffer, opts))
                        buffer = ""
                    depth++
                    state = BRACKET
                ELSE IF c == '.' AND opts.AllowDots:
                    IF buffer != "":
                        segments.append(createSegment(buffer, opts))
                        buffer = ""
                    depth++
                ELSE:
                    buffer += c

            CASE BRACKET:
                IF c == ']':
                    segments.append(createBracketSegment(buffer, opts))
                    buffer = ""
                    state = IDENT
                ELSE:
                    buffer += c

    // Handle remaining buffer
    IF buffer != "":
        segments.append(createSegment(buffer, opts))

    RETURN segments

FUNCTION createBracketSegment(content, opts):
    // Empty brackets []
    IF content == "":
        RETURN Segment{value: "", isEmpty: true}

    // Check if numeric index
    IF opts.ParseArrays:
        index = tryParseInt(content)
        IF index >= 0 AND index <= opts.ArrayLimit:
            IF toString(index) == content:  // no leading zeros
                RETURN Segment{value: content, isIndex: true, index: index}

    // String key
    RETURN Segment{value: content}
```

**Complexity**: O(k) where k = key length

---

### Algorithm 4: URL Decoding

**Purpose**: Decode percent-encoded strings

**Input**: Encoded string, charset
**Output**: Decoded string

```
FUNCTION decode(input, charset):
    // Fast path: check if decoding needed
    needsDecode = false
    FOR c IN input:
        IF c == '%' OR c == '+':
            needsDecode = true
            BREAK

    IF NOT needsDecode:
        RETURN input  // zero-copy

    result = StringBuilder(capacity: len(input))

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
                byte = (hi << 4) | lo
                IF charset == ISO88591:
                    // ISO-8859-1: byte value = Unicode codepoint
                    result.appendRune(rune(byte))
                ELSE:
                    // UTF-8: accumulate bytes
                    result.appendByte(byte)
                i += 3
            ELSE:
                result.append(c)
                i++
        ELSE:
            result.append(c)
            i++

    RETURN result.toString()

// Lookup table for hex decoding (branchless)
hexTable[256] = {
    '0'-'9': 0-9,
    'a'-'f': 10-15,
    'A'-'F': 10-15,
    others: -1
}

FUNCTION hexValue(c):
    RETURN hexTable[c]
```

**Complexity**: O(n) with lookup table for hex conversion

---

### Algorithm 5: Tree Construction

**Purpose**: Build nested structure from path-value pairs

**Input**: Path segments, value, existing tree, options
**Output**: Modified tree

```
FUNCTION insertPath(root, path, value, opts):
    // Security: always block __proto__
    FOR segment IN path:
        IF segment.value == "__proto__":
            RETURN  // silently skip entire path

    // Prototype check
    IF NOT opts.PlainObjects:
        FOR segment IN path:
            IF isPrototypeKey(segment.value) AND NOT opts.AllowPrototypes:
                RETURN  // silently skip

    // Navigate to parent of final segment
    current = root
    FOR i = 0 TO len(path) - 2:
        segment = path[i]
        nextSegment = path[i+1]
        current = ensureContainer(current, segment, nextSegment, opts)
        IF current == nil:
            RETURN

    // Insert at final segment
    finalSegment = path[len(path)-1]
    insertValue(current, finalSegment, value, opts)

FUNCTION ensureContainer(current, segment, nextSegment, opts):
    SWITCH type(current):
        CASE map[string]any:
            key = segment.value
            IF segment.isEmpty:
                key = "0"  // empty bracket without parent

            existing = current[key]
            IF existing != nil:
                RETURN existing

            // Create appropriate container
            IF (nextSegment.isIndex OR nextSegment.isEmpty) AND opts.ParseArrays:
                newContainer = []any{}
            ELSE:
                newContainer = map[string]any{}

            current[key] = newContainer
            RETURN newContainer

        CASE []any:
            IF segment.isIndex:
                // Extend slice if needed
                WHILE len(current) <= segment.index:
                    current = append(current, nil)

                IF current[segment.index] == nil:
                    // Create container
                    IF (nextSegment.isIndex OR nextSegment.isEmpty) AND opts.ParseArrays:
                        current[segment.index] = []any{}
                    ELSE:
                        current[segment.index] = map[string]any{}

                RETURN current[segment.index]
            ELSE:
                // String key on slice - convert to map
                mapVersion = sliceToMap(current)
                // Update parent reference
                RETURN ensureContainer(mapVersion, segment, nextSegment, opts)

FUNCTION insertValue(container, segment, value, opts):
    SWITCH type(container):
        CASE map[string]any:
            key = segment.value
            existing = container[key]

            // Empty brackets = array append
            IF segment.isEmpty AND opts.ParseArrays:
                IF opts.AllowEmptyArrays AND (value == "" OR value == null):
                    IF existing == nil:
                        container[key] = []any{}
                    RETURN

                IF existing == nil:
                    container[key] = []any{value}
                ELSE IF isSlice(existing):
                    container[key] = append(existing, value)
                ELSE:
                    container[key] = []any{existing, value}
                RETURN

            // Handle duplicates and mixed types
            IF existing == nil:
                container[key] = value
            ELSE:
                container[key] = mergeValues(existing, value, opts)

        CASE []any:
            IF segment.isEmpty:
                // Empty brackets = append
                IF opts.AllowEmptyArrays AND (value == "" OR value == null):
                    RETURN  // don't append empty
                container = append(container, value)
            ELSE IF segment.isIndex:
                // Extend and set
                WHILE len(container) <= segment.index:
                    container = append(container, nil)
                IF container[segment.index] == nil:
                    container[segment.index] = value
                ELSE:
                    container[segment.index] = mergeValues(container[segment.index], value, opts)

FUNCTION mergeValues(existing, new, opts):
    // Duplicate handling mode
    SWITCH opts.Duplicates:
        CASE "first":
            RETURN existing
        CASE "last":
            RETURN new
        CASE "combine":
            // Fall through to merge logic

    // JS qs merge semantics
    IF isMap(existing) AND isPrimitive(new):
        // Add primitive as key with value true
        IF isString(new) AND (opts.AllowPrototypes OR NOT isPrototypeKey(new)):
            existing[new] = true
        RETURN existing

    IF isSlice(existing):
        RETURN append(existing, new)

    IF isPrimitive(existing):
        RETURN []any{existing, new}

    RETURN new
```

**Complexity**: O(d) where d = path depth, plus potential slice extensions

---

### Algorithm 6: Comma Value Processing

**Purpose**: Split comma-separated values while preserving encoded commas

**Key Insight**: Split by LITERAL commas before URL decoding

```
FUNCTION processCommaValue(rawValue, charset, opts):
    IF NOT opts.Comma:
        RETURN decode(rawValue, charset)

    // Split by literal commas only (not %2C)
    parts = splitByLiteralComma(rawValue)

    IF len(parts) == 1:
        RETURN decode(parts[0], charset)

    result = []any{}
    FOR part IN parts:
        result.append(decode(part, charset))

    RETURN result

FUNCTION splitByLiteralComma(s):
    result = []
    current = StringBuilder()

    FOR i = 0 TO len(s):
        c = s[i]
        IF c == ',':
            result.append(current.toString())
            current.reset()
        ELSE:
            current.append(c)

    result.append(current.toString())
    RETURN result
```

**Complexity**: O(n)

---

## Options Integration Matrix

| Option | Phase | Implementation |
|--------|-------|----------------|
| IgnoreQueryPrefix | 1-Preprocessing | Strip leading '?' |
| CharsetSentinel | 1-Preprocessing | Detect before tokenization |
| Charset | 4-Value Processing | Pass to decoder |
| Delimiter | 2-Tokenization | Use as split character |
| DelimiterRegexp | 2-Tokenization | Use regex split |
| ParameterLimit | 2-Tokenization | Stop after N tokens |
| ThrowOnLimitExceeded | 2-Tokenization | Return error vs silent |
| Depth | 3-Key Parsing | Limit segment count |
| StrictDepth | 3-Key Parsing | Error vs literal remainder |
| AllowDots | 3-Key Parsing | Treat '.' as separator |
| DecodeDotInKeys | 3-Key Parsing | Decode %2E in segments |
| ParseArrays | 3-Key Parsing, 5-Tree | Enable index detection, array creation |
| ArrayLimit | 3-Key Parsing | Max index for arrays |
| Comma | 4-Value Processing | Split by literal comma |
| InterpretNumericEntities | 4-Value Processing | Convert &#NNN; |
| StrictNullHandling | 4-Value Processing | null vs "" for no-value |
| AllowPrototypes | 5-Tree Construction | Allow prototype keys |
| PlainObjects | 5-Tree Construction | Skip prototype check |
| Duplicates | 5-Tree Construction | combine/first/last |
| AllowEmptyArrays | 5-Tree Construction | [] vs [""] for empty |
| AllowSparse | 6-Post Processing | Keep nil holes |
| Decoder | 4-Value Processing | Custom decode function |

---

## File Structure

```
v2/
├── parse.go              # Public API, ParseOptions, functional options
├── parse_tokenizer.go    # Phase 2: Tokenization
├── parse_keypath.go      # Phase 3: Key path FSM
├── parse_decode.go       # Phase 4: URL decoding, charset, comma
├── parse_tree.go         # Phase 5: Tree construction
├── parse_fast.go         # Main integration, phases 1 & 6
└── utils.go              # Shared utilities (Compact, Merge for legacy)
```

---

## Performance Characteristics

| Operation | Complexity | Allocations |
|-----------|------------|-------------|
| Tokenization | O(n) | 1 slice |
| Key parsing | O(k) per key | 1 slice per key |
| URL decoding | O(v) per value | 0-1 string per value |
| Tree insertion | O(d) per path | 0-d maps/slices |
| Total | O(n + k*t + d*t) | O(t) where t = token count |

Where:
- n = input length
- k = average key length
- d = average nesting depth
- t = number of tokens

---

## Security Considerations

1. **Prototype Pollution**: `__proto__` always blocked, other prototype keys configurable
2. **Parameter Limit**: Prevents DoS via excessive parameters
3. **Array Limit**: Prevents memory exhaustion via large indices
4. **Depth Limit**: Prevents stack overflow in nested structures
