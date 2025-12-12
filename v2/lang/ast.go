package lang

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
