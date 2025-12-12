// Copyright 2025 Zaytra
// SPDX-License-Identifier: Apache-2.0

package qs

import "testing"

func TestFormatConstants(t *testing.T) {
	tests := []struct {
		name     string
		format   Format
		expected string
	}{
		{"RFC1738", FormatRFC1738, "RFC1738"},
		{"RFC3986", FormatRFC3986, "RFC3986"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.format) != tt.expected {
				t.Errorf("Format %s = %q, want %q", tt.name, tt.format, tt.expected)
			}
		})
	}
}

func TestDefaultFormat(t *testing.T) {
	if DefaultFormat != FormatRFC3986 {
		t.Errorf("DefaultFormat = %q, want %q", DefaultFormat, FormatRFC3986)
	}
}

func TestFormattersExist(t *testing.T) {
	formats := []Format{FormatRFC1738, FormatRFC3986}

	for _, format := range formats {
		if _, ok := Formatters[format]; !ok {
			t.Errorf("Formatters[%q] not found", format)
		}
	}
}

func TestFormatRFC1738(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"space encoded", "hello%20world", "hello+world"},
		{"multiple spaces", "a%20b%20c", "a+b+c"},
		{"no spaces", "helloworld", "helloworld"},
		{"empty string", "", ""},
		{"only space", "%20", "+"},
		{"mixed content", "foo%20bar%3Dbaz", "foo+bar%3Dbaz"},
		{"already plus", "hello+world", "hello+world"},
	}

	formatter := Formatters[FormatRFC1738]
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter(tt.input)
			if result != tt.expected {
				t.Errorf("RFC1738(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatRFC3986(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"space encoded", "hello%20world", "hello%20world"},
		{"multiple spaces", "a%20b%20c", "a%20b%20c"},
		{"no spaces", "helloworld", "helloworld"},
		{"empty string", "", ""},
		{"with plus", "hello+world", "hello+world"},
		{"mixed content", "foo%20bar%3Dbaz", "foo%20bar%3Dbaz"},
	}

	formatter := Formatters[FormatRFC3986]
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter(tt.input)
			if result != tt.expected {
				t.Errorf("RFC3986(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetFormatter(t *testing.T) {
	tests := []struct {
		name           string
		format         Format
		input          string
		expectedOutput string
	}{
		{"RFC1738 known", FormatRFC1738, "a%20b", "a+b"},
		{"RFC3986 known", FormatRFC3986, "a%20b", "a%20b"},
		{"unknown format fallback", Format("UNKNOWN"), "a%20b", "a%20b"},
		{"empty format fallback", Format(""), "a%20b", "a%20b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := GetFormatter(tt.format)
			result := formatter(tt.input)
			if result != tt.expectedOutput {
				t.Errorf("GetFormatter(%q)(%q) = %q, want %q", tt.format, tt.input, result, tt.expectedOutput)
			}
		})
	}
}

func TestFormatterFuncType(t *testing.T) {
	// Verify that formatter functions match the FormatterFunc type
	var _ FormatterFunc = Formatters[FormatRFC1738]
	var _ FormatterFunc = Formatters[FormatRFC3986]
	var _ FormatterFunc = GetFormatter(FormatRFC1738)
}
