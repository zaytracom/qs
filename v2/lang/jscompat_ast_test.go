package lang

import "testing"

func TestAST_EqualsInsideBrackets_SplitsOnCloseBracketEquals(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()

	qs, _, err := Parse(arena, "a[>=]=23&a[<=>]==23&a[==]=23", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if qs.ParamLen != 3 {
		t.Fatalf("params: %d", qs.ParamLen)
	}

	// a[>=]=23
	{
		p := arena.Params[0]
		if !p.HasEquals {
			t.Fatal("p0: expected HasEquals=true")
		}
		if got := arena.GetString(p.Key.Raw); got != "a[>=]" {
			t.Fatalf("p0 key raw: %q", got)
		}
		v := arena.Values[p.ValueIdx]
		if got := arena.GetString(v.Raw); got != "23" {
			t.Fatalf("p0 val raw: %q", got)
		}
	}

	// a[<=>]==23 -> value begins with "="
	{
		p := arena.Params[1]
		if got := arena.GetString(p.Key.Raw); got != "a[<=>]" {
			t.Fatalf("p1 key raw: %q", got)
		}
		v := arena.Values[p.ValueIdx]
		if got := arena.GetString(v.Raw); got != "=23" {
			t.Fatalf("p1 val raw: %q", got)
		}
	}
}

func TestAST_ParameterLimit_Truncates(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ParameterLimit = 2

	qs, _, err := Parse(arena, "a=b&c=d&e=f", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if qs.ParamLen != 2 {
		t.Fatalf("ParamLen: got %d", qs.ParamLen)
	}
	if got := arena.GetString(arena.Params[0].Key.Raw); got != "a" {
		t.Fatalf("p0 key: %q", got)
	}
	if got := arena.GetString(arena.Params[1].Key.Raw); got != "c" {
		t.Fatalf("p1 key: %q", got)
	}
}

func TestAST_ParameterLimit_ThrowOnExceeded(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.ParameterLimit = 2
	cfg.Flags |= FlagThrowOnLimitExceeded

	_, _, err := Parse(arena, "a=b&c=d&e=f", cfg)
	if err != ErrParameterLimitExceeded {
		t.Fatalf("err: got %v", err)
	}
}

func TestAST_CustomDelimiter_Semicolon(t *testing.T) {
	arena := NewArena(8)
	cfg := DefaultConfig()
	cfg.Delimiter = ';'

	qs, _, err := Parse(arena, "a=1;b=2;c=3", cfg)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if qs.ParamLen != 3 {
		t.Fatalf("ParamLen: %d", qs.ParamLen)
	}
	if got := arena.GetString(arena.Params[2].Key.Raw); got != "c" {
		t.Fatalf("p2 key: %q", got)
	}
}
