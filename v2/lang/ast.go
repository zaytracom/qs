package lang

// DecodeString extracts and decodes a string (lazy decode).
// It allocates only if decoding is needed.
func (a *Arena) DecodeString(s Span, charset Charset) string {
	raw := a.GetString(s)

	// Fast path: no decoding needed.
	needsDecode := false
	for i := 0; i < len(raw); i++ {
		if raw[i] == '%' || raw[i] == '+' {
			needsDecode = true
			break
		}
	}
	if !needsDecode {
		return raw
	}

	// Decode %XX and '+' -> space.
	out := make([]byte, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch c {
		case '+':
			out = append(out, ' ')
		case '%':
			if i+2 >= len(raw) {
				// Invalid escape: keep as-is (graceful).
				out = append(out, '%')
				continue
			}
			hi := fromHex(raw[i+1])
			lo := fromHex(raw[i+2])
			if hi < 0 || lo < 0 {
				// Invalid escape: keep as-is (graceful).
				out = append(out, '%')
				continue
			}
			out = append(out, byte(hi<<4|lo))
			i += 2
		default:
			out = append(out, c)
		}
	}

	// For ISO-8859-1, bytes map directly to Unicode code points 0x00..0xFF.
	// For UTF-8, bytes are interpreted as UTF-8.
	// Converting []byte -> string uses UTF-8; ISO callers should only pass
	// input that was encoded as ISO-8859-1 bytes.
	if charset == CharsetISO88591 {
		runes := make([]rune, len(out))
		for i, b := range out {
			runes[i] = rune(b)
		}
		return string(runes)
	}
	return string(out)
}

// Span references a substring in the original input without copying.
// Off is a byte offset into Arena.Source.
type Span struct {
	Off uint32
	Len uint16
}

// SegmentKind identifies the type of key segment.
type SegmentKind uint8

const (
	// SegIdent is an identifier segment, e.g. "foo".
	SegIdent SegmentKind = iota
	// SegIndex is a numeric index segment, e.g. "0", "1".
	SegIndex
	// SegEmpty is an empty bracket segment: [].
	SegEmpty
	// SegLiteral is a literal remainder after depth limit.
	SegLiteral
)

// Notation identifies how a segment was specified.
type Notation uint8

const (
	// NotationRoot is the first segment with no prefix.
	NotationRoot Notation = iota
	// NotationBracket is a [segment] segment.
	NotationBracket
	// NotationDot is a .segment segment.
	NotationDot
)

// Segment represents a single path component in a key.
type Segment struct {
	Kind     SegmentKind
	Notation Notation
	Span     Span
	Index    int32 // parsed index value, -1 if not SegIndex
}

// ValueKind identifies the type of value.
type ValueKind uint8

const (
	// ValSimple is a single value.
	ValSimple ValueKind = iota
	// ValComma is a comma-separated list.
	ValComma
	// ValNull is a "null" value from missing "=" with StrictNullHandling.
	ValNull
)

// Value represents a parameter value.
type Value struct {
	Kind     ValueKind
	Raw      Span
	PartsOff uint16 // offset into Arena.ValueParts for comma values
	PartsLen uint8  // number of parts (for ValComma)
}

// Key represents a parameter key with its path segments.
type Key struct {
	SegStart uint16
	SegLen   uint8
	Raw      Span
}

// Param represents a single key=value parameter.
type Param struct {
	Key       Key
	ValueIdx  uint16 // index into Arena.Values; 0xFFFF means "no value node"
	HasEquals bool
}

// QueryString is the root AST node.
type QueryString struct {
	HasPrefix  bool
	ParamStart uint16
	ParamLen   uint16
}

// Arena holds all AST nodes, enabling zero-allocation parsing when reused.
type Arena struct {
	Source string

	Params     []Param
	Segments   []Segment
	Values     []Value
	ValueParts []Span
}

// NewArena allocates an arena with a capacity sized for typical inputs.
// Reuse the arena across parses via (*Arena).Reset to avoid allocations.
func NewArena(estimatedParams int) *Arena {
	if estimatedParams < 0 {
		estimatedParams = 0
	}
	return &Arena{
		Params:     make([]Param, 0, estimatedParams),
		Segments:   make([]Segment, 0, estimatedParams*3),
		Values:     make([]Value, 0, estimatedParams),
		ValueParts: make([]Span, 0, estimatedParams),
	}
}

// Reset clears the arena for reuse without deallocating.
func (a *Arena) Reset(source string) {
	a.Source = source
	a.Params = a.Params[:0]
	a.Segments = a.Segments[:0]
	a.Values = a.Values[:0]
	a.ValueParts = a.ValueParts[:0]
}

// GetString returns the raw substring referenced by span (no decoding).
func (a *Arena) GetString(s Span) string {
	start := int(s.Off)
	end := start + int(s.Len)
	return a.Source[start:end]
}
