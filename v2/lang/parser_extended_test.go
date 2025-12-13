package lang

import (
	"strings"
	"testing"
)

// =============================================================================
// BASIC PARSING TESTS
// =============================================================================

func TestParse_EmptyInput(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	qs, _, err := Parse(arena, "", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if qs.ParamLen != 0 {
		t.Fatalf("expected 0 params, got %d", qs.ParamLen)
	}
}

func TestParse_QueryPrefix(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagIgnoreQueryPrefix

	qs, _, err := Parse(arena, "?a=b&c=d", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !qs.HasPrefix {
		t.Fatal("expected HasPrefix=true")
	}
	if qs.ParamLen != 2 {
		t.Fatalf("expected 2 params, got %d", qs.ParamLen)
	}
	if got := arena.GetString(arena.Params[0].Key.Raw); got != "a" {
		t.Fatalf("expected key 'a', got %q", got)
	}
}

func TestParse_QueryPrefixWithoutFlag(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	// FlagIgnoreQueryPrefix NOT set

	qs, _, err := Parse(arena, "?a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !qs.HasPrefix {
		t.Fatal("expected HasPrefix=true")
	}
	// Without the flag, '?' is part of the key
	if qs.ParamLen != 1 {
		t.Fatalf("expected 1 param, got %d", qs.ParamLen)
	}
	if got := arena.GetString(arena.Params[0].Key.Raw); got != "?a" {
		t.Fatalf("expected key '?a', got %q", got)
	}
}

func TestParse_OnlyQueryPrefix(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagIgnoreQueryPrefix

	qs, _, err := Parse(arena, "?", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if qs.ParamLen != 0 {
		t.Fatalf("expected 0 params, got %d", qs.ParamLen)
	}
}

func TestParse_KeyWithoutValue(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "foo", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if p.HasEquals {
		t.Fatal("expected HasEquals=false")
	}
	if p.ValueIdx != noValue {
		t.Fatalf("expected no value, got idx %d", p.ValueIdx)
	}
}

func TestParse_KeyWithEmptyValue(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "foo=", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if !p.HasEquals {
		t.Fatal("expected HasEquals=true")
	}
	if p.ValueIdx == noValue {
		t.Fatal("expected value node")
	}
	v := arena.Values[p.ValueIdx]
	if v.Kind != ValSimple {
		t.Fatalf("expected ValSimple, got %v", v.Kind)
	}
	if arena.GetString(v.Raw) != "" {
		t.Fatalf("expected empty value, got %q", arena.GetString(v.Raw))
	}
}

func TestParse_StrictNullHandling(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagStrictNullHandling

	_, _, err := Parse(arena, "foo&bar=baz", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// "foo" should have a null value
	p0 := arena.Params[0]
	if p0.HasEquals {
		t.Fatal("foo: expected HasEquals=false")
	}
	if p0.ValueIdx == noValue {
		t.Fatal("foo: expected value node for null")
	}
	v0 := arena.Values[p0.ValueIdx]
	if v0.Kind != ValNull {
		t.Fatalf("foo: expected ValNull, got %v", v0.Kind)
	}

	// "bar=baz" should be normal
	p1 := arena.Params[1]
	if !p1.HasEquals {
		t.Fatal("bar: expected HasEquals=true")
	}
}

func TestParse_EmptyKey_Skipped(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "=value&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Empty key should be skipped
	if len(arena.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(arena.Params))
	}
	if got := arena.GetString(arena.Params[0].Key.Raw); got != "a" {
		t.Fatalf("expected key 'a', got %q", got)
	}
}

func TestParse_MultipleDelimiters(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a=1&&b=2&&&c=3", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Empty segments between delimiters should be skipped
	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}
}

func TestParse_CustomDelimiter(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Delimiter = ';'

	_, _, err := Parse(arena, "a=1;b=2;c=3", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}
}

// =============================================================================
// BRACKET NOTATION TESTS
// =============================================================================

func TestParse_EmptyBrackets(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[]=1&a[]=2", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}

	// Both should have SegEmpty
	for i, p := range arena.Params {
		if p.Key.SegLen != 2 {
			t.Fatalf("param %d: expected 2 segments, got %d", i, p.Key.SegLen)
		}
		seg1 := arena.Segments[p.Key.SegStart+1]
		if seg1.Kind != SegEmpty {
			t.Fatalf("param %d: expected SegEmpty, got %v", i, seg1.Kind)
		}
		if seg1.Notation != NotationBracket {
			t.Fatalf("param %d: expected NotationBracket, got %v", i, seg1.Notation)
		}
	}
}

func TestParse_NestedBrackets(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[b][c][d]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if p.Key.SegLen != 4 {
		t.Fatalf("expected 4 segments, got %d", p.Key.SegLen)
	}

	expected := []struct {
		text     string
		notation Notation
	}{
		{"a", NotationRoot},
		{"b", NotationBracket},
		{"c", NotationBracket},
		{"d", NotationBracket},
	}

	for i, exp := range expected {
		seg := arena.Segments[p.Key.SegStart+uint16(i)]
		if arena.GetString(seg.Span) != exp.text {
			t.Errorf("seg %d: expected %q, got %q", i, exp.text, arena.GetString(seg.Span))
		}
		if seg.Notation != exp.notation {
			t.Errorf("seg %d: expected notation %v, got %v", i, exp.notation, seg.Notation)
		}
	}
}

func TestParse_EncodedBrackets(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	// %5B = '[', %5D = ']'
	_, _, err := Parse(arena, "a%5Bb%5D=c", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if p.Key.SegLen != 2 {
		t.Fatalf("expected 2 segments, got %d", p.Key.SegLen)
	}
	seg1 := arena.Segments[p.Key.SegStart+1]
	if seg1.Notation != NotationBracket {
		t.Fatalf("expected NotationBracket, got %v", seg1.Notation)
	}
}

func TestParse_MixedBracketsCase(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	// Mixed case: %5b (lowercase)
	_, _, err := Parse(arena, "a%5bb%5d=c", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if p.Key.SegLen != 2 {
		t.Fatalf("expected 2 segments, got %d", p.Key.SegLen)
	}
}

func TestParse_UnclosedBracket(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[b=c", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Unclosed bracket should be treated as literal
	p := arena.Params[0]
	if got := arena.GetString(p.Key.Raw); got != "a[b" {
		t.Fatalf("expected 'a[b', got %q", got)
	}
}

func TestParse_NestedUnclosedBracket(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	// '[' inside brackets is not allowed
	_, _, err := Parse(arena, "a[[b]]=c", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	// Should not parse nested open bracket
	if got := arena.GetString(p.Key.Raw); got != "a[[b]]" {
		t.Fatalf("expected 'a[[b]]', got %q", got)
	}
}

func TestParse_KeyStartingWithBracket(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "[a]=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	// [a] should be treated as first bracket segment with no root
	// Actually according to spec, bracket key is allowed as root
	if got := arena.GetString(p.Key.Raw); got != "[a]" {
		t.Fatalf("expected '[a]', got %q", got)
	}
}

// =============================================================================
// ARRAY INDEX TESTS
// =============================================================================

func TestParse_ArrayIndices(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[0]=x&a[1]=y&a[2]=z", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	for i := 0; i < 3; i++ {
		p := arena.Params[i]
		seg := arena.Segments[p.Key.SegStart+1]
		if seg.Kind != SegIndex {
			t.Errorf("param %d: expected SegIndex, got %v", i, seg.Kind)
		}
		if seg.Index != int32(i) {
			t.Errorf("param %d: expected index %d, got %d", i, i, seg.Index)
		}
	}
}

func TestParse_ArrayIndexBeyondLimit(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ArrayLimit = 5

	_, _, err := Parse(arena, "a[10]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	seg := arena.Segments[p.Key.SegStart+1]
	// Index beyond limit should be treated as identifier
	if seg.Kind != SegIdent {
		t.Fatalf("expected SegIdent, got %v", seg.Kind)
	}
	if seg.Index != -1 {
		t.Fatalf("expected index -1, got %d", seg.Index)
	}
}

func TestParse_ArrayIndexAtLimit(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ArrayLimit = 10

	_, _, err := Parse(arena, "a[10]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	seg := arena.Segments[p.Key.SegStart+1]
	if seg.Kind != SegIndex {
		t.Fatalf("expected SegIndex, got %v", seg.Kind)
	}
	if seg.Index != 10 {
		t.Fatalf("expected index 10, got %d", seg.Index)
	}
}

func TestParse_LeadingZeroIndex(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[01]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	seg := arena.Segments[p.Key.SegStart+1]
	// Leading zero makes it non-canonical, should be identifier
	if seg.Kind != SegIdent {
		t.Fatalf("expected SegIdent (leading zero), got %v", seg.Kind)
	}
}

func TestParse_ZeroIndex(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[0]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	seg := arena.Segments[p.Key.SegStart+1]
	if seg.Kind != SegIndex {
		t.Fatalf("expected SegIndex, got %v", seg.Kind)
	}
	if seg.Index != 0 {
		t.Fatalf("expected index 0, got %d", seg.Index)
	}
}

func TestParse_ParseArraysDisabled(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ParseArrays = false

	_, _, err := Parse(arena, "a[0]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	seg := arena.Segments[p.Key.SegStart+1]
	// With ParseArrays=false, indices should be identifiers
	if seg.Kind != SegIdent {
		t.Fatalf("expected SegIdent, got %v", seg.Kind)
	}
}

// =============================================================================
// DOT NOTATION TESTS
// =============================================================================

func TestParse_DotsDisabled(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	// FlagAllowDots NOT set

	_, _, err := Parse(arena, "a.b.c=d", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	// Without AllowDots, the whole key is one segment
	if p.Key.SegLen != 1 {
		t.Fatalf("expected 1 segment, got %d", p.Key.SegLen)
	}
	if got := arena.GetString(p.Key.Raw); got != "a.b.c" {
		t.Fatalf("expected 'a.b.c', got %q", got)
	}
}

func TestParse_EncodedDots(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots | FlagDecodeDotInKeys

	// %2E = '.'
	_, _, err := Parse(arena, "a%2Eb=c", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	if p.Key.SegLen != 2 {
		t.Fatalf("expected 2 segments, got %d", p.Key.SegLen)
	}
}

func TestParse_ConsecutiveDots(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	_, _, err := Parse(arena, "a..b=c", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Consecutive dots should not create empty segments
	p := arena.Params[0]
	// Implementation behavior may vary - just check no crash
	t.Logf("segments: %d, key: %q", p.Key.SegLen, arena.GetString(p.Key.Raw))
}

func TestParse_TrailingDot(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	_, _, err := Parse(arena, "a.b.=c", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	t.Logf("segments: %d, key: %q", p.Key.SegLen, arena.GetString(p.Key.Raw))
}

func TestParse_DotBeforeBracket(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	_, _, err := Parse(arena, "a.b[c]=d", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	if p.Key.SegLen != 3 {
		t.Fatalf("expected 3 segments, got %d", p.Key.SegLen)
	}

	s0 := arena.Segments[p.Key.SegStart]
	s1 := arena.Segments[p.Key.SegStart+1]
	s2 := arena.Segments[p.Key.SegStart+2]

	if s0.Notation != NotationRoot {
		t.Errorf("s0: expected NotationRoot, got %v", s0.Notation)
	}
	if s1.Notation != NotationDot {
		t.Errorf("s1: expected NotationDot, got %v", s1.Notation)
	}
	if s2.Notation != NotationBracket {
		t.Errorf("s2: expected NotationBracket, got %v", s2.Notation)
	}
}

// =============================================================================
// DEPTH LIMIT TESTS
// =============================================================================

func TestParse_DepthLimit_Brackets(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Depth = 2

	_, _, err := Parse(arena, "a[b][c][d][e]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	// Should have: a, [b], [c], literal[d][e]
	found := false
	for i := uint16(0); i < uint16(p.Key.SegLen); i++ {
		seg := arena.Segments[p.Key.SegStart+i]
		if seg.Kind == SegLiteral {
			found = true
			lit := arena.GetString(seg.Span)
			if !strings.Contains(lit, "[d]") {
				t.Errorf("literal should contain [d], got %q", lit)
			}
		}
	}
	if !found {
		t.Error("expected a SegLiteral segment")
	}
}

func TestParse_DepthZero_FlatMode(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Depth = 0

	_, _, err := Parse(arena, "a[b][c]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	// Depth 0 = flat mode, entire key as single segment
	if p.Key.SegLen != 1 {
		t.Fatalf("expected 1 segment (flat mode), got %d", p.Key.SegLen)
	}
	if got := arena.GetString(p.Key.Raw); got != "a[b][c]" {
		t.Fatalf("expected 'a[b][c]', got %q", got)
	}
}

// =============================================================================
// COMMA VALUES TESTS
// =============================================================================

func TestParse_CommaValues_NoCommaFlag(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	// FlagComma NOT set

	_, _, err := Parse(arena, "a=x,y,z", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	v := arena.Values[p.ValueIdx]
	// Without FlagComma, commas are part of the value
	if v.Kind != ValSimple {
		t.Fatalf("expected ValSimple, got %v", v.Kind)
	}
	if got := arena.GetString(v.Raw); got != "x,y,z" {
		t.Fatalf("expected 'x,y,z', got %q", got)
	}
}

func TestParse_CommaValues_Single(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma

	_, _, err := Parse(arena, "a=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	v := arena.Values[p.ValueIdx]
	// Single value without comma should be ValSimple
	if v.Kind != ValSimple {
		t.Fatalf("expected ValSimple, got %v", v.Kind)
	}
}

func TestParse_CommaValues_Empty(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma

	_, _, err := Parse(arena, "a=,", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	v := arena.Values[p.ValueIdx]
	if v.Kind != ValComma {
		t.Fatalf("expected ValComma, got %v", v.Kind)
	}
	if v.PartsLen != 2 {
		t.Fatalf("expected 2 parts, got %d", v.PartsLen)
	}
}

func TestParse_CommaValues_Many(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma

	_, _, err := Parse(arena, "a=1,2,3,4,5", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	v := arena.Values[p.ValueIdx]
	if v.Kind != ValComma {
		t.Fatalf("expected ValComma, got %v", v.Kind)
	}
	if v.PartsLen != 5 {
		t.Fatalf("expected 5 parts, got %d", v.PartsLen)
	}

	for i := uint16(0); i < uint16(v.PartsLen); i++ {
		part := arena.ValueParts[v.PartsOff+i]
		expected := string('1' + byte(i))
		if got := arena.GetString(part); got != expected {
			t.Errorf("part %d: expected %q, got %q", i, expected, got)
		}
	}
}

// =============================================================================
// CHARSET SENTINEL TESTS
// =============================================================================

func TestParse_CharsetSentinel_ISO(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagCharsetSentinel

	// ISO-8859-1 sentinel: &#10003; encoded
	_, cs, err := Parse(arena, "utf8=%26%2310003%3B&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cs != CharsetISO88591 {
		t.Fatalf("expected CharsetISO88591, got %v", cs)
	}
}

func TestParse_CharsetSentinel_Unknown(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagCharsetSentinel

	// Unknown utf8 value should be treated as normal param
	_, cs, err := Parse(arena, "utf8=unknown&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Default charset
	if cs != CharsetUTF8 {
		t.Fatalf("expected CharsetUTF8, got %v", cs)
	}
	// utf8=unknown should be kept as param
	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}
}

func TestParse_CharsetSentinel_OnlyFirst(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagCharsetSentinel

	// Only first utf8= is checked
	_, cs, err := Parse(arena, "utf8=%E2%9C%93&utf8=%26%2310003%3B&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// First is UTF-8
	if cs != CharsetUTF8 {
		t.Fatalf("expected CharsetUTF8, got %v", cs)
	}
	// Second utf8= should be kept as regular param
	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}
}

// =============================================================================
// PROTOTYPE KEY TESTS
// =============================================================================

func TestParse_ProtoIsNormalKey(t *testing.T) {
	// In Go there's no prototype pollution, so __proto__ is just a normal key
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "__proto__=x&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Both params should be parsed - __proto__ is a normal key
	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}
}

func TestParse_PrototypeKeysAreNormal(t *testing.T) {
	// In Go, these are just normal keys (no prototype pollution)
	arena := NewArena(8)
	cfg := DefaultConfig()

	keys := []string{
		"constructor",
		"prototype",
		"toString",
		"hasOwnProperty",
	}

	for _, key := range keys {
		arena.Reset("")
		_, _, err := Parse(arena, key+"=x&ok=y", cfg)
		if err != nil {
			t.Fatalf("Parse %s: %v", key, err)
		}
		// Both params should be parsed
		if len(arena.Params) != 2 {
			t.Errorf("%s: expected 2 params, got %d", key, len(arena.Params))
		}
	}
}

func TestParse_ProtoInNested(t *testing.T) {
	// In Go, __proto__ in nested position is also just a normal key
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[__proto__]=x&b=y", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Both params should be parsed
	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}
}

func TestParse_ProtoEncoded(t *testing.T) {
	// Encoded __proto__ is also just a normal key in Go
	arena := NewArena(8)
	cfg := DefaultConfig()

	// %5F = '_'
	_, _, err := Parse(arena, "%5F%5Fproto%5F%5F=x&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Both params should be parsed
	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}
}

// =============================================================================
// PARAMETER LIMIT TESTS
// =============================================================================

func TestParse_ParameterLimit(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ParameterLimit = 3

	_, _, err := Parse(arena, "a=1&b=2&c=3&d=4&e=5", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Should only have first 3 params
	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}
}

func TestParse_ParameterLimit_ThrowOnExceeded(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ParameterLimit = 2
	cfg.Flags |= FlagThrowOnLimitExceeded

	_, _, err := Parse(arena, "a=1&b=2&c=3", cfg)
	if err != ErrParameterLimitExceeded {
		t.Fatalf("expected ErrParameterLimitExceeded, got %v", err)
	}
}

func TestParse_ParameterLimit_Zero(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ParameterLimit = 0 // should mean unlimited

	params := make([]string, 100)
	for i := range params {
		params[i] = string(rune('a'+i%26)) + "=" + string(rune('0'+i%10))
	}
	input := strings.Join(params, "&")

	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 100 {
		t.Fatalf("expected 100 params, got %d", len(arena.Params))
	}
}

// =============================================================================
// STRICT PROFILE TESTS
// =============================================================================

func TestParse_StrictProfile_ValidEncoding(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Profile = ProfileStrict

	_, _, err := Parse(arena, "a=%20%21%7E", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
}

func TestParse_StrictProfile_InvalidEncoding(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Profile = ProfileStrict

	// Invalid: %ZZ is not valid hex
	_, _, err := Parse(arena, "a=%ZZ", cfg)
	if err != ErrInvalidPercentCode {
		t.Fatalf("expected ErrInvalidPercentCode, got %v", err)
	}
}

func TestParse_StrictProfile_TruncatedEncoding(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Profile = ProfileStrict

	// Truncated: % at end
	_, _, err := Parse(arena, "a=test%", cfg)
	if err != ErrInvalidPercentCode {
		t.Fatalf("expected ErrInvalidPercentCode, got %v", err)
	}
}

func TestParse_StrictProfile_TruncatedEncoding2(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Profile = ProfileStrict

	// Truncated: %2 at end
	_, _, err := Parse(arena, "a=test%2", cfg)
	if err != ErrInvalidPercentCode {
		t.Fatalf("expected ErrInvalidPercentCode, got %v", err)
	}
}

func TestParse_FastProfile_InvalidEncoding(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Profile = ProfileFast

	// Fast mode should not validate
	_, _, err := Parse(arena, "a=%ZZ", cfg)
	if err != nil {
		t.Fatalf("Parse: %v (fast mode should not validate)", err)
	}
}

// =============================================================================
// EDGE CASES AND STRESS TESTS
// =============================================================================

func TestParse_VeryLongKey(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	key := strings.Repeat("a", 10000)
	_, _, err := Parse(arena, key+"=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(arena.Params))
	}
}

func TestParse_VeryLongValue(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	value := strings.Repeat("x", 10000)
	_, _, err := Parse(arena, "key="+value, cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	v := arena.Values[arena.Params[0].ValueIdx]
	if int(v.Raw.Len) != 10000 {
		t.Fatalf("expected value len 10000, got %d", v.Raw.Len)
	}
}

func TestParse_ManyParams(t *testing.T) {
	arena := NewArena(1024)
	cfg := DefaultConfig()
	cfg.ParameterLimit = 0

	var parts []string
	for i := 0; i < 1000; i++ {
		parts = append(parts, "k"+string(rune('a'+i%26))+"=v")
	}
	input := strings.Join(parts, "&")

	_, _, err := Parse(arena, input, cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 1000 {
		t.Fatalf("expected 1000 params, got %d", len(arena.Params))
	}
}

func TestParse_DeepNesting(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Depth = 100

	// a[b][c][d]...[z]=value (26 levels)
	key := "a"
	for c := 'b'; c <= 'z'; c++ {
		key += "[" + string(c) + "]"
	}

	_, _, err := Parse(arena, key+"=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	if p.Key.SegLen != 26 {
		t.Fatalf("expected 26 segments, got %d", p.Key.SegLen)
	}
}

func TestParse_SpecialCharsInValue(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	// Various special chars that should be in value
	_, _, err := Parse(arena, "a=hello+world&b=%20%21%40", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	v0 := arena.Values[arena.Params[0].ValueIdx]
	if got := arena.GetString(v0.Raw); got != "hello+world" {
		t.Errorf("value 0: expected 'hello+world', got %q", got)
	}
}

func TestParse_NilArena(t *testing.T) {
	cfg := DefaultConfig()
	_, _, err := Parse(nil, "a=b", cfg)
	if err != ErrNilArena {
		t.Fatalf("expected ErrNilArena, got %v", err)
	}
}

func TestParse_ArenaReuse(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	// First parse
	_, _, err := Parse(arena, "a=1&b=2&c=3", cfg)
	if err != nil {
		t.Fatalf("Parse 1: %v", err)
	}
	if len(arena.Params) != 3 {
		t.Fatalf("Parse 1: expected 3 params, got %d", len(arena.Params))
	}

	// Second parse should reset
	_, _, err = Parse(arena, "x=y", cfg)
	if err != nil {
		t.Fatalf("Parse 2: %v", err)
	}
	if len(arena.Params) != 1 {
		t.Fatalf("Parse 2: expected 1 param, got %d", len(arena.Params))
	}
	if got := arena.GetString(arena.Params[0].Key.Raw); got != "x" {
		t.Fatalf("Parse 2: expected key 'x', got %q", got)
	}
}

// =============================================================================
// SPECIFICATION COMPLIANCE TESTS
// =============================================================================

func TestSpec_Example_NestedObjects(t *testing.T) {
	// Spec §9.2: user[name]=John&user[age]=30
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "user[name]=John&user[age]=30", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}

	// Both should have user as root, then name/age as bracket
	for i, p := range arena.Params {
		if p.Key.SegLen != 2 {
			t.Errorf("param %d: expected 2 segments, got %d", i, p.Key.SegLen)
		}
		root := arena.Segments[p.Key.SegStart]
		if arena.GetString(root.Span) != "user" {
			t.Errorf("param %d: expected root 'user', got %q", i, arena.GetString(root.Span))
		}
	}
}

func TestSpec_Example_DotNotation(t *testing.T) {
	// Spec §9.3: user.name=John&user.address.city=NYC (AllowDots=true)
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	_, _, err := Parse(arena, "user.name=John&user.address.city=NYC", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Second param should have 3 segments: user.address.city
	p1 := arena.Params[1]
	if p1.Key.SegLen != 3 {
		t.Fatalf("expected 3 segments, got %d", p1.Key.SegLen)
	}

	expected := []string{"user", "address", "city"}
	for i, exp := range expected {
		seg := arena.Segments[p1.Key.SegStart+uint16(i)]
		if arena.GetString(seg.Span) != exp {
			t.Errorf("seg %d: expected %q, got %q", i, exp, arena.GetString(seg.Span))
		}
	}
}

func TestSpec_Example_ArrayIndices(t *testing.T) {
	// Spec §9.4: colors[0]=red&colors[1]=blue&colors[2]=green
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "colors[0]=red&colors[1]=blue&colors[2]=green", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	for i, p := range arena.Params {
		seg := arena.Segments[p.Key.SegStart+1]
		if seg.Kind != SegIndex {
			t.Errorf("param %d: expected SegIndex, got %v", i, seg.Kind)
		}
		if seg.Index != int32(i) {
			t.Errorf("param %d: expected index %d, got %d", i, i, seg.Index)
		}
	}
}

func TestSpec_Example_EmptyBrackets(t *testing.T) {
	// Spec §9.5: tags[]=js&tags[]=go&tags[]=rust
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "tags[]=js&tags[]=go&tags[]=rust", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(arena.Params))
	}

	for i, p := range arena.Params {
		seg := arena.Segments[p.Key.SegStart+1]
		if seg.Kind != SegEmpty {
			t.Errorf("param %d: expected SegEmpty, got %v", i, seg.Kind)
		}
	}
}

func TestSpec_Example_CommaSeparated(t *testing.T) {
	// Spec §9.6: ids=1,2,3,4,5 (Comma=true)
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma

	_, _, err := Parse(arena, "ids=1,2,3,4,5", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	v := arena.Values[arena.Params[0].ValueIdx]
	if v.Kind != ValComma {
		t.Fatalf("expected ValComma, got %v", v.Kind)
	}
	if v.PartsLen != 5 {
		t.Fatalf("expected 5 parts, got %d", v.PartsLen)
	}
}

func TestSpec_Example_DepthLimiting(t *testing.T) {
	// Spec §9.9: a[b][c][d]=e with Depth=2
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Depth = 2

	_, _, err := Parse(arena, "a[b][c][d]=e", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	// Should have: a, [b], [c], then literal remainder
	hasLiteral := false
	for i := uint16(0); i < uint16(p.Key.SegLen); i++ {
		seg := arena.Segments[p.Key.SegStart+i]
		if seg.Kind == SegLiteral {
			hasLiteral = true
			if !strings.Contains(arena.GetString(seg.Span), "[d]") {
				t.Errorf("literal should contain [d]: %q", arena.GetString(seg.Span))
			}
		}
	}
	if !hasLiteral {
		t.Error("expected SegLiteral for depth-exceeded remainder")
	}
}

// =============================================================================
// BENCHMARK
// =============================================================================

// =============================================================================
// ADDITIONAL COVERAGE TESTS
// =============================================================================

func TestParse_CharsetSentinel_CaseInsensitive(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagCharsetSentinel

	// Lowercase hex in sentinel value
	_, cs, err := Parse(arena, "utf8=%e2%9c%93&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cs != CharsetUTF8 {
		t.Fatalf("expected CharsetUTF8, got %v", cs)
	}
}

func TestParse_TooManyParams(t *testing.T) {
	// This tests the uint16 overflow check
	// We can't easily trigger this without huge input, but we test the limit logic
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ParameterLimit = 65535

	// Just ensure it doesn't crash
	_, _, err := Parse(arena, "a=1&b=2", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
}

func TestParse_SpanBoundsCheck(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	// Normal input should work
	_, _, err := Parse(arena, "key=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Verify span is correctly bounded
	span := arena.Params[0].Key.Raw
	if int(span.Off)+int(span.Len) > len(arena.Source) {
		t.Fatal("span extends beyond source")
	}
}

func TestParse_DeepNesting_WithDots(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots
	cfg.Depth = 3

	_, _, err := Parse(arena, "a.b.c.d.e=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	// Should have literal segment after depth exceeded
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

func TestParse_StrictDepth_WithDots(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots | FlagStrictDepth
	cfg.Depth = 2

	_, _, err := Parse(arena, "a.b.c.d=value", cfg)
	if err != ErrDepthLimitExceeded {
		t.Fatalf("expected ErrDepthLimitExceeded, got %v", err)
	}
}

func TestParse_ProtoNormalKey(t *testing.T) {
	// In Go, __proto__ is just a normal key (no prototype pollution)
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "__proto__=x&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Both params should be parsed - __proto__ is a normal key in Go
	if len(arena.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(arena.Params))
	}
}

func TestParse_IdentSegmentAsIndex_DotNotation(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	// .0 after a segment should be parsed as index
	_, _, err := Parse(arena, "a.0=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	if p.Key.SegLen != 2 {
		t.Fatalf("expected 2 segments, got %d", p.Key.SegLen)
	}

	seg1 := arena.Segments[p.Key.SegStart+1]
	if seg1.Kind != SegIndex {
		t.Fatalf("expected SegIndex, got %v", seg1.Kind)
	}
	if seg1.Index != 0 {
		t.Fatalf("expected index 0, got %d", seg1.Index)
	}
}

func TestParse_RootSegmentNotIndex(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	// Root segment should not be parsed as index even if numeric
	_, _, err := Parse(arena, "123=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := arena.Params[0]
	seg := arena.Segments[p.Key.SegStart]
	// Root segment with numeric value should still be SegIdent (per qs behavior)
	if seg.Kind != SegIdent {
		t.Fatalf("expected SegIdent for root, got %v", seg.Kind)
	}
}

func TestParse_EmptyBracketBlocked(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	// Empty bracket segment should not be blocked
	_, _, err := Parse(arena, "a[]=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(arena.Params))
	}
}

func TestParse_StrictValidation_Key(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Profile = ProfileStrict

	// Invalid percent encoding in key
	_, _, err := Parse(arena, "a%ZZb=value", cfg)
	if err != ErrInvalidPercentCode {
		t.Fatalf("expected ErrInvalidPercentCode, got %v", err)
	}
}

func TestParse_ValueParts_Limit(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma

	// Many comma-separated values
	parts := make([]string, 200)
	for i := range parts {
		parts[i] = string(rune('a' + i%26))
	}
	value := "key=" + joinComma(parts)

	_, _, err := Parse(arena, value, cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	v := arena.Values[arena.Params[0].ValueIdx]
	if v.Kind != ValComma {
		t.Fatalf("expected ValComma, got %v", v.Kind)
	}
	if int(v.PartsLen) != 200 {
		t.Fatalf("expected 200 parts, got %d", v.PartsLen)
	}
}

func joinComma(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "," + parts[i]
	}
	return result
}

func TestParse_NewArena_NegativeCapacity(t *testing.T) {
	arena := NewArena(-5)
	if arena == nil {
		t.Fatal("NewArena should not return nil")
	}
	// Should have 0 capacity
	if cap(arena.Params) != 0 {
		t.Errorf("expected 0 capacity, got %d", cap(arena.Params))
	}
}

func TestParse_DotAtKeyEnd(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	// Dot at the end of key (before =)
	_, _, err := Parse(arena, "a.=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Should parse without error
	if len(arena.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(arena.Params))
	}
}

func TestParse_BracketAfterDot(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	// a.[b] - dot followed by bracket
	_, _, err := Parse(arena, "a.[b]=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Implementation specific - just check no crash
	t.Logf("params: %d", len(arena.Params))
}

func TestParse_OnlyDot(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	_, _, err := Parse(arena, ".=value", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Leading dot - implementation specific
	t.Logf("params: %d", len(arena.Params))
}

func TestParse_MultipleDuplicateKeys(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a=1&a=2&a=3&a=4", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// All should be kept (semantic layer handles duplicates)
	if len(arena.Params) != 4 {
		t.Fatalf("expected 4 params, got %d", len(arena.Params))
	}
}

func BenchmarkParse_Simple(b *testing.B) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	input := "a=b&c=d&e=f"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = Parse(arena, input, cfg)
	}
}

func BenchmarkParse_Nested(b *testing.B) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	input := "user[profile][name]=John&user[profile][age]=30&user[settings][theme]=dark"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = Parse(arena, input, cfg)
	}
}

func BenchmarkParse_Complex(b *testing.B) {
	arena := NewArena(32)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots | FlagComma
	input := "a[b].c=x,y,z&d[0]=1&d[1]=2&e.f.g=value&arr[]=a&arr[]=b&arr[]=c"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = Parse(arena, input, cfg)
	}
}

func BenchmarkParse_DotsNoDepthExceeded(b *testing.B) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots
	cfg.Depth = 5
	input := "a.b.c=value&x.y.z=other"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = Parse(arena, input, cfg)
	}
}

func BenchmarkParse_DotsDepthExceeded(b *testing.B) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots
	cfg.Depth = 2
	input := "a.b.c.d.e=value&x.y.z.w=other"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = Parse(arena, input, cfg)
	}
}

func BenchmarkParse_DotsDepthExceeded_NoBracketConversion(b *testing.B) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots | FlagAllowDotsNoBracketConversion
	cfg.Depth = 2
	input := "a.b.c.d.e=value&x.y.z.w=other"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = Parse(arena, input, cfg)
	}
}

func BenchmarkParseBytes_DotsDepthExceeded(b *testing.B) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots
	cfg.Depth = 2
	input := []byte("a.b.c.d.e=value&x.y.z.w=other")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseBytes(arena, input, cfg)
	}
}

func BenchmarkParseBytes_DotsDepthExceeded_NoBracketConversion(b *testing.B) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots | FlagAllowDotsNoBracketConversion
	cfg.Depth = 2
	input := []byte("a.b.c.d.e=value&x.y.z.w=other")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseBytes(arena, input, cfg)
	}
}

// Large complex query string simulating real-world API request
var largeComplexInput = []byte(
	"user[profile][firstName]=John&user[profile][lastName]=Doe&user[profile][email]=john.doe%40example.com&" +
		"user[profile][phone]=%2B1-555-123-4567&user[settings][theme]=dark&user[settings][language]=en-US&" +
		"user[settings][notifications][email]=true&user[settings][notifications][sms]=false&" +
		"user[settings][notifications][push]=true&user[preferences][timezone]=America%2FNew_York&" +
		"filters[status][]=active&filters[status][]=pending&filters[status][]=review&" +
		"filters[category][0]=electronics&filters[category][1]=computers&filters[category][2]=accessories&" +
		"filters[price][min]=100&filters[price][max]=5000&filters[price][currency]=USD&" +
		"filters[date][from]=2024-01-01&filters[date][to]=2024-12-31&" +
		"pagination[page]=1&pagination[limit]=50&pagination[offset]=0&" +
		"sort[field]=createdAt&sort[order]=desc&sort[nulls]=last&" +
		"search[query]=laptop%20computer&search[fields][]=title&search[fields][]=description&search[fields][]=tags&" +
		"meta[requestId]=550e8400-e29b-41d4-a716-446655440000&meta[timestamp]=1702483200&meta[version]=2.1.0&" +
		"flags[includeDeleted]=false&flags[expandRefs]=true&flags[validate]=strict&" +
		"nested.dot.notation.field=value1&another.deeply.nested.path.here=value2&" +
		"arr[]=first&arr[]=second&arr[]=third&arr[]=fourth&arr[]=fifth&" +
		"matrix[0][0]=a&matrix[0][1]=b&matrix[1][0]=c&matrix[1][1]=d&" +
		"empty=&nullish&special[chars]=hello%20world%21%40%23%24%25",
)

func BenchmarkParseBytes_LargeComplex(b *testing.B) {
	arena := NewArena(64)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseBytes(arena, largeComplexInput, cfg)
	}
}

func BenchmarkParseBytes_LargeComplex_DepthLimited(b *testing.B) {
	arena := NewArena(64)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots
	cfg.Depth = 3

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseBytes(arena, largeComplexInput, cfg)
	}
}

func BenchmarkParse_LargeComplex(b *testing.B) {
	arena := NewArena(64)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots
	input := string(largeComplexInput)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = Parse(arena, input, cfg)
	}
}
