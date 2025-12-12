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

func TestFunctionalOptions(t *testing.T) {
	t.Run("WithAllowDots", func(t *testing.T) {
		opts := applyParseOptions(WithAllowDots(true))
		if !opts.AllowDots {
			t.Error("WithAllowDots(true) should set AllowDots to true")
		}
	})

	t.Run("WithAllowEmptyArrays", func(t *testing.T) {
		opts := applyParseOptions(WithAllowEmptyArrays(true))
		if !opts.AllowEmptyArrays {
			t.Error("WithAllowEmptyArrays(true) should set AllowEmptyArrays to true")
		}
	})

	t.Run("WithAllowPrototypes", func(t *testing.T) {
		opts := applyParseOptions(WithAllowPrototypes(true))
		if !opts.AllowPrototypes {
			t.Error("WithAllowPrototypes(true) should set AllowPrototypes to true")
		}
	})

	t.Run("WithAllowSparse", func(t *testing.T) {
		opts := applyParseOptions(WithAllowSparse(true))
		if !opts.AllowSparse {
			t.Error("WithAllowSparse(true) should set AllowSparse to true")
		}
	})

	t.Run("WithArrayLimit", func(t *testing.T) {
		opts := applyParseOptions(WithArrayLimit(100))
		if opts.ArrayLimit != 100 {
			t.Errorf("WithArrayLimit(100) = %d, want 100", opts.ArrayLimit)
		}
	})

	t.Run("WithArrayLimit zero", func(t *testing.T) {
		opts := applyParseOptions(WithArrayLimit(0))
		if opts.ArrayLimit != 0 {
			t.Errorf("WithArrayLimit(0) = %d, want 0", opts.ArrayLimit)
		}
	})

	t.Run("WithCharset", func(t *testing.T) {
		opts := applyParseOptions(WithCharset(CharsetISO88591))
		if opts.Charset != CharsetISO88591 {
			t.Errorf("WithCharset(ISO88591) = %q, want %q", opts.Charset, CharsetISO88591)
		}
	})

	t.Run("WithCharsetSentinel", func(t *testing.T) {
		opts := applyParseOptions(WithCharsetSentinel(true))
		if !opts.CharsetSentinel {
			t.Error("WithCharsetSentinel(true) should set CharsetSentinel to true")
		}
	})

	t.Run("WithComma", func(t *testing.T) {
		opts := applyParseOptions(WithComma(true))
		if !opts.Comma {
			t.Error("WithComma(true) should set Comma to true")
		}
	})

	t.Run("WithDecodeDotInKeys enables AllowDots", func(t *testing.T) {
		opts := applyParseOptions(WithDecodeDotInKeys(true))
		if !opts.DecodeDotInKeys {
			t.Error("WithDecodeDotInKeys(true) should set DecodeDotInKeys to true")
		}
		if !opts.AllowDots {
			t.Error("WithDecodeDotInKeys(true) should also enable AllowDots")
		}
	})

	t.Run("WithDecoder", func(t *testing.T) {
		customDecoder := func(s string, c Charset, k string) (string, error) {
			return s, nil
		}
		opts := applyParseOptions(WithDecoder(customDecoder))
		if opts.Decoder == nil {
			t.Error("WithDecoder should set Decoder")
		}
	})

	t.Run("WithDelimiter", func(t *testing.T) {
		opts := applyParseOptions(WithDelimiter(";"))
		if opts.Delimiter != ";" {
			t.Errorf("WithDelimiter(;) = %q, want ;", opts.Delimiter)
		}
	})

	t.Run("WithDelimiterRegexp clears Delimiter", func(t *testing.T) {
		re := regexp.MustCompile("[&;]")
		opts := applyParseOptions(WithDelimiterRegexp(re))
		if opts.DelimiterRegexp != re {
			t.Error("WithDelimiterRegexp should set DelimiterRegexp")
		}
		if opts.Delimiter != "" {
			t.Error("WithDelimiterRegexp should clear Delimiter")
		}
	})

	t.Run("WithDelimiter clears DelimiterRegexp", func(t *testing.T) {
		re := regexp.MustCompile("[&;]")
		opts := applyParseOptions(WithDelimiterRegexp(re), WithDelimiter(";"))
		if opts.DelimiterRegexp != nil {
			t.Error("WithDelimiter should clear DelimiterRegexp")
		}
		if opts.Delimiter != ";" {
			t.Errorf("Delimiter = %q, want ;", opts.Delimiter)
		}
	})

	t.Run("WithDepth", func(t *testing.T) {
		opts := applyParseOptions(WithDepth(10))
		if opts.Depth != 10 {
			t.Errorf("WithDepth(10) = %d, want 10", opts.Depth)
		}
	})

	t.Run("WithDepth zero", func(t *testing.T) {
		opts := applyParseOptions(WithDepth(0))
		if opts.Depth != 0 {
			t.Errorf("WithDepth(0) = %d, want 0", opts.Depth)
		}
	})

	t.Run("WithDuplicates", func(t *testing.T) {
		opts := applyParseOptions(WithDuplicates(DuplicateLast))
		if opts.Duplicates != DuplicateLast {
			t.Errorf("WithDuplicates(last) = %q, want %q", opts.Duplicates, DuplicateLast)
		}
	})

	t.Run("WithIgnoreQueryPrefix", func(t *testing.T) {
		opts := applyParseOptions(WithIgnoreQueryPrefix(true))
		if !opts.IgnoreQueryPrefix {
			t.Error("WithIgnoreQueryPrefix(true) should set IgnoreQueryPrefix to true")
		}
	})

	t.Run("WithInterpretNumericEntities", func(t *testing.T) {
		opts := applyParseOptions(WithInterpretNumericEntities(true))
		if !opts.InterpretNumericEntities {
			t.Error("WithInterpretNumericEntities(true) should set InterpretNumericEntities to true")
		}
	})

	t.Run("WithParameterLimit", func(t *testing.T) {
		opts := applyParseOptions(WithParameterLimit(500))
		if opts.ParameterLimit != 500 {
			t.Errorf("WithParameterLimit(500) = %d, want 500", opts.ParameterLimit)
		}
	})

	t.Run("WithParseArrays false", func(t *testing.T) {
		opts := applyParseOptions(WithParseArrays(false))
		if opts.ParseArrays {
			t.Error("WithParseArrays(false) should set ParseArrays to false")
		}
	})

	t.Run("WithPlainObjects", func(t *testing.T) {
		opts := applyParseOptions(WithPlainObjects(true))
		if !opts.PlainObjects {
			t.Error("WithPlainObjects(true) should set PlainObjects to true")
		}
	})

	t.Run("WithStrictDepth", func(t *testing.T) {
		opts := applyParseOptions(WithStrictDepth(true))
		if !opts.StrictDepth {
			t.Error("WithStrictDepth(true) should set StrictDepth to true")
		}
	})

	t.Run("WithStrictNullHandling", func(t *testing.T) {
		opts := applyParseOptions(WithStrictNullHandling(true))
		if !opts.StrictNullHandling {
			t.Error("WithStrictNullHandling(true) should set StrictNullHandling to true")
		}
	})

	t.Run("WithThrowOnLimitExceeded", func(t *testing.T) {
		opts := applyParseOptions(WithThrowOnLimitExceeded(true))
		if !opts.ThrowOnLimitExceeded {
			t.Error("WithThrowOnLimitExceeded(true) should set ThrowOnLimitExceeded to true")
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		opts := applyParseOptions(
			WithAllowDots(true),
			WithDepth(10),
			WithComma(true),
			WithDelimiter(";"),
		)
		if !opts.AllowDots {
			t.Error("AllowDots should be true")
		}
		if opts.Depth != 10 {
			t.Errorf("Depth = %d, want 10", opts.Depth)
		}
		if !opts.Comma {
			t.Error("Comma should be true")
		}
		if opts.Delimiter != ";" {
			t.Errorf("Delimiter = %q, want ;", opts.Delimiter)
		}
	})

	t.Run("no options returns defaults", func(t *testing.T) {
		opts := applyParseOptions()
		defaults := DefaultParseOptions()
		if opts.ArrayLimit != defaults.ArrayLimit {
			t.Errorf("ArrayLimit = %d, want %d", opts.ArrayLimit, defaults.ArrayLimit)
		}
		if opts.Depth != defaults.Depth {
			t.Errorf("Depth = %d, want %d", opts.Depth, defaults.Depth)
		}
		if opts.ParseArrays != defaults.ParseArrays {
			t.Errorf("ParseArrays = %v, want %v", opts.ParseArrays, defaults.ParseArrays)
		}
	})

	t.Run("defaults are correct with functional options", func(t *testing.T) {
		// This is the key test - verifies that functional options solve the zero-value problem
		opts := applyParseOptions() // no options = all defaults

		if opts.ParseArrays != true {
			t.Error("ParseArrays default should be true")
		}
		if opts.ArrayLimit != 20 {
			t.Errorf("ArrayLimit default should be 20, got %d", opts.ArrayLimit)
		}
		if opts.Depth != 5 {
			t.Errorf("Depth default should be 5, got %d", opts.Depth)
		}
		if opts.ParameterLimit != 1000 {
			t.Errorf("ParameterLimit default should be 1000, got %d", opts.ParameterLimit)
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

// ============================================
// Tests for Core Parse Function (Phase 2.2)
// ============================================

func TestParseBasic(t *testing.T) {
	t.Run("parses simple key-value", func(t *testing.T) {
		result, err := Parse("a=b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "b" {
			t.Errorf("result[a] = %v, want 'b'", result["a"])
		}
	})

	t.Run("parses multiple params", func(t *testing.T) {
		result, err := Parse("a=b&c=d")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "b" {
			t.Errorf("result[a] = %v, want 'b'", result["a"])
		}
		if result["c"] != "d" {
			t.Errorf("result[c] = %v, want 'd'", result["c"])
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		result, err := Parse("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("result should be empty, got %v", result)
		}
	})

	t.Run("handles empty value", func(t *testing.T) {
		result, err := Parse("a=")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "" {
			t.Errorf("result[a] = %v, want ''", result["a"])
		}
	})

	t.Run("handles key without value", func(t *testing.T) {
		result, err := Parse("a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "" {
			t.Errorf("result[a] = %v, want ''", result["a"])
		}
	})

	t.Run("handles key without value with strictNullHandling", func(t *testing.T) {
		result, err := Parse("a", WithStrictNullHandling(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != nil {
			t.Errorf("result[a] = %v, want nil", result["a"])
		}
	})
}

func TestParseQueryPrefix(t *testing.T) {
	t.Run("does not strip ? by default", func(t *testing.T) {
		result, err := Parse("?a=b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := result["?a"]; !ok {
			t.Errorf("result should have key '?a', got %v", result)
		}
	})

	t.Run("strips ? with IgnoreQueryPrefix", func(t *testing.T) {
		result, err := Parse("?a=b", WithIgnoreQueryPrefix(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "b" {
			t.Errorf("result[a] = %v, want 'b'", result["a"])
		}
	})
}

func TestParseDelimiter(t *testing.T) {
	t.Run("uses & as default delimiter", func(t *testing.T) {
		result, err := Parse("a=b&c=d")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 keys, got %d", len(result))
		}
	})

	t.Run("uses custom delimiter", func(t *testing.T) {
		result, err := Parse("a=b;c=d", WithDelimiter(";"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "b" {
			t.Errorf("result[a] = %v, want 'b'", result["a"])
		}
		if result["c"] != "d" {
			t.Errorf("result[c] = %v, want 'd'", result["c"])
		}
	})

	t.Run("uses regexp delimiter", func(t *testing.T) {
		re := regexp.MustCompile("[&;]")
		result, err := Parse("a=b&c=d;e=f", WithDelimiterRegexp(re))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("expected 3 keys, got %d", len(result))
		}
		if result["a"] != "b" {
			t.Errorf("result[a] = %v, want 'b'", result["a"])
		}
		if result["c"] != "d" {
			t.Errorf("result[c] = %v, want 'd'", result["c"])
		}
		if result["e"] != "f" {
			t.Errorf("result[e] = %v, want 'f'", result["e"])
		}
	})
}

func TestParseParameterLimit(t *testing.T) {
	t.Run("respects parameter limit", func(t *testing.T) {
		result, err := Parse("a=1&b=2&c=3&d=4&e=5", WithParameterLimit(3))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("expected 3 keys, got %d", len(result))
		}
	})

	t.Run("returns error when limit exceeded with throwOnLimitExceeded", func(t *testing.T) {
		_, err := Parse("a=1&b=2&c=3", WithParameterLimit(2), WithThrowOnLimitExceeded(true))
		if err != ErrParameterLimitExceeded {
			t.Errorf("expected ErrParameterLimitExceeded, got %v", err)
		}
	})

	t.Run("does not error when at limit with throwOnLimitExceeded", func(t *testing.T) {
		result, err := Parse("a=1&b=2", WithParameterLimit(2), WithThrowOnLimitExceeded(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 keys, got %d", len(result))
		}
	})
}

func TestParseDuplicates(t *testing.T) {
	t.Run("combines duplicates by default", func(t *testing.T) {
		result, err := Parse("a=1&a=2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result[a] should be array, got %T", result["a"])
		}
		if len(arr) != 2 || arr[0] != "1" || arr[1] != "2" {
			t.Errorf("result[a] = %v, want [1, 2]", arr)
		}
	})

	t.Run("keeps first with DuplicateFirst", func(t *testing.T) {
		result, err := Parse("a=1&a=2", WithDuplicates(DuplicateFirst))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "1" {
			t.Errorf("result[a] = %v, want '1'", result["a"])
		}
	})

	t.Run("keeps last with DuplicateLast", func(t *testing.T) {
		result, err := Parse("a=1&a=2", WithDuplicates(DuplicateLast))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "2" {
			t.Errorf("result[a] = %v, want '2'", result["a"])
		}
	})
}

func TestParseComma(t *testing.T) {
	t.Run("does not split by comma by default", func(t *testing.T) {
		result, err := Parse("a=1,2,3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "1,2,3" {
			t.Errorf("result[a] = %v, want '1,2,3'", result["a"])
		}
	})

	t.Run("splits by comma with Comma option", func(t *testing.T) {
		result, err := Parse("a=1,2,3", WithComma(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		arr, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result[a] should be array, got %T", result["a"])
		}
		if len(arr) != 3 {
			t.Errorf("expected 3 elements, got %d", len(arr))
		}
		if arr[0] != "1" || arr[1] != "2" || arr[2] != "3" {
			t.Errorf("result[a] = %v, want [1, 2, 3]", arr)
		}
	})
}

func TestParseURLEncoding(t *testing.T) {
	t.Run("decodes URL-encoded values", func(t *testing.T) {
		result, err := Parse("a=%20b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != " b" {
			t.Errorf("result[a] = %q, want ' b'", result["a"])
		}
	})

	t.Run("decodes + as space", func(t *testing.T) {
		result, err := Parse("a=hello+world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "hello world" {
			t.Errorf("result[a] = %q, want 'hello world'", result["a"])
		}
	})

	t.Run("decodes URL-encoded keys", func(t *testing.T) {
		result, err := Parse("a%20b=c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a b"] != "c" {
			t.Errorf("result['a b'] = %v, want 'c'", result["a b"])
		}
	})

	t.Run("decodes URL-encoded brackets and parses nested", func(t *testing.T) {
		result, err := Parse("a%5Bb%5D=c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// URL-encoded brackets are decoded and parsed as nested objects
		a, ok := result["a"].(map[string]any)
		if !ok {
			t.Fatalf("result['a'] should be map, got %T: %v", result["a"], result)
		}
		if a["b"] != "c" {
			t.Errorf("result['a']['b'] = %v, want 'c'", a["b"])
		}
	})
}

func TestParseCharsetSentinel(t *testing.T) {
	t.Run("detects UTF-8 charset sentinel", func(t *testing.T) {
		result, err := Parse("utf8=%E2%9C%93&a=b", WithCharsetSentinel(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The sentinel itself should be skipped
		if _, ok := result["utf8"]; ok {
			t.Error("utf8 sentinel should not be in result")
		}
		if result["a"] != "b" {
			t.Errorf("result[a] = %v, want 'b'", result["a"])
		}
	})

	t.Run("detects ISO-8859-1 charset sentinel", func(t *testing.T) {
		result, err := Parse("utf8=%26%2310003%3B&a=b", WithCharsetSentinel(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The sentinel itself should be skipped
		if _, ok := result["utf8"]; ok {
			t.Error("utf8 sentinel should not be in result")
		}
		if result["a"] != "b" {
			t.Errorf("result[a] = %v, want 'b'", result["a"])
		}
	})
}

func TestParseInterpretNumericEntities(t *testing.T) {
	t.Run("interprets numeric entities with ISO-8859-1", func(t *testing.T) {
		// The & must be URL-encoded as %26 to not be treated as delimiter
		result, err := Parse("a=%26%239786%3B",
			WithCharset(CharsetISO88591),
			WithInterpretNumericEntities(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "☺" {
			t.Errorf("result[a] = %q, want '☺'", result["a"])
		}
	})

	t.Run("does not interpret entities without option", func(t *testing.T) {
		// The & must be URL-encoded as %26 to not be treated as delimiter
		result, err := Parse("a=%26%239786%3B", WithCharset(CharsetISO88591))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a"] != "&#9786;" {
			t.Errorf("result[a] = %q, want '&#9786;'", result["a"])
		}
	})
}

func TestParseCustomDecoder(t *testing.T) {
	t.Run("uses custom decoder", func(t *testing.T) {
		customDecoder := func(s string, cs Charset, kind string) (string, error) {
			return "decoded:" + s, nil
		}
		result, err := Parse("a=b", WithDecoder(customDecoder))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["decoded:a"] != "decoded:b" {
			t.Errorf("result = %v", result)
		}
	})
}

func TestParseBracketNotation(t *testing.T) {
	t.Run("parses nested objects with bracket notation", func(t *testing.T) {
		result, err := Parse("a[b]=c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].(map[string]any)
		if !ok {
			t.Fatalf("result['a'] should be map, got %T", result["a"])
		}
		if a["b"] != "c" {
			t.Errorf("result['a']['b'] = %v, want 'c'", a["b"])
		}
	})

	t.Run("parses deeply nested objects", func(t *testing.T) {
		result, err := Parse("a[b][c]=d")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].(map[string]any)
		if !ok {
			t.Fatalf("result['a'] should be map, got %T", result["a"])
		}
		b, ok := a["b"].(map[string]any)
		if !ok {
			t.Fatalf("result['a']['b'] should be map, got %T", a["b"])
		}
		if b["c"] != "d" {
			t.Errorf("result['a']['b']['c'] = %v, want 'd'", b["c"])
		}
	})

	t.Run("handles empty bracket notation as array", func(t *testing.T) {
		result, err := Parse("a[]=b&a[]=c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result['a'] should be array, got %T: %v", result["a"], result)
		}
		if len(a) != 2 || a[0] != "b" || a[1] != "c" {
			t.Errorf("result['a'] = %v, want ['b', 'c']", a)
		}
	})
}

func TestParseNestedObjects(t *testing.T) {
	t.Run("parses a[b][c][d]=e", func(t *testing.T) {
		result, err := Parse("a[b][c][d]=e")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a := result["a"].(map[string]any)
		b := a["b"].(map[string]any)
		c := b["c"].(map[string]any)
		if c["d"] != "e" {
			t.Errorf("result['a']['b']['c']['d'] = %v, want 'e'", c["d"])
		}
	})

	t.Run("respects depth limit", func(t *testing.T) {
		// Default depth is 5, so 5 bracket segments are parsed
		result, err := Parse("a[b][c][d][e][f][g]=h")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// parent + 5 brackets = 6 levels, then remaining is literal
		a := result["a"].(map[string]any)
		b := a["b"].(map[string]any)
		c := b["c"].(map[string]any)
		d := c["d"].(map[string]any)
		e := d["e"].(map[string]any)
		f := e["f"].(map[string]any)
		// After depth 5 brackets, the 6th bracket becomes literal key
		if f["[g]"] != "h" {
			t.Errorf("expected f['[g]'] = 'h', got %v", f)
		}
	})

	t.Run("custom depth limit", func(t *testing.T) {
		result, err := Parse("a[b][c]=d", WithDepth(1))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a := result["a"].(map[string]any)
		b := a["b"].(map[string]any)
		// Only 1 bracket parsed, rest is literal
		if b["[c]"] != "d" {
			t.Errorf("expected b['[c]'] = 'd', got %v", b)
		}
	})

	t.Run("strictDepth returns error when exceeded", func(t *testing.T) {
		_, err := Parse("a[b][c]=d", WithDepth(1), WithStrictDepth(true))
		if err != ErrDepthLimitExceeded {
			t.Errorf("expected ErrDepthLimitExceeded, got %v", err)
		}
	})
}

func TestParseDotNotation(t *testing.T) {
	t.Run("parses dot notation when enabled", func(t *testing.T) {
		result, err := Parse("a.b.c=d", WithAllowDots(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a := result["a"].(map[string]any)
		b := a["b"].(map[string]any)
		if b["c"] != "d" {
			t.Errorf("result['a']['b']['c'] = %v, want 'd'", b["c"])
		}
	})

	t.Run("does not parse dot notation by default", func(t *testing.T) {
		result, err := Parse("a.b.c=d")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["a.b.c"] != "d" {
			t.Errorf("result['a.b.c'] = %v, want 'd'", result["a.b.c"])
		}
	})

	t.Run("mixes dot and bracket notation", func(t *testing.T) {
		result, err := Parse("a.b[c]=d", WithAllowDots(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a := result["a"].(map[string]any)
		b := a["b"].(map[string]any)
		if b["c"] != "d" {
			t.Errorf("result['a']['b']['c'] = %v, want 'd'", b["c"])
		}
	})
}

func TestParseArrays(t *testing.T) {
	t.Run("parses indexed arrays", func(t *testing.T) {
		result, err := Parse("a[0]=b&a[1]=c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result['a'] should be array, got %T: %v", result["a"], result)
		}
		if len(a) < 2 || a[0] != "b" || a[1] != "c" {
			t.Errorf("result['a'] = %v, want ['b', 'c']", a)
		}
	})

	t.Run("parses bracket arrays", func(t *testing.T) {
		result, err := Parse("a[]=b&a[]=c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result['a'] should be array, got %T: %v", result["a"], result)
		}
		if len(a) != 2 || a[0] != "b" || a[1] != "c" {
			t.Errorf("result['a'] = %v, want ['b', 'c']", a)
		}
	})

	t.Run("converts to object when index exceeds arrayLimit", func(t *testing.T) {
		result, err := Parse("a[100]=b", WithArrayLimit(20))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].(map[string]any)
		if !ok {
			t.Fatalf("result['a'] should be map (index > arrayLimit), got %T: %v", result["a"], result)
		}
		if a["100"] != "b" {
			t.Errorf("result['a']['100'] = %v, want 'b'", a["100"])
		}
	})

	t.Run("does not parse arrays when parseArrays is false", func(t *testing.T) {
		result, err := Parse("a[0]=b", WithParseArrays(false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].(map[string]any)
		if !ok {
			t.Fatalf("result['a'] should be map, got %T: %v", result["a"], result)
		}
		if a["0"] != "b" {
			t.Errorf("result['a']['0'] = %v, want 'b'", a["0"])
		}
	})

	t.Run("handles sparse arrays", func(t *testing.T) {
		result, err := Parse("a[0]=b&a[2]=c", WithAllowSparse(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result['a'] should be array, got %T: %v", result["a"], result)
		}
		// With sparse, we preserve nil gaps
		if len(a) < 3 {
			t.Fatalf("array too short: %v", a)
		}
		if a[0] != "b" || a[2] != "c" {
			t.Errorf("result['a'] = %v, want sparse array with b at 0, c at 2", a)
		}
	})

	t.Run("compacts sparse arrays by default", func(t *testing.T) {
		result, err := Parse("a[1]=b&a[3]=c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result['a'] should be array, got %T: %v", result["a"], result)
		}
		// Sparse arrays are compacted by default
		if len(a) != 2 || a[0] != "b" || a[1] != "c" {
			t.Errorf("result['a'] = %v, want compacted ['b', 'c']", a)
		}
	})

	t.Run("parses nested arrays", func(t *testing.T) {
		result, err := Parse("a[0][0]=b&a[0][1]=c&a[1][0]=d")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result['a'] should be array, got %T: %v", result["a"], result)
		}
		a0, ok := a[0].([]any)
		if !ok {
			t.Fatalf("result['a'][0] should be array, got %T", a[0])
		}
		if a0[0] != "b" || a0[1] != "c" {
			t.Errorf("result['a'][0] = %v, want ['b', 'c']", a0)
		}
	})

	t.Run("parses arrays with objects", func(t *testing.T) {
		result, err := Parse("a[0][b]=c&a[1][d]=e")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result['a'] should be array, got %T: %v", result["a"], result)
		}
		a0, ok := a[0].(map[string]any)
		if !ok {
			t.Fatalf("result['a'][0] should be map, got %T", a[0])
		}
		if a0["b"] != "c" {
			t.Errorf("result['a'][0]['b'] = %v, want 'c'", a0["b"])
		}
	})
}

func TestParsePrototypeProtection(t *testing.T) {
	t.Run("ignores __proto__ by default", func(t *testing.T) {
		result, err := Parse("__proto__[a]=b")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// __proto__ should be ignored
		if _, ok := result["__proto__"]; ok {
			t.Error("__proto__ should be ignored by default")
		}
	})

	t.Run("allows __proto__ with AllowPrototypes", func(t *testing.T) {
		result, err := Parse("__proto__=bad", WithAllowPrototypes(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result["__proto__"] != "bad" {
			t.Errorf("result['__proto__'] = %v, want 'bad'", result["__proto__"])
		}
	})

	t.Run("ignores constructor by default", func(t *testing.T) {
		result, err := Parse("constructor[prototype]=bad")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := result["constructor"]; ok {
			t.Error("constructor should be ignored by default")
		}
	})
}

func TestParseAllowEmptyArrays(t *testing.T) {
	t.Run("creates empty array with AllowEmptyArrays", func(t *testing.T) {
		result, err := Parse("a[]=", WithAllowEmptyArrays(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		a, ok := result["a"].([]any)
		if !ok {
			t.Fatalf("result['a'] should be array, got %T: %v", result["a"], result)
		}
		if len(a) != 0 {
			t.Errorf("result['a'] = %v, want empty array", a)
		}
	})
}

func TestParseDecodeDotInKeys(t *testing.T) {
	t.Run("decodes %2E as dot in keys", func(t *testing.T) {
		result, err := Parse("a%2Eb=c", WithDecodeDotInKeys(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// DecodeDotInKeys enables allowDots, so a.b becomes nested
		// But %2E in the key itself should become a literal dot
		a, ok := result["a"].(map[string]any)
		if !ok {
			t.Fatalf("result['a'] should be map, got %T: %v", result["a"], result)
		}
		// The key "a%2Eb" with decodeDotInKeys and allowDots:
		// First %2E is decoded to "." -> "a.b"
		// Then with allowDots, a.b -> a[b]
		if a["b"] != "c" {
			t.Errorf("result = %v", result)
		}
	})
}

func TestInterpretNumericEntitiesFunc(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"&#9786;", "☺"},
		{"&#65;", "A"},
		{"hello &#9786; world", "hello ☺ world"},
		{"no entities here", "no entities here"},
		{"&#65;&#66;&#67;", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := interpretNumericEntitiesFunc(tt.input)
			if result != tt.expected {
				t.Errorf("interpretNumericEntitiesFunc(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
