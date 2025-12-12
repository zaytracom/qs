# Query String Language Specification

## Version 1.0

---

## Table of Contents

1. [Overview](#1-overview)
2. [Lexical Structure](#2-lexical-structure)
3. [Syntax Grammar](#3-syntax-grammar)
4. [Semantic Rules](#4-semantic-rules)
5. [Dialects](#5-dialects)
6. [AST Definition](#6-ast-definition)
7. [Encoding Rules](#7-encoding-rules)
8. [Configuration Options Reference](#8-configuration-options-reference)
9. [Examples](#9-examples)
10. [BNF Grammar](#10-bnf-grammar)

---

## 1. Overview

The Query String (QS) Language is a format for encoding structured data as URL query strings. It supports:

- Flat key-value pairs
- Nested objects via bracket or dot notation
- Arrays via indexed brackets, empty brackets, or comma-separated values
- Multiple encoding formats (RFC 3986, RFC 1738)
- Multiple character sets (UTF-8, ISO-8859-1)

The language has multiple **dialects** controlled by configuration options, allowing flexibility in parsing and serialization.

---

## 2. Lexical Structure

### 2.1 Character Set

The base character set is ASCII. Extended characters are represented via:
- **UTF-8**: Percent-encoded UTF-8 byte sequences
- **ISO-8859-1**: Direct byte-to-Unicode mapping (0x00-0xFF → U+0000-00FF)

### 2.2 Tokens

```
DELIMITER     ::= '&' | ';' | ',' | <custom>
EQUALS        ::= '='
LBRACKET      ::= '[' | '%5B'
RBRACKET      ::= ']' | '%5D'
DOT           ::= '.' | '%2E'
COMMA         ::= ','
PLUS          ::= '+'

IDENTIFIER    ::= (CHAR - RESERVED)+
VALUE         ::= (CHAR)*

RESERVED      ::= DELIMITER | EQUALS | LBRACKET | RBRACKET

CHAR          ::= UNRESERVED | PERCENT_ENCODED
UNRESERVED    ::= ALPHA | DIGIT | '-' | '_' | '.' | '~'
PERCENT_ENCODED ::= '%' HEXDIG HEXDIG
HEXDIG        ::= '0'-'9' | 'A'-'F' | 'a'-'f'
```

### 2.3 Special Tokens

| Token | Encoded Form | Description |
|-------|-------------|-------------|
| `[` | `%5B` | Left bracket for nesting |
| `]` | `%5D` | Right bracket for nesting |
| `&` | `%26` | Parameter delimiter |
| `=` | `%3D` | Key-value separator |
| `.` | `%2E` | Dot notation separator |
| `,` | `%2C` | Comma for value lists |
| `+` | `%2B` | Plus sign (or space in RFC 1738) |
| ` ` | `%20` or `+` | Space character |

---

## 3. Syntax Grammar

### 3.1 Top-Level Structure

```
QueryString ::= ε | '?' QueryBody | QueryBody
QueryBody   ::= Parameter (DELIMITER Parameter)*
Parameter   ::= Key ('=' Value)?
```

### 3.2 Key Structure

```
Key         ::= RootKey Segment*
RootKey     ::= Identifier | BracketKey
Segment     ::= BracketSegment | DotSegment

BracketSegment ::= '[' SegmentContent ']'
DotSegment     ::= '.' Identifier           /* Only when AllowDots=true */

SegmentContent ::= ε | Index | Identifier
Index          ::= Digit+                    /* 0 to ArrayLimit */
Identifier     ::= (Char - Reserved)+

BracketKey     ::= '[' SegmentContent ']'   /* Key starting with bracket */
```

### 3.3 Value Structure

```
Value       ::= SimpleValue | CommaList
SimpleValue ::= (Char)*
CommaList   ::= SimpleValue (',' SimpleValue)*   /* Only when Comma=true */
```

### 3.4 Depth Constraints

Nesting depth is limited by the `Depth` option (default: 5). Segments beyond the depth limit are treated as literal key text.

```
/* With Depth=2 */
a[b][c][d]=e  →  a[b][c] has literal key "[d]" = "e"
```

---

## 4. Semantic Rules

### 4.1 Object Creation

Objects are created when a key segment is a non-numeric identifier:

```
a[b]=c          →  { a: { b: "c" } }
a[b][c]=d       →  { a: { b: { c: "d" } } }
a.b.c=d         →  { a: { b: { c: "d" } } }  /* AllowDots=true */
```

### 4.2 Array Creation

Arrays are created when:

1. **Empty brackets** are used:
   ```
   a[]=b&a[]=c   →  { a: ["b", "c"] }
   ```

2. **Numeric indices** within `ArrayLimit` are used:
   ```
   a[0]=b&a[1]=c →  { a: ["b", "c"] }
   ```

3. **Comma-separated values** with `Comma=true`:
   ```
   a=b,c,d       →  { a: ["b", "c", "d"] }
   ```

### 4.3 Type Coercion Rules

| Existing | New Access | Result |
|----------|-----------|--------|
| `string` | `array[]` | `[string, newValue]` |
| `array` | `string key` | convert to object, add key |
| `object` | `primitive` | `object[primitive] = true` |

### 4.4 Index Handling

| Index | Condition | Result |
|-------|-----------|--------|
| `≤ ArrayLimit` | `ParseArrays=true` | Array with index |
| `> ArrayLimit` | - | Object with string key |
| Non-numeric | - | Object with string key |

### 4.5 Sparse Arrays

```
a[1]=b&a[3]=c
/* AllowSparse=false (default) */  →  { a: ["b", "c"] }
/* AllowSparse=true */             →  { a: [nil, "b", nil, "c"] }
```

### 4.6 Duplicate Handling

| Mode | `a=1&a=2` Result |
|------|-----------------|
| `combine` (default) | `{ a: ["1", "2"] }` |
| `first` | `{ a: "1" }` |
| `last` | `{ a: "2" }` |

### 4.7 Null Handling

| Input | `StrictNullHandling=false` | `StrictNullHandling=true` |
|-------|---------------------------|--------------------------|
| `a` | `{ a: "" }` | `{ a: null }` |
| `a=` | `{ a: "" }` | `{ a: "" }` |
| `a=b` | `{ a: "b" }` | `{ a: "b" }` |

### 4.8 Empty Arrays

```
a[]=
/* AllowEmptyArrays=false (default) */  →  { a: [""] }
/* AllowEmptyArrays=true */             →  { a: [] }
```

---

## 5. Dialects

The QS language supports multiple dialects controlled by configuration options.

### 5.1 Standard Dialect (Default)

```yaml
Delimiter: "&"
Depth: 5
ArrayLimit: 20
ParseArrays: true
AllowDots: false
Comma: false
StrictNullHandling: false
```

**Example:**
```
a[b][c]=d&a[b][e]=f&arr[0]=x&arr[1]=y
→ { a: { b: { c: "d", e: "f" } }, arr: ["x", "y"] }
```

### 5.2 Dot Notation Dialect

```yaml
AllowDots: true
```

**Example:**
```
a.b.c=d&a.b.e=f
→ { a: { b: { c: "d", e: "f" } } }
```

### 5.3 Comma-Separated Dialect

```yaml
Comma: true
```

**Example:**
```
tags=a,b,c&ids=1,2,3
→ { tags: ["a", "b", "c"], ids: ["1", "2", "3"] }
```

### 5.4 Semicolon Delimiter Dialect

```yaml
Delimiter: ";"
```

**Example:**
```
a=1;b=2;c=3
→ { a: "1", b: "2", c: "3" }
```

### 5.5 Flat Dialect (No Nesting)

```yaml
Depth: 0
ParseArrays: false
```

**Example:**
```
a[b]=c&arr[0]=x
→ { "a[b]": "c", "arr[0]": "x" }
```

### 5.6 Strict Null Dialect

```yaml
StrictNullHandling: true
```

**Example:**
```
enabled&disabled=false&count=0
→ { enabled: null, disabled: "false", count: "0" }
```

### 5.7 Sparse Array Dialect

```yaml
AllowSparse: true
```

**Example:**
```
a[0]=x&a[5]=y
→ { a: [x, nil, nil, nil, nil, y] }
```

### 5.8 Combined Dialects

Dialects can be combined:

```yaml
AllowDots: true
Comma: true
StrictNullHandling: true
```

**Example:**
```
user.tags=admin,editor&user.active
→ { user: { tags: ["admin", "editor"], active: null } }
```

---

## 6. AST Definition

The AST is designed for **zero-allocation parsing**. All nodes are views into the original query string, using spans (offset + length) instead of copied strings. This enables:

- Zero heap allocations during parsing
- Excellent cache locality
- Lazy decoding (only when values are accessed)
- Arena-based memory management

### 6.1 Core Types (Zero-Alloc)

```go
// Span references a substring in the original input without copying
type Span struct {
    Off uint32  // byte offset into source string
    Len uint16  // byte length
}

// SegmentKind identifies the type of key segment
type SegmentKind uint8
const (
    SegIdent   SegmentKind = iota  // identifier: "foo"
    SegIndex                        // numeric index: "0", "1", "42"
    SegEmpty                        // empty brackets: []
    SegLiteral                      // literal remainder after depth limit
)

// Notation identifies how the segment was specified
type Notation uint8
const (
    NotationRoot    Notation = iota  // first segment (no prefix)
    NotationBracket                   // [segment]
    NotationDot                       // .segment
)

// Segment represents a single path component in a key
type Segment struct {
    Kind     SegmentKind
    Notation Notation
    Span     Span   // view into source for the segment value
    Index    int32  // parsed index value, -1 if not SegIndex
}

// ValueKind identifies the type of value
type ValueKind uint8
const (
    ValSimple ValueKind = iota  // single value
    ValComma                     // comma-separated list
    ValNull                      // no "=" present (with StrictNullHandling)
)

// Value represents a parameter value
type Value struct {
    Kind     ValueKind
    Raw      Span    // view into source for entire value
    PartsOff uint16  // offset into Arena.ValueParts for comma values
    PartsLen uint8   // number of parts (for ValComma)
}

// Key represents a parameter key with its path segments
type Key struct {
    SegStart uint16  // offset into Arena.Segments
    SegLen   uint8   // number of segments
    Raw      Span    // view into source for entire key
}

// Param represents a single key=value parameter
type Param struct {
    Key       Key
    ValueIdx  uint16  // index into Arena.Values, 0xFFFF = no value
    HasEquals bool    // distinguishes "a" from "a="
}

// QueryString is the root AST node
type QueryString struct {
    HasPrefix  bool    // started with "?"
    ParamStart uint16  // offset into Arena.Params
    ParamLen   uint16  // number of parameters
}
```

### 6.2 Arena (Memory Pool)

```go
// Arena holds all AST nodes, enabling zero-allocation parsing
type Arena struct {
    Source     string    // original query string (reference, not copy)

    Params     []Param   // all parameters
    Segments   []Segment // all key segments
    Values     []Value   // all values
    ValueParts []Span    // spans for comma-separated value parts
}

// Pre-allocate with estimated capacity for reuse
func NewArena(estimatedParams int) *Arena {
    return &Arena{
        Params:     make([]Param, 0, estimatedParams),
        Segments:   make([]Segment, 0, estimatedParams*3),
        Values:     make([]Value, 0, estimatedParams),
        ValueParts: make([]Span, 0, estimatedParams),
    }
}

// Reset clears the arena for reuse without deallocating
func (a *Arena) Reset(source string) {
    a.Source = source
    a.Params = a.Params[:0]
    a.Segments = a.Segments[:0]
    a.Values = a.Values[:0]
    a.ValueParts = a.ValueParts[:0]
}

// GetString extracts a string from a span (allocates only when called)
func (a *Arena) GetString(s Span) string {
    return a.Source[s.Off : s.Off+uint32(s.Len)]
}

// DecodeString extracts and decodes a string (lazy decode)
func (a *Arena) DecodeString(s Span, charset Charset) string {
    raw := a.Source[s.Off : s.Off+uint32(s.Len)]
    return decode(raw, charset)  // only allocates if decoding needed
}
```

### 6.3 Dialect Flags (Bitmask)

```go
// Flags control parsing behavior via bitmask (branch-prediction friendly)
type Flags uint32

const (
    FlagAllowDots            Flags = 1 << iota  // parse "." as separator
    FlagAllowEmptyArrays                         // a[]= creates []
    FlagAllowPrototypes                          // allow constructor, prototype
    FlagAllowSparse                              // preserve array gaps
    FlagComma                                    // split values on ","
    FlagStrictNullHandling                       // "a" → null
    FlagStrictDepth                              // error on depth exceeded
    FlagIgnoreQueryPrefix                        // strip leading "?"
    FlagCharsetSentinel                          // auto-detect charset
    FlagInterpretNumericEntities                 // convert &#NNN;
    FlagDecodeDotInKeys                          // decode %2E in keys
    FlagThrowOnLimitExceeded                     // error on limits
)

// Check flag efficiently
func (f Flags) Has(flag Flags) bool {
    return f&flag != 0
}
```

### 6.4 Conceptual AST (For Documentation)

For documentation purposes, the logical AST structure is:

```typescript
// Conceptual representation (NOT the runtime format)
type ASTNode =
  | QueryStringNode
  | ParameterNode
  | KeyNode
  | SegmentNode
  | ValueNode

interface QueryStringNode {
  type: "QueryString"
  parameters: ParameterNode[]
  hasPrefix: boolean
}

interface ParameterNode {
  type: "Parameter"
  key: KeyNode
  value: ValueNode | null
  hasEquals: boolean
}

interface KeyNode {
  type: "Key"
  segments: SegmentNode[]
  raw: string
}

interface SegmentNode {
  type: "Segment"
  kind: "identifier" | "index" | "empty" | "literal"
  value: string            // decoded segment value
  index?: number           // for kind="index"
  notation: "root" | "bracket" | "dot"
}

interface ValueNode {
  type: "Value"
  kind: "simple" | "comma-list" | "null"
  values: string[]         // single element for simple, multiple for comma-list
  raw: string              // original encoded value
}
```

### 6.5 Example AST (Zero-Alloc)

**Input:** `a[b][0]=x,y&c.d=z` (with `AllowDots=true`, `Comma=true`)

```go
// Source: "a[b][0]=x,y&c.d=z"
//          0123456789...

arena := &Arena{
    Source: "a[b][0]=x,y&c.d=z",

    // Segments array (all key segments)
    Segments: []Segment{
        // Parameter 0: a[b][0]
        {Kind: SegIdent,   Notation: NotationRoot,    Span: {Off: 0, Len: 1}, Index: -1},  // "a"
        {Kind: SegIdent,   Notation: NotationBracket, Span: {Off: 2, Len: 1}, Index: -1},  // "b"
        {Kind: SegIndex,   Notation: NotationBracket, Span: {Off: 5, Len: 1}, Index: 0},   // "0"
        // Parameter 1: c.d
        {Kind: SegIdent,   Notation: NotationRoot,    Span: {Off: 12, Len: 1}, Index: -1}, // "c"
        {Kind: SegIdent,   Notation: NotationDot,     Span: {Off: 14, Len: 1}, Index: -1}, // "d"
    },

    // Values array
    Values: []Value{
        {Kind: ValComma,  Raw: {Off: 8, Len: 3}, PartsOff: 0, PartsLen: 2},  // "x,y"
        {Kind: ValSimple, Raw: {Off: 16, Len: 1}, PartsOff: 0, PartsLen: 0}, // "z"
    },

    // ValueParts for comma-separated values
    ValueParts: []Span{
        {Off: 8, Len: 1},   // "x"
        {Off: 10, Len: 1},  // "y"
    },

    // Parameters array
    Params: []Param{
        {Key: Key{SegStart: 0, SegLen: 3, Raw: {Off: 0, Len: 7}}, ValueIdx: 0, HasEquals: true},
        {Key: Key{SegStart: 3, SegLen: 2, Raw: {Off: 12, Len: 3}}, ValueIdx: 1, HasEquals: true},
    },
}

// Root query string
qs := QueryString{
    HasPrefix:  false,
    ParamStart: 0,
    ParamLen:   2,
}
```

### 6.6 Accessing Values (Lazy Decode)

```go
// Get first parameter's key segments
param := arena.Params[0]
for i := uint16(0); i < uint16(param.Key.SegLen); i++ {
    seg := arena.Segments[param.Key.SegStart + i]
    // Only allocate string when actually needed
    name := arena.GetString(seg.Span)      // "a", "b", "0"
    decoded := arena.DecodeString(seg.Span, CharsetUTF8)
}

// Get comma-separated values
val := arena.Values[param.ValueIdx]
if val.Kind == ValComma {
    for i := uint16(0); i < uint16(val.PartsLen); i++ {
        part := arena.ValueParts[val.PartsOff + i]
        decoded := arena.DecodeString(part, CharsetUTF8)  // "x", "y"
    }
}
```

### 6.7 Performance Guarantees

```
┌─────────────────────────────────────────────────────────────────┐
│                    PERFORMANCE CONTRACT                          │
├─────────────────────────────────────────────────────────────────┤
│  Phase              │ Heap Allocations │ Notes                  │
├─────────────────────┼──────────────────┼────────────────────────┤
│  Parsing            │ 0                │ Views into source      │
│  AST Construction   │ 0                │ Arena-based storage    │
│  Value Access       │ 0                │ Span references        │
│  String Decode      │ 0-1 per call     │ Lazy, only if needed   │
│  Semantic Build     │ N                │ Creates Go objects     │
│  Serialization      │ N                │ Creates output string  │
└─────────────────────┴──────────────────┴────────────────────────┘

Key Principles:
- All AST nodes are VIEWS into the original query string
- Decoding is LAZY - only happens when string values are accessed
- Arena can be REUSED across multiple parse calls
- Semantic object construction is a SEPARATE phase (allowed to allocate)
```

### 6.8 Runtime Profile Modes

```go
type Profile uint8

const (
    ProfileFast   Profile = iota  // Skip validation, maximum speed
    ProfileStrict                  // Full RFC validation
)

// Fast mode: skip unnecessary checks
// - No charset validation
// - No depth pre-check
// - No parameter count validation until limit hit

// Strict mode: full validation
// - Validate all percent-encoding
// - Pre-check depth limits
// - Validate charset compatibility
```

---

## 7. Encoding Rules

### 7.1 RFC 3986 (Default)

Unreserved characters: `A-Z a-z 0-9 - _ . ~`

All other characters are percent-encoded.

```
Space     → %20
!         → %21
#         → %23
$         → %24
&         → %26
'         → %27
(         → %28
)         → %29
*         → %2A
+         → %2B
,         → %2C
/         → %2F
:         → %3A
;         → %3B
=         → %3D
?         → %3F
@         → %40
[         → %5B
]         → %5D
```

### 7.2 RFC 1738

Same as RFC 3986, except:
```
Space     → +
+         → %2B
```

### 7.3 UTF-8 Encoding

Multi-byte UTF-8 sequences are percent-encoded byte-by-byte:

```
ö (U+00F6) → %C3%B6    (UTF-8: 0xC3 0xB6)
☺ (U+263A) → %E2%98%BA  (UTF-8: 0xE2 0x98 0xBA)
```

### 7.4 ISO-8859-1 Encoding

Characters in the range U+0000-U+00FF are encoded as single bytes:

```
ö (U+00F6) → %F6       (ISO-8859-1: 0xF6)
§ (U+00A7) → %A7       (ISO-8859-1: 0xA7)
```

### 7.5 Charset Sentinel

The charset sentinel indicates which encoding was used:

```
UTF-8:      utf8=%E2%9C%93      (✓ encoded)
ISO-8859-1: utf8=%26%2310003%3B (&#10003; encoded)
```

---

## 8. Configuration Options Reference

### 8.1 Parse Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `AllowDots` | bool | `false` | Parse `.` as nesting separator |
| `AllowEmptyArrays` | bool | `false` | `a[]=` creates `[]` instead of `[""]` |
| `AllowPrototypes` | bool | `false` | Allow `constructor`, `prototype` keys |
| `AllowSparse` | bool | `false` | Preserve array gaps |
| `ArrayLimit` | int | `20` | Max index for array notation |
| `Charset` | enum | `UTF-8` | `UTF-8` or `ISO-8859-1` |
| `CharsetSentinel` | bool | `false` | Auto-detect charset from sentinel |
| `Comma` | bool | `false` | Split values on `,` |
| `DecodeDotInKeys` | bool | `false` | Decode `%2E` as `.` in keys |
| `Decoder` | func | `nil` | Custom decoder function |
| `Delimiter` | string | `&` | Parameter separator |
| `DelimiterRegexp` | regexp | `nil` | Regex for splitting parameters |
| `Depth` | int | `5` | Max nesting depth |
| `Duplicates` | enum | `combine` | `combine`, `first`, or `last` |
| `IgnoreQueryPrefix` | bool | `false` | Strip leading `?` |
| `InterpretNumericEntities` | bool | `false` | Convert `&#NNN;` to chars |
| `ParameterLimit` | int | `1000` | Max number of parameters |
| `ParseArrays` | bool | `true` | Enable array parsing |
| `PlainObjects` | bool | `false` | Skip prototype check |
| `StrictDepth` | bool | `false` | Error on depth exceeded |
| `StrictNullHandling` | bool | `false` | `a` → `null` instead of `""` |
| `ThrowOnLimitExceeded` | bool | `false` | Error on limit exceeded |

### 8.2 Stringify Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `AddQueryPrefix` | bool | `false` | Add leading `?` |
| `AllowDots` | bool | `false` | Use dot notation |
| `AllowEmptyArrays` | bool | `false` | Output `key[]` for empty arrays |
| `ArrayFormat` | enum | `indices` | `indices`, `brackets`, `repeat`, `comma` |
| `Charset` | enum | `UTF-8` | Output charset |
| `CharsetSentinel` | bool | `false` | Add charset sentinel |
| `CommaRoundTrip` | bool | `false` | Single-element arrays use `[]` |
| `Delimiter` | string | `&` | Output delimiter |
| `Encode` | bool | `true` | Enable encoding |
| `EncodeDotInKeys` | bool | `false` | Encode `.` as `%2E` in keys |
| `Encoder` | func | `nil` | Custom encoder function |
| `EncodeValuesOnly` | bool | `false` | Only encode values |
| `Filter` | func/[]string | `nil` | Filter/transform values |
| `Format` | enum | `RFC3986` | `RFC3986` or `RFC1738` |
| `SerializeDate` | func | RFC3339 | Date serialization function |
| `SkipNulls` | bool | `false` | Skip null values |
| `Sort` | func | `nil` | Key sorting function |
| `StrictNullHandling` | bool | `false` | `null` → `a` instead of `a=` |

---

## 9. Examples

### 9.1 Basic Parsing

```
Input:  a=b&c=d
Output: { a: "b", c: "d" }
```

### 9.2 Nested Objects

```
Input:  user[name]=John&user[age]=30
Output: { user: { name: "John", age: "30" } }
```

### 9.3 Dot Notation (AllowDots=true)

```
Input:  user.name=John&user.address.city=NYC
Output: { user: { name: "John", address: { city: "NYC" } } }
```

### 9.4 Arrays with Indices

```
Input:  colors[0]=red&colors[1]=blue&colors[2]=green
Output: { colors: ["red", "blue", "green"] }
```

### 9.5 Arrays with Empty Brackets

```
Input:  tags[]=js&tags[]=go&tags[]=rust
Output: { tags: ["js", "go", "rust"] }
```

### 9.6 Comma-Separated Values (Comma=true)

```
Input:  ids=1,2,3,4,5
Output: { ids: ["1", "2", "3", "4", "5"] }
```

### 9.7 Mixed Structures

```
Input:  users[0][name]=Alice&users[0][roles][]=admin&users[1][name]=Bob
Output: {
  users: [
    { name: "Alice", roles: ["admin"] },
    { name: "Bob" }
  ]
}
```

### 9.8 URL Encoded

```
Input:  message=Hello%20World%21&emoji=%F0%9F%98%80
Output: { message: "Hello World!", emoji: "😀" }
```

### 9.9 Depth Limiting (Depth=2)

```
Input:  a[b][c][d]=e
Output: { a: { b: { "[c][d]": "e" } } }
```

### 9.10 Sparse Arrays (AllowSparse=true)

```
Input:  a[0]=first&a[5]=last
Output: { a: ["first", nil, nil, nil, nil, "last"] }
```

---

## 10. BNF Grammar

### 10.1 Complete BNF

```bnf
<query-string>    ::= <empty> | "?" <query-body> | <query-body>

<query-body>      ::= <parameter> | <parameter> <delimiter> <query-body>

<parameter>       ::= <key> | <key> "=" | <key> "=" <value>

<key>             ::= <root-key> <segments>

<root-key>        ::= <identifier> | "[" <segment-content> "]"

<segments>        ::= <empty> | <segment> <segments>

<segment>         ::= <bracket-segment> | <dot-segment>

<bracket-segment> ::= "[" <segment-content> "]"

<dot-segment>     ::= "." <identifier>

<segment-content> ::= <empty> | <index> | <identifier>

<index>           ::= <digit> | <digit> <index>

<identifier>      ::= <id-char> | <id-char> <identifier>

<value>           ::= <simple-value> | <comma-list>

<simple-value>    ::= <empty> | <value-char> <simple-value>

<comma-list>      ::= <simple-value> "," <simple-value>
                    | <simple-value> "," <comma-list>

<delimiter>       ::= "&" | ";" | ","

<id-char>         ::= <unreserved> | <percent-encoded>

<value-char>      ::= <unreserved> | <percent-encoded> | <sub-delim>

<unreserved>      ::= <alpha> | <digit> | "-" | "_" | "." | "~"

<percent-encoded> ::= "%" <hexdig> <hexdig>

<sub-delim>       ::= "!" | "$" | "'" | "(" | ")" | "*" | "+" | ","

<alpha>           ::= "A" | "B" | ... | "Z" | "a" | "b" | ... | "z"

<digit>           ::= "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"

<hexdig>          ::= <digit> | "A" | "B" | "C" | "D" | "E" | "F"
                    | "a" | "b" | "c" | "d" | "e" | "f"

<empty>           ::= ""
```

### 10.2 Extended BNF (EBNF) with Dialect Annotations

```ebnf
(* Query String Grammar with Dialect Support *)

QueryString = [ "?" ] , QueryBody ;                   (* IgnoreQueryPrefix *)

QueryBody = Parameter , { Delimiter , Parameter } ;

Parameter = Key , [ "=" , [ Value ] ] ;               (* StrictNullHandling *)

Key = RootKey , { Segment } ;                         (* Depth limit *)

RootKey = Identifier | BracketedKey ;

BracketedKey = "[" , SegmentContent , "]" ;

Segment = BracketSegment
        | DotSegment ;                                (* AllowDots *)

BracketSegment = "[" , SegmentContent , "]" ;

DotSegment = "." , Identifier ;                       (* Only with AllowDots *)

SegmentContent = (* empty *)
               | Index                                (* ParseArrays, ArrayLimit *)
               | Identifier ;

Index = Digit , { Digit } ;                           (* Must be <= ArrayLimit *)

Identifier = IdentChar , { IdentChar } ;

Value = SimpleValue
      | CommaList ;                                   (* Comma option *)

CommaList = SimpleValue , "," , SimpleValue , { "," , SimpleValue } ;

Delimiter = "&"                                       (* default *)
          | ";"                                       (* custom *)
          | "," ;                                     (* custom *)

SimpleValue = { ValueChar } ;

IdentChar = UnreservedChar | PercentEncoded ;

ValueChar = UnreservedChar | PercentEncoded | SubDelim ;

UnreservedChar = Letter | Digit | "-" | "_" | "." | "~" ;

PercentEncoded = "%" , HexDigit , HexDigit ;

SubDelim = "!" | "$" | "'" | "(" | ")" | "*" | "+" ;

Letter = "A" | "B" | ... | "Z" | "a" | "b" | ... | "z" ;

Digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;

HexDigit = Digit | "A" | "B" | "C" | "D" | "E" | "F"
                 | "a" | "b" | "c" | "d" | "e" | "f" ;
```

---

## Appendix A: Security Considerations

### A.1 Prototype Pollution Prevention

The following keys are blocked by default:
- `__proto__` - **ALWAYS blocked** (security critical)
- `constructor` - blocked unless `AllowPrototypes=true`
- `prototype` - blocked unless `AllowPrototypes=true`
- `hasOwnProperty` - blocked unless `AllowPrototypes=true`

**Implementation:** Check at segment level BEFORE semantic build:

```go
// Check during key parsing, not tree construction
func (p *Parser) parseSegment() (Segment, error) {
    seg := p.readSegment()

    // __proto__ is ALWAYS blocked
    if p.arena.SpanEquals(seg.Span, "__proto__") {
        return Segment{}, ErrProtoBlocked
    }

    // Other prototype keys blocked unless allowed
    if !p.flags.Has(FlagAllowPrototypes) {
        if p.arena.SpanEqualsAny(seg.Span, prototypeKeys) {
            return Segment{}, ErrPrototypeKey
        }
    }

    return seg, nil
}
```

### A.2 Denial of Service Protection

| Limit | Default | Purpose |
|-------|---------|---------|
| `ParameterLimit` | 1000 | Prevents excessive parameter parsing |
| `ArrayLimit` | 20 | Prevents memory exhaustion via large indices |
| `Depth` | 5 | Prevents stack overflow in nested structures |

### A.3 Input Validation

All inputs should be treated as untrusted. The parser:
- Validates all percent-encoded sequences
- Limits recursion depth
- Bounds array indices
- Sanitizes prototype-related keys

### A.4 Memory Safety

The zero-alloc design provides additional security benefits:
- No unbounded allocations from malicious input
- Arena size is predictable and controllable
- No GC pressure from parse operations
- Span bounds are always validated against source length

---

## Appendix B: Compatibility Notes

### B.1 Node.js `qs` Library Compatibility

This specification is designed to be compatible with the Node.js `qs` library (version 6.x). Key compatibility features:

- Same default options
- Same parsing rules for nested objects and arrays
- Same handling of edge cases (depth limits, array limits, duplicates)
- Same encoding/decoding rules

### B.2 URL Standard Compatibility

The encoding rules follow:
- RFC 3986 (URI Generic Syntax)
- RFC 1738 (URL specification) for optional `+` space encoding
- WHATWG URL Standard for form data encoding

---

## Appendix C: Parser Implementation Guide (Zero-Alloc)

### C.1 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           ZERO-ALLOC PARSER                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐                                                     │
│  │   Input String  │  (kept as reference, not copied)                    │
│  └────────┬────────┘                                                     │
│           │                                                              │
│           ▼                                                              │
│  ┌─────────────────────────────────────────────────────────┐            │
│  │              PHASE 1: SCAN (Zero-Alloc)                  │            │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │            │
│  │  │ Preprocess  │→ │  Tokenize   │→ │  Key Path FSM   │  │            │
│  │  │ (strip ?)   │  │ (find &/=)  │  │ (find []/.)    │  │            │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │            │
│  │                                                          │            │
│  │  Output: Spans (offsets into input), NOT strings         │            │
│  └─────────────────────────────────────────────────────────┘            │
│           │                                                              │
│           ▼                                                              │
│  ┌─────────────────────────────────────────────────────────┐            │
│  │                PHASE 2: AST (Zero-Alloc)                 │            │
│  │                                                          │            │
│  │  Arena {                                                 │            │
│  │      Source:     string     // reference to input        │            │
│  │      Params:     []Param    // pre-allocated slice       │            │
│  │      Segments:   []Segment  // pre-allocated slice       │            │
│  │      Values:     []Value    // pre-allocated slice       │            │
│  │      ValueParts: []Span     // pre-allocated slice       │            │
│  │  }                                                       │            │
│  │                                                          │            │
│  │  All nodes are VIEWS (Spans) into Source                 │            │
│  └─────────────────────────────────────────────────────────┘            │
│           │                                                              │
│           │  (optional, allocates)                                       │
│           ▼                                                              │
│  ┌─────────────────────────────────────────────────────────┐            │
│  │             PHASE 3: SEMANTIC BUILD (Allocates)          │            │
│  │                                                          │            │
│  │  - Lazy decode strings (only when accessed)              │            │
│  │  - Create nested map[string]any / []any                  │            │
│  │  - Handle duplicates, sparse arrays, nulls               │            │
│  │                                                          │            │
│  │  Output: map[string]any                                  │            │
│  └─────────────────────────────────────────────────────────┘            │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### C.2 Key Path Parser State Machine

```
States: ROOT, IDENT, BRACKET, DOT

┌──────────────────────────────────────────────────────────────┐
│                                                              │
│    ┌──────┐                    ┌──────────┐                 │
│    │      │◄───── ']' ─────────│          │                 │
│    │IDENT │                    │ BRACKET  │                 │
│    │      │────── '[' ────────►│          │                 │
│    └──┬───┘                    └────┬─────┘                 │
│       │                             │                        │
│       │ '.' (if AllowDots)          │ any char               │
│       ▼                             ▼                        │
│   emit segment,               accumulate (track span)        │
│   start new span                                             │
│                                                              │
└──────────────────────────────────────────────────────────────┘

CRITICAL: Handle empty brackets explicitly!

if src[i] == ']' && i == bracketStart {
    // Empty segment: a[] → SegEmpty
    emit(Segment{Kind: SegEmpty, Span: {Off: i, Len: 0}})
}
```

### C.3 Tokenizer Implementation (Zero-Alloc)

```go
// Single-pass tokenizer that produces Spans
func (p *Parser) tokenize() {
    src := p.arena.Source
    keyStart := uint32(0)
    state := stateKey

    for i := uint32(0); i <= uint32(len(src)); i++ {
        var c byte
        if i < uint32(len(src)) {
            c = src[i]
        }

        switch {
        case c == '&' || c == 0 || p.isDelimiter(c):
            // Emit parameter
            if i > keyStart || state == stateValue {
                p.emitParam(keyStart, i, state)
            }
            keyStart = i + 1
            state = stateKey

        case c == '=' && state == stateKey:
            // Key ends, value starts
            p.currentKeyEnd = i
            state = stateValue
        }
    }
}

// Zero allocation - just records spans
func (p *Parser) emitParam(start, end uint32, state int) {
    param := Param{
        Key:       p.parseKey(start, p.currentKeyEnd),
        HasEquals: state == stateValue,
    }
    if state == stateValue {
        param.ValueIdx = p.emitValue(p.currentKeyEnd+1, end)
    } else {
        param.ValueIdx = 0xFFFF // no value
    }
    p.arena.Params = append(p.arena.Params, param)
}
```

### C.4 Lazy Decoder

```go
// Decoder interface - pluggable, lazy
type Decoder func(src Span, arena *Arena, charset Charset) string

// Default decoder - allocates only if decoding needed
func DefaultDecoder(span Span, arena *Arena, charset Charset) string {
    raw := arena.Source[span.Off : span.Off+uint32(span.Len)]

    // Fast path: check if decoding needed
    needsDecode := false
    for i := 0; i < len(raw); i++ {
        if raw[i] == '%' || raw[i] == '+' {
            needsDecode = true
            break
        }
    }

    if !needsDecode {
        return raw  // Zero allocation - return slice of source
    }

    // Slow path: decode (allocates)
    return decodePercent(raw, charset)
}
```

### C.5 Complexity Analysis

| Operation | Time | Space (Heap) | Notes |
|-----------|------|--------------|-------|
| Tokenization | O(n) | **0** | Single pass, spans only |
| Key Parsing | O(k) per key | **0** | FSM, spans only |
| AST Construction | O(p) | **0** | Append to pre-allocated slices |
| Value Access | O(1) | **0** | Span lookup |
| String Decode | O(v) | **0-1** | Only if % or + present |
| Semantic Build | O(p × d) | O(p × d) | Creates Go objects |

Where:
- n = input length
- k = average key length
- p = parameter count
- d = average nesting depth
- v = value length

### C.6 Arena Reuse Pattern

```go
// Pool for arena reuse across requests
var arenaPool = sync.Pool{
    New: func() any {
        return NewArena(32) // typical parameter count
    },
}

func Parse(input string) (map[string]any, error) {
    arena := arenaPool.Get().(*Arena)
    defer arenaPool.Put(arena)

    arena.Reset(input)

    // Parse into arena (zero alloc)
    if err := parseIntoArena(input, arena); err != nil {
        return nil, err
    }

    // Build semantic object (allocates, but arena stays clean)
    return buildObject(arena)
}
```

### C.7 Memory Layout

```
Arena Memory Layout (cache-friendly):

┌─────────────────────────────────────────────────────────────┐
│ Source: "a[b]=1&c=2"  (reference to input, not owned)       │
├─────────────────────────────────────────────────────────────┤
│ Params:   [Param0][Param1]...     (contiguous)              │
├─────────────────────────────────────────────────────────────┤
│ Segments: [Seg0][Seg1][Seg2]...   (contiguous)              │
├─────────────────────────────────────────────────────────────┤
│ Values:   [Val0][Val1]...         (contiguous)              │
├─────────────────────────────────────────────────────────────┤
│ ValueParts: [Span0][Span1]...     (contiguous)              │
└─────────────────────────────────────────────────────────────┘

All slices are pre-allocated and reused.
No pointer chasing - excellent cache locality.
```

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-01 | Initial specification |
| 1.1 | 2025-01 | Zero-alloc AST redesign, arena-based memory, performance guarantees |
