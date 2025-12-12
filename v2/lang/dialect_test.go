package lang

import (
	"strings"
	"testing"
)

// =============================================================================
// DIALECT TESTS - Testing AST parser with real query strings from jscompat tests
// =============================================================================

// =============================================================================
// STANDARD DIALECT (§5.1)
// =============================================================================

func TestDialect_Standard_DeepNested(t *testing.T) {
	// From TestDeepNestedComplex
	// user[profile][name]=John%20Doe&user[profile][emails][0]=john%40example.com...
	input := "user[profile][name]=John&user[profile][settings][theme]=dark&user[tags][0]=admin"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	// Check first param: user[profile][name]
	p0 := arena.Params[0]
	if p0.Key.SegLen != 3 {
		t.Errorf("param 0: expected 3 segments, got %d", p0.Key.SegLen)
	}

	seg0 := arena.Segments[p0.Key.SegStart]
	seg1 := arena.Segments[p0.Key.SegStart+1]
	seg2 := arena.Segments[p0.Key.SegStart+2]

	if arena.GetString(seg0.Span) != "user" || seg0.Notation != NotationRoot {
		t.Errorf("seg0: expected 'user' root, got %q %v", arena.GetString(seg0.Span), seg0.Notation)
	}
	if arena.GetString(seg1.Span) != "profile" || seg1.Notation != NotationBracket {
		t.Errorf("seg1: expected 'profile' bracket, got %q %v", arena.GetString(seg1.Span), seg1.Notation)
	}
	if arena.GetString(seg2.Span) != "name" || seg2.Notation != NotationBracket {
		t.Errorf("seg2: expected 'name' bracket, got %q %v", arena.GetString(seg2.Span), seg2.Notation)
	}
}

func TestDialect_Standard_ArrayIndices(t *testing.T) {
	// From TestArrayFormatIndices
	input := "colors[0]=red&colors[1]=green&colors[2]=blue&numbers[0]=1&numbers[1]=2"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 5 {
		t.Fatalf("expected 5 params, got %d", len(arena.Params))
	}

	// Check that indices are parsed correctly
	for i := 0; i < 3; i++ {
		p := arena.Params[i]
		seg1 := arena.Segments[p.Key.SegStart+1]
		if seg1.Kind != SegIndex {
			t.Errorf("param %d: expected SegIndex, got %v", i, seg1.Kind)
		}
		if seg1.Index != int32(i) {
			t.Errorf("param %d: expected index %d, got %d", i, i, seg1.Index)
		}
	}
}

func TestDialect_Standard_Brackets(t *testing.T) {
	// From TestArrayFormatBrackets
	input := "colors[]=red&colors[]=green&colors[]=blue"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	// All should have SegEmpty for []
	for i, p := range arena.Params {
		seg1 := arena.Segments[p.Key.SegStart+1]
		if seg1.Kind != SegEmpty {
			t.Errorf("param %d: expected SegEmpty, got %v", i, seg1.Kind)
		}
	}
}

func TestDialect_Standard_Repeat(t *testing.T) {
	// From TestArrayFormatRepeat: colors=red&colors=green&colors=blue
	input := "colors=red&colors=green&colors=blue"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// All three params with same key
	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	for i, p := range arena.Params {
		if p.Key.SegLen != 1 {
			t.Errorf("param %d: expected 1 segment, got %d", i, p.Key.SegLen)
		}
		seg := arena.Segments[p.Key.SegStart]
		if arena.GetString(seg.Span) != "colors" {
			t.Errorf("param %d: expected 'colors', got %q", i, arena.GetString(seg.Span))
		}
	}
}

// =============================================================================
// DOT NOTATION DIALECT (§5.2)
// =============================================================================

func TestDialect_DotNotation_Basic(t *testing.T) {
	// From TestDotNotationEncoded
	input := "config.api.key=secret123&config.nested.deep.value=42"
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}

	// First: config.api.key -> 3 segments
	p0 := arena.Params[0]
	if p0.Key.SegLen != 3 {
		t.Errorf("param 0: expected 3 segments, got %d", p0.Key.SegLen)
	}

	seg0 := arena.Segments[p0.Key.SegStart]
	seg1 := arena.Segments[p0.Key.SegStart+1]
	seg2 := arena.Segments[p0.Key.SegStart+2]

	if seg0.Notation != NotationRoot {
		t.Errorf("seg0: expected NotationRoot, got %v", seg0.Notation)
	}
	if seg1.Notation != NotationDot {
		t.Errorf("seg1: expected NotationDot, got %v", seg1.Notation)
	}
	if seg2.Notation != NotationDot {
		t.Errorf("seg2: expected NotationDot, got %v", seg2.Notation)
	}

	// Second: config.nested.deep.value -> 4 segments
	p1 := arena.Params[1]
	if p1.Key.SegLen != 4 {
		t.Errorf("param 1: expected 4 segments, got %d", p1.Key.SegLen)
	}
}

func TestDialect_DotNotation_Mixed(t *testing.T) {
	// Mixed dot and bracket notation
	input := "user.profile[settings].theme=dark&user[tags].0=admin"
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// First: user.profile[settings].theme
	p0 := arena.Params[0]
	if p0.Key.SegLen != 4 {
		t.Fatalf("param 0: expected 4 segments, got %d", p0.Key.SegLen)
	}

	notations := []Notation{NotationRoot, NotationDot, NotationBracket, NotationDot}
	for i, expected := range notations {
		seg := arena.Segments[p0.Key.SegStart+uint16(i)]
		if seg.Notation != expected {
			t.Errorf("seg %d: expected %v, got %v", i, expected, seg.Notation)
		}
	}
}

func TestDialect_DotNotation_EncodedDot(t *testing.T) {
	// Encoded dot %2E
	input := "a%2Eb=c"
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots | FlagDecodeDotInKeys

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	p0 := arena.Params[0]
	if p0.Key.SegLen != 2 {
		t.Fatalf("expected 2 segments, got %d", p0.Key.SegLen)
	}
}

// =============================================================================
// COMMA-SEPARATED DIALECT (§5.3)
// =============================================================================

func TestDialect_Comma_Values(t *testing.T) {
	// From TestArrayFormatComma
	input := "colors=red,green,blue&numbers=1,2,3,4,5"
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}

	// First value: red,green,blue -> ValComma with 3 parts
	v0 := arena.Values[arena.Params[0].ValueIdx]
	if v0.Kind != ValComma {
		t.Errorf("value 0: expected ValComma, got %v", v0.Kind)
	}
	if v0.PartsLen != 3 {
		t.Errorf("value 0: expected 3 parts, got %d", v0.PartsLen)
	}

	// Check parts
	expectedParts := []string{"red", "green", "blue"}
	for i, expected := range expectedParts {
		part := arena.ValueParts[v0.PartsOff+uint16(i)]
		if arena.GetString(part) != expected {
			t.Errorf("part %d: expected %q, got %q", i, expected, arena.GetString(part))
		}
	}

	// Second value: 1,2,3,4,5 -> ValComma with 5 parts
	v1 := arena.Values[arena.Params[1].ValueIdx]
	if v1.Kind != ValComma {
		t.Errorf("value 1: expected ValComma, got %v", v1.Kind)
	}
	if v1.PartsLen != 5 {
		t.Errorf("value 1: expected 5 parts, got %d", v1.PartsLen)
	}
}

func TestDialect_Comma_SingleValue(t *testing.T) {
	// Single value without comma should be ValSimple
	input := "a=single"
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	v := arena.Values[arena.Params[0].ValueIdx]
	if v.Kind != ValSimple {
		t.Errorf("expected ValSimple for single value, got %v", v.Kind)
	}
}

func TestDialect_Comma_WithArrayLimit(t *testing.T) {
	// From TestCommaWithArrayLimit
	input := "a=1,2,3,4,5"
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma
	cfg.ArrayLimit = 2

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	v := arena.Values[arena.Params[0].ValueIdx]
	if v.Kind != ValComma {
		t.Errorf("expected ValComma, got %v", v.Kind)
	}
	// Comma parsing is not affected by ArrayLimit (ArrayLimit is for bracket indices)
	if v.PartsLen != 5 {
		t.Errorf("expected 5 parts, got %d", v.PartsLen)
	}
}

// =============================================================================
// SEMICOLON DELIMITER DIALECT (§5.4)
// =============================================================================

func TestDialect_Semicolon_Basic(t *testing.T) {
	// From TestCustomDelimiter
	input := "a=1;b=2;c=3"
	cfg := DefaultConfig()
	cfg.Delimiter = ';'

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		seg := arena.Segments[arena.Params[i].Key.SegStart]
		if arena.GetString(seg.Span) != exp {
			t.Errorf("param %d: expected %q, got %q", i, exp, arena.GetString(seg.Span))
		}
	}
}

func TestDialect_Semicolon_Nested(t *testing.T) {
	// Semicolon with nested keys
	input := "a[b]=1;c[d][e]=2"
	cfg := DefaultConfig()
	cfg.Delimiter = ';'

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}

	// First: a[b] -> 2 segments
	if arena.Params[0].Key.SegLen != 2 {
		t.Errorf("param 0: expected 2 segments, got %d", arena.Params[0].Key.SegLen)
	}

	// Second: c[d][e] -> 3 segments
	if arena.Params[1].Key.SegLen != 3 {
		t.Errorf("param 1: expected 3 segments, got %d", arena.Params[1].Key.SegLen)
	}
}

// =============================================================================
// FLAT DIALECT (§5.5) - Depth=0
// =============================================================================

func TestDialect_Flat_NoNesting(t *testing.T) {
	// From TestParseArraysFalse concept - flat mode treats brackets as literal
	input := "a[b][c]=value&d[0]=x"
	cfg := DefaultConfig()
	cfg.Depth = 0

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}

	// First param should be single segment with full key "a[b][c]"
	p0 := arena.Params[0]
	if p0.Key.SegLen != 1 {
		t.Errorf("param 0: expected 1 segment (flat mode), got %d", p0.Key.SegLen)
	}
	seg := arena.Segments[p0.Key.SegStart]
	if arena.GetString(seg.Span) != "a[b][c]" {
		t.Errorf("expected 'a[b][c]', got %q", arena.GetString(seg.Span))
	}
}

func TestDialect_Flat_WithDots(t *testing.T) {
	// Flat mode with dots enabled still doesn't parse
	input := "a.b.c=value"
	cfg := DefaultConfig()
	cfg.Depth = 0
	cfg.Flags |= FlagAllowDots

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	p0 := arena.Params[0]
	if p0.Key.SegLen != 1 {
		t.Errorf("expected 1 segment (flat mode), got %d", p0.Key.SegLen)
	}
	seg := arena.Segments[p0.Key.SegStart]
	if arena.GetString(seg.Span) != "a.b.c" {
		t.Errorf("expected 'a.b.c', got %q", arena.GetString(seg.Span))
	}
}

// =============================================================================
// STRICT NULL DIALECT (§5.6)
// =============================================================================

func TestDialect_StrictNull_KeyWithoutEquals(t *testing.T) {
	// From TestSparseArraysNulls concept
	input := "enabled&disabled=false&count=0"
	cfg := DefaultConfig()
	cfg.Flags |= FlagStrictNullHandling

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	// First: "enabled" without = -> ValNull
	p0 := arena.Params[0]
	if p0.HasEquals {
		t.Error("param 0: expected HasEquals=false")
	}
	v0 := arena.Values[p0.ValueIdx]
	if v0.Kind != ValNull {
		t.Errorf("param 0: expected ValNull, got %v", v0.Kind)
	}

	// Second: "disabled=false" -> ValSimple
	p1 := arena.Params[1]
	if !p1.HasEquals {
		t.Error("param 1: expected HasEquals=true")
	}
	v1 := arena.Values[p1.ValueIdx]
	if v1.Kind != ValSimple {
		t.Errorf("param 1: expected ValSimple, got %v", v1.Kind)
	}
}

func TestDialect_StrictNull_EmptyValue(t *testing.T) {
	// a= should still be ValSimple (empty string), not ValNull
	input := "a=&b"
	cfg := DefaultConfig()
	cfg.Flags |= FlagStrictNullHandling

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// a= -> ValSimple (empty)
	p0 := arena.Params[0]
	if !p0.HasEquals {
		t.Error("param 0: expected HasEquals=true")
	}
	v0 := arena.Values[p0.ValueIdx]
	if v0.Kind != ValSimple {
		t.Errorf("param 0: expected ValSimple, got %v", v0.Kind)
	}

	// b -> ValNull
	p1 := arena.Params[1]
	if p1.HasEquals {
		t.Error("param 1: expected HasEquals=false")
	}
	v1 := arena.Values[p1.ValueIdx]
	if v1.Kind != ValNull {
		t.Errorf("param 1: expected ValNull, got %v", v1.Kind)
	}
}

// =============================================================================
// SPARSE ARRAY DIALECT (§5.7) - FlagAllowSparse
// Note: AllowSparse is semantic layer, but we test that indices are parsed
// =============================================================================

func TestDialect_SparseArray_LargeGaps(t *testing.T) {
	// Sparse array with gaps: a[0]=x&a[5]=y
	input := "a[0]=first&a[5]=last"
	cfg := DefaultConfig()
	cfg.ArrayLimit = 20

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}

	// Both should have SegIndex
	seg0 := arena.Segments[arena.Params[0].Key.SegStart+1]
	seg1 := arena.Segments[arena.Params[1].Key.SegStart+1]

	if seg0.Kind != SegIndex || seg0.Index != 0 {
		t.Errorf("param 0: expected SegIndex 0, got %v %d", seg0.Kind, seg0.Index)
	}
	if seg1.Kind != SegIndex || seg1.Index != 5 {
		t.Errorf("param 1: expected SegIndex 5, got %v %d", seg1.Kind, seg1.Index)
	}
}

// =============================================================================
// COMBINED DIALECTS (§5.8)
// =============================================================================

func TestDialect_Combined_DotsAndComma(t *testing.T) {
	// AllowDots + Comma
	input := "user.tags=admin,editor&user.active"
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots | FlagComma | FlagStrictNullHandling

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}

	// First: user.tags=admin,editor
	p0 := arena.Params[0]
	if p0.Key.SegLen != 2 {
		t.Errorf("param 0: expected 2 segments, got %d", p0.Key.SegLen)
	}
	v0 := arena.Values[p0.ValueIdx]
	if v0.Kind != ValComma {
		t.Errorf("param 0: expected ValComma, got %v", v0.Kind)
	}
	if v0.PartsLen != 2 {
		t.Errorf("param 0: expected 2 parts, got %d", v0.PartsLen)
	}

	// Second: user.active (no =) -> ValNull
	p1 := arena.Params[1]
	if p1.Key.SegLen != 2 {
		t.Errorf("param 1: expected 2 segments, got %d", p1.Key.SegLen)
	}
	v1 := arena.Values[p1.ValueIdx]
	if v1.Kind != ValNull {
		t.Errorf("param 1: expected ValNull, got %v", v1.Kind)
	}
}

func TestDialect_Combined_DotsAndBrackets(t *testing.T) {
	// Mixed dots and brackets
	input := "users[0].name=Alice&users[0].roles[0]=admin&users[1].name=Bob"
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	// First: users[0].name -> 3 segments
	p0 := arena.Params[0]
	if p0.Key.SegLen != 3 {
		t.Errorf("param 0: expected 3 segments, got %d", p0.Key.SegLen)
	}

	// Check segment types
	seg0 := arena.Segments[p0.Key.SegStart]
	seg1 := arena.Segments[p0.Key.SegStart+1]
	seg2 := arena.Segments[p0.Key.SegStart+2]

	if seg0.Notation != NotationRoot || arena.GetString(seg0.Span) != "users" {
		t.Errorf("seg0: expected root 'users', got %v %q", seg0.Notation, arena.GetString(seg0.Span))
	}
	if seg1.Notation != NotationBracket || seg1.Kind != SegIndex || seg1.Index != 0 {
		t.Errorf("seg1: expected bracket index 0, got %v %v %d", seg1.Notation, seg1.Kind, seg1.Index)
	}
	if seg2.Notation != NotationDot || arena.GetString(seg2.Span) != "name" {
		t.Errorf("seg2: expected dot 'name', got %v %q", seg2.Notation, arena.GetString(seg2.Span))
	}
}

// =============================================================================
// REAL-WORLD QUERY STRINGS FROM JSCOMPAT TESTS
// =============================================================================

func TestRealWorld_APIQuery(t *testing.T) {
	// From TestRealWorldAPI
	input := "filters[status][0]=active&filters[status][1]=pending&filters[created][$gte]=2024-01-01&filters[created][$lte]=2024-12-31&pagination[page]=1&pagination[limit]=25"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 6 {
		t.Fatalf("expected 6 params, got %d", len(arena.Params))
	}

	// Check filters[status][0]
	p0 := arena.Params[0]
	if p0.Key.SegLen != 3 {
		t.Errorf("param 0: expected 3 segments, got %d", p0.Key.SegLen)
	}

	// Check that $gte is parsed as identifier (not special)
	p2 := arena.Params[2]
	seg2_2 := arena.Segments[p2.Key.SegStart+2]
	if arena.GetString(seg2_2.Span) != "$gte" {
		t.Errorf("expected '$gte', got %q", arena.GetString(seg2_2.Span))
	}
}

func TestRealWorld_SpecialChars(t *testing.T) {
	// From TestSpecialCharsUnicode (encoded)
	input := "message=Hello%2C%20World!&unicode=%E6%97%A5%E6%9C%AC%E8%AA%9E&ampersand=tom%26jerry"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	// Values should be raw (not decoded by AST parser)
	v0 := arena.Values[arena.Params[0].ValueIdx]
	if arena.GetString(v0.Raw) != "Hello%2C%20World!" {
		t.Errorf("expected raw value 'Hello%%2C%%20World!', got %q", arena.GetString(v0.Raw))
	}
}

func TestRealWorld_MegaComplex(t *testing.T) {
	// Simplified version of TestMegaComplex
	input := "users[0].name=Alice&users[0].roles[0]=admin&filters.active=true&tags[0]=important&tags[1]=urgent"
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 5 {
		t.Fatalf("expected 5 params, got %d", len(arena.Params))
	}

	// Verify mixed notation parsing
	// users[0].name -> root, bracket(index), dot
	p0 := arena.Params[0]
	if p0.Key.SegLen != 3 {
		t.Fatalf("param 0: expected 3 segments, got %d", p0.Key.SegLen)
	}

	seg0 := arena.Segments[p0.Key.SegStart]
	seg1 := arena.Segments[p0.Key.SegStart+1]
	seg2 := arena.Segments[p0.Key.SegStart+2]

	if seg0.Kind != SegIdent || seg0.Notation != NotationRoot {
		t.Errorf("seg0: expected ident root")
	}
	if seg1.Kind != SegIndex || seg1.Notation != NotationBracket {
		t.Errorf("seg1: expected index bracket")
	}
	if seg2.Kind != SegIdent || seg2.Notation != NotationDot {
		t.Errorf("seg2: expected ident dot")
	}
}

// =============================================================================
// DEPTH LIMIT TESTS
// =============================================================================

func TestDialect_Depth_WithBrackets(t *testing.T) {
	// From TestDepthLimits: a[b][c][d][e][f][g]=deep with depth 3
	input := "a[b][c][d][e][f][g]=deep"
	cfg := DefaultConfig()
	cfg.Depth = 3

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	p := arena.Params[0]

	// Should have segments up to depth, then literal remainder
	hasLiteral := false
	for i := uint16(0); i < uint16(p.Key.SegLen); i++ {
		seg := arena.Segments[p.Key.SegStart+i]
		if seg.Kind == SegLiteral {
			hasLiteral = true
			// Literal should contain remaining brackets
			lit := arena.GetString(seg.Span)
			if !strings.Contains(lit, "[") {
				t.Errorf("literal should contain brackets: %q", lit)
			}
		}
	}

	if !hasLiteral {
		t.Error("expected SegLiteral for depth-exceeded remainder")
	}
}

func TestDialect_Depth_WithDots(t *testing.T) {
	// From TestAllowDotsWithDepthStrict: a.b.c.d=e with depth 2
	input := "a.b.c.d=e"
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots
	cfg.Depth = 2

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	p := arena.Params[0]

	// Should have literal for remaining .c.d
	hasLiteral := false
	for i := uint16(0); i < uint16(p.Key.SegLen); i++ {
		seg := arena.Segments[p.Key.SegStart+i]
		if seg.Kind == SegLiteral {
			hasLiteral = true
		}
	}

	if !hasLiteral {
		t.Error("expected SegLiteral for depth-exceeded remainder")
	}
}

func TestDialect_StrictDepth_Error(t *testing.T) {
	// From TestStrictDepth
	input := "a[b][c][d]=e"
	cfg := DefaultConfig()
	cfg.Depth = 2
	cfg.Flags |= FlagStrictDepth

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != ErrDepthLimitExceeded {
		t.Fatalf("expected ErrDepthLimitExceeded, got %v", err)
	}
}

// =============================================================================
// PARAMETER LIMIT TESTS
// =============================================================================

func TestDialect_ParameterLimit(t *testing.T) {
	// From TestParameterLimit
	input := "a=1&b=2&c=3&d=4&e=5"
	cfg := DefaultConfig()
	cfg.ParameterLimit = 3

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(arena.Params) != 3 {
		t.Errorf("expected 3 params (limit), got %d", len(arena.Params))
	}
}

func TestDialect_ParameterLimit_Throw(t *testing.T) {
	// From TestThrowOnLimitExceeded
	input := "a=1&b=2&c=3"
	cfg := DefaultConfig()
	cfg.ParameterLimit = 2
	cfg.Flags |= FlagThrowOnLimitExceeded

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != ErrParameterLimitExceeded {
		t.Fatalf("expected ErrParameterLimitExceeded, got %v", err)
	}
}

// =============================================================================
// CHARSET SENTINEL TESTS
// =============================================================================

func TestDialect_CharsetSentinel_UTF8(t *testing.T) {
	// From TestCharsetSentinel
	input := "utf8=%E2%9C%93&a=b"
	cfg := DefaultConfig()
	cfg.Flags |= FlagCharsetSentinel

	arena := NewArena(32)
	_, cs, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if cs != CharsetUTF8 {
		t.Errorf("expected CharsetUTF8, got %v", cs)
	}

	// utf8= param should be skipped
	if len(arena.Params) != 1 {
		t.Errorf("expected 1 param (utf8 skipped), got %d", len(arena.Params))
	}
}

func TestDialect_CharsetSentinel_ISO(t *testing.T) {
	// ISO-8859-1 sentinel
	input := "utf8=%26%2310003%3B&a=b"
	cfg := DefaultConfig()
	cfg.Flags |= FlagCharsetSentinel

	arena := NewArena(32)
	_, cs, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if cs != CharsetISO88591 {
		t.Errorf("expected CharsetISO88591, got %v", cs)
	}
}

// =============================================================================
// PROTOTYPE BLOCKING TESTS
// =============================================================================

func TestDialect_ProtoBlocked(t *testing.T) {
	// From TestAllowPrototypes
	input := "a[__proto__][b]=c&normal=value"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// __proto__ param should be skipped
	if len(arena.Params) != 1 {
		t.Errorf("expected 1 param (__proto__ blocked), got %d", len(arena.Params))
	}

	seg := arena.Segments[arena.Params[0].Key.SegStart]
	if arena.GetString(seg.Span) != "normal" {
		t.Errorf("expected 'normal', got %q", arena.GetString(seg.Span))
	}
}

func TestDialect_ProtoAllowed(t *testing.T) {
	// With FlagAllowPrototypes (but __proto__ still blocked!)
	input := "constructor=x&prototype=y"
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowPrototypes

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// With AllowPrototypes, constructor and prototype are allowed
	if len(arena.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(arena.Params))
	}
}

// =============================================================================
// QUERY PREFIX TESTS
// =============================================================================

func TestDialect_QueryPrefix_Ignored(t *testing.T) {
	// From TestQueryPrefix
	input := "?a=b&c=d"
	cfg := DefaultConfig()
	cfg.Flags |= FlagIgnoreQueryPrefix

	arena := NewArena(32)
	qs, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !qs.HasPrefix {
		t.Error("expected HasPrefix=true")
	}

	if len(arena.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(arena.Params))
	}

	// First key should be "a", not "?a"
	seg := arena.Segments[arena.Params[0].Key.SegStart]
	if arena.GetString(seg.Span) != "a" {
		t.Errorf("expected 'a', got %q", arena.GetString(seg.Span))
	}
}

func TestDialect_QueryPrefix_NotIgnored(t *testing.T) {
	// Without flag, ? is part of the key
	input := "?a=b"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	seg := arena.Segments[arena.Params[0].Key.SegStart]
	if arena.GetString(seg.Span) != "?a" {
		t.Errorf("expected '?a', got %q", arena.GetString(seg.Span))
	}
}

// =============================================================================
// ARRAY LIMIT TESTS
// =============================================================================

func TestDialect_ArrayLimit_BeyondLimit(t *testing.T) {
	// From TestArrayLimit: a[100]=b with arrayLimit 50
	input := "a[100]=b"
	cfg := DefaultConfig()
	cfg.ArrayLimit = 50

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Index 100 > limit 50, so should be SegIdent not SegIndex
	seg := arena.Segments[arena.Params[0].Key.SegStart+1]
	if seg.Kind != SegIdent {
		t.Errorf("expected SegIdent (beyond limit), got %v", seg.Kind)
	}
	if seg.Index != -1 {
		t.Errorf("expected Index=-1, got %d", seg.Index)
	}
}

func TestDialect_ArrayLimit_AtLimit(t *testing.T) {
	input := "a[50]=b"
	cfg := DefaultConfig()
	cfg.ArrayLimit = 50

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Index 50 == limit 50, should be SegIndex
	seg := arena.Segments[arena.Params[0].Key.SegStart+1]
	if seg.Kind != SegIndex {
		t.Errorf("expected SegIndex (at limit), got %v", seg.Kind)
	}
	if seg.Index != 50 {
		t.Errorf("expected Index=50, got %d", seg.Index)
	}
}

// =============================================================================
// DUPLICATES HANDLING (AST level - just counts params)
// =============================================================================

func TestDialect_Duplicates_AllKept(t *testing.T) {
	// From TestDuplicatesHandling - AST keeps all params
	input := "a=1&a=2&a=3"
	cfg := DefaultConfig()

	arena := NewArena(32)
	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// All 3 should be in AST
	if len(arena.Params) != 3 {
		t.Errorf("expected 3 params (all duplicates kept), got %d", len(arena.Params))
	}

	// All have same key
	for i, p := range arena.Params {
		seg := arena.Segments[p.Key.SegStart]
		if arena.GetString(seg.Span) != "a" {
			t.Errorf("param %d: expected 'a', got %q", i, arena.GetString(seg.Span))
		}
	}
}
