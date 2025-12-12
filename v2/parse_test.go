// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import (
	"regexp"
	"testing"
)

func TestDefaultParseOptions(t *testing.T) {
	opts := DefaultParseOptions()

	// Check all default values
	if opts.AllowDots != false {
		t.Errorf("AllowDots = %v, want false", opts.AllowDots)
	}
	if opts.AllowEmptyArrays != false {
		t.Errorf("AllowEmptyArrays = %v, want false", opts.AllowEmptyArrays)
	}
	if opts.AllowPrototypes != false {
		t.Errorf("AllowPrototypes = %v, want false", opts.AllowPrototypes)
	}
	if opts.AllowSparse != false {
		t.Errorf("AllowSparse = %v, want false", opts.AllowSparse)
	}
	if opts.ArrayLimit != 20 {
		t.Errorf("ArrayLimit = %d, want 20", opts.ArrayLimit)
	}
	if opts.Charset != CharsetUTF8 {
		t.Errorf("Charset = %q, want %q", opts.Charset, CharsetUTF8)
	}
	if opts.CharsetSentinel != false {
		t.Errorf("CharsetSentinel = %v, want false", opts.CharsetSentinel)
	}
	if opts.Comma != false {
		t.Errorf("Comma = %v, want false", opts.Comma)
	}
	if opts.DecodeDotInKeys != false {
		t.Errorf("DecodeDotInKeys = %v, want false", opts.DecodeDotInKeys)
	}
	if opts.Decoder != nil {
		t.Errorf("Decoder = %v, want nil", opts.Decoder)
	}
	if opts.Delimiter != "&" {
		t.Errorf("Delimiter = %q, want %q", opts.Delimiter, "&")
	}
	if opts.DelimiterRegexp != nil {
		t.Errorf("DelimiterRegexp = %v, want nil", opts.DelimiterRegexp)
	}
	if opts.Depth != 5 {
		t.Errorf("Depth = %d, want 5", opts.Depth)
	}
	if opts.Duplicates != DuplicateCombine {
		t.Errorf("Duplicates = %q, want %q", opts.Duplicates, DuplicateCombine)
	}
	if opts.IgnoreQueryPrefix != false {
		t.Errorf("IgnoreQueryPrefix = %v, want false", opts.IgnoreQueryPrefix)
	}
	if opts.InterpretNumericEntities != false {
		t.Errorf("InterpretNumericEntities = %v, want false", opts.InterpretNumericEntities)
	}
	if opts.ParameterLimit != 1000 {
		t.Errorf("ParameterLimit = %d, want 1000", opts.ParameterLimit)
	}
	if opts.ParseArrays != true {
		t.Errorf("ParseArrays = %v, want true", opts.ParseArrays)
	}
	if opts.PlainObjects != false {
		t.Errorf("PlainObjects = %v, want false", opts.PlainObjects)
	}
	if opts.StrictDepth != false {
		t.Errorf("StrictDepth = %v, want false", opts.StrictDepth)
	}
	if opts.StrictNullHandling != false {
		t.Errorf("StrictNullHandling = %v, want false", opts.StrictNullHandling)
	}
	if opts.ThrowOnLimitExceeded != false {
		t.Errorf("ThrowOnLimitExceeded = %v, want false", opts.ThrowOnLimitExceeded)
	}
}

func TestNormalizeParseOptions(t *testing.T) {
	t.Run("nil options returns defaults", func(t *testing.T) {
		opts, err := normalizeParseOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.ArrayLimit != DefaultArrayLimit {
			t.Errorf("ArrayLimit = %d, want %d", opts.ArrayLimit, DefaultArrayLimit)
		}
		if opts.Depth != DefaultDepth {
			t.Errorf("Depth = %d, want %d", opts.Depth, DefaultDepth)
		}
		if opts.ParameterLimit != DefaultParameterLimit {
			t.Errorf("ParameterLimit = %d, want %d", opts.ParameterLimit, DefaultParameterLimit)
		}
	})

	t.Run("empty charset defaults to UTF-8", func(t *testing.T) {
		opts, err := normalizeParseOptions(&ParseOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Charset != CharsetUTF8 {
			t.Errorf("Charset = %q, want %q", opts.Charset, CharsetUTF8)
		}
	})

	t.Run("invalid charset returns error", func(t *testing.T) {
		_, err := normalizeParseOptions(&ParseOptions{Charset: "invalid"})
		if err != ErrInvalidCharset {
			t.Errorf("err = %v, want ErrInvalidCharset", err)
		}
	})

	t.Run("valid charsets are accepted", func(t *testing.T) {
		for _, charset := range []Charset{CharsetUTF8, CharsetISO88591} {
			opts, err := normalizeParseOptions(&ParseOptions{Charset: charset})
			if err != nil {
				t.Errorf("unexpected error for charset %q: %v", charset, err)
			}
			if opts.Charset != charset {
				t.Errorf("Charset = %q, want %q", opts.Charset, charset)
			}
		}
	})

	t.Run("empty duplicates defaults to combine", func(t *testing.T) {
		opts, err := normalizeParseOptions(&ParseOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Duplicates != DuplicateCombine {
			t.Errorf("Duplicates = %q, want %q", opts.Duplicates, DuplicateCombine)
		}
	})

	t.Run("invalid duplicates returns error", func(t *testing.T) {
		_, err := normalizeParseOptions(&ParseOptions{Duplicates: "invalid"})
		if err != ErrInvalidDuplicates {
			t.Errorf("err = %v, want ErrInvalidDuplicates", err)
		}
	})

	t.Run("valid duplicates are accepted", func(t *testing.T) {
		for _, dup := range []DuplicateHandling{DuplicateCombine, DuplicateFirst, DuplicateLast} {
			opts, err := normalizeParseOptions(&ParseOptions{Duplicates: dup})
			if err != nil {
				t.Errorf("unexpected error for duplicates %q: %v", dup, err)
			}
			if opts.Duplicates != dup {
				t.Errorf("Duplicates = %q, want %q", opts.Duplicates, dup)
			}
		}
	})

	t.Run("zero numeric values get defaults", func(t *testing.T) {
		opts, err := normalizeParseOptions(&ParseOptions{
			ArrayLimit:     0,
			Depth:          0,
			ParameterLimit: 0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.ArrayLimit != DefaultArrayLimit {
			t.Errorf("ArrayLimit = %d, want %d", opts.ArrayLimit, DefaultArrayLimit)
		}
		if opts.Depth != DefaultDepth {
			t.Errorf("Depth = %d, want %d", opts.Depth, DefaultDepth)
		}
		if opts.ParameterLimit != DefaultParameterLimit {
			t.Errorf("ParameterLimit = %d, want %d", opts.ParameterLimit, DefaultParameterLimit)
		}
	})

	t.Run("custom numeric values are preserved", func(t *testing.T) {
		opts, err := normalizeParseOptions(&ParseOptions{
			ArrayLimit:     50,
			Depth:          10,
			ParameterLimit: 500,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.ArrayLimit != 50 {
			t.Errorf("ArrayLimit = %d, want 50", opts.ArrayLimit)
		}
		if opts.Depth != 10 {
			t.Errorf("Depth = %d, want 10", opts.Depth)
		}
		if opts.ParameterLimit != 500 {
			t.Errorf("ParameterLimit = %d, want 500", opts.ParameterLimit)
		}
	})

	t.Run("empty delimiter defaults to &", func(t *testing.T) {
		opts, err := normalizeParseOptions(&ParseOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Delimiter != DefaultDelimiter {
			t.Errorf("Delimiter = %q, want %q", opts.Delimiter, DefaultDelimiter)
		}
	})

	t.Run("custom delimiter is preserved", func(t *testing.T) {
		opts, err := normalizeParseOptions(&ParseOptions{Delimiter: ";"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Delimiter != ";" {
			t.Errorf("Delimiter = %q, want %q", opts.Delimiter, ";")
		}
	})

	t.Run("delimiter regexp takes precedence", func(t *testing.T) {
		re := regexp.MustCompile("[&;]")
		opts, err := normalizeParseOptions(&ParseOptions{DelimiterRegexp: re})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.DelimiterRegexp != re {
			t.Error("DelimiterRegexp was not preserved")
		}
		// Delimiter should remain empty when regexp is set
		if opts.Delimiter != "" {
			t.Errorf("Delimiter = %q, want empty when regexp is set", opts.Delimiter)
		}
	})

	t.Run("decodeDotInKeys enables allowDots", func(t *testing.T) {
		opts, err := normalizeParseOptions(&ParseOptions{
			DecodeDotInKeys: true,
			AllowDots:       false,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !opts.AllowDots {
			t.Error("AllowDots should be true when DecodeDotInKeys is true")
		}
	})
}

func TestParseOptionsWithMethods(t *testing.T) {
	base := DefaultParseOptions()

	t.Run("WithAllowDots", func(t *testing.T) {
		opts := base.WithAllowDots(true)
		if !opts.AllowDots {
			t.Error("WithAllowDots(true) should set AllowDots to true")
		}
		if base.AllowDots {
			t.Error("original options should not be modified")
		}
	})

	t.Run("WithArrayLimit", func(t *testing.T) {
		opts := base.WithArrayLimit(100)
		if opts.ArrayLimit != 100 {
			t.Errorf("WithArrayLimit(100) = %d, want 100", opts.ArrayLimit)
		}
	})

	t.Run("WithCharset", func(t *testing.T) {
		opts := base.WithCharset(CharsetISO88591)
		if opts.Charset != CharsetISO88591 {
			t.Errorf("WithCharset(ISO88591) = %q, want %q", opts.Charset, CharsetISO88591)
		}
	})

	t.Run("WithComma", func(t *testing.T) {
		opts := base.WithComma(true)
		if !opts.Comma {
			t.Error("WithComma(true) should set Comma to true")
		}
	})

	t.Run("WithDelimiter", func(t *testing.T) {
		opts := base.WithDelimiter(";")
		if opts.Delimiter != ";" {
			t.Errorf("WithDelimiter(;) = %q, want ;", opts.Delimiter)
		}
	})

	t.Run("WithDepth", func(t *testing.T) {
		opts := base.WithDepth(10)
		if opts.Depth != 10 {
			t.Errorf("WithDepth(10) = %d, want 10", opts.Depth)
		}
	})

	t.Run("WithDuplicates", func(t *testing.T) {
		opts := base.WithDuplicates(DuplicateLast)
		if opts.Duplicates != DuplicateLast {
			t.Errorf("WithDuplicates(last) = %q, want %q", opts.Duplicates, DuplicateLast)
		}
	})

	t.Run("WithIgnoreQueryPrefix", func(t *testing.T) {
		opts := base.WithIgnoreQueryPrefix(true)
		if !opts.IgnoreQueryPrefix {
			t.Error("WithIgnoreQueryPrefix(true) should set IgnoreQueryPrefix to true")
		}
	})

	t.Run("WithParameterLimit", func(t *testing.T) {
		opts := base.WithParameterLimit(500)
		if opts.ParameterLimit != 500 {
			t.Errorf("WithParameterLimit(500) = %d, want 500", opts.ParameterLimit)
		}
	})

	t.Run("WithParseArrays", func(t *testing.T) {
		opts := base.WithParseArrays(false)
		if opts.ParseArrays {
			t.Error("WithParseArrays(false) should set ParseArrays to false")
		}
	})

	t.Run("WithStrictNullHandling", func(t *testing.T) {
		opts := base.WithStrictNullHandling(true)
		if !opts.StrictNullHandling {
			t.Error("WithStrictNullHandling(true) should set StrictNullHandling to true")
		}
	})

	t.Run("chaining", func(t *testing.T) {
		opts := base.
			WithAllowDots(true).
			WithDepth(10).
			WithComma(true).
			WithDelimiter(";")

		if !opts.AllowDots || opts.Depth != 10 || !opts.Comma || opts.Delimiter != ";" {
			t.Error("chained With methods should work correctly")
		}
	})
}

func TestDuplicateHandlingConstants(t *testing.T) {
	if DuplicateCombine != "combine" {
		t.Errorf("DuplicateCombine = %q, want %q", DuplicateCombine, "combine")
	}
	if DuplicateFirst != "first" {
		t.Errorf("DuplicateFirst = %q, want %q", DuplicateFirst, "first")
	}
	if DuplicateLast != "last" {
		t.Errorf("DuplicateLast = %q, want %q", DuplicateLast, "last")
	}
}

func TestParseErrors(t *testing.T) {
	// Test that error variables are properly defined
	errors := []error{
		ErrInvalidAllowEmptyArrays,
		ErrInvalidDecodeDotInKeys,
		ErrInvalidDecoder,
		ErrInvalidCharset,
		ErrInvalidDuplicates,
		ErrInvalidThrowOnLimit,
		ErrParameterLimitExceeded,
		ErrArrayLimitExceeded,
		ErrDepthLimitExceeded,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("error variable should not be nil")
		}
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	}
}
