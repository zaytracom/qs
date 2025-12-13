package lang

// needsDecode checks if bytes contain % or + that need decoding.
func needsDecode(b []byte) bool {
	for _, c := range b {
		if c == '%' || c == '+' {
			return true
		}
	}
	return false
}

// decodeInPlace decodes %XX and '+' in place.
// Result is always <= len(b), so it's safe.
// Returns the decoded slice (subslice of dst).
func decodeInPlace(dst, src []byte) []byte {
	dst = dst[:0]
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch c {
		case '+':
			dst = append(dst, ' ')
		case '%':
			if i+2 >= len(src) {
				dst = append(dst, '%')
				continue
			}
			hi := fromHex(src[i+1])
			lo := fromHex(src[i+2])
			if hi < 0 || lo < 0 {
				dst = append(dst, '%')
				continue
			}
			dst = append(dst, byte(hi<<4|lo))
			i += 2
		default:
			dst = append(dst, c)
		}
	}
	return dst
}

// DecodeBytes decodes %XX and '+' from span into arena's scratch buffer.
// Returns the decoded bytes (valid until next decode call).
// Zero-allocation for simple cases (no encoding).
func (a *Arena) DecodeBytes(s Span) []byte {
	raw := a.GetBytes(s)
	if !needsDecode(raw) {
		return raw
	}
	// Ensure scratch has enough capacity
	if cap(a.scratch) < len(raw) {
		a.scratch = make([]byte, 0, len(raw)*2)
	}
	a.scratch = decodeInPlace(a.scratch, raw)
	return a.scratch
}

// DecodeString extracts and decodes a string (lazy decode).
// It allocates only if decoding is needed.
func (a *Arena) DecodeString(s Span, charset Charset) string {
	raw := a.GetBytes(s)

	// Fast path: no decoding needed.
	if !needsDecode(raw) {
		return string(raw)
	}

	// Decode into scratch buffer
	if cap(a.scratch) < len(raw) {
		a.scratch = make([]byte, 0, len(raw)*2)
	}
	out := decodeInPlace(a.scratch, raw)

	// For ISO-8859-1, bytes map directly to Unicode code points 0x00..0xFF.
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
	Source []byte

	Params     []Param
	Segments   []Segment
	Values     []Value
	ValueParts []Span

	// Scratch buffer for decoding - reused to avoid allocations.
	scratch []byte
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
		scratch:    make([]byte, 0, 256),
	}
}

// Reset clears the arena for reuse without deallocating.
func (a *Arena) Reset(source string) {
	// Convert string to []byte. This is the only allocation for input.
	a.Source = []byte(source)
	a.Params = a.Params[:0]
	a.Segments = a.Segments[:0]
	a.Values = a.Values[:0]
	a.ValueParts = a.ValueParts[:0]
	a.scratch = a.scratch[:0]
}

// ResetBytes clears the arena for reuse with []byte input (zero-copy).
func (a *Arena) ResetBytes(source []byte) {
	a.Source = source
	a.Params = a.Params[:0]
	a.Segments = a.Segments[:0]
	a.Values = a.Values[:0]
	a.ValueParts = a.ValueParts[:0]
	a.scratch = a.scratch[:0]
}

// GetBytes returns the raw bytes referenced by span (no decoding, zero-copy).
func (a *Arena) GetBytes(s Span) []byte {
	start := int(s.Off)
	end := start + int(s.Len)
	return a.Source[start:end]
}

// GetString returns the raw substring referenced by span (no decoding).
// This allocates a new string.
func (a *Arena) GetString(s Span) string {
	return string(a.GetBytes(s))
}
