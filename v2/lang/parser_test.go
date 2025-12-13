package lang

import "testing"

func TestParse_Basic(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	qs, cs, err := Parse(arena, "a=b&c=d", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cs != CharsetUTF8 {
		t.Fatalf("charset: got %v", cs)
	}
	if qs.ParamLen != 2 {
		t.Fatalf("params: got %d", qs.ParamLen)
	}

	p0 := arena.Params[0]
	if !p0.HasEquals {
		t.Fatalf("p0 HasEquals: false")
	}
	if got := arena.GetString(p0.Key.Raw); got != "a" {
		t.Fatalf("p0 key raw: %q", got)
	}
	if p0.Key.SegLen != 1 {
		t.Fatalf("p0 seg len: %d", p0.Key.SegLen)
	}
	seg0 := arena.Segments[p0.Key.SegStart]
	if seg0.Kind != SegIdent || seg0.Notation != NotationRoot || seg0.Index != -1 {
		t.Fatalf("p0 seg0: %+v", seg0)
	}
	if got := arena.GetString(seg0.Span); got != "a" {
		t.Fatalf("p0 seg0 text: %q", got)
	}
	if p0.ValueIdx == noValue {
		t.Fatalf("p0 missing value")
	}
	v0 := arena.Values[p0.ValueIdx]
	if v0.Kind != ValSimple {
		t.Fatalf("p0 value kind: %v", v0.Kind)
	}
	if got := arena.GetString(v0.Raw); got != "b" {
		t.Fatalf("p0 value: %q", got)
	}
}

func TestParse_BracketsAndIndex(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[b][0]=x", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if p.Key.SegLen != 3 {
		t.Fatalf("seg len: %d", p.Key.SegLen)
	}
	s0 := arena.Segments[p.Key.SegStart]
	s1 := arena.Segments[p.Key.SegStart+1]
	s2 := arena.Segments[p.Key.SegStart+2]
	if arena.GetString(s0.Span) != "a" || s0.Notation != NotationRoot || s0.Kind != SegIdent {
		t.Fatalf("s0: %+v", s0)
	}
	if arena.GetString(s1.Span) != "b" || s1.Notation != NotationBracket || s1.Kind != SegIdent {
		t.Fatalf("s1: %+v", s1)
	}
	if arena.GetString(s2.Span) != "0" || s2.Notation != NotationBracket || s2.Kind != SegIndex || s2.Index != 0 {
		t.Fatalf("s2: %+v", s2)
	}
}

func TestParse_EqualsInsideBrackets_KeyValueSplit(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "a[>=]=23&a[<=>]==23&a[==]=23", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(arena.Params) != 3 {
		t.Fatalf("params: %d", len(arena.Params))
	}

	// a[>=]=23 -> segments: a, >= ; value: 23
	{
		p := arena.Params[0]
		if arena.GetString(p.Key.Raw) != "a[>=]" {
			t.Fatalf("p0 key: %q", arena.GetString(p.Key.Raw))
		}
		if p.Key.SegLen != 2 {
			t.Fatalf("p0 seglen: %d", p.Key.SegLen)
		}
		s1 := arena.Segments[p.Key.SegStart+1]
		if arena.GetString(s1.Span) != ">=" {
			t.Fatalf("p0 seg1: %q", arena.GetString(s1.Span))
		}
		v := arena.Values[p.ValueIdx]
		if arena.GetString(v.Raw) != "23" {
			t.Fatalf("p0 val: %q", arena.GetString(v.Raw))
		}
	}

	// a[<=>]==23 -> value should include leading "=" (i.e. "=23")
	{
		p := arena.Params[1]
		v := arena.Values[p.ValueIdx]
		if arena.GetString(v.Raw) != "=23" {
			t.Fatalf("p1 val: %q", arena.GetString(v.Raw))
		}
	}
}

func TestParse_AllowDots(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	_, _, err := Parse(arena, "a.b=c", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if p.Key.SegLen != 2 {
		t.Fatalf("seg len: %d", p.Key.SegLen)
	}
	s0 := arena.Segments[p.Key.SegStart]
	s1 := arena.Segments[p.Key.SegStart+1]
	if arena.GetString(s0.Span) != "a" || s0.Notation != NotationRoot {
		t.Fatalf("s0: %+v", s0)
	}
	if arena.GetString(s1.Span) != "b" || s1.Notation != NotationDot {
		t.Fatalf("s1: %+v", s1)
	}
}

func TestParse_DotAfterBracket(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots

	_, _, err := Parse(arena, "a[b].c=d", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if p.Key.SegLen != 3 {
		t.Fatalf("seg len: %d", p.Key.SegLen)
	}
	s0 := arena.Segments[p.Key.SegStart]
	s1 := arena.Segments[p.Key.SegStart+1]
	s2 := arena.Segments[p.Key.SegStart+2]
	if arena.GetString(s0.Span) != "a" || s0.Notation != NotationRoot {
		t.Fatalf("s0: %+v", s0)
	}
	if arena.GetString(s1.Span) != "b" || s1.Notation != NotationBracket {
		t.Fatalf("s1: %+v", s1)
	}
	if arena.GetString(s2.Span) != "c" || s2.Notation != NotationDot {
		t.Fatalf("s2: %+v", s2)
	}
}

func TestParse_DepthLimit_LiteralRemainder(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots
	cfg.Depth = 1

	_, _, err := Parse(arena, "a.b.c=d", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := arena.Params[0]
	if p.Key.SegLen != 3 {
		t.Fatalf("seg len: %d", p.Key.SegLen)
	}
	s2 := arena.Segments[p.Key.SegStart+2]
	if s2.Kind != SegLiteral {
		t.Fatalf("literal kind: %+v", s2)
	}
	if got := arena.GetString(s2.Span); got != ".c" {
		t.Fatalf("literal text: %q", got)
	}
}

func TestParse_StrictDepthErrors(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagAllowDots | FlagStrictDepth
	cfg.Depth = 1

	_, _, err := Parse(arena, "a.b.c=d", cfg)
	if err != ErrDepthLimitExceeded {
		t.Fatalf("err: got %v", err)
	}
}

func TestParse_CommaValues(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagComma

	_, _, err := Parse(arena, "a=x,y&b=z", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p0 := arena.Params[0]
	v0 := arena.Values[p0.ValueIdx]
	if v0.Kind != ValComma || v0.PartsLen != 2 {
		t.Fatalf("v0: %+v", v0)
	}
	if got := arena.GetString(arena.ValueParts[v0.PartsOff+0]); got != "x" {
		t.Fatalf("part0: %q", got)
	}
	if got := arena.GetString(arena.ValueParts[v0.PartsOff+1]); got != "y" {
		t.Fatalf("part1: %q", got)
	}
}

func TestParse_CharsetSentinel_SkipsParam(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Flags |= FlagCharsetSentinel
	cfg.Charset = CharsetISO88591

	_, cs, err := Parse(arena, "utf8=%E2%9C%93&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cs != CharsetUTF8 {
		t.Fatalf("charset: got %v", cs)
	}
	if len(arena.Params) != 1 {
		t.Fatalf("params: %d", len(arena.Params))
	}
	if got := arena.GetString(arena.Params[0].Key.Raw); got != "a" {
		t.Fatalf("key: %q", got)
	}
}

func TestParse_ProtoKeyNormal(t *testing.T) {
	// In Go, __proto__ is just a normal key (no prototype pollution)
	arena := NewArena(8)
	cfg := DefaultConfig()

	_, _, err := Parse(arena, "__proto__[x]=y&a=b", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// Both params should be parsed - __proto__ is a normal key
	if len(arena.Params) != 2 {
		t.Fatalf("params: %d", len(arena.Params))
	}
}

func TestParse_NoAllocs(t *testing.T) {
	arena := NewArena(32)
	cfg := DefaultConfig()

	allocs := testing.AllocsPerRun(1000, func() {
		_, _, err := Parse(arena, "a[b]=1&c=d&e=f&g[h][0]=x", cfg)
		if err != nil {
			panic(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("allocs: got %v", allocs)
	}
}
